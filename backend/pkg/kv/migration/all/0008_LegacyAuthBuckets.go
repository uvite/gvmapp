package all

import (
	"github.com/uvite/gvmapp/backend/pkg/kv/migration"
)

var Migration0008_LegacyAuthBuckets = migration.CreateBuckets(
	"Create Legacy authorization buckets",
	[]byte("legacy/authorizationsv1"), []byte("legacy/authorizationindexv1"))
