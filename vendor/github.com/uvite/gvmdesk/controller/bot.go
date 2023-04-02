package gtools

import (
	"context"
	"fmt"
	"github.com/spf13/afero"
	"github.com/uvite/gvm/engine"
	"github.com/uvite/gvm/pkg/lib"
	vite "github.com/uvite/gvm/tart/floats"
	"github.com/uvite/gvmbot/pkg/types"
	"github.com/uvite/gvmdesk/internal"
	"os"
	"time"
)

// /策略现场
type GvmOptions struct {
	Exchange string `json:"exchange" form:"exchange"`
	Symbol   string `json:"symbol" form:"symbol"`
	Interval string `json:"interval" form:"interval"`
	Sname    string `json:"sname" form:"sname"`
	Code     string `json:"code" form:"code"`
}
type GvmBotTest struct {
	GvmOptions
	close  *vite.Slice
	high   *vite.Slice
	low    *vite.Slice
	open   *vite.Slice
	volume *vite.Slice
}

func (a *App) RunTestCode(alert internal.AlertItem) {

	var options GvmBotTest

	options.close = &vite.Slice{}
	options.high = &vite.Slice{}
	options.low = &vite.Slice{}
	options.open = &vite.Slice{}
	options.volume = &vite.Slice{}
	options.Code = alert.Content
	fmt.Printf("[bot]:%+v", options)
	fmt.Printf("[alert]:%+v", alert)

	gvm, _ := engine.NewGvm()
	pwd, _ := os.Getwd()
	fs := afero.NewOsFs()

	filepath := fmt.Sprintf("%s/js/test/%s", pwd, "file")
	rtOpts := lib.RuntimeOptions{}
	r, err := gvm.GetSimpleRunner(filepath, fmt.Sprintf(`
			import {Nats} from 'k6/x/nats';
			import ta from 'k6/x/ta';
			import {sleep} from 'k6'; 

			%s

			`, options.Code),
		fs, rtOpts)

	fmt.Println("\n[1=====]", options.Code, "\n")

	gvm.Runner = r
	gvm.Runtime = r.Bundle.Vm

	ctx, cancel := context.WithCancel(a.Ctx)
	//GvmBaseApi.Cancel = cancel
	defer cancel()
	gvm.Ctx = ctx
	gvm.Init()
	fmt.Println("[2]", err)

	gvm.Set("close", options.close)
	gvm.Set("open", options.open)
	gvm.Set("low", options.low)
	gvm.Set("high", options.high)
	gvm.Set("volume", options.volume)
	gvm.Set("symbol", options.Symbol)
	fmt.Println("[3]", err)

	a.Exchange.Stream.OnKLineClosed(func(kline types.KLine) {
		//log.Infof("kline closed: %s", kline.String())
		//fmt.Println("[4]", err)

		if kline.Symbol == alert.Symbol && kline.Interval.String() == alert.Interval {
			gvm.Vu.RunOnce()
		}

		//runtime.EventsEmit(a.Ctx, fmt.Sprintf(string("closed-%s-%s"), symbol, interval))
	})

	a.Exchange.Stream.OnKLine(func(kline types.KLine) {

		if kline.Symbol == alert.Symbol && kline.Interval.String() == alert.Interval {
			//fmt.Println("234324")
			gvm.Vu.RunDefault()
		}
		//fmt.Println("[5]", err)
		//
		//fmt.Println(kline.Symbol, alert.Symbol, kline.Interval.String(), alert.Interval)
		//log.Infof("kline: %s", kline.String())
		//runtime.EventsEmit(a.Ctx, fmt.Sprintf(string("real-%s-%s"), symbol, interval))
	})
	//gvm.Vu.RunOnce()

}
func (a *App) RunTestFile(alert internal.AlertItem) {

	var options GvmBotTest

	options.close = &vite.Slice{}
	options.high = &vite.Slice{}
	options.low = &vite.Slice{}
	options.open = &vite.Slice{}
	options.volume = &vite.Slice{}
	options.Code = alert.Content
	//fmt.Printf("[bot]:%+v", options)
	fmt.Printf("[alert]:%+v \n", alert)
	file := alert.Path
	gvm, _ := engine.NewGvm()
	fmt.Println(file)
	err := gvm.LoadFile(file)
	fmt.Println(err)
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	gvm.Ctx = a.Ctx
	gvm.Init()

	fmt.Println("[2]", err)

	gvm.Set("close", options.close)
	gvm.Set("open", options.open)
	gvm.Set("low", options.low)
	gvm.Set("high", options.high)
	gvm.Set("volume", options.volume)
	gvm.Set("symbol", alert.Symbol)
	gvm.Set("interval", alert.Interval)
	fmt.Println("[3]", err)

	now := time.Now()
	kLines, err := a.Exchange.Session.Exchange.QueryKLines(a.Ctx, alert.Symbol, types.Interval(alert.Interval), types.KLineQueryOptions{
		Limit:   100,
		EndTime: &now,
	})
	if err != nil {
		fmt.Println(err)
	}
	//log.Infof("kLines from RESTful API")
	for _, kline := range kLines {
		//log.Info(kline.String())
		//fmt.Println(kline.String())
		options.close.Push(kline.Close.Float64())
		options.high.Push(kline.High.Float64())
		options.low.Push(kline.Low.Float64())
		options.open.Push(kline.Open.Float64())
		options.volume.Push(kline.Volume.Float64())
	}
	//store, ok := a.Exchange.Session.MarketDataStore(alert.Symbol)
	//if !ok {
	//	panic("cannot get 1m history")
	//}
	//klines, ok := store.KLinesOfInterval(types.Interval(alert.Interval))
	//fmt.Println(klines, ok)
	//klineLength := len(*klines)
	//
	//if !ok || klineLength == 0 {
	//	errors.New("klines not exists")
	//}

	a.Exchange.Stream.OnKLineClosed(func(kline types.KLine) {
		options.close.Push(kline.Close.Float64())
		options.high.Push(kline.High.Float64())
		options.low.Push(kline.Low.Float64())
		options.open.Push(kline.Open.Float64())
		options.volume.Push(kline.Volume.Float64())

		//log.Infof("kline closed: %s", kline.String())
		//fmt.Println("[4]", err)
		//gvm.Vu.RunDefault()
		//if kline.Symbol == alert.Symbol && kline.Interval.String() == alert.Interval {
		//	gvm.Vu.RunDefault()
		//}

		//runtime.EventsEmit(a.Ctx, fmt.Sprintf(string("closed-%s-%s"), symbol, interval))
	})
	go gvm.Vu.RunDefault()
	a.Exchange.Stream.OnKLine(func(kline types.KLine) {

		//v, err := gvm.Vu.RunDefault()
		//fmt.Println(v, err)
		//if kline.Symbol == alert.Symbol && kline.Interval.String() == alert.Interval {
		//	//fmt.Println("234324")
		//	gvm.Vu.RunDefault()
		//}
		//fmt.Println("[5]", err)
		//
		//fmt.Println(kline.Symbol, alert.Symbol, kline.Interval.String(), alert.Interval)
		//log.Infof("kline: %s", kline.String())
		//runtime.EventsEmit(a.Ctx, fmt.Sprintf(string("real-%s-%s"), symbol, interval))
	})
	//gvm.Vu.RunOnce()

}
