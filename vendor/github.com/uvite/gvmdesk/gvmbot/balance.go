package gvmbot

import (
	"context"
	"fmt"

	"github.com/uvite/gvmbot/pkg/bbgo"
	"github.com/uvite/gvmbot/pkg/types"
)

func (ex Exchange) GetBalance() (balance *types.Balance, ok error) {
	ctx := context.Background()
	environ := bbgo.NewEnvironment()
	if err := environ.ConfigureDatabase(ctx); err != nil {

	}

	if err := environ.ConfigureExchangeSessions(ex.UserConfig); err != nil {

	}
	fmt.Println("[0]", ex.SessionName)

	if len(ex.SessionName) > 0 {
		session, ok := environ.Session(ex.SessionName)
		if !ok {

		}

		a, err := session.Exchange.QueryAccountBalances(ctx)

		if err != nil {

		}

		balance, ok := a.Copy()["USDT"]
		fmt.Printf("%s", balance.String())
		//db := global.GVA_DB
		//dt := &bots.GvmBalance{
		//	ExchangeCode: ex.exchangeId,
		//	Available:    balance.Available.String(),
		//	Locked:       balance.Locked.String(),
		//}
		//db.Create(dt)

	}

	return &types.Balance{}, nil
}
