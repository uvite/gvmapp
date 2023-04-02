package types

import "github.com/uvite/gvmbot/pkg/datatype/floats"

var _ Series = floats.Slice([]float64{}).Addr()
