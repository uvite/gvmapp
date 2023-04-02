package binanceapi

import (
	"time"

	"github.com/c9s/requestgen"

	"github.com/uvite/gvmbot/pkg/fixedpoint"
	"github.com/uvite/gvmbot/pkg/types"
)

type MarginInterestRate struct {
	Asset             string                     `json:"asset"`
	DailyInterestRate fixedpoint.Value           `json:"dailyInterestRate"`
	Timestamp         types.MillisecondTimestamp `json:"timestamp"`
	VipLevel          int                        `json:"vipLevel"`
}

//go:generate requestgen -method GET -url "/sapi/v1/margin/interestRateHistory" -type GetMarginInterestRateHistoryRequest -responseType []MarginInterestRate
type GetMarginInterestRateHistoryRequest struct {
	client requestgen.APIClient

	asset     string     `param:"asset"`
	startTime *time.Time `param:"startTime,milliseconds"`
	endTime   *time.Time `param:"endTime,milliseconds"`
}

func (c *RestClient) NewGetMarginInterestRateHistoryRequest() *GetMarginInterestRateHistoryRequest {
	return &GetMarginInterestRateHistoryRequest{client: c}
}
