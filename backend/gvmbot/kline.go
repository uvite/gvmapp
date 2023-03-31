package gvmbot

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/uvite/gvmbot/pkg/bbgo"
	"github.com/uvite/gvmbot/pkg/cmd/cmdutil"
	"github.com/uvite/gvmbot/pkg/types"
	"syscall"
	"time"
)

func (ex Exchange) SubKline() (err error) {
	ctx := context.Background()

	environ := bbgo.NewEnvironment()
	if err := environ.ConfigureExchangeSessions(ex.UserConfig); err != nil {
		return err
	}

	session, ok := environ.Session(ex.SessionName)
	if !ok {
		return fmt.Errorf("session %s not found", ex.SessionName)
	}
	interval := "1m"
	symbol := "BTCUSDT"
	now := time.Now()
	kLines, err := session.Exchange.QueryKLines(ctx, ex.Symbol, types.Interval(interval), types.KLineQueryOptions{
		Limit:   50,
		EndTime: &now,
	})
	if err != nil {
		return err
	}
	log.Infof("kLines from RESTful API")
	for _, k := range kLines {
		log.Info(k.String())
	}

	s := session.Exchange.NewStream()
	s.SetPublicOnly()

	s.Subscribe(types.KLineChannel, symbol, types.SubscribeOptions{Interval: types.Interval(interval)})
	s.Subscribe(types.KLineChannel, symbol, types.SubscribeOptions{Interval: types.Interval("5m")})
	s.Subscribe(types.KLineChannel, symbol, types.SubscribeOptions{Interval: types.Interval("15m")})
	s.Subscribe(types.KLineChannel, symbol, types.SubscribeOptions{Interval: types.Interval("30m")})
	s.Subscribe(types.KLineChannel, symbol, types.SubscribeOptions{Interval: types.Interval("1h")})
	s.Subscribe(types.KLineChannel, "ETHUSDT", types.SubscribeOptions{Interval: types.Interval(interval)})
	s.Subscribe(types.KLineChannel, "ETHUSDT", types.SubscribeOptions{Interval: types.Interval("5m")})
	s.Subscribe(types.KLineChannel, "ETHUSDT", types.SubscribeOptions{Interval: types.Interval("15m")})
	s.Subscribe(types.KLineChannel, "ETHUSDT", types.SubscribeOptions{Interval: types.Interval("30m")})
	s.Subscribe(types.KLineChannel, "ETHUSDT", types.SubscribeOptions{Interval: types.Interval("1h")})

	s.Subscribe(types.KLineChannel, "BNBUSDT", types.SubscribeOptions{Interval: types.Interval(interval)})
	s.Subscribe(types.KLineChannel, "BNBUSDT", types.SubscribeOptions{Interval: types.Interval("5m")})
	s.Subscribe(types.KLineChannel, "BNBUSDT", types.SubscribeOptions{Interval: types.Interval("15m")})
	s.Subscribe(types.KLineChannel, "BNBUSDT", types.SubscribeOptions{Interval: types.Interval("30m")})
	s.Subscribe(types.KLineChannel, "BNBUSDT", types.SubscribeOptions{Interval: types.Interval("1h")})

	s.Subscribe(types.KLineChannel, "LTCUSDT", types.SubscribeOptions{Interval: types.Interval(interval)})
	s.Subscribe(types.KLineChannel, "LTCUSDT", types.SubscribeOptions{Interval: types.Interval("5m")})
	s.Subscribe(types.KLineChannel, "LTCUSDT", types.SubscribeOptions{Interval: types.Interval("15m")})
	s.Subscribe(types.KLineChannel, "LTCUSDT", types.SubscribeOptions{Interval: types.Interval("30m")})
	s.Subscribe(types.KLineChannel, "LTCUSDT", types.SubscribeOptions{Interval: types.Interval("1h")})

	s.OnKLineClosed(func(kline types.KLine) {
		log.Infof("kline closed: %s", kline.String())
	})

	s.OnKLine(func(kline types.KLine) {
		log.Infof("kline: %s", kline.String())
	})

	log.Infof("connecting...")
	if err := s.Connect(ctx); err != nil {
		return err
	}

	log.Infof("connected")
	defer func() {
		log.Infof("closing connection...")
		if err := s.Close(); err != nil {
			log.WithError(err).Errorf("connection close error")
		}
	}()

	cmdutil.WaitForSignal(ctx, syscall.SIGINT, syscall.SIGTERM)
	return nil
}
