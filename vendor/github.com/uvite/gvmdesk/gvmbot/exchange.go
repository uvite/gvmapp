package gvmbot

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"github.com/uvite/gvmbot/pkg/bbgo"
	"github.com/uvite/gvmbot/pkg/types"
	"strings"
)

type Exchange struct {
	UserConfig  *bbgo.Config
	Stream      types.Stream
	Session      *bbgo.ExchangeSession
	SessionName string
	Symbol      string
	ExchangeId  string
	okQty       bool `json:"okLimit"` //操蛋的okex 张币转换

}

func New(dotenvFile string, configFile string, exchangeId string) *Exchange {

	if err := godotenv.Overload(dotenvFile); err != nil {
		fmt.Println(err, "error loading dotenv file")
	}
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	//
	exchange, symbol := viper.GetString("exchange"),
		viper.GetString("symbol")
	fmt.Println("[exchange]", exchange, symbol, exchangeId)
	userConfig, err := bbgo.Load(configFile, false)
	if err != nil {
		fmt.Println(err)
	}
	return &Exchange{
		SessionName: exchange,
		Symbol:      symbol,
		UserConfig:  userConfig,
		ExchangeId:  exchangeId,
	}

}
