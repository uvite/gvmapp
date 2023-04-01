package services

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/uvite/gvmapp/backend/gvmbot"
	"github.com/uvite/gvmapp/backend/util"
	"github.com/uvite/gvmbot/pkg/types"

	"github.com/uvite/gvm/engine"
	vite "github.com/uvite/gvm/tart/floats"

	taskmodel "github.com/uvite/gvmapp/backend/pkg/model"
	"github.com/uvite/gvmapp/backend/services/bot"
	"path/filepath"

	"time"
)

var (
	newVoice = bot.NewVoice()
)

type BotService struct {
	Ctx context.Context

	//*ExchangeService
	Exchange *gvmbot.Exchange
	Gvm      *engine.Gvm
	Task     taskmodel.Task

	Symbol   string
	Interval string

	close  *vite.Slice
	high   *vite.Slice
	low    *vite.Slice
	open   *vite.Slice
	volume *vite.Slice
	price  *vite.Slice
}

func NewBotService() *BotService {
	Bot := &BotService{}
	return Bot
}

// 创建一个新的机器人
func (b *BotService) NewBot(task *taskmodel.Task) {
	gvm, _ := engine.NewGvm()
	b.Gvm = gvm
	b.close = &vite.Slice{}
	b.high = &vite.Slice{}
	b.low = &vite.Slice{}
	b.open = &vite.Slice{}
	b.volume = &vite.Slice{}
	b.price = &vite.Slice{}
	configData := util.GetConfigDir()

	file := filepath.Join(configData, "js", (task.ID.String() + ".js"))

	symbol := task.Symbol
	interval := task.Interval
	b.Symbol = symbol
	b.Interval = interval

	err := gvm.LoadFile(file)
	if err != nil {
		log.Error(err)
	}
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	gvm.Ctx = b.Ctx
	gvm.Init()
	gvm.Set("close", b.close)
	gvm.Set("open", b.open)
	gvm.Set("low", b.low)
	gvm.Set("high", b.high)
	gvm.Set("volume", b.volume)
	gvm.Set("price", b.price)
	gvm.Set("symbol", symbol)
	gvm.Set("interval", interval)

	gvm.Set("alert", newVoice.Alert)

	log.Infof("gvm set", b.Symbol)

}
func (e *BotService) GetKline() {

	now := time.Now()
	kLines, err := e.Exchange.Session.Exchange.QueryKLines(e.Ctx, e.Symbol, types.Interval(e.Interval), types.KLineQueryOptions{
		Limit:   1500,
		EndTime: &now,
	})
	if err != nil {
		fmt.Println(err)
	}
	log.Infof("kLines from RESTful API")
	for _, kline := range kLines {
		//log.Info(kline.String())
		//fmt.Println(kline.String())
		e.close.Push(kline.Close.Float64())
		e.high.Push(kline.High.Float64())
		e.low.Push(kline.Low.Float64())
		e.open.Push(kline.Open.Float64())
		e.volume.Push(kline.Volume.Float64())
		e.Gvm.Run()

	}
}

func (e *BotService) OnklineClose() {
	log.Info("onklineclose")
	e.Exchange.Stream.OnKLineClosed(func(kline types.KLine) {
		fmt.Println("e.Symbol", e.Symbol)
		if kline.Symbol == e.Symbol && kline.Interval == types.Interval(e.Interval) {
			e.close.Push(kline.Close.Float64())
			e.high.Push(kline.High.Float64())
			e.low.Push(kline.Low.Float64())
			e.open.Push(kline.Open.Float64())
			e.volume.Push(kline.Volume.Float64())
			//db.price = fixedpoint.Value(kline.Close.Float64())

			e.Gvm.Run()
		}

	})

}
func (e *BotService) Onkline() {
	log.Info("Onkline")

	e.Exchange.Stream.OnKLine(func(kline types.KLine) {

		fmt.Println("e.Symbol kline", e.Symbol)
		if kline.Symbol == e.Symbol && kline.Interval == types.Interval(e.Interval) {
			if e.price.Len() > 5 {
				e.price.Pop(0)
			}

			e.price.Push(kline.Close.Float64())
			//fmt.Println(db.price.Len(), db.price.Tail(3))
		}
	})

}

func (b *BotService) SetExchange(service *ExchangeService) {
	b.Exchange = service.Exchange
}
