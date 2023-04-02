package u8

import (
	"fmt"
	"github.com/uvite/gvmbot/pkg/fixedpoint"
	"github.com/uvite/gvmbot/pkg/types"
)

func (s *Strategy) CheckStopLoss() bool {
	flag := false
	//百分比止损
	if s.UseStopLoss {
		//stoploss := s.StopLoss.Float64()
		base := s.p.GetBase()

		//quantity := base.Abs()
		price := s.getLastPrice()

		if base.Sign() > 0 {
			if price.Compare(s.StopPrice) < 0 {
				flag = true
				log.Infoln("止损：", price, s.StopPrice)
			}
		}
		if base.Sign() < 0 {
			if price.Compare(s.StopPrice) > 0 {
				flag = true
				log.Infoln("止损：", price, s.StopPrice)

			}
		}
		//fmt.Println("s.p,quantity", s.p, quantity)
		//fmt.Println("stop loss", s.StopPrice, s.highestPrice, s.lowestPrice)
		//if s.sellPrice > 0 && s.StopPrice.Float64() <= s.highestPrice ||
		//	s.buyPrice > 0 && s.StopPrice.Float64() >= s.lowestPrice {
		//	fmt.Println("stop loss active ")
		//
		//	flag = true
		//}
	}
	//s.calculateNetValue()
	//最高最低点止损
	if s.UseHighLow {
		hlwindow := s.HighLowWindow
		Highest := types.Highest(s.high, hlwindow)
		Lowest := types.Lowest(s.low, hlwindow)
		//fmt.Println("UseHighLow:", s.sellPrice, Highest, Lowest)
		if s.sellPrice > 0 && Highest <= s.highestPrice ||
			s.buyPrice > 0 && Lowest >= s.lowestPrice {
			fmt.Println("UseHighLow active ")

			flag = true
		}
	}

	if s.UseAtr {
		atr := s.atr.Last()
		//fmt.Println("atr stop ", s.sellPrice, atr, s.sellPrice+1*atr, s.highestPrice)
		if s.sellPrice > 0 && s.sellPrice+1*atr <= s.highestPrice ||
			s.buyPrice > 0 && s.buyPrice-2*atr >= s.lowestPrice {
			flag = true
			fmt.Println("atr stop active ")

		}
	}
	return flag
}

// 计算净值
func (s *Strategy) calculateNetValue() bool {
	s.UpdateBalance()
	balances := s.GeneralOrderExecutor.Session().GetAccount().Balances()
	flag := balances[s.Market.QuoteCurrency].Available.Compare(s.OpenningBalances.Mul(fixedpoint.NewFromFloat(0.8)))

	if flag <= 0 {
		//退出订单
		s.Exit("最大止损")
		s.Running = false
		return false
	} else {
		s.Running = true
	}
	decs := fmt.Sprintf("期初资金:%s,停机资金:%s,是否达到:%t,机器人状态:%t",
		s.OpenningBalances,
		s.OpenningBalances.Mul(fixedpoint.NewFromFloat(0.8)),
		flag <= 0, s.Running,
	)

	log.Infoln(decs)
	return true
	//fmt.Printf("可用余额:%s,开单方向:%s,杠杆倍数:%s,开单数量:%s,现价:%s,止损价：%s", balances[s.Market.QuoteCurrency].Available,
	//direct, leverage, quoteAmount, price, stopPrice)

}
