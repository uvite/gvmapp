package all

import "github.com/uvite/gvmapp/backend/pkg/kv/migration"

// Migration0006_DeleteBucketSessionsv1 removes the sessionsv1 bucket
// from the backing kv store.
var Migration0006_DeleteBucketSessionsv1 = migration.DeleteBuckets("delete sessionsv1 bucket", []byte("sessionsv1"))
