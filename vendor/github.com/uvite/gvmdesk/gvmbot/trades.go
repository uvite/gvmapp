package gvmbot

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/uvite/gvmbot/pkg/bbgo"
	"github.com/uvite/gvmbot/pkg/types"
	"time"
)

func (ex Exchange) GetTrades() (err error) {
	ctx := context.Background()
	environ := bbgo.NewEnvironment()
	if err := environ.ConfigureDatabase(ctx); err != nil {

	}

	if err := environ.ConfigureExchangeSessions(ex.UserConfig); err != nil {

	}
	fmt.Println("[0]", ex.SessionName)

	session, ok := environ.Session(ex.SessionName)
	if !ok {

	}

	var limit int64
	limit = 10
	now := time.Now()
	since := now.Add(-24 * time.Hour)

	tradeHistoryService, ok := session.Exchange.(types.ExchangeTradeHistoryService)
	if !ok {
		// skip exchanges that does not support trading history services
		log.Warnf("exchange %s does not implement ExchangeTradeHistoryService, skip syncing closed orders (tradesCmd)", session.Exchange.Name())
		return nil
	}

	trades, err := tradeHistoryService.QueryTrades(ctx, ex.Symbol, &types.TradeQueryOptions{
		StartTime:   &since,
		Limit:       limit,
		LastTradeID: 0,
	})
	if err != nil {
		return err
	}

	//db := global.GVA_DB
	//if db==nil{
	//	global.GVA_DB = initialize.Gorm()
	//	db = global.GVA_DB
	//}

	//data := []bots.GvmTrades{}
	log.Infof("%d trades", len(trades))
	for _, trade := range trades {
		//dt := bots.GvmTrades{
		//	ExchangeId: ex.exchangeId,
		//	Trade:      trade,
		//}
		//db.Create(dt)
		//data = append(data, dt)

		log.Infof("TRADE %s %s %4s %s @ %s orderID %d %s amount %v , fee %v %s ",
			trade.Exchange.String(),
			trade.Symbol,
			trade.Side,
			trade.Quantity.FormatString(4),
			trade.Price.FormatString(3),
			trade.OrderID,
			trade.Time.Time().Format(time.StampMilli),
			trade.QuoteQuantity,
			trade.Fee,
			trade.FeeCurrency)
	}
	//db.Create(data)

	return nil
}
