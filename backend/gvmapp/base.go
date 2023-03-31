package gvmapp

import (
	"github.com/uvite/gvmapp/backend/pkg/bot"
	"github.com/uvite/gvmapp/backend/util"
)

//	"TUSDUSDT", "IOTAUSDT", "XLMUSDT", "ONTUSDT", "TRXUSDT", "ETCUSDT", "ICXUSDT", "VENUSDT", "NULSUSDT", "VETUSDT"}

func (a *App) AppSetting() *util.Resp {
	resultMap := make(map[string]interface{}, 0)
	resultMap["symbols"] = bot.Symbols
	resultMap["intervals"] = bot.Intervals
	return util.Success(resultMap)
}
