package all

import (
	"github.com/influxdata/influxdb/v2/dbrp"
	"github.com/uvite/gvmapp/backend/pkg/kv"
)

var Migration0012_DBRPByOrgIndex = kv.NewIndexMigration(dbrp.ByOrgIDIndexMapping, kv.WithIndexMigrationCleanup)
