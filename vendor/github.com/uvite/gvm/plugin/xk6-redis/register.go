// Package redis only exists to register the redis extension
package redis

import (
	"github.com/uvite/gvm/plugin/xk6-redis/redis"

	"github.com/uvite/gvm/pkg/js/modules"
)

// Register the extension on module initialization, available to
// import from JS as "k6/x/redis".
func init() {
	modules.Register("k6/x/redis", new(redis.RootModule))
}
