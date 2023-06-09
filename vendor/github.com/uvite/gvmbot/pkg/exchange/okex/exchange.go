package okex

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"github.com/uvite/gvmbot/pkg/exchange/okex/okexapi"
	"github.com/uvite/gvmbot/pkg/fixedpoint"
	"github.com/uvite/gvmbot/pkg/types"
)

var marketDataLimiter = rate.NewLimiter(rate.Every(time.Second/10), 1)
var orderLimiter = rate.NewLimiter(5, 2)
var queryTradeLimiter = rate.NewLimiter(1, 2)

// OKB is the platform currency of OKEx, pre-allocate static string here
const OKB = "OKB"

var log = logrus.WithFields(logrus.Fields{
	"exchange": "okex",
})

func init() {
	_ = types.Exchange(&Exchange{})
	_ = types.MarginExchange(&Exchange{})
	_ = types.FuturesExchange(&Exchange{})
}

type Exchange struct {
	types.MarginSettings
	types.FuturesSettings
	key, secret, passphrase string

	client *okexapi.RestClient
}

func (e *Exchange) QueryPositions(ctx context.Context, symbol string) ([]*types.Positions, error) {
	//instrumentID := toLocalSymbol(symbol)
	req := e.client.TradeService.NewGetAccountPostions().InstrumentType(okexapi.InstrumentTypeSwap)
	postions, err := req.Do(ctx)
	if err != nil {
		return nil, err
	}

	return convertPosition(postions)
}

func New(key, secret, passphrase string) *Exchange {
	client := okexapi.NewClient()

	if len(key) > 0 && len(secret) > 0 {
		client.Auth(key, secret, passphrase)
	}

	return &Exchange{
		key: key,
		// pragma: allowlist nextline secret
		secret:     secret,
		passphrase: passphrase,
		client:     client,
	}
}

func (e *Exchange) Name() types.ExchangeName {
	return types.ExchangeOKEx
}

func (e *Exchange) QueryMarkets(ctx context.Context) (types.MarketMap, error) {
	//var instruments = []okexapi.Instrument{}
	var mt okexapi.InstrumentType
	if e.IsFutures {
		mt = okexapi.InstrumentTypeSwap

	} else {
		mt = okexapi.InstrumentTypeSpot
	}
	instruments, err := e.client.PublicDataService.NewGetInstrumentsRequest().
		InstrumentType(mt).
		Do(ctx)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	markets := types.MarketMap{}

	for _, instrument := range instruments {
		symbol := toGlobalSymbol(instrument.InstrumentID)
		market := types.Market{
			Symbol:      symbol,
			LocalSymbol: instrument.InstrumentID,

			QuoteCurrency: instrument.QuoteCurrency,
			BaseCurrency:  instrument.BaseCurrency,

			// convert tick size OKEx to precision
			PricePrecision:  int(-math.Log10(instrument.TickSize.Float64())),
			VolumePrecision: int(-math.Log10(instrument.LotSize.Float64())),

			// TickSize: OKEx's price tick, for BTC-USDT it's "0.1"
			TickSize: instrument.TickSize,

			// Quantity step size, for BTC-USDT, it's "0.00000001"
			StepSize: fixedpoint.MustNewFromString(instrument.ContractValue),

			// for BTC-USDT, it's "0.00001"
			MinQuantity: instrument.MinSize,

			// OKEx does not offer minimal notional, use 1 USD here.
			MinNotional: fixedpoint.One,
			MinAmount:   fixedpoint.One,

			CtVal: fixedpoint.MustNewFromString(instrument.ContractValue),
		}

		if e.IsFutures {
			market.QuoteCurrency = instrument.SettleCurrency
			market.BaseCurrency = instrument.ContractValueCurrency
		}

		markets[symbol] = market
	}

	return markets, nil
}

func (e *Exchange) QueryTicker(ctx context.Context, symbol string) (*types.Ticker, error) {
	symbol = toLocalSymbol(symbol)

	marketTicker, err := e.client.MarketTicker(symbol)
	if err != nil {
		return nil, err
	}

	return toGlobalTicker(*marketTicker), nil
}

func (e *Exchange) QueryTickers(ctx context.Context, symbols ...string) (map[string]types.Ticker, error) {
	instrumentType := okexapi.InstrumentTypeSpot
	if e.IsFutures {
		instrumentType = okexapi.InstrumentTypeFutures
	}

	marketTickers, err := e.client.MarketTickers(instrumentType)
	if err != nil {
		return nil, err
	}

	tickers := make(map[string]types.Ticker)
	for _, marketTicker := range marketTickers {
		symbol := toGlobalSymbol(marketTicker.InstrumentID)
		ticker := toGlobalTicker(marketTicker)
		tickers[symbol] = *ticker
	}

	if len(symbols) == 0 {
		return tickers, nil
	}

	selectedTickers := make(map[string]types.Ticker, len(symbols))
	for _, symbol := range symbols {
		if ticker, ok := tickers[symbol]; ok {
			selectedTickers[symbol] = ticker
		}
	}

	return selectedTickers, nil
}

func (e *Exchange) PlatformFeeCurrency() string {
	return OKB
}

func (e *Exchange) QueryAccount(ctx context.Context) (*types.Account, error) {
	accountBalance, err := e.client.AccountBalances()
	if err != nil {
		return nil, err
	}

	var account = types.Account{
		AccountType: "SPOT",
	}

	var balanceMap = toGlobalBalance(accountBalance)
	account.UpdateBalances(balanceMap)
	return &account, nil
}

func (e *Exchange) QueryAccountBalances(ctx context.Context) (types.BalanceMap, error) {
	accountBalances, err := e.client.AccountBalances()
	if err != nil {
		return nil, err
	}

	var balanceMap = toGlobalBalance(accountBalances)
	return balanceMap, nil
}

func (e *Exchange) SubmitOrder(ctx context.Context, order types.SubmitOrder) (*types.Order, error) {

	//fmt.Println("order------:", order)
	orderType, err := toLocalOrderType(order.Type)
	if err != nil {
		return nil, err
	}

	switch order.Type {
	case types.OrderTypeTakeProfitLimit, types.OrderTypeTakeProfitMarket, types.OrderTypeStopLimit, types.OrderTypeStopMarket:

		createdOrder, err := e.SubmitAlgoOrder(ctx, order)
		return createdOrder, err
	}

	orderReq := e.client.TradeService.NewPlaceOrderRequest()
	param := order.Params

	orderReq.InstrumentID(toLocalSymbol(order.Symbol))
	orderReq.Side(toLocalSideType(order.Side))

	//if order.Market.Symbol != "" {
	//	fmt.Println(111, order.Market.FormatQuantity(order.Quantity), order.Quantity)
	//	orderReq.Quantity(order.Market.FormatQuantity(order.Quantity))
	//} else {
	//	// TODO report error
	//	fmt.Println(222)
	//	orderReq.Quantity(order.Quantity.FormatString(8))
	//}

	//orderReq.Quantity(order.Quantity.FormatString(8))
	//fmt.Println("order.Market.FormatPrice(order.Price)", order.Market.FormatPrice(order.Price))
	// set price field for limit orders
	switch order.Type {
	case types.OrderTypeStopLimit, types.OrderTypeLimit, types.OrderTypeLimitMaker:
		if order.Market.Symbol != "" {
			fmt.Println(33)
			orderReq.Price(order.Market.FormatPrice(order.Price))
		} else {
			// TODO report error
			fmt.Println(444)
			orderReq.Price(order.Price.FormatString(8))
		}
	case types.OrderTypeMarket, types.OrderTypeStopMarket, types.OrderTypeTakeProfitMarket:
		//okex 市价交易的，货币类型
		if !e.IsFutures {
			orderReq.TgtCcy("base_ccy")

		}

	}

	switch order.TimeInForce {
	case "FOK":
		orderReq.OrderType(okexapi.OrderTypeFOK)
	case "IOC":
		orderReq.OrderType(okexapi.OrderTypeIOC)
	default:
		orderReq.OrderType(orderType)
	}

	if order.Params != nil {
		if param["tdMode"] != nil {
			orderReq.TradeMode(param["tdMode"].(string))

		}
		if param["instId"] != nil {
			orderReq.InstrumentID(param["instId"].(string))

		}
		//orderReq.InstrumentID(param["instId"].(string))
		if param["ccy"] != nil {
			orderReq.CCY(param["ccy"].(string))

		}

	}

	if e.IsFutures {

		//
		if order.ReduceOnly {
			orderReq.Quantity(order.Quantity.String())

			//fmt.Println("close ", order.Quantity.Div(fixedpoint.NewFromFloat(0.1)).Round(0, fixedpoint.Down).String())
			orderReq.ReduceOnly(order.ReduceOnly)

			//orderReq.Quantity(order.Quantity.Div(fixedpoint.NewFromFloat()).Round(0, fixedpoint.Down).String())

		} else {
			//这个地方是个坑 ，测试网和实盘不一样
			if okexapi.PaperTrade() {
				//q := order.Quantity.Div(order.Market.TickSize).Int64()
				//orderReq.Quantity(strconv.FormatInt(q, 10))

				q := order.Quantity.Round(1, fixedpoint.Down).Div(order.Market.TickSize).Int64()
				fmt.Println("\n[9]", order.Quantity, order.Quantity.Round(1, fixedpoint.Down))
				fmt.Println("[12]", q, order.Market.TickSize)
				fmt.Println("[22]", strconv.FormatInt(q, 10))
				orderReq.Quantity(strconv.FormatInt(q, 10))
			} else {
				q := order.Quantity.Round(1, fixedpoint.Down).Div(order.Market.StepSize).Int64()
				orderReq.Quantity(strconv.FormatInt(q, 10))
			}

		}

		orderReq.TradeMode("cross")

	} else {
		orderReq.TgtCcy("base_ccy")
		orderReq.TradeMode("cash")
	}
	//

	//orderReq("cash")
	//fmt.Println(3333)
	orderHead, err := orderReq.Do(ctx)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var orderID int64
	if orderHead.OrderID != "" {
		orderID, err = strconv.ParseInt(orderHead.OrderID, 10, 64)

		if err != nil {
			return nil, err
		}
	} else {
		return nil, nil
	}

	return &types.Order{
		SubmitOrder:      order,
		Exchange:         types.ExchangeOKEx,
		OrderID:          uint64(orderID),
		Status:           types.OrderStatusNew,
		ExecutedQuantity: fixedpoint.Zero,
		IsWorking:        true,
		CreationTime:     types.Time(time.Now()),
		UpdateTime:       types.Time(time.Now()),
		IsMargin:         false,
		IsIsolated:       false,
	}, nil

}
func (e *Exchange) SubmitAlgoOrder(ctx context.Context, order types.SubmitOrder) (*types.Order, error) {

	orderReq := e.client.TradeService.NewPlaceAlgoOrderRequest()

	orderType, err := toLocalOrderType(order.Type)
	if err != nil {
		return nil, err
	}

	orderReq.InstrumentID(toLocalSymbol(order.Symbol))
	orderReq.Side(toLocalSideType(order.Side))

	if order.Market.Symbol != "" {
		orderReq.Quantity(order.Market.FormatQuantity(order.Quantity))
	} else {

		orderReq.Quantity(order.Quantity.FormatString(8))
	}

	switch order.Type {
	case types.OrderTypeTakeProfitLimit:
		orderReq.SetTpTriggerPxType("last")
		orderReq.SetTpTriggerPx(order.Market.FormatPrice(order.Price))
		orderReq.SetTpOrdPx(order.Market.FormatPrice(order.StopPrice))

	case types.OrderTypeTakeProfitMarket:
		orderReq.SetTpTriggerPxType("last")
		orderReq.SetTpTriggerPx(order.Market.FormatPrice(order.StopPrice))
		orderReq.SetTpOrdPx("-1")

	case types.OrderTypeStopLimit:
		orderReq.SetSlTriggerPxType("last")
		orderReq.SetSlTriggerPx(order.Market.FormatPrice(order.Price))
		orderReq.SetSlOrdPx(order.Market.FormatPrice(order.StopPrice))
	case types.OrderTypeStopMarket:
		orderReq.SetSlTriggerPxType("last")
		//orderReq.SetSlTriggerPx("-1")
		orderReq.SetSlTriggerPx(order.Market.FormatPrice(order.StopPrice))
		orderReq.SetSlOrdPx("-1")

	}

	switch order.TimeInForce {
	case "FOK":
		orderReq.OrderType(okexapi.OrderTypeFOK)
	case "IOC":
		orderReq.OrderType(okexapi.OrderTypeIOC)
	default:
		orderReq.OrderType(orderType)
	}

	if e.IsFutures {
		orderReq.TradeMode("cross")
		switch order.Side {
		case types.SideTypeBuy:
			if order.ReduceOnly {
				orderReq.PosSide(okexapi.PosSideTypeSell)
			} else {
				orderReq.PosSide(okexapi.PosSideTypeBuy)
			}
		case types.SideTypeSell:
			if order.ReduceOnly {
				orderReq.PosSide(okexapi.PosSideTypeBuy)
			} else {
				orderReq.PosSide(okexapi.PosSideTypeSell)
			}

		}

		if order.ReduceOnly {

			//fmt.Println("close ", order.Quantity.Div(fixedpoint.NewFromFloat(0.1)).Round(0, fixedpoint.Down).String())
			orderReq.ReduceOnly(order.ReduceOnly)
			q := order.Quantity.Round(1, fixedpoint.Down).Div(order.Market.TickSize).Int64()
			orderReq.Quantity(strconv.FormatInt(q, 10))
			//orderReq.Quantity(order.Quantity.Div(fixedpoint.NewFromFloat(0.1)).Round(0, fixedpoint.Down).String())

		} else {
			orderReq.Quantity(order.Quantity.Div(fixedpoint.NewFromFloat(0.1)).Round(0, fixedpoint.Down).String())

		}

	} else {
		orderReq.TgtCcy("base_ccy")
		orderReq.TradeMode("cash")
	}

	orderReq.OrderType(okexapi.OrderTypeConditional)
	orderHead, err := orderReq.Do(ctx)
	fmt.Println("[algo--error]", err)
	if err != nil {

		return nil, err
	}
	//spew.Dump(orderHead)
	var orderID int64
	if orderHead.AlgoId != "" {
		orderID, err = strconv.ParseInt(orderHead.AlgoId, 10, 64)

		if err != nil {
			return nil, err
		}
	} else {
		return nil, nil
	}

	return &types.Order{
		SubmitOrder:      order,
		Exchange:         types.ExchangeOKEx,
		OrderID:          uint64(orderID),
		Status:           types.OrderStatusNew,
		ExecutedQuantity: fixedpoint.Zero,
		IsWorking:        true,
		CreationTime:     types.Time(time.Now()),
		UpdateTime:       types.Time(time.Now()),
		IsMargin:         false,
		IsIsolated:       false,
	}, nil

}

func (e *Exchange) QueryOpenOrders(ctx context.Context, symbol string) (orders []types.Order, err error) {
	instrumentID := toLocalSymbol(symbol)
	req := e.client.TradeService.NewGetPendingOrderRequest().InstrumentType(okexapi.InstrumentTypeSpot).InstrumentID(instrumentID)
	orderDetails, err := req.Do(ctx)
	if err != nil {
		return orders, err
	}

	orders, err = toGlobalOrders(orderDetails)
	return orders, err
}

func (e *Exchange) CancelOrders(ctx context.Context, orders ...types.Order) error {
	if len(orders) == 0 {
		return nil
	}

	var reqs []*okexapi.CancelOrderRequest
	for _, order := range orders {
		if len(order.Symbol) == 0 {
			return errors.New("symbol is required for canceling an okex order")
		}

		req := e.client.TradeService.NewCancelOrderRequest()
		req.InstrumentID(toLocalSymbol(order.Symbol))
		req.OrderID(strconv.FormatUint(order.OrderID, 10))
		if len(order.ClientOrderID) > 0 {
			req.ClientOrderID(order.ClientOrderID)
		}
		reqs = append(reqs, req)
	}

	batchReq := e.client.TradeService.NewBatchCancelOrderRequest()
	batchReq.Add(reqs...)
	_, err := batchReq.Do(ctx)
	return err
}

func (e *Exchange) NewStream() types.Stream {
	return NewStream(e.client)
}

func (e *Exchange) QueryKLines(ctx context.Context, symbol string, interval types.Interval, options types.KLineQueryOptions) ([]types.KLine, error) {
	if err := marketDataLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	intervalParam := toLocalInterval(interval.String())

	req := e.client.MarketDataService.NewCandlesticksRequest(toLocalSymbol(symbol))
	req.Bar(intervalParam)

	if options.StartTime != nil {
		req.After(options.StartTime.UnixMilli())
	}

	if options.EndTime != nil {
		req.After(options.EndTime.UnixMilli())
	}
	req.Limit(options.Limit)

	candles, err := req.Do(ctx)
	if err != nil {
		return nil, err
	}

	var klines []types.KLine
	for _, candle := range candles {
		klines = append(klines, types.KLine{
			Exchange:    types.ExchangeOKEx,
			Symbol:      symbol,
			Interval:    interval,
			Open:        candle.Open,
			High:        candle.High,
			Low:         candle.Low,
			Close:       candle.Close,
			LastTradeID: 0,
			Closed:      true,
			Volume:      candle.Volume,
			QuoteVolume: candle.VolumeInCurrency,
			StartTime:   types.Time(candle.Time),
			EndTime:     types.Time(candle.Time.Add(interval.Duration() - time.Millisecond)),
		})
	}
	klines = types.SortKLinesAscending(klines)

	return klines, nil

}

func (e *Exchange) QueryTrades(ctx context.Context, symbol string, options *types.TradeQueryOptions) ([]types.Trade, error) {
	if err := queryTradeLimiter.Wait(ctx); err != nil {
		return nil, err
	}
	var remoteTrades []okexapi.OrderDetails

	var params = url.Values{}
	params.Add("instType", "SWAP")
	history := e.client.TradeService.NewGetHistoryTrade()
	history.InstrumentType(okexapi.InstrumentTypeSwap)
	history.InstrumentID(symbol)

	remoteTrades, err := history.Do(ctx)

	if err != nil {
		return nil, err
	}
	trades, err := toGlobalTrades(remoteTrades)

	trades = types.SortTradesAscending(trades)

	return trades, nil

}

// lastOrderID is not supported on MAX
func (e *Exchange) QueryClosedOrders(ctx context.Context, symbol string, since, until time.Time, lastOrderID uint64) ([]types.Order, error) {
	log.Warn("!!!MAX EXCHANGE API NOTICE!!!")
	log.Warn("the since/until conditions will not be effected on closed orders query, max exchange does not support time-range-based query")
	return nil, nil
}
