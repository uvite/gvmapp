package u8

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/uvite/gvmbot/pkg/dynamic"
	"github.com/uvite/gvmbot/pkg/indicator"
	"github.com/uvite/gvmbot/pkg/style"
	"io"
	"math"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/uvite/gvmbot/pkg/bbgo"
	"github.com/uvite/gvmbot/pkg/datatype/floats"
	"github.com/uvite/gvmbot/pkg/fixedpoint"
	"github.com/uvite/gvmbot/pkg/types"

	"github.com/uvite/gvm/engine"
	vite "github.com/uvite/gvm/tart/floats"
)

const ID = "u8"

var log = logrus.WithField("strategy", ID)
var Four fixedpoint.Value = fixedpoint.NewFromInt(4)
var Three fixedpoint.Value = fixedpoint.NewFromInt(3)
var Two fixedpoint.Value = fixedpoint.NewFromInt(2)
var Delta fixedpoint.Value = fixedpoint.NewFromFloat(0.01)
var Fee = 0.0008 // taker fee % * 2, for upper bound
type Side string

const (
	SideShort = Side("Short")
	SideLong  = Side("Long")
)

type LimtStop struct {
	limit fixedpoint.Value
	stop  fixedpoint.Value
}

func init() {
	bbgo.RegisterStrategy(ID, &Strategy{})
}

func filterErrors(errs []error) (es []error) {
	for _, e := range errs {
		if _, ok := e.(types.ZeroAssetError); ok {
			continue
		}
		if bbgo.ErrExceededSubmitOrderRetryLimit == e {
			continue
		}
		es = append(es, e)
	}
	return es
}

type Strategy struct {
	Symbol string `json:"symbol"`
	Gvm    *engine.Gvm
	bbgo.OpenPositionOptions
	bbgo.StrategyController
	types.Market
	types.IntervalWindow
	bbgo.SourceSelector

	*bbgo.Environment
	*types.Position    `persistence:"position"`
	*types.ProfitStats `persistence:"profit_stats"`
	*types.TradeStats  `persistence:"trade_stats"`

	p *types.Position

	MinInterval   types.Interval `json:"minInterval"`                   // minimum interval referred for doing stoploss/trailing exists and updating highest/lowest
	Debug         bool           `json:"debug" modifiable:"true"`       // to print debug message or not
	UseStopLoss   bool           `json:"useStopLoss" modifiable:"true"` // whether to use stoploss rate to do stoploss
	UseHighLow    bool           `json:"useHighLow" modifiable:"true"`  // whether to use stoploss rate to do stoploss
	HighLowWindow int            `json:"highLowWindow"`                 // trendLine is used for rebalancing the position. When trendLine goes up, hold base, otherwise hold quote

	UseAtr    bool `json:"useAtr" modifiable:"true"` // use atr as stoploss
	WindowATR int  `json:"windowATR"`

	StopLoss  fixedpoint.Value `json:"stoploss" modifiable:"true"` // stoploss rate
	ctx       context.Context
	close     *vite.Slice
	high      *vite.Slice
	low       *vite.Slice
	open      *vite.Slice
	volume    *vite.Slice
	long      Side
	short     Side
	atr       *indicator.ATR
	trendLine types.UpdatableSeriesExtend
	beta      float64 // last beta value from trendline's linear regression (previous slope of the trendline)

	counter int
	//elapsed    *types.Queue
	//priceLines *types.Queue

	//ma         types.UpdatableSeriesExtend
	//stdevHigh              *indicator.StdDev
	//stdevLow               *indicator.StdDev
	//drift                  *DriftMA

	midPrice     fixedpoint.Value // the midPrice is the average of bestBid and bestAsk in public orderbook
	Price        fixedpoint.Value // the midPrice is the average of bestBid and bestAsk in public orderbook
	lock         sync.RWMutex     `ignore:"true"` // lock for midPrice
	positionLock sync.RWMutex     `ignore:"true"` // lock for highest/lowest and p
	pendingLock  sync.Mutex       `ignore:"true"`
	startTime    time.Time        // trading start time

	maxCounterBuyCanceled  int            // the largest counter of the order on the buy side been cancelled. meaning the latest cancelled buy order.
	maxCounterSellCanceled int            // the largest counter of the order on the sell side been cancelled. meaning the latest cancelled sell order.
	orderPendingCounter    map[uint64]int // records the timepoint when the orders are created, using the counter at the time.

	PredictOffset             int     `json:"predictOffset"`                          // the lookback length for the prediction using linear regression
	HighLowVarianceMultiplier float64 `json:"hlVarianceMultiplier" modifiable:"true"` // modifier to set the limit order price
	NoTrailingStopLoss        bool    `json:"noTrailingStopLoss" modifiable:"true"`   // turn off the trailing exit and stoploss

	HLRangeWindow         int `json:"hlRangeWindow"`         // ma window for kline high/low changes
	SmootherWindow        int `json:"smootherWindow"`        // window that controls the smoothness of drift
	FisherTransformWindow int `json:"fisherTransformWindow"` // fisher transform window to filter drift's negative signals
	ATRWindow             int `json:"atrWindow"`

	// window for atr indicator
	PendingMinInterval int `json:"pendingMinInterval" modifiable:"true"` // if order not be traded for pendingMinInterval of time, cancel it.

	NoRebalance bool `json:"noRebalance" modifiable:"true"` // disable rebalance

	TrendWindow             int       `json:"trendWindow"`                       // trendLine is used for rebalancing the position. When trendLine goes up, hold base, otherwise hold quote
	RebalanceFilter         float64   `json:"rebalanceFilter" modifiable:"true"` // beta filter on the Linear Regression of trendLine
	TrailingCallbackRate    []float64 `json:"trailingCallbackRate" modifiable:"true"`
	TrailingActivationRatio []float64 `json:"trailingActivationRatio" modifiable:"true"`

	buyPrice     float64 `persistence:"buy_price"`     // price when a long position is opened
	sellPrice    float64 `persistence:"sell_price"`    // price when a short position is opened
	highestPrice float64 `persistence:"highest_price"` // highestPrice when the position is opened
	lowestPrice  float64 `persistence:"lowest_price"`  // lowestPrice when the position is opened

	// This is not related to trade but for statistics graph generation
	// Will deduct fee in percentage from every trade
	//GraphPNLDeductFee bool   `json:"graphPNLDeductFee"`
	//CanvasPath        string `json:"canvasPath"`       // backtest related. the path to store the indicator graph
	//GraphPNLPath      string `json:"graphPNLPath"`     // backtest related. the path to store the pnl % graph per trade graph.
	//GraphCumPNLPath   string `json:"graphCumPNLPath"`  // backtest related. the path to store the asset changes in graph
	//GraphElapsedPath  string `json:"graphElapsedPath"` // the path to store the elapsed time in ms
	//GenerateGraph     bool   `json:"generateGraph"`    // whether to generate graph when shutdown

	ExitMethods bbgo.ExitMethodSet `json:"exits"`
	Session     *bbgo.ExchangeSession
	*bbgo.GeneralOrderExecutor

	getLastPrice   func() fixedpoint.Value
	longLimitStop  *LimtStop
	shortLimitStop *LimtStop

	LongLeverage  fixedpoint.Value `json:"longLeverage"`
	ShortLeverage fixedpoint.Value `json:"shortLeverage"`
	StopPrice     fixedpoint.Value

	Running          bool
	OpenningBalances fixedpoint.Value `json:"openningBalances"`
}

func (s *Strategy) ID() string {
	return ID
}

func (s *Strategy) InstanceID() string {
	return fmt.Sprintf("%s:%s:%v", ID, s.Symbol, bbgo.IsBackTesting)
}

func (s *Strategy) Subscribe(session *bbgo.ExchangeSession) {
	//fmt.Println(s.MinInterval, "====", s.Interval)
	session.Subscribe(types.KLineChannel, s.Symbol, types.SubscribeOptions{
		Interval: s.MinInterval,
	})
	session.Subscribe(types.KLineChannel, s.Symbol, types.SubscribeOptions{
		Interval: s.Interval,
	})

	if !bbgo.IsBackTesting {
		session.Subscribe(types.BookTickerChannel, s.Symbol, types.SubscribeOptions{})
	}
	s.ExitMethods.SetAndSubscribe(session, s)

}

func (s *Strategy) CurrentPosition() *types.Position {
	return s.Position
}

const closeOrderRetryLimit = 5

func (s *Strategy) initIndicators(store *bbgo.MarketDataStore) error {
	s.atr = &indicator.ATR{IntervalWindow: types.IntervalWindow{Interval: s.Interval, Window: s.WindowATR}}
	s.trendLine = &indicator.EWMA{IntervalWindow: types.IntervalWindow{Interval: s.Interval, Window: s.TrendWindow}}
	s.close = &vite.Slice{}
	s.high = &vite.Slice{}
	s.low = &vite.Slice{}
	s.open = &vite.Slice{}
	s.volume = &vite.Slice{}
	s.long = SideLong
	s.short = SideShort

	klines, ok := store.KLinesOfInterval(s.Interval)

	klineLength := len(*klines)

	if !ok || klineLength == 0 {
		return errors.New("klines not exists")
	}
	tmpK := (*klines)[klineLength-1]
	s.startTime = tmpK.StartTime.Time().Add(tmpK.Interval.Duration())

	for _, kline := range *klines {

		//fmt.Println("kline.StartTime", kline.StartTime)
		s.AddKline(kline)

	}
	return nil
}
func (s *Strategy) AddKline(kline types.KLine) {
	source := s.GetSource(&kline).Float64()

	s.close.Push(kline.Close.Float64())
	s.high.Push(kline.High.Float64())
	s.low.Push(kline.Low.Float64())
	s.open.Push(kline.Open.Float64())
	s.volume.Push(kline.Volume.Float64())

	s.atr.PushK(kline)
	s.trendLine.Update(source)

	//closes.Push(kline.Close.Float64())
	//high.Push(kline.High.Float64())
	//low.Push(kline.Low.Float64())
	//open.Push(kline.Open.Float64())

	//fmt.Println(s.atr.Last())
	//fmt.Println("s.close.Last():", s.close.Last())

}
func Plot(line *vite.Slice, args ...any) {

	fmt.Println(line.Last(), args)
}

func (s *Strategy) Engine() {
	gvm, _ := engine.NewGvm()

	pwd, _ := os.Getwd()

	filepath := fmt.Sprintf("%s/%s", pwd, "bbgo.js")
	err := gvm.LoadFile(filepath)
	fmt.Println(err)
	//gvm.Ctx = ctx
	s.Gvm = gvm

}
func (s *Strategy) InitEngine() {
	//初始化之前设置ctx
	s.Gvm.Set("close", s.close)
	s.Gvm.Set("open", s.open)
	s.Gvm.Set("low", s.low)
	s.Gvm.Set("high", s.high)
	s.Gvm.Set("volume", s.volume)
	s.Gvm.Set("symbol", s.Symbol)
	s.Gvm.Set("price", &s.Price)
	s.Gvm.Set("strategy", s)
	s.Gvm.Set("postion", s.Position)
	s.Gvm.Set("plot", Plot)
	s.Gvm.Set("log", log)

}

func (s *Strategy) RunCode(file string, options map[string]interface{}) (goja.Value, error) {
	vm, _ := engine.NewGvm()
	fmt.Println(file)
	err := vm.LoadFile(file)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	vm.Ctx = ctx
	vm.Init()

	fmt.Println(err)
	//
	vm.Set("close", s.close)
	vm.Set("open", s.open)
	vm.Set("low", s.low)
	vm.Set("high", s.high)
	vm.Set("volume", s.volume)
	vm.Set("symbol", s.Symbol)
	vm.Set("price", &s.Price)
	vm.Set("postion", s.Position)
	vm.Set("plot", Plot)
	vm.Set("strategy", s)
	vm.Set("log", log)
	vm.Set("options", options)
	value, ok := vm.Vu.RunDefault()
	fmt.Println(ok)
	if ok != nil {
		return nil, ok
	}
	return value, ok

}
func (s *Strategy) Run(ctx context.Context, orderExecutor bbgo.OrderExecutor, session *bbgo.ExchangeSession) error {
	s.ctx = ctx
	//ectx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	s.Engine()
	s.Gvm.Ctx = ctx
	s.Gvm.Init()
	s.InitEngine()
	//杠杆倍数
	if s.Leverage == fixedpoint.Zero {
		s.Leverage = fixedpoint.One
	}

	instanceID := s.InstanceID()
	// Will be set by persistence if there's any from DB
	if s.Position == nil {
		s.Position = types.NewPositionFromMarket(s.Market)
		s.p = types.NewPositionFromMarket(s.Market)
	} else {
		s.p = types.NewPositionFromMarket(s.Market)
		s.p.Base = s.Position.Base
		s.p.Quote = s.Position.Quote
		s.p.AverageCost = s.Position.AverageCost
	}
	if s.ProfitStats == nil {
		s.ProfitStats = types.NewProfitStats(s.Market)
	}
	if s.TradeStats == nil {
		s.TradeStats = types.NewTradeStats(s.Symbol)
	}
	// StrategyController
	s.Status = types.StrategyStatusRunning
	s.Running = true
	s.OnSuspend(func() {
		_ = s.GeneralOrderExecutor.GracefulCancel(ctx)
	})

	s.OnEmergencyStop(func() {
		_ = s.GeneralOrderExecutor.GracefulCancel(ctx)
		_ = s.ClosePosition(ctx, fixedpoint.One)
	})

	s.GeneralOrderExecutor = bbgo.NewGeneralOrderExecutor(session, s.Symbol, ID, instanceID, s.Position)
	s.GeneralOrderExecutor.BindEnvironment(s.Environment)
	s.GeneralOrderExecutor.BindProfitStats(s.ProfitStats)
	s.GeneralOrderExecutor.BindTradeStats(s.TradeStats)
	s.GeneralOrderExecutor.TradeCollector().OnPositionUpdate(func(position *types.Position) {
		bbgo.Sync(ctx, s)
	})
	s.GeneralOrderExecutor.Bind()

	s.orderPendingCounter = make(map[uint64]int)

	// Exit methods from config
	for _, method := range s.ExitMethods {
		method.Bind(session, s.GeneralOrderExecutor)
	}

	profit := floats.Slice{1., 1.}
	price, _ := s.Session.LastPrice(s.Symbol)
	initAsset := s.CalcAssetValue(price).Float64()

	cumProfit := floats.Slice{initAsset, initAsset}
	modify := func(p float64) float64 {
		return p
	}

	s.GeneralOrderExecutor.TradeCollector().OnTrade(func(trade types.Trade, _profit, _netProfit fixedpoint.Value) {
		//spew.Dump(trade)
		//fmt.Println("fuck", trade)
		s.p.AddTrade(trade)
		price := trade.Price.Float64()

		//fmt.Println(s.p.IsLong(), s.p.IsShort())
		//fmt.Println(s.p.String())
		s.pendingLock.Lock()
		delete(s.orderPendingCounter, trade.OrderID)
		s.pendingLock.Unlock()

		if s.buyPrice > 0 {
			profit.Update(modify(price / s.buyPrice))
			cumProfit.Update(s.CalcAssetValue(trade.Price).Float64())
		} else if s.sellPrice > 0 {
			profit.Update(modify(s.sellPrice / price))
			cumProfit.Update(s.CalcAssetValue(trade.Price).Float64())
		}
		s.positionLock.Lock()
		if s.p.IsDust(trade.Price) {
			s.buyPrice = 0
			s.sellPrice = 0
			s.highestPrice = 0
			s.lowestPrice = 0
		} else if s.p.IsLong() {
			s.buyPrice = s.p.ApproximateAverageCost.Float64()
			s.sellPrice = 0
			s.highestPrice = math.Max(s.buyPrice, s.highestPrice)
			s.lowestPrice = s.buyPrice
		} else if s.p.IsShort() {
			s.sellPrice = s.p.ApproximateAverageCost.Float64()
			s.buyPrice = 0
			s.highestPrice = s.sellPrice
			if s.lowestPrice == 0 {
				s.lowestPrice = s.sellPrice
			} else {
				s.lowestPrice = math.Min(s.lowestPrice, s.sellPrice)
			}
		}
		s.positionLock.Unlock()

	})

	//s.frameKLine = &types.KLine{}
	//s.klineMin = &types.KLine{}

	//s.priceLines = types.NewQueue(300)
	//s.elapsed = types.NewQueue(60000)

	s.initTickerFunctions(ctx)
	s.startTime = s.Environment.StartTime()
	s.TradeStats.SetIntervalProfitCollector(types.NewIntervalProfitCollector(types.Interval1d, s.startTime))
	s.TradeStats.SetIntervalProfitCollector(types.NewIntervalProfitCollector(types.Interval1w, s.startTime))

	// event trigger order: s.Interval => Interval1m
	store, ok := session.MarketDataStore(s.Symbol)
	if !ok {
		panic("cannot get 1m history")
	}
	if err := s.initIndicators(store); err != nil {
		log.WithError(err).Errorf("initIndicator failed")
		return nil
	}
	//s.Gvm.Ctx = ctx
	//v, e := s.Gvm.Run()
	//fmt.Println(v, e)
	s.InitEngine()
	store.OnKLineClosed(func(kline types.KLine) {
		//var fn func(int642 int64) int64
		//err := s.JsVm.ExportTo(s.JsVm.Get("A"), &fn)
		//if err != nil {
		//	panic(err)
		//}
		//
		//fmt.Println(fn(4))

		s.counter = int(kline.StartTime.Time().Add(kline.Interval.Duration()).Sub(s.startTime).Milliseconds()) / s.MinInterval.Milliseconds()
		//s.minutesCounter = int(kline.StartTime.Time().Add(kline.Interval.Duration()).Sub(s.startTime).Minutes())
		//if s.Stop {
		//	if s.StopTime.Before(time.Now()) {
		//		s.Stop = false
		//		bbgo.Notify("静默期 。。。。。")
		//
		//	}
		//}

		if kline.Interval == s.Interval {
			//fmt.Println(kline)
			s.AddKline(kline)

			if bbgo.IsBackTesting {
				//fmt.Println("kline close")
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				s.Gvm.Ctx = ctx
				s.Gvm.Init()
				s.InitEngine()
				v, e := s.Gvm.Run()
				fmt.Println(v, e)
			} else {
				s.Gvm.Run()
			}
			//
			//	//initVU, err := r.NewVU(ctx, 1, 1, make(chan metrics.SampleContainer, 100))
			//	//require.NoError(t, err)
			//	//vu := initVU.Activate(&lib.VUActivationParams{RunContext: ctx})
			//	//err = vu.RunOnce()

			//	ch := make(chan metrics.SampleContainer, 100)
			//	initVU, err := runner.NewVU(ctx, 1, 1, ch)
			//
			//	vu := initVU.Activate(&lib.VUActivationParams{RunContext: ctx})
			//	value, ok := vu.RunDefault()
			//	fmt.Println(value, ok, err)
			//	////fmt.Println("====")
			//	//if ok := s.JsVm.Vu.RunOnce(); ok != nil {
			//	//
			//	//	fmt.Errorf("jsvm run err : %w", ok)
			//	//
			//	//}
			//}
		}
		if kline.Interval == s.MinInterval {
			s.klineHandlerMin(ctx, kline, s.counter)
		}
	})
	s.longLimitStop = &LimtStop{
		limit: 0.0,
		stop:  0.0,
	}
	s.shortLimitStop = &LimtStop{
		limit: 0.0,
		stop:  0.0,
	}
	s.CheckLimitStop()

	bbgo.OnShutdown(ctx, func(ctx context.Context, wg *sync.WaitGroup) {

		var buffer bytes.Buffer

		s.Print(&buffer, true, true)

		fmt.Fprintln(&buffer, "--- NonProfitable Dates ---")
		for _, daypnl := range s.TradeStats.IntervalProfits[types.Interval1d].GetNonProfitableIntervals() {
			fmt.Fprintf(&buffer, "%s\n", daypnl)
		}
		fmt.Fprintln(&buffer, s.TradeStats.BriefString())

		os.Stdout.Write(buffer.Bytes())

		//if s.GenerateGraph {
		//	s.Draw(s.frameKLine.StartTime, &profit, &cumProfit)
		//}
		wg.Done()
	})

	return nil
}
func (s *Strategy) Print(f io.Writer, pretty bool, withColor ...bool) {
	var tableStyle *table.Style
	if pretty {
		tableStyle = style.NewDefaultTableStyle()
	}
	dynamic.PrintConfig(s, f, tableStyle, len(withColor) > 0 && withColor[0], dynamic.DefaultWhiteList()...)
}
