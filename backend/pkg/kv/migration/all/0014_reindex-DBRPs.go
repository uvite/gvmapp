package all

import (
	"github.com/influxdata/influxdb/v2/dbrp"
	"github.com/uvite/gvmapp/backend/pkg/kv"
)

var Migration0014_ReindexDBRPs = kv.NewIndexMigration(dbrp.ByOrgIDIndexMapping)
