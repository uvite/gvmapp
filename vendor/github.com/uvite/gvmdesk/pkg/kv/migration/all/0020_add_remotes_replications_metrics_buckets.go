package all

import "github.com/uvite/gvmdesk/pkg/kv/migration"

var (
	remoteMetricsBucket       = []byte("remotesv2")
	replicationsMetricsBucket = []byte("replicationsv2")
)

var Migration0020_Add_remotes_replications_metrics_buckets = migration.CreateBuckets(
	"create remotes and replications metrics buckets",
	remoteMetricsBucket,
	replicationsMetricsBucket,
)
