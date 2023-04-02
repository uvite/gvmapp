package gtools

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/uvite/gvmbot/pkg/bbgo"
	"github.com/uvite/gvmbot/pkg/types"
	"github.com/uvite/gvmdesk/gvmbot"
	"os"
)

type SymbolInterval struct {
	Symbol   string
	Interval string
}

var SymbolIntervals = []SymbolInterval{}

func (a *App) InitExchange() {
	//go func() {
	//	for {
	//
	//		time.Sleep(1 * time.Second)
	//	}
	//}()
	a.Kline()
}

// 读取坚果云笔记本
func (a *App) Kline() {
	pwd, _ := os.Getwd()
	fmt.Println(pwd)
	filepath := fmt.Sprintf("%s/adev/%s", pwd, ".env.local")
	configpath := fmt.Sprintf("%s/adev/%s", pwd, "bbgo.yaml")

	ex := gvmbot.New(filepath, configpath, "abc")
	//ctx := a.Ctx

	environ := bbgo.NewEnvironment()
	if err := environ.ConfigureExchangeSessions(ex.UserConfig); err != nil {

	}

	session, ok := environ.Session(ex.SessionName)
	if !ok {

	}
	//interval := "1m"
	////symbol := "BTCUSDT"
	//now := time.Now()
	//kLines, err := session.Exchange.QueryKLines(ctx, ex.Symbol, types.Interval(interval), types.KLineQueryOptions{
	//	Limit:   50,
	//	EndTime: &now,
	//})
	//if err != nil {
	//	fmt.Println(err)
	//}
	//log.Infof("kLines from RESTful API")
	//for _, k := range kLines {
	//	log.Info(k.String())
	//}

	s := session.Exchange.NewStream()
	s.SetPublicOnly()
	ex.Stream = s
	ex.Session = session

	//s.Subscribe(types.KLineChannel, ex.Symbol, types.SubscribeOptions{Interval: types.Interval(interval)})
	//
	//if err := s.Connect(ctx); err != nil {
	//	fmt.Println(err)
	//}

	log.Infof("connected")
	//defer func() {
	//	log.Infof("closing connection...")
	//	if err := s.Close(); err != nil {
	//		log.WithError(err).Errorf("connection close error")
	//	}
	//}()
	a.Exchange = ex

}

func (a *App) AddSymbolInterval(symbol string, interval string) RespDate {
	//if err := a.Exchange.Stream.Close(); err != nil {
	//	fmt.Println(err)
	//}
	a.Exchange.Stream.Subscribe(types.KLineChannel, symbol, types.SubscribeOptions{Interval: types.Interval(interval)})

	a.Exchange.Stream.OnKLineClosed(func(kline types.KLine) {
		//log.Infof("kline closed: %s", kline.String())


		//log.Infof("real-%s-%s kline: %s", symbol, interval, kline.String())
		//runtime.EventsEmit(a.Ctx, fmt.Sprintf(string("closed-%s-%s"), symbol, interval))
	})
	//
	a.Exchange.Stream.OnKLine(func(kline types.KLine) {
		//log.Infof("real-%s-%s kline: %s", symbol, interval, kline.String())
		//runtime.EventsEmit(a.Ctx, fmt.Sprintf(string("real-%s-%s"), symbol, interval))
	})
	si := &SymbolInterval{
		Symbol:   symbol,
		Interval: interval,
	}

	return RespDate{Code: 200, Data: si}

}
