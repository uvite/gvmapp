package services

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/uvite/gvmapp/backend/gvmbot"
	"github.com/uvite/gvmapp/backend/util"
	"github.com/uvite/gvmbot/pkg/bbgo"
	"github.com/uvite/gvmbot/pkg/types"
)

type SymbolInterval struct {
	Symbol   string
	Interval string
}

var SymbolIntervals = []SymbolInterval{}
var (
	Symbols   = []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "BCCUSDT", "NEOUSDT", "LTCUSDT", "QTUMUSDT", "ADAUSDT", "XRPUSDT", "EOSUSDT"}
	Intervals = []string{"1m", "3m", "5m", "15m", "30m", "1h", "4h", "8h", "1d"}
)

type ExchangeService struct {
	Ctx context.Context

	Exchange *gvmbot.Exchange
}

func NewExchangeService() *ExchangeService {

	exchangeService := &ExchangeService{}

	return exchangeService
}

// 初始化交易所
func (e *ExchangeService) Init() {

	filepath := fmt.Sprintf("%s/%s", util.GetEnvDir(), ".env.local")
	configpath := fmt.Sprintf("%s/%s", util.GetEnvDir(), "bbgo.yaml")
	fmt.Println(filepath, configpath)
	ex := gvmbot.New(filepath, configpath, "abc")
	environ := bbgo.NewEnvironment()
	if err := environ.ConfigureExchangeSessions(ex.UserConfig); err != nil {

	}

	session, ok := environ.Session(ex.SessionName)
	if !ok {

	}

	s := session.Exchange.NewStream()
	s.SetPublicOnly()
	ex.Stream = s
	ex.Session = session
	e.Exchange = ex
	e.Subscript()

}

func (eb *ExchangeService) Subscript() {

	for _, symbol := range Symbols {
		for _, interval := range Intervals {
			eb.AddSymbolInterval(symbol, interval)
		}
	}

	if err := eb.Exchange.Stream.Connect(eb.Ctx); err != nil {
		log.Error(err)
	}

}

func (eb *ExchangeService) AddSymbolInterval(symbol string, interval string) {

	eb.Exchange.Stream.Subscribe(types.KLineChannel, symbol, types.SubscribeOptions{Interval: types.Interval(interval)})

	eb.Exchange.Stream.OnKLineClosed(func(kline types.KLine) {
		//log.Infof("kline closed: %s", kline.String())

		//log.Infof("real-%s-%s kline: %s", symbol, interval, kline.String())
		//runtime.EventsEmit(a.Ctx, fmt.Sprintf(string("closed-%s-%s"), symbol, interval))
	})
	//
	eb.Exchange.Stream.OnKLine(func(kline types.KLine) {
		//log.Infof("real-%s-%s kline: %s", symbol, interval, kline.String())
		//runtime.EventsEmit(a.Ctx, fmt.Sprintf(string("real-%s-%s"), symbol, interval))
	})
	//si := &SymbolInterval{
	//	Symbol:   symbol,
	//	Interval: interval,
	//}
}
