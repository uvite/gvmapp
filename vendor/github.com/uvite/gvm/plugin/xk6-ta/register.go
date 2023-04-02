// Package timers is here just to register the k6/x/events module
package ta

import (
	"github.com/uvite/gvm/pkg/js/modules"
	"github.com/uvite/gvm/plugin/xk6-ta/ta"
)

func init() {

	modules.Register("k6/x/ta", new(ta.RootModule))
}
