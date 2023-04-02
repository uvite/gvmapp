package all

import "github.com/uvite/gvmdesk/pkg/kv/migration"

var Migration0009_LegacyAuthPasswordBuckets = migration.CreateBuckets(
	"Create legacy auth password bucket",
	[]byte("legacy/authorizationPasswordv1"))
