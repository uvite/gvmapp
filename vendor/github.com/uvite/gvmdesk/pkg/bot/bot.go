package bot

import (
	"context"
	"fmt"
	"github.com/uvite/gvm/engine"
	vite "github.com/uvite/gvm/tart/floats"
	"github.com/uvite/gvmbot/pkg/types"
	taskmodel "github.com/uvite/gvmdesk/pkg/model"
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

type Deskbot struct {
	Gvm      *engine.Gvm
	Task     taskmodel.Task
	Exbot    *Exbot
	Symbol   string
	Interval string

	close  *vite.Slice
	high   *vite.Slice
	low    *vite.Slice
	open   *vite.Slice
	volume *vite.Slice
}

func NewBot(ctx context.Context, task taskmodel.Task, eb *Exbot) *Deskbot {

	gvm, _ := engine.NewGvm()
	deskbot := &Deskbot{
		Gvm:   gvm,
		Task:  task,
		Exbot: eb,
	}

	deskbot.close = &vite.Slice{}
	deskbot.high = &vite.Slice{}
	deskbot.low = &vite.Slice{}
	deskbot.open = &vite.Slice{}
	deskbot.volume = &vite.Slice{}

	fmt.Printf("[task]:%+v \n", task)
	file := task.Metadata["path"].(string)
	symbol := task.Metadata["symbol"].(string)
	interval := task.Metadata["interval"].(string)
	deskbot.Symbol = symbol
	deskbot.Interval = interval
	//gvm, _ := engine.NewGvm()
	fmt.Println(file)
	err := gvm.LoadFile(file)
	fmt.Println(err)
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	gvm.Ctx = ctx
	gvm.Init()

	fmt.Println("[2]", err)

	gvm.Set("close", deskbot.close)
	gvm.Set("open", deskbot.open)
	gvm.Set("low", deskbot.low)
	gvm.Set("high", deskbot.high)
	gvm.Set("volume", deskbot.volume)
	gvm.Set("symbol", symbol)
	gvm.Set("interval", interval)
	fmt.Println("[3]", err)

	now := time.Now()
	kLines, err := eb.Exchange.Session.Exchange.QueryKLines(ctx, symbol, types.Interval(interval), types.KLineQueryOptions{
		Limit:   1500,
		EndTime: &now,
	})
	if err != nil {
		fmt.Println(err)
	}
	//log.Infof("kLines from RESTful API")
	for _, kline := range kLines {
		//log.Info(kline.String())
		//fmt.Println(kline.String())
		deskbot.close.Push(kline.Close.Float64())
		deskbot.high.Push(kline.High.Float64())
		deskbot.low.Push(kline.Low.Float64())
		deskbot.open.Push(kline.Open.Float64())
		deskbot.volume.Push(kline.Volume.Float64())
	}
	return deskbot
}

func (db *Deskbot) OnklineClose() {
	db.Exbot.Exchange.Stream.OnKLineClosed(func(kline types.KLine) {
		if kline.Symbol == db.Symbol && kline.Interval == types.Interval(db.Interval) {
			db.close.Push(kline.Close.Float64())
			db.high.Push(kline.High.Float64())
			db.low.Push(kline.Low.Float64())
			db.open.Push(kline.Open.Float64())
			db.volume.Push(kline.Volume.Float64())
			db.Gvm.Run()
		}

	})

}
func (db *Deskbot) Onkline() {
	db.Exbot.Exchange.Stream.OnKLine(func(kline types.KLine) {
		//db.close.Push(kline.Close.Float64())
		//db.high.Push(kline.High.Float64())
		//db.low.Push(kline.Low.Float64())
		//db.open.Push(kline.Open.Float64())
		//db.volume.Push(kline.Volume.Float64())
		db.Gvm.Run()
	})

}
