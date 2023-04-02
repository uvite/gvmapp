package types

import (
	"time"

	"github.com/uvite/gvmbot/pkg/fixedpoint"
)

type FundingRate struct {
	FundingRate fixedpoint.Value
	FundingTime time.Time
	Time        time.Time
}
