// Generated by tmpl
// https://github.com/benbjohnson/tmpl
//
// DO NOT EDIT!
// Source: compact.gen.go.tmpl

package tsm1

import (
	"sort"

	"github.com/influxdata/influxdb/v2/tsdb"
)

// merge combines the next set of blocks into merged blocks.
func (k *tsmBatchKeyIterator) mergeFloat() {
	// No blocks left, or pending merged values, we're done
	if len(k.blocks) == 0 && len(k.merged) == 0 && k.mergedFloatValues.Len() == 0 {
		return
	}

	sort.Stable(k.blocks)

	dedup := k.mergedFloatValues.Len() != 0
	if len(k.blocks) > 0 && !dedup {
		// If we have more than one block or any partially tombstoned blocks, we many need to dedup
		dedup = len(k.blocks[0].tombstones) > 0 || k.blocks[0].partiallyRead()

		// Quickly scan each block to see if any overlap with the prior block, if they overlap then
		// we need to dedup as there may be duplicate points now
		for i := 1; !dedup && i < len(k.blocks); i++ {
			dedup = k.blocks[i].partiallyRead() ||
				k.blocks[i].overlapsTimeRange(k.blocks[i-1].minTime, k.blocks[i-1].maxTime) ||
				len(k.blocks[i].tombstones) > 0
		}

	}

	k.merged = k.combineFloat(dedup)
}

// combine returns a new set of blocks using the current blocks in the buffers.  If dedup
// is true, all the blocks will be decoded, dedup and sorted in in order.  If dedup is false,
// only blocks that are smaller than the chunk size will be decoded and combined.
func (k *tsmBatchKeyIterator) combineFloat(dedup bool) blocks {
	if dedup {
		for k.mergedFloatValues.Len() < k.size && len(k.blocks) > 0 {
			for len(k.blocks) > 0 && k.blocks[0].read() {
				k.blocks = k.blocks[1:]
			}

			if len(k.blocks) == 0 {
				break
			}
			first := k.blocks[0]
			minTime := first.minTime
			maxTime := first.maxTime

			// Adjust the min time to the start of any overlapping blocks.
			for i := 0; i < len(k.blocks); i++ {
				if k.blocks[i].overlapsTimeRange(minTime, maxTime) && !k.blocks[i].read() {
					if k.blocks[i].minTime < minTime {
						minTime = k.blocks[i].minTime
					}
					if k.blocks[i].maxTime > minTime && k.blocks[i].maxTime < maxTime {
						maxTime = k.blocks[i].maxTime
					}
				}
			}

			// We have some overlapping blocks so decode all, append in order and then dedup
			for i := 0; i < len(k.blocks); i++ {
				if !k.blocks[i].overlapsTimeRange(minTime, maxTime) || k.blocks[i].read() {
					continue
				}

				var v tsdb.FloatArray
				var err error
				if err = DecodeFloatArrayBlock(k.blocks[i].b, &v); err != nil {
					k.handleDecodeError(err, "float")
					return nil
				}

				// Invariant: v.MaxTime() == k.blocks[i].maxTime
				if k.blocks[i].maxTime != v.MaxTime() {
					if maxTime == k.blocks[i].maxTime {
						maxTime = v.MaxTime()
					}
					k.blocks[i].maxTime = v.MaxTime()
				}

				// Remove values we already read
				v.Exclude(k.blocks[i].readMin, k.blocks[i].readMax)

				// Filter out only the values for overlapping block
				v.Include(minTime, maxTime)
				if v.Len() > 0 {
					// Record that we read a subset of the block
					k.blocks[i].markRead(v.MinTime(), v.MaxTime())
				}

				// Apply each tombstone to the block
				for _, ts := range k.blocks[i].tombstones {
					v.Exclude(ts.Min, ts.Max)
				}

				k.mergedFloatValues.Merge(&v)
			}
		}

		// Since we combined multiple blocks, we could have more values than we should put into
		// a single block.  We need to chunk them up into groups and re-encode them.
		return k.chunkFloat(nil)
	}
	var i int

	for ; i < len(k.blocks); i++ {

		// skip this block if it's values were already read
		if k.blocks[i].read() {
			continue
		}

		// if this block is already full, just add it as is
		count, err := BlockCount(k.blocks[i].b)
		if err != nil {
			k.AppendError(err)
			continue
		}

		if count < k.size {
			break
		}

		k.merged = append(k.merged, k.blocks[i])
	}

	if k.fast {
		for i < len(k.blocks) {
			// skip this block if it's values were already read
			if k.blocks[i].read() {
				i++
				continue
			}

			k.merged = append(k.merged, k.blocks[i])
			i++
		}
	}

	// if we only have 1 blocks left, just append it as is and avoid decoding/recoding
	if i == len(k.blocks)-1 {
		if !k.blocks[i].read() {
			k.merged = append(k.merged, k.blocks[i])
		}
		i++
	}

	// The remaining blocks can be combined and we know that they do not overlap and
	// so we can just append each, sort and re-encode.
	for i < len(k.blocks) && k.mergedFloatValues.Len() < k.size {
		if k.blocks[i].read() {
			i++
			continue
		}

		var v tsdb.FloatArray
		if err := DecodeFloatArrayBlock(k.blocks[i].b, &v); err != nil {
			k.handleDecodeError(err, "float")
			return nil
		}

		// Invariant: v.MaxTime() == k.blocks[i].maxTime
		if k.blocks[i].maxTime != v.MaxTime() {
			k.blocks[i].maxTime = v.MaxTime()
		}

		// Apply each tombstone to the block
		for _, ts := range k.blocks[i].tombstones {
			v.Exclude(ts.Min, ts.Max)
		}

		k.blocks[i].markRead(k.blocks[i].minTime, k.blocks[i].maxTime)

		k.mergedFloatValues.Merge(&v)
		i++
	}

	k.blocks = k.blocks[i:]

	return k.chunkFloat(k.merged)
}

func (k *tsmBatchKeyIterator) chunkFloat(dst blocks) blocks {
	if k.mergedFloatValues.Len() > k.size {
		var values tsdb.FloatArray
		values.Timestamps = k.mergedFloatValues.Timestamps[:k.size]
		minTime, maxTime := values.Timestamps[0], values.Timestamps[len(values.Timestamps)-1]
		values.Values = k.mergedFloatValues.Values[:k.size]

		cb, err := EncodeFloatArrayBlock(&values, nil) // TODO(edd): pool this buffer
		if err != nil {
			k.handleEncodeError(err, "float")
			return nil
		}

		dst = append(dst, &block{
			minTime: minTime,
			maxTime: maxTime,
			key:     k.key,
			b:       cb,
		})
		k.mergedFloatValues.Timestamps = k.mergedFloatValues.Timestamps[k.size:]
		k.mergedFloatValues.Values = k.mergedFloatValues.Values[k.size:]
		return dst
	}

	// Re-encode the remaining values into the last block
	if k.mergedFloatValues.Len() > 0 {
		minTime, maxTime := k.mergedFloatValues.Timestamps[0], k.mergedFloatValues.Timestamps[len(k.mergedFloatValues.Timestamps)-1]
		cb, err := EncodeFloatArrayBlock(k.mergedFloatValues, nil) // TODO(edd): pool this buffer
		if err != nil {
			k.handleEncodeError(err, "float")
			return nil
		}

		dst = append(dst, &block{
			minTime: minTime,
			maxTime: maxTime,
			key:     k.key,
			b:       cb,
		})
		k.mergedFloatValues.Timestamps = k.mergedFloatValues.Timestamps[:0]
		k.mergedFloatValues.Values = k.mergedFloatValues.Values[:0]
	}
	return dst
}

// merge combines the next set of blocks into merged blocks.
func (k *tsmBatchKeyIterator) mergeInteger() {
	// No blocks left, or pending merged values, we're done
	if len(k.blocks) == 0 && len(k.merged) == 0 && k.mergedIntegerValues.Len() == 0 {
		return
	}

	sort.Stable(k.blocks)

	dedup := k.mergedIntegerValues.Len() != 0
	if len(k.blocks) > 0 && !dedup {
		// If we have more than one block or any partially tombstoned blocks, we many need to dedup
		dedup = len(k.blocks[0].tombstones) > 0 || k.blocks[0].partiallyRead()

		// Quickly scan each block to see if any overlap with the prior block, if they overlap then
		// we need to dedup as there may be duplicate points now
		for i := 1; !dedup && i < len(k.blocks); i++ {
			dedup = k.blocks[i].partiallyRead() ||
				k.blocks[i].overlapsTimeRange(k.blocks[i-1].minTime, k.blocks[i-1].maxTime) ||
				len(k.blocks[i].tombstones) > 0
		}

	}

	k.merged = k.combineInteger(dedup)
}

// combine returns a new set of blocks using the current blocks in the buffers.  If dedup
// is true, all the blocks will be decoded, dedup and sorted in in order.  If dedup is false,
// only blocks that are smaller than the chunk size will be decoded and combined.
func (k *tsmBatchKeyIterator) combineInteger(dedup bool) blocks {
	if dedup {
		for k.mergedIntegerValues.Len() < k.size && len(k.blocks) > 0 {
			for len(k.blocks) > 0 && k.blocks[0].read() {
				k.blocks = k.blocks[1:]
			}

			if len(k.blocks) == 0 {
				break
			}
			first := k.blocks[0]
			minTime := first.minTime
			maxTime := first.maxTime

			// Adjust the min time to the start of any overlapping blocks.
			for i := 0; i < len(k.blocks); i++ {
				if k.blocks[i].overlapsTimeRange(minTime, maxTime) && !k.blocks[i].read() {
					if k.blocks[i].minTime < minTime {
						minTime = k.blocks[i].minTime
					}
					if k.blocks[i].maxTime > minTime && k.blocks[i].maxTime < maxTime {
						maxTime = k.blocks[i].maxTime
					}
				}
			}

			// We have some overlapping blocks so decode all, append in order and then dedup
			for i := 0; i < len(k.blocks); i++ {
				if !k.blocks[i].overlapsTimeRange(minTime, maxTime) || k.blocks[i].read() {
					continue
				}

				var v tsdb.IntegerArray
				var err error
				if err = DecodeIntegerArrayBlock(k.blocks[i].b, &v); err != nil {
					k.handleDecodeError(err, "integer")
					return nil
				}

				// Invariant: v.MaxTime() == k.blocks[i].maxTime
				if k.blocks[i].maxTime != v.MaxTime() {
					if maxTime == k.blocks[i].maxTime {
						maxTime = v.MaxTime()
					}
					k.blocks[i].maxTime = v.MaxTime()
				}

				// Remove values we already read
				v.Exclude(k.blocks[i].readMin, k.blocks[i].readMax)

				// Filter out only the values for overlapping block
				v.Include(minTime, maxTime)
				if v.Len() > 0 {
					// Record that we read a subset of the block
					k.blocks[i].markRead(v.MinTime(), v.MaxTime())
				}

				// Apply each tombstone to the block
				for _, ts := range k.blocks[i].tombstones {
					v.Exclude(ts.Min, ts.Max)
				}

				k.mergedIntegerValues.Merge(&v)
			}
		}

		// Since we combined multiple blocks, we could have more values than we should put into
		// a single block.  We need to chunk them up into groups and re-encode them.
		return k.chunkInteger(nil)
	}
	var i int

	for ; i < len(k.blocks); i++ {

		// skip this block if it's values were already read
		if k.blocks[i].read() {
			continue
		}

		// if this block is already full, just add it as is
		count, err := BlockCount(k.blocks[i].b)
		if err != nil {
			k.AppendError(err)
			continue
		}

		if count < k.size {
			break
		}

		k.merged = append(k.merged, k.blocks[i])
	}

	if k.fast {
		for i < len(k.blocks) {
			// skip this block if it's values were already read
			if k.blocks[i].read() {
				i++
				continue
			}

			k.merged = append(k.merged, k.blocks[i])
			i++
		}
	}

	// if we only have 1 blocks left, just append it as is and avoid decoding/recoding
	if i == len(k.blocks)-1 {
		if !k.blocks[i].read() {
			k.merged = append(k.merged, k.blocks[i])
		}
		i++
	}

	// The remaining blocks can be combined and we know that they do not overlap and
	// so we can just append each, sort and re-encode.
	for i < len(k.blocks) && k.mergedIntegerValues.Len() < k.size {
		if k.blocks[i].read() {
			i++
			continue
		}

		var v tsdb.IntegerArray
		if err := DecodeIntegerArrayBlock(k.blocks[i].b, &v); err != nil {
			k.handleDecodeError(err, "integer")
			return nil
		}

		// Invariant: v.MaxTime() == k.blocks[i].maxTime
		if k.blocks[i].maxTime != v.MaxTime() {
			k.blocks[i].maxTime = v.MaxTime()
		}

		// Apply each tombstone to the block
		for _, ts := range k.blocks[i].tombstones {
			v.Exclude(ts.Min, ts.Max)
		}

		k.blocks[i].markRead(k.blocks[i].minTime, k.blocks[i].maxTime)

		k.mergedIntegerValues.Merge(&v)
		i++
	}

	k.blocks = k.blocks[i:]

	return k.chunkInteger(k.merged)
}

func (k *tsmBatchKeyIterator) chunkInteger(dst blocks) blocks {
	if k.mergedIntegerValues.Len() > k.size {
		var values tsdb.IntegerArray
		values.Timestamps = k.mergedIntegerValues.Timestamps[:k.size]
		minTime, maxTime := values.Timestamps[0], values.Timestamps[len(values.Timestamps)-1]
		values.Values = k.mergedIntegerValues.Values[:k.size]

		cb, err := EncodeIntegerArrayBlock(&values, nil) // TODO(edd): pool this buffer
		if err != nil {
			k.handleEncodeError(err, "integer")
			return nil
		}

		dst = append(dst, &block{
			minTime: minTime,
			maxTime: maxTime,
			key:     k.key,
			b:       cb,
		})
		k.mergedIntegerValues.Timestamps = k.mergedIntegerValues.Timestamps[k.size:]
		k.mergedIntegerValues.Values = k.mergedIntegerValues.Values[k.size:]
		return dst
	}

	// Re-encode the remaining values into the last block
	if k.mergedIntegerValues.Len() > 0 {
		minTime, maxTime := k.mergedIntegerValues.Timestamps[0], k.mergedIntegerValues.Timestamps[len(k.mergedIntegerValues.Timestamps)-1]
		cb, err := EncodeIntegerArrayBlock(k.mergedIntegerValues, nil) // TODO(edd): pool this buffer
		if err != nil {
			k.handleEncodeError(err, "integer")
			return nil
		}

		dst = append(dst, &block{
			minTime: minTime,
			maxTime: maxTime,
			key:     k.key,
			b:       cb,
		})
		k.mergedIntegerValues.Timestamps = k.mergedIntegerValues.Timestamps[:0]
		k.mergedIntegerValues.Values = k.mergedIntegerValues.Values[:0]
	}
	return dst
}

// merge combines the next set of blocks into merged blocks.
func (k *tsmBatchKeyIterator) mergeUnsigned() {
	// No blocks left, or pending merged values, we're done
	if len(k.blocks) == 0 && len(k.merged) == 0 && k.mergedUnsignedValues.Len() == 0 {
		return
	}

	sort.Stable(k.blocks)

	dedup := k.mergedUnsignedValues.Len() != 0
	if len(k.blocks) > 0 && !dedup {
		// If we have more than one block or any partially tombstoned blocks, we many need to dedup
		dedup = len(k.blocks[0].tombstones) > 0 || k.blocks[0].partiallyRead()

		// Quickly scan each block to see if any overlap with the prior block, if they overlap then
		// we need to dedup as there may be duplicate points now
		for i := 1; !dedup && i < len(k.blocks); i++ {
			dedup = k.blocks[i].partiallyRead() ||
				k.blocks[i].overlapsTimeRange(k.blocks[i-1].minTime, k.blocks[i-1].maxTime) ||
				len(k.blocks[i].tombstones) > 0
		}

	}

	k.merged = k.combineUnsigned(dedup)
}

// combine returns a new set of blocks using the current blocks in the buffers.  If dedup
// is true, all the blocks will be decoded, dedup and sorted in in order.  If dedup is false,
// only blocks that are smaller than the chunk size will be decoded and combined.
func (k *tsmBatchKeyIterator) combineUnsigned(dedup bool) blocks {
	if dedup {
		for k.mergedUnsignedValues.Len() < k.size && len(k.blocks) > 0 {
			for len(k.blocks) > 0 && k.blocks[0].read() {
				k.blocks = k.blocks[1:]
			}

			if len(k.blocks) == 0 {
				break
			}
			first := k.blocks[0]
			minTime := first.minTime
			maxTime := first.maxTime

			// Adjust the min time to the start of any overlapping blocks.
			for i := 0; i < len(k.blocks); i++ {
				if k.blocks[i].overlapsTimeRange(minTime, maxTime) && !k.blocks[i].read() {
					if k.blocks[i].minTime < minTime {
						minTime = k.blocks[i].minTime
					}
					if k.blocks[i].maxTime > minTime && k.blocks[i].maxTime < maxTime {
						maxTime = k.blocks[i].maxTime
					}
				}
			}

			// We have some overlapping blocks so decode all, append in order and then dedup
			for i := 0; i < len(k.blocks); i++ {
				if !k.blocks[i].overlapsTimeRange(minTime, maxTime) || k.blocks[i].read() {
					continue
				}

				var v tsdb.UnsignedArray
				var err error
				if err = DecodeUnsignedArrayBlock(k.blocks[i].b, &v); err != nil {
					k.handleDecodeError(err, "unsigned")
					return nil
				}

				// Invariant: v.MaxTime() == k.blocks[i].maxTime
				if k.blocks[i].maxTime != v.MaxTime() {
					if maxTime == k.blocks[i].maxTime {
						maxTime = v.MaxTime()
					}
					k.blocks[i].maxTime = v.MaxTime()
				}

				// Remove values we already read
				v.Exclude(k.blocks[i].readMin, k.blocks[i].readMax)

				// Filter out only the values for overlapping block
				v.Include(minTime, maxTime)
				if v.Len() > 0 {
					// Record that we read a subset of the block
					k.blocks[i].markRead(v.MinTime(), v.MaxTime())
				}

				// Apply each tombstone to the block
				for _, ts := range k.blocks[i].tombstones {
					v.Exclude(ts.Min, ts.Max)
				}

				k.mergedUnsignedValues.Merge(&v)
			}
		}

		// Since we combined multiple blocks, we could have more values than we should put into
		// a single block.  We need to chunk them up into groups and re-encode them.
		return k.chunkUnsigned(nil)
	}
	var i int

	for ; i < len(k.blocks); i++ {

		// skip this block if it's values were already read
		if k.blocks[i].read() {
			continue
		}

		// if this block is already full, just add it as is
		count, err := BlockCount(k.blocks[i].b)
		if err != nil {
			k.AppendError(err)
			continue
		}

		if count < k.size {
			break
		}

		k.merged = append(k.merged, k.blocks[i])
	}

	if k.fast {
		for i < len(k.blocks) {
			// skip this block if it's values were already read
			if k.blocks[i].read() {
				i++
				continue
			}

			k.merged = append(k.merged, k.blocks[i])
			i++
		}
	}

	// if we only have 1 blocks left, just append it as is and avoid decoding/recoding
	if i == len(k.blocks)-1 {
		if !k.blocks[i].read() {
			k.merged = append(k.merged, k.blocks[i])
		}
		i++
	}

	// The remaining blocks can be combined and we know that they do not overlap and
	// so we can just append each, sort and re-encode.
	for i < len(k.blocks) && k.mergedUnsignedValues.Len() < k.size {
		if k.blocks[i].read() {
			i++
			continue
		}

		var v tsdb.UnsignedArray
		if err := DecodeUnsignedArrayBlock(k.blocks[i].b, &v); err != nil {
			k.handleDecodeError(err, "unsigned")
			return nil
		}

		// Invariant: v.MaxTime() == k.blocks[i].maxTime
		if k.blocks[i].maxTime != v.MaxTime() {
			k.blocks[i].maxTime = v.MaxTime()
		}

		// Apply each tombstone to the block
		for _, ts := range k.blocks[i].tombstones {
			v.Exclude(ts.Min, ts.Max)
		}

		k.blocks[i].markRead(k.blocks[i].minTime, k.blocks[i].maxTime)

		k.mergedUnsignedValues.Merge(&v)
		i++
	}

	k.blocks = k.blocks[i:]

	return k.chunkUnsigned(k.merged)
}

func (k *tsmBatchKeyIterator) chunkUnsigned(dst blocks) blocks {
	if k.mergedUnsignedValues.Len() > k.size {
		var values tsdb.UnsignedArray
		values.Timestamps = k.mergedUnsignedValues.Timestamps[:k.size]
		minTime, maxTime := values.Timestamps[0], values.Timestamps[len(values.Timestamps)-1]
		values.Values = k.mergedUnsignedValues.Values[:k.size]

		cb, err := EncodeUnsignedArrayBlock(&values, nil) // TODO(edd): pool this buffer
		if err != nil {
			k.handleEncodeError(err, "unsigned")
			return nil
		}

		dst = append(dst, &block{
			minTime: minTime,
			maxTime: maxTime,
			key:     k.key,
			b:       cb,
		})
		k.mergedUnsignedValues.Timestamps = k.mergedUnsignedValues.Timestamps[k.size:]
		k.mergedUnsignedValues.Values = k.mergedUnsignedValues.Values[k.size:]
		return dst
	}

	// Re-encode the remaining values into the last block
	if k.mergedUnsignedValues.Len() > 0 {
		minTime, maxTime := k.mergedUnsignedValues.Timestamps[0], k.mergedUnsignedValues.Timestamps[len(k.mergedUnsignedValues.Timestamps)-1]
		cb, err := EncodeUnsignedArrayBlock(k.mergedUnsignedValues, nil) // TODO(edd): pool this buffer
		if err != nil {
			k.handleEncodeError(err, "unsigned")
			return nil
		}

		dst = append(dst, &block{
			minTime: minTime,
			maxTime: maxTime,
			key:     k.key,
			b:       cb,
		})
		k.mergedUnsignedValues.Timestamps = k.mergedUnsignedValues.Timestamps[:0]
		k.mergedUnsignedValues.Values = k.mergedUnsignedValues.Values[:0]
	}
	return dst
}

// merge combines the next set of blocks into merged blocks.
func (k *tsmBatchKeyIterator) mergeString() {
	// No blocks left, or pending merged values, we're done
	if len(k.blocks) == 0 && len(k.merged) == 0 && k.mergedStringValues.Len() == 0 {
		return
	}

	sort.Stable(k.blocks)

	dedup := k.mergedStringValues.Len() != 0
	if len(k.blocks) > 0 && !dedup {
		// If we have more than one block or any partially tombstoned blocks, we many need to dedup
		dedup = len(k.blocks[0].tombstones) > 0 || k.blocks[0].partiallyRead()

		// Quickly scan each block to see if any overlap with the prior block, if they overlap then
		// we need to dedup as there may be duplicate points now
		for i := 1; !dedup && i < len(k.blocks); i++ {
			dedup = k.blocks[i].partiallyRead() ||
				k.blocks[i].overlapsTimeRange(k.blocks[i-1].minTime, k.blocks[i-1].maxTime) ||
				len(k.blocks[i].tombstones) > 0
		}

	}

	k.merged = k.combineString(dedup)
}

// combine returns a new set of blocks using the current blocks in the buffers.  If dedup
// is true, all the blocks will be decoded, dedup and sorted in in order.  If dedup is false,
// only blocks that are smaller than the chunk size will be decoded and combined.
func (k *tsmBatchKeyIterator) combineString(dedup bool) blocks {
	if dedup {
		for k.mergedStringValues.Len() < k.size && len(k.blocks) > 0 {
			for len(k.blocks) > 0 && k.blocks[0].read() {
				k.blocks = k.blocks[1:]
			}

			if len(k.blocks) == 0 {
				break
			}
			first := k.blocks[0]
			minTime := first.minTime
			maxTime := first.maxTime

			// Adjust the min time to the start of any overlapping blocks.
			for i := 0; i < len(k.blocks); i++ {
				if k.blocks[i].overlapsTimeRange(minTime, maxTime) && !k.blocks[i].read() {
					if k.blocks[i].minTime < minTime {
						minTime = k.blocks[i].minTime
					}
					if k.blocks[i].maxTime > minTime && k.blocks[i].maxTime < maxTime {
						maxTime = k.blocks[i].maxTime
					}
				}
			}

			// We have some overlapping blocks so decode all, append in order and then dedup
			for i := 0; i < len(k.blocks); i++ {
				if !k.blocks[i].overlapsTimeRange(minTime, maxTime) || k.blocks[i].read() {
					continue
				}

				var v tsdb.StringArray
				var err error
				if err = DecodeStringArrayBlock(k.blocks[i].b, &v); err != nil {
					k.handleDecodeError(err, "string")
					return nil
				}

				// Invariant: v.MaxTime() == k.blocks[i].maxTime
				if k.blocks[i].maxTime != v.MaxTime() {
					if maxTime == k.blocks[i].maxTime {
						maxTime = v.MaxTime()
					}
					k.blocks[i].maxTime = v.MaxTime()
				}

				// Remove values we already read
				v.Exclude(k.blocks[i].readMin, k.blocks[i].readMax)

				// Filter out only the values for overlapping block
				v.Include(minTime, maxTime)
				if v.Len() > 0 {
					// Record that we read a subset of the block
					k.blocks[i].markRead(v.MinTime(), v.MaxTime())
				}

				// Apply each tombstone to the block
				for _, ts := range k.blocks[i].tombstones {
					v.Exclude(ts.Min, ts.Max)
				}

				k.mergedStringValues.Merge(&v)
			}
		}

		// Since we combined multiple blocks, we could have more values than we should put into
		// a single block.  We need to chunk them up into groups and re-encode them.
		return k.chunkString(nil)
	}
	var i int

	for ; i < len(k.blocks); i++ {

		// skip this block if it's values were already read
		if k.blocks[i].read() {
			continue
		}

		// if this block is already full, just add it as is
		count, err := BlockCount(k.blocks[i].b)
		if err != nil {
			k.AppendError(err)
			continue
		}

		if count < k.size {
			break
		}

		k.merged = append(k.merged, k.blocks[i])
	}

	if k.fast {
		for i < len(k.blocks) {
			// skip this block if it's values were already read
			if k.blocks[i].read() {
				i++
				continue
			}

			k.merged = append(k.merged, k.blocks[i])
			i++
		}
	}

	// if we only have 1 blocks left, just append it as is and avoid decoding/recoding
	if i == len(k.blocks)-1 {
		if !k.blocks[i].read() {
			k.merged = append(k.merged, k.blocks[i])
		}
		i++
	}

	// The remaining blocks can be combined and we know that they do not overlap and
	// so we can just append each, sort and re-encode.
	for i < len(k.blocks) && k.mergedStringValues.Len() < k.size {
		if k.blocks[i].read() {
			i++
			continue
		}

		var v tsdb.StringArray
		if err := DecodeStringArrayBlock(k.blocks[i].b, &v); err != nil {
			k.handleDecodeError(err, "string")
			return nil
		}

		// Invariant: v.MaxTime() == k.blocks[i].maxTime
		if k.blocks[i].maxTime != v.MaxTime() {
			k.blocks[i].maxTime = v.MaxTime()
		}

		// Apply each tombstone to the block
		for _, ts := range k.blocks[i].tombstones {
			v.Exclude(ts.Min, ts.Max)
		}

		k.blocks[i].markRead(k.blocks[i].minTime, k.blocks[i].maxTime)

		k.mergedStringValues.Merge(&v)
		i++
	}

	k.blocks = k.blocks[i:]

	return k.chunkString(k.merged)
}

func (k *tsmBatchKeyIterator) chunkString(dst blocks) blocks {
	if k.mergedStringValues.Len() > k.size {
		var values tsdb.StringArray
		values.Timestamps = k.mergedStringValues.Timestamps[:k.size]
		minTime, maxTime := values.Timestamps[0], values.Timestamps[len(values.Timestamps)-1]
		values.Values = k.mergedStringValues.Values[:k.size]

		cb, err := EncodeStringArrayBlock(&values, nil) // TODO(edd): pool this buffer
		if err != nil {
			k.handleEncodeError(err, "string")
			return nil
		}

		dst = append(dst, &block{
			minTime: minTime,
			maxTime: maxTime,
			key:     k.key,
			b:       cb,
		})
		k.mergedStringValues.Timestamps = k.mergedStringValues.Timestamps[k.size:]
		k.mergedStringValues.Values = k.mergedStringValues.Values[k.size:]
		return dst
	}

	// Re-encode the remaining values into the last block
	if k.mergedStringValues.Len() > 0 {
		minTime, maxTime := k.mergedStringValues.Timestamps[0], k.mergedStringValues.Timestamps[len(k.mergedStringValues.Timestamps)-1]
		cb, err := EncodeStringArrayBlock(k.mergedStringValues, nil) // TODO(edd): pool this buffer
		if err != nil {
			k.handleEncodeError(err, "string")
			return nil
		}

		dst = append(dst, &block{
			minTime: minTime,
			maxTime: maxTime,
			key:     k.key,
			b:       cb,
		})
		k.mergedStringValues.Timestamps = k.mergedStringValues.Timestamps[:0]
		k.mergedStringValues.Values = k.mergedStringValues.Values[:0]
	}
	return dst
}

// merge combines the next set of blocks into merged blocks.
func (k *tsmBatchKeyIterator) mergeBoolean() {
	// No blocks left, or pending merged values, we're done
	if len(k.blocks) == 0 && len(k.merged) == 0 && k.mergedBooleanValues.Len() == 0 {
		return
	}

	sort.Stable(k.blocks)

	dedup := k.mergedBooleanValues.Len() != 0
	if len(k.blocks) > 0 && !dedup {
		// If we have more than one block or any partially tombstoned blocks, we many need to dedup
		dedup = len(k.blocks[0].tombstones) > 0 || k.blocks[0].partiallyRead()

		// Quickly scan each block to see if any overlap with the prior block, if they overlap then
		// we need to dedup as there may be duplicate points now
		for i := 1; !dedup && i < len(k.blocks); i++ {
			dedup = k.blocks[i].partiallyRead() ||
				k.blocks[i].overlapsTimeRange(k.blocks[i-1].minTime, k.blocks[i-1].maxTime) ||
				len(k.blocks[i].tombstones) > 0
		}

	}

	k.merged = k.combineBoolean(dedup)
}

// combine returns a new set of blocks using the current blocks in the buffers.  If dedup
// is true, all the blocks will be decoded, dedup and sorted in in order.  If dedup is false,
// only blocks that are smaller than the chunk size will be decoded and combined.
func (k *tsmBatchKeyIterator) combineBoolean(dedup bool) blocks {
	if dedup {
		for k.mergedBooleanValues.Len() < k.size && len(k.blocks) > 0 {
			for len(k.blocks) > 0 && k.blocks[0].read() {
				k.blocks = k.blocks[1:]
			}

			if len(k.blocks) == 0 {
				break
			}
			first := k.blocks[0]
			minTime := first.minTime
			maxTime := first.maxTime

			// Adjust the min time to the start of any overlapping blocks.
			for i := 0; i < len(k.blocks); i++ {
				if k.blocks[i].overlapsTimeRange(minTime, maxTime) && !k.blocks[i].read() {
					if k.blocks[i].minTime < minTime {
						minTime = k.blocks[i].minTime
					}
					if k.blocks[i].maxTime > minTime && k.blocks[i].maxTime < maxTime {
						maxTime = k.blocks[i].maxTime
					}
				}
			}

			// We have some overlapping blocks so decode all, append in order and then dedup
			for i := 0; i < len(k.blocks); i++ {
				if !k.blocks[i].overlapsTimeRange(minTime, maxTime) || k.blocks[i].read() {
					continue
				}

				var v tsdb.BooleanArray
				var err error
				if err = DecodeBooleanArrayBlock(k.blocks[i].b, &v); err != nil {
					k.handleDecodeError(err, "boolean")
					return nil
				}

				// Invariant: v.MaxTime() == k.blocks[i].maxTime
				if k.blocks[i].maxTime != v.MaxTime() {
					if maxTime == k.blocks[i].maxTime {
						maxTime = v.MaxTime()
					}
					k.blocks[i].maxTime = v.MaxTime()
				}

				// Remove values we already read
				v.Exclude(k.blocks[i].readMin, k.blocks[i].readMax)

				// Filter out only the values for overlapping block
				v.Include(minTime, maxTime)
				if v.Len() > 0 {
					// Record that we read a subset of the block
					k.blocks[i].markRead(v.MinTime(), v.MaxTime())
				}

				// Apply each tombstone to the block
				for _, ts := range k.blocks[i].tombstones {
					v.Exclude(ts.Min, ts.Max)
				}

				k.mergedBooleanValues.Merge(&v)
			}
		}

		// Since we combined multiple blocks, we could have more values than we should put into
		// a single block.  We need to chunk them up into groups and re-encode them.
		return k.chunkBoolean(nil)
	}
	var i int

	for ; i < len(k.blocks); i++ {

		// skip this block if it's values were already read
		if k.blocks[i].read() {
			continue
		}

		// if this block is already full, just add it as is
		count, err := BlockCount(k.blocks[i].b)
		if err != nil {
			k.AppendError(err)
			continue
		}

		if count < k.size {
			break
		}

		k.merged = append(k.merged, k.blocks[i])
	}

	if k.fast {
		for i < len(k.blocks) {
			// skip this block if it's values were already read
			if k.blocks[i].read() {
				i++
				continue
			}

			k.merged = append(k.merged, k.blocks[i])
			i++
		}
	}

	// if we only have 1 blocks left, just append it as is and avoid decoding/recoding
	if i == len(k.blocks)-1 {
		if !k.blocks[i].read() {
			k.merged = append(k.merged, k.blocks[i])
		}
		i++
	}

	// The remaining blocks can be combined and we know that they do not overlap and
	// so we can just append each, sort and re-encode.
	for i < len(k.blocks) && k.mergedBooleanValues.Len() < k.size {
		if k.blocks[i].read() {
			i++
			continue
		}

		var v tsdb.BooleanArray
		if err := DecodeBooleanArrayBlock(k.blocks[i].b, &v); err != nil {
			k.handleDecodeError(err, "boolean")
			return nil
		}

		// Invariant: v.MaxTime() == k.blocks[i].maxTime
		if k.blocks[i].maxTime != v.MaxTime() {
			k.blocks[i].maxTime = v.MaxTime()
		}

		// Apply each tombstone to the block
		for _, ts := range k.blocks[i].tombstones {
			v.Exclude(ts.Min, ts.Max)
		}

		k.blocks[i].markRead(k.blocks[i].minTime, k.blocks[i].maxTime)

		k.mergedBooleanValues.Merge(&v)
		i++
	}

	k.blocks = k.blocks[i:]

	return k.chunkBoolean(k.merged)
}

func (k *tsmBatchKeyIterator) chunkBoolean(dst blocks) blocks {
	if k.mergedBooleanValues.Len() > k.size {
		var values tsdb.BooleanArray
		values.Timestamps = k.mergedBooleanValues.Timestamps[:k.size]
		minTime, maxTime := values.Timestamps[0], values.Timestamps[len(values.Timestamps)-1]
		values.Values = k.mergedBooleanValues.Values[:k.size]

		cb, err := EncodeBooleanArrayBlock(&values, nil) // TODO(edd): pool this buffer
		if err != nil {
			k.handleEncodeError(err, "boolean")
			return nil
		}

		dst = append(dst, &block{
			minTime: minTime,
			maxTime: maxTime,
			key:     k.key,
			b:       cb,
		})
		k.mergedBooleanValues.Timestamps = k.mergedBooleanValues.Timestamps[k.size:]
		k.mergedBooleanValues.Values = k.mergedBooleanValues.Values[k.size:]
		return dst
	}

	// Re-encode the remaining values into the last block
	if k.mergedBooleanValues.Len() > 0 {
		minTime, maxTime := k.mergedBooleanValues.Timestamps[0], k.mergedBooleanValues.Timestamps[len(k.mergedBooleanValues.Timestamps)-1]
		cb, err := EncodeBooleanArrayBlock(k.mergedBooleanValues, nil) // TODO(edd): pool this buffer
		if err != nil {
			k.handleEncodeError(err, "boolean")
			return nil
		}

		dst = append(dst, &block{
			minTime: minTime,
			maxTime: maxTime,
			key:     k.key,
			b:       cb,
		})
		k.mergedBooleanValues.Timestamps = k.mergedBooleanValues.Timestamps[:0]
		k.mergedBooleanValues.Values = k.mergedBooleanValues.Values[:0]
	}
	return dst
}
