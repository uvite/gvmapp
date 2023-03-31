package gvmbot

import (
	"context"
	"fmt"

	"github.com/uvite/gvmbot/pkg/bbgo"
)

func (ex Exchange) GetAccount() (err error) {
	ctx := context.Background()
	environ := bbgo.NewEnvironment()
	if err := environ.ConfigureDatabase(ctx); err != nil {

	}

	if err := environ.ConfigureExchangeSessions(ex.UserConfig); err != nil {

	}
	fmt.Println("[0]", ex.SessionName)

	session, ok := environ.Session(ex.SessionName)
	if !ok {

	}

	if err != nil {
		return err
	}

	a, err := session.Exchange.QueryAccount(ctx)

	if err != nil {
	}
	//db := global.GVA_DB
	balance, ok := a.Balance("USDT")
	fmt.Printf("%f,%f", balance.Available.Float64(), balance.Locked.Float64())
	fmt.Printf("%s", balance.String())
	fmt.Printf("%s", ex.ExchangeId)
	//dt := &bots.GvmBalance{
	//	ExchangeId: ex.exchangeId,
	//	Available:  balance.Available.String(),
	//	Locked:     balance.Locked.String(),
	//}
	//db.Create(dt)

	return nil
}
