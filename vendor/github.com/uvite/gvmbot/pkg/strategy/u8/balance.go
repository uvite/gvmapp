package u8

import (
	"context"
	"fmt"
	"github.com/uvite/gvmbot/pkg/fixedpoint"
	"github.com/uvite/gvmbot/pkg/types"
	"math"
)

func (s *Strategy) UpdateBalance() {
	account, err := s.Session.UpdateAccount(s.ctx)
	fmt.Println(account, err)
	balances := s.Session.GetAccount().Balances()

	desc := fmt.Sprintf("当前:%s,可用:%s,锁定:%s",
		balances[s.Market.QuoteCurrency].Currency,
		balances[s.Market.QuoteCurrency].Available,
		balances[s.Market.QuoteCurrency].Locked,
	)

	log.Info(desc)

}

// Sending new rebalance orders cost too much.
// Modify the position instead to expect the strategy itself rebalance on Close
func (s *Strategy) Rebalance(ctx context.Context) {
	price := s.getLastPrice()
	_, beta := types.LinearRegression(s.trendLine, 3)
	if math.Abs(beta) > s.RebalanceFilter && math.Abs(s.beta) > s.RebalanceFilter || math.Abs(s.beta) < s.RebalanceFilter && math.Abs(beta) < s.RebalanceFilter {
		return
	}
	balances := s.GeneralOrderExecutor.Session().GetAccount().Balances()
	baseBalance := balances[s.Market.BaseCurrency].Total()
	quoteBalance := balances[s.Market.QuoteCurrency].Total()
	total := baseBalance.Add(quoteBalance.Div(price))
	percentage := fixedpoint.One.Sub(Delta)
	log.Infof("rebalance beta %f %v", beta, s.p)
	if beta > s.RebalanceFilter {
		if total.Mul(percentage).Compare(baseBalance) > 0 {
			q := total.Mul(percentage).Sub(baseBalance)
			s.p.Lock()
			defer s.p.Unlock()
			s.p.Base = q.Neg()
			s.p.Quote = q.Mul(price)
			s.p.AverageCost = price
		}
	} else if beta <= -s.RebalanceFilter {
		if total.Mul(percentage).Compare(quoteBalance.Div(price)) > 0 {
			q := total.Mul(percentage).Sub(quoteBalance.Div(price))
			s.p.Lock()
			defer s.p.Unlock()
			s.p.Base = q
			s.p.Quote = q.Mul(price).Neg()
			s.p.AverageCost = price
		}
	} else {
		if total.Div(Two).Compare(quoteBalance.Div(price)) > 0 {
			q := total.Div(Two).Sub(quoteBalance.Div(price))
			s.p.Lock()
			defer s.p.Unlock()
			s.p.Base = q
			s.p.Quote = q.Mul(price).Neg()
			s.p.AverageCost = price
		} else if total.Div(Two).Compare(baseBalance) > 0 {
			q := total.Div(Two).Sub(baseBalance)
			s.p.Lock()
			defer s.p.Unlock()
			s.p.Base = q.Neg()
			s.p.Quote = q.Mul(price)
			s.p.AverageCost = price
		} else {
			s.p.Lock()
			defer s.p.Unlock()
			s.p.Reset()
		}
	}
	log.Infof("rebalanceafter %v %v %v", baseBalance, quoteBalance, s.p)
	s.beta = beta
}

func (s *Strategy) CalcAssetValue(price fixedpoint.Value) fixedpoint.Value {
	balances := s.Session.GetAccount().Balances()
	return balances[s.Market.BaseCurrency].Total().Mul(price).Add(balances[s.Market.QuoteCurrency].Total())
}
