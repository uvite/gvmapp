package services

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/uvite/gvmapp/backend/pkg/executor"
	"github.com/uvite/gvmapp/backend/services/bot"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"sync"
)

type PoolService struct {
	Ctx context.Context

	Exchange *ExchangeService

	currentCancel sync.Map
}

func NewPoolService() *PoolService {
	Pool := &PoolService{}

	return Pool
}

func (l *PoolService) Listen() {
	log.Info("alert Listening ")
	runtime.EventsOn(l.Ctx, "service.pool.startbot", func(data ...interface{}) {
		fmt.Printf("%+v", data)

		promise := data[0].(*executor.Promise)
		fmt.Printf("%+v", promise)
		log.Info("开始", promise.Id)
		l.StartBot(promise)
	})
	runtime.EventsOn(l.Ctx, "service.pool.closebot", func(data ...interface{}) {
		fmt.Printf("data %+v \n", data)

		promise := data[0].(*executor.Promise)
		fmt.Printf("promise%+v\n", promise)
		log.Info("关闭", promise.Id)

		l.CloseBot(promise)
	})
}
func (b *PoolService) StartBot(promise *executor.Promise) {
	log.Info("[start bot]", promise.Id)
	_, ok := b.currentCancel.Load(promise.Id)
	if ok {
		return
	}
	bot := NewBotService()
	bot.Ctx = b.Ctx
	bot.SetExchange(b.Exchange)
	bot.NewBot(promise.Task)
	bot.GetKline()
	bot.OnklineClose()
	bot.Onkline()
	b.currentCancel.Store(promise.Id, bot.Gvm.CancelFun)

	//b.Bots = append(b.Bots, bot)
}
func (b *PoolService) CloseBot(promise *executor.Promise) {

	close, ok := b.currentCancel.Load(promise.Id)
	log.Info("[close bot]", close, ok, promise.Id)
	if ok {
		close.(context.CancelFunc)()
	}
	b.currentCancel.Delete(promise.Id)

	//b.Bots = append(b.Bots, bot)
}

func (b *PoolService) SetExchange(service *ExchangeService) {
	b.Exchange = service
}

func (b *PoolService) Speak(msg string) {
	newVoice := bot.NewVoice()
	newVoice.Alert(msg)
}
