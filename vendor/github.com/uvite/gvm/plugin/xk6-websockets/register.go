// Package websockets exist just to register the websockets extension
package websockets

import (
	"github.com/uvite/gvm/pkg/js/modules"
	"github.com/uvite/gvm/plugin/xk6-websockets/websockets"
)

func init() {
	modules.Register("k6/x/websockets", new(websockets.RootModule))
}
