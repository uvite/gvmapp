package gvmbot

import (
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"
	"github.com/uvite/gvmbot/pkg/bbgo"
	"github.com/uvite/gvmbot/pkg/exchange/binance"
	"github.com/uvite/gvmbot/pkg/exchange/okex/okexapi"
	"github.com/uvite/gvmbot/pkg/fixedpoint"
	"github.com/uvite/gvmbot/pkg/strategy/u8"
	"github.com/uvite/gvmbot/pkg/types"
)

type Options struct {
	qty     fixedpoint.Value `json:"qty"`
	price   fixedpoint.Value `json:"price"`
	limit   fixedpoint.Value `json:"limit"`
	stop    fixedpoint.Value `json:"stop"`
	comment string           `json:"comment"`
	tag     string           `json:"tag"`
	okLimit bool             `json:"okLimit"` //oklimit 为限价单，非algo
}

func (ex Exchange) Entry(id string, side u8.Side, data map[string]interface{}) {

	params := u8.Keys(data)

	option := &Options{}
	for _, k := range params {
		switch k {
		case "qty":
			option.qty, _ = fixedpoint.NewFromString(data["qty"].(string))
		case "price":
			option.price, _ = fixedpoint.NewFromString(data["price"].(string))

		case "limit":
			option.limit, _ = fixedpoint.NewFromString(data["limit"].(string))
		case "stop":
			option.stop, _ = fixedpoint.NewFromString(data["stop"].(string))
		case "okLimit":
			option.okLimit = data["okLimit"].(bool)
		case "comment":
			option.comment = data["comment"].(string)
		case "tag":
			option.tag = data["tag"].(string)

		}
	}
	fmt.Printf("%+v", option)
	ex.OpenOrder(side, option)

}

func (ex Exchange) OpenOrder(side u8.Side, options *Options) {
	direct := types.SideTypeBuy
	redirect := types.SideTypeBuy
	if side == u8.SideLong {
		direct = types.SideTypeBuy
		redirect = types.SideTypeSell
	} else if side == u8.SideShort {
		direct = types.SideTypeSell
		redirect = types.SideTypeBuy
	}
	quantity := options.qty

	createdOrder := types.SubmitOrder{
		Symbol:   ex.Symbol,
		Side:     direct,
		Quantity: quantity,
		Tag:      options.comment,
	}
	//有价格为限价单，没有为市价单
	if options.price > 0 {
		createdOrder.Type = types.OrderTypeLimit
		createdOrder.Price = options.price
	} else {
		createdOrder.Type = types.OrderTypeMarket
	}

	//减仓挂单操作 止赢
	if options.limit > 0 {
		//ok限价 非algo
		if !options.okLimit {
			if options.price > 0 {
				createdOrder.Type = types.OrderTypeTakeProfitLimit

			} else {
				createdOrder.Type = types.OrderTypeTakeProfitMarket
			}
		}
		createdOrder.Side = redirect
		createdOrder.StopPrice = options.limit
		createdOrder.ReduceOnly = true
	}
	//减仓挂单操作 止损
	if options.stop > 0 {

		//ok限价 非algo
		if !options.okLimit {

			if options.price > 0 {
				createdOrder.Type = types.OrderTypeStopLimit

			} else {
				createdOrder.Type = types.OrderTypeStopMarket
			}
		}

		createdOrder.Side = redirect
		createdOrder.StopPrice = options.stop
		createdOrder.ReduceOnly = true
	}

	ex.SubmitOrder(createdOrder)

	return

}

// 市价平仓
func (ex Exchange) Exit(id string, side u8.Side, data map[string]interface{}) {

	params := u8.Keys(data)

	option := &Options{}
	for _, k := range params {
		switch k {
		case "qty":
			option.qty, _ = fixedpoint.NewFromString(data["qty"].(string))
		//case "price":
		//	option.price, _ = fixedpoint.NewFromString(data["price"].(string))
		//
		//case "limit":
		//	option.limit, _ = fixedpoint.NewFromString(data["limit"].(string))
		//case "stop":
		//	option.stop, _ = fixedpoint.NewFromString(data["stop"].(string))
		//case "okLimit":
		//	option.okLimit = data["okLimit"].(bool)
		case "comment":
			option.comment = data["comment"].(string)
		case "tag":
			option.tag = data["tag"].(string)

		}
	}
	fmt.Printf("%+v", option)
	ex.CloseOrder(side, option)

}

func (ex Exchange) CloseOrder(side u8.Side, options *Options) {
	direct := types.SideTypeBuy

	if side == u8.SideLong {

		direct = types.SideTypeSell
	} else if side == u8.SideShort {

		direct = types.SideTypeBuy
	}
	quantity := options.qty

	createdOrder := types.SubmitOrder{
		Symbol:     ex.Symbol,
		Side:       direct,
		Quantity:   quantity,
		Tag:        options.comment,
		ReduceOnly: true,
	}
	createdOrder.Type = types.OrderTypeMarket
	ex.okQty = true
	ex.SubmitOrder(createdOrder)

	return

}

func (ex Exchange) SubmitOrder(createdOrder types.SubmitOrder) (err error) {
	ctx := context.Background()
	environ := bbgo.NewEnvironment()

	if err := environ.ConfigureExchangeSessions(ex.UserConfig); err != nil {

	}
	fmt.Println("[0]", ex.SessionName)

	session, ok := environ.Session(ex.SessionName)
	if !ok {

	}

	if ex.SessionName == "okex" {
		markets, err := session.Exchange.QueryMarkets(ctx)
		if err != nil {
			fmt.Println(err)
		}
		market, ok := markets[ex.Symbol]
		fmt.Println(market)
		if !ok {
			fmt.Errorf("market %s is not defined  ", market)
		}
		if ex.okQty {
			q := createdOrder.Quantity.Round(1, fixedpoint.Down).Div(market.TickSize)

			fmt.Println("[q]", q)
			createdOrder.Quantity = q
		}
		createdOrder.Market = market
		createdOrder.TimeInForce = "GTC"

	}
	spew.Dump("=====")
	fmt.Printf("%+v", createdOrder)
	a, err := session.Exchange.SubmitOrder(ctx, createdOrder)
	fmt.Println(a)
	fmt.Println(err)
	return nil
}

func (e *Exchange) Cancel(id string, args ...any) {

}

/*
*
strategy.cancel_all() → void
*/
func (ex *Exchange) CancelAll(args ...any) {
	ctx := context.Background()
	if ex.SessionName == "binance" {
		key, secret := viper.GetString("binance-api-key"), viper.GetString("binance-api-secret")

		var exchange = binance.New(key, secret)
		exchange.IsFutures = true

		err := exchange.FClient.NewCancelAllOpenOrdersService().Symbol(ex.Symbol).Do(ctx)
		fmt.Println(err)
	} else if ex.SessionName == "okex" {
		key, secret, passphrase := viper.GetString("okex-api-key"),
			viper.GetString("okex-api-secret"),
			viper.GetString("okex-api-passphrase")
		if len(key) == 0 || len(secret) == 0 {

		}

		client := okexapi.NewClient()
		client.Auth(key, secret, passphrase)
		res, _ := client.CancelPendingAlgos(ex.Symbol)
		//	orderid := gjson.Get(data1.String(), "data[0].algoId")
		//
		fmt.Println("fuck 止损单：", res)

		//res,err:=client.CancelAlgos(algoId)
		//fmt.Println(res,err)
	}

}
