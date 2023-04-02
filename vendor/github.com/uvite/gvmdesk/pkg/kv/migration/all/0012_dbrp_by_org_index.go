package all

import (
	"github.com/influxdata/influxdb/v2/dbrp"
	"github.com/uvite/gvmdesk/pkg/kv"
)

var Migration0012_DBRPByOrgIndex = kv.NewIndexMigration(dbrp.ByOrgIDIndexMapping, kv.WithIndexMigrationCleanup)
