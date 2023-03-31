package gvmbot

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/uvite/gvmbot/pkg/bbgo"
	"github.com/uvite/gvmbot/pkg/types"
)

func (ex Exchange) GetPostions() (err error) {
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

	postions, err := session.Exchange.QueryPositions(ctx, ex.Symbol)

	if err != nil {
		return err
	}


	//if db==nil{
	//	global.GVA_DB = initialize.Gorm()
	//	db = global.GVA_DB
	//}

	//data := []bots.GvmTrades{}
	log.Infof("%d postions", len(postions),ex.ExchangeId)
	for _, postion := range postions {

		//data = append(data, dt)

		log.Infof("Postion %+v  ",
			postion,
		)
	}
	//db.Create(data)

	return nil
}

func (ex Exchange) RealPostions() (postionsList []*types.Positions,err error) {
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

	postions, err := session.Exchange.QueryPositions(ctx, ex.Symbol)

	if err != nil {
		return nil,err
	}


	return postions,nil
}
