package websockets

import (
	"github.com/dop251/goja"
	"github.com/uvite/gvm/pkg/js/common"
)

// must is a small helper that will panic if err is not nil.
func must(rt *goja.Runtime, err error) {
	if err != nil {
		common.Throw(rt, err)
	}
}
