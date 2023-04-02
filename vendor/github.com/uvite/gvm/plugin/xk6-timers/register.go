// Package timers is here just to register the k6/x/events module
package timers

import (
	"github.com/uvite/gvm/pkg/js/modules"
	"github.com/uvite/gvm/plugin/xk6-timers/timers"
)

func init() {
	modules.Register("k6/x/timers", new(timers.RootModule))
}
