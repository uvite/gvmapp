package all

import (
	"github.com/influxdata/influxdb/v2/telegraf"
	"github.com/uvite/gvmapp/backend/pkg/kv"
)

// Migration0010_AddIndexTelegrafByOrg adds the index telegraf configs by organization ID
var Migration0010_AddIndexTelegrafByOrg = kv.NewIndexMigration(telegraf.ByOrganizationIndexMapping, kv.WithIndexMigrationCleanup)
