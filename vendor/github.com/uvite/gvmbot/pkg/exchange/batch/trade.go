package batch

import (
	"context"
	"time"

	"github.com/uvite/gvmbot/pkg/types"
)

var closedErrChan = make(chan error)

func init() {
	close(closedErrChan)
}

type TradeBatchQuery struct {
	types.ExchangeTradeHistoryService
}

func (e TradeBatchQuery) Query(ctx context.Context, symbol string, options *types.TradeQueryOptions) (c chan types.Trade, errC chan error) {
	if options.EndTime == nil {
		now := time.Now()
		options.EndTime = &now
	}

	startTime := *options.StartTime
	endTime := *options.EndTime
	query := &AsyncTimeRangedBatchQuery{
		Type: types.Trade{},
		Q: func(startTime, endTime time.Time) (interface{}, error) {
			return e.ExchangeTradeHistoryService.QueryTrades(ctx, symbol, options)
		},
		T: func(obj interface{}) time.Time {
			return time.Time(obj.(types.Trade).Time)
		},
		ID: func(obj interface{}) string {
			trade := obj.(types.Trade)
			if trade.ID > options.LastTradeID {
				options.LastTradeID = trade.ID
			}
			return trade.Key().String()
		},
	}

	c = make(chan types.Trade, 100)
	errC = query.Query(ctx, c, startTime, endTime)
	return c, errC
}
