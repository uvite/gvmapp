package bot

import (
	"context"
	"fmt"
	"github.com/uvite/gvmbot/pkg/bbgo"
	"github.com/uvite/gvmbot/pkg/types"
	"github.com/uvite/gvmdesk/gvmbot"
	"os"
	"sync"
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

// Launcher represents the main program execution.
type Exbot struct {
	wg       sync.WaitGroup
	cancel   func()
	doneChan <-chan struct{}

	Exchange *gvmbot.Exchange
	Ctx      context.Context
}

func NewExbot(ctx context.Context) *Exbot {
	return &Exbot{
		Ctx: ctx,
	}
}

// 初始化交易所
func (eb *Exbot) InitExchange() {
	pwd, _ := os.Getwd()
	fmt.Println(pwd)
	filepath := fmt.Sprintf("%s/adev/%s", pwd, ".env.local")
	configpath := fmt.Sprintf("%s/adev/%s", pwd, "bbgo.yaml")
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
	eb.Exchange = ex

}

func (eb *Exbot) Subscript() {
	//symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "BCCUSDT", "NEOUSDT", "LTCUSDT", "QTUMUSDT", "ADAUSDT", "XRPUSDT", "EOSUSDT"}
	////	"TUSDUSDT", "IOTAUSDT", "XLMUSDT", "ONTUSDT", "TRXUSDT", "ETCUSDT", "ICXUSDT", "VENUSDT", "NULSUSDT", "VETUSDT"}
	//intervals := []string{"1m", "3m", "5m", "15m", "30m", "1h", "4h", "8h", "1d"}

	for _, symbol := range  Symbols {
		for _, interval := range  Intervals {
			eb.AddSymbolInterval(symbol, interval)
		}
		//fmt.Println(k)
		//app.AddSymbolInterval(k, "1m")
		//app.AddSymbolInterval(k, "3m")
		//app.AddSymbolInterval(k, "5m")
		//app.AddSymbolInterval(k, "15m")
	}

	if err := eb.Exchange.Stream.Connect(eb.Ctx); err != nil {
		fmt.Println(err)
	}

	//
	//item := internal.AlertItem{
	//	Title:    "rsi",
	//	Symbol:   symbol,
	//	Interval: interval,
	//	Path:     "/hein/gvmdesk/js/4.js",
	//}
	//app.RunTestFile(item)
}

func (eb *Exbot) AddSymbolInterval(symbol string, interval string) {

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
