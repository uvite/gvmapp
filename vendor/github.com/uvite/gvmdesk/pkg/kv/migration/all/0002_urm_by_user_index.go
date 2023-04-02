package all

import (
	"github.com/influxdata/influxdb/v2/tenant/index"
	"github.com/uvite/gvmdesk/pkg/kv"
)

// Migration0002_AddURMByUserIndex creates the URM by user index and populates missing entries based on the source.
var Migration0002_AddURMByUserIndex = kv.NewIndexMigration(index.URMByUserIndexMapping, kv.WithIndexMigrationCleanup)
