package all

import (
	"github.com/influxdata/influxdb/v2/v1/services/meta"
	"github.com/uvite/gvmapp/backend/pkg/kv/migration"
)

var Migration0007_CreateMetaDataBucket = migration.CreateBuckets(
	"Create TSM metadata buckets",
	meta.BucketName)
