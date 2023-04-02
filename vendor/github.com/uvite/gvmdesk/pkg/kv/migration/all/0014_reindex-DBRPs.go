package all

import (
	"github.com/influxdata/influxdb/v2/dbrp"
	"github.com/uvite/gvmdesk/pkg/kv"
)

var Migration0014_ReindexDBRPs = kv.NewIndexMigration(dbrp.ByOrgIDIndexMapping)
