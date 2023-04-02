package launcher

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	platform "github.com/influxdata/influxdb/v2"
	"github.com/influxdata/influxdb/v2/http"
	sqliteMigrations "github.com/influxdata/influxdb/v2/sqlite/migrations"
	"github.com/uvite/gvmdesk/pkg/bolt"
	"github.com/uvite/gvmdesk/pkg/bot"
	"github.com/uvite/gvmdesk/pkg/kv/migration"
	"github.com/uvite/gvmdesk/pkg/kv/migration/all"

	taskmodel "github.com/uvite/gvmdesk/pkg/model"
	nethttp "net/http"
	"time"

	"github.com/influxdata/influxdb/v2/kit/feature"

	"github.com/influxdata/influxdb/v2/kit/prom"
	"github.com/influxdata/influxdb/v2/kit/tracing"

	"github.com/uvite/gvmdesk/pkg/kv"
	tenant "github.com/uvite/gvmdesk/pkg/tenant"

	"github.com/influxdata/influxdb/v2/query/fluxlang"

	//"github.com/influxdata/influxdb/v2/query/fluxlang"
	"github.com/influxdata/influxdb/v2/snowflake"
	"github.com/influxdata/influxdb/v2/sqlite"
	"github.com/influxdata/influxdb/v2/task/backend/scheduler"
	"github.com/uvite/gvmdesk/pkg/executor"
	"path/filepath"
	"strings"
	"sync"
	//// needed for tsm1
	//_ "github.com/influxdata/influxdb/v2/tsdb/engine/tsm1"
	//
	//// needed for tsi1
	//_ "github.com/influxdata/influxdb/v2/tsdb/index/tsi1"
	pzap "github.com/influxdata/influxdb/v2/zap"
	"github.com/opentracing/opentracing-go"
	jaegerconfig "github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
)

const (
	// DiskStore stores all REST resources to disk in boltdb and sqlite.
	DiskStore = "disk"
	// BoltStore also stores all REST resources to disk in boltdb and sqlite. Kept for backwards-compatibility.
	BoltStore = "bolt"
	// MemoryStore stores all REST resources in memory (useful for testing).
	MemoryStore = "memory"

	// LogTracing enables tracing via zap logs
	LogTracing = "log"
	// JaegerTracing enables tracing via the Jaeger client library
	JaegerTracing = "jaeger"
)

type labeledCloser struct {
	label  string
	closer func(context.Context) error
}

// Launcher represents the main program execution.
type Launcher struct {
	wg       sync.WaitGroup
	cancel   func()
	doneChan <-chan struct{}
	closers  []labeledCloser

	flagger feature.Flagger

	kvStore   kv.Store
	KvService *kv.Service
	sqlStore  *sqlite.SqlStore

	Executor *executor.Executor
	reg      *prom.Registry
	log      *zap.Logger

	httpPort int

	TSC   taskmodel.TaskService
	Exbot *bot.Exbot
}

type stoppingScheduler interface {
	scheduler.Scheduler
	Stop()
}

// NewLauncher returns a new instance of Launcher with a no-op logger.
func NewLauncher() *Launcher {
	return &Launcher{
		log: zap.NewNop(),
	}
}

// Shutdown shuts down the HTTP server and waits for all services to clean up.
func (m *Launcher) Shutdown(ctx context.Context) error {
	var errs []string

	// Shut down subsystems in the reverse order of their registration.
	for i := len(m.closers); i > 0; i-- {
		lc := m.closers[i-1]
		m.log.Info("Stopping subsystem", zap.String("subsystem", lc.label))
		if err := lc.closer(ctx); err != nil {
			m.log.Error("Failed to stop subsystem", zap.String("subsystem", lc.label), zap.Error(err))
			errs = append(errs, err.Error())
		}
	}

	m.wg.Wait()

	// N.B. We ignore any errors here because Sync is known to fail with EINVAL
	// when logging to Stdout on certain OS's.
	//
	// Uber made the same change within the core of the logger implementation.
	// See: https://github.com/uber-go/zap/issues/328
	_ = m.log.Sync()

	if len(errs) > 0 {
		return fmt.Errorf("failed to shut down server: [%s]", strings.Join(errs, ","))
	}
	return nil
}

func (m *Launcher) Done() <-chan struct{} {
	return m.doneChan
}

func (m *Launcher) Run(ctx context.Context, opts *InfluxdOpts) (err error) {
	span, ctx := tracing.StartSpanFromContext(ctx)
	defer span.Finish()

	ctx, m.cancel = context.WithCancel(ctx)
	m.doneChan = ctx.Done()

	m.initTracing(opts)
	//fmt.Println(opts)
	// Open KV and SQL stores.
	procID, err := m.openMetaStores(ctx, opts)
	if err != nil {
		return err
	}
	fmt.Println(procID)

	tenantStore := tenant.NewStore(m.kvStore)
	ts := tenant.NewSystem(tenantStore, m.log.With(zap.String("store", "new")))

	serviceConfig := kv.ServiceConfig{
		FluxLanguageService: fluxlang.DefaultService,
	}
	//bucketHTTPServer := ts.NewBucketHTTPHandler(m.log, labelSvc)

	m.KvService = kv.NewService(m.log.With(zap.String("store", "kv")), m.kvStore, ts, serviceConfig)

	var taskSvc taskmodel.TaskService
	{
		// create the task stack

		combinedTaskService := NewAnalyticalStorage(
			m.log.With(zap.String("service", "task-analytical-store")),
			m.KvService,
		)

		executor := executor.NewExecutor(
			m.log.With(zap.String("service", "task-executor")),

			combinedTaskService,
			m.Exbot,
		)
		err = executor.LoadExistingScheduleRuns(ctx)
		fmt.Println(444, err)
		if err != nil {
			m.log.Fatal("could not load existing scheduled runs", zap.Error(err))
		}
		m.Executor = executor
	}
	fmt.Println(333)
	m.TSC = taskSvc

	//errorHandler := kithttp.NewErrorHandler(m.log.With(zap.String("handler", "error_logger")))
	Router := gin.Default()
	Router.GET("/", func(c *gin.Context) {
		time.Sleep(5 * time.Second)
		c.String(nethttp.StatusOK, "Welcome Gin Server")
	})

	if err := m.runHTTP(opts, Router); err != nil {
		return err
	}

	return nil
}

// initTracing sets up the global tracer for the influxd process.
// Any errors encountered during setup are logged, but don't crash the process.
func (m *Launcher) initTracing(opts *InfluxdOpts) {
	switch opts.TracingType {
	case LogTracing:
		m.log.Info("Tracing via zap logging")
		opentracing.SetGlobalTracer(pzap.NewTracer(m.log, snowflake.NewIDGenerator()))

	case JaegerTracing:
		m.log.Info("Tracing via Jaeger")
		cfg, err := jaegerconfig.FromEnv()
		if err != nil {
			m.log.Error("Failed to get Jaeger client config from environment variables", zap.Error(err))
			return
		}
		tracer, closer, err := cfg.NewTracer()
		if err != nil {
			m.log.Error("Failed to instantiate Jaeger tracer", zap.Error(err))
			return
		}
		m.closers = append(m.closers, labeledCloser{
			label: "Jaeger tracer",
			closer: func(context.Context) error {
				return closer.Close()
			},
		})
		opentracing.SetGlobalTracer(tracer)
	}
}
func (m *Launcher) openMetaStores(ctx context.Context, opts *InfluxdOpts) (string, error) {
	type flushableKVStore interface {
		kv.SchemaStore
		http.Flusher
	}
	var kvStore flushableKVStore
	var sqlStore *sqlite.SqlStore

	var procID string
	var err error
	switch opts.StoreType {
	case BoltStore:
		m.log.Warn("Using --store=bolt is deprecated. Use --store=disk instead.")
		fallthrough
	case DiskStore:
		boltClient := bolt.NewClient(m.log.With(zap.String("service", "bolt")))
		boltClient.Path = opts.BoltPath

		if err := boltClient.Open(ctx); err != nil {
			m.log.Error("Failed opening bolt", zap.Error(err))
			return "", err
		}
		m.closers = append(m.closers, labeledCloser{
			label: "bolt",
			closer: func(context.Context) error {
				return boltClient.Close()
			},
		})

		procID = boltClient.ID().String()

		boltKV := bolt.NewKVStore(m.log.With(zap.String("service", "kvstore-bolt")), opts.BoltPath)
		boltKV.WithDB(boltClient.DB())
		kvStore = boltKV

		// If a sqlite-path is not specified, store sqlite db in the same directory as bolt with the default filename.
		if opts.SqLitePath == "" {
			opts.SqLitePath = filepath.Join(filepath.Dir(opts.BoltPath), sqlite.DefaultFilename)
		}
		sqlStore, err = sqlite.NewSqlStore(opts.SqLitePath, m.log.With(zap.String("service", "sqlite")))
		if err != nil {
			m.log.Error("Failed opening sqlite store", zap.Error(err))
			return "", err
		}

	case MemoryStore:
		//kvStore = inmem.NewKVStore()
		//sqlStore, err = sqlite.NewSqlStore(sqlite.InmemPath, m.log.With(zap.String("service", "sqlite")))
		//if err != nil {
		//	m.log.Error("Failed opening sqlite store", zap.Error(err))
		//	return "", err
		//}

	default:
		err := fmt.Errorf("unknown store type %s; expected disk or memory", opts.StoreType)
		m.log.Error("Failed opening metadata store", zap.Error(err))
		return "", err
	}

	m.closers = append(m.closers, labeledCloser{
		label: "sqlite",
		closer: func(context.Context) error {
			return sqlStore.Close()
		},
	})

	if err != nil {
		m.log.Error("Failed to initialize kv migrator", zap.Error(err))
		return "", err
	}

	// Apply migrations to the KV and SQL metadata stores.
	kvMigrator, err := migration.NewMigrator(
		m.log.With(zap.String("service", "KV migrations")),
		kvStore,
		all.Migrations[:]...,
	)
	if err != nil {
		m.log.Error("Failed to initialize kv migrator", zap.Error(err))
		return "", err
	}
	sqlMigrator := sqlite.NewMigrator(sqlStore, m.log.With(zap.String("service", "SQL migrations")))

	// If we're migrating a persistent data store, take a backup of the pre-migration state for rollback.
	if opts.StoreType == DiskStore || opts.StoreType == BoltStore {
		backupPattern := "%s.pre-%s-upgrade.backup"
		info := platform.GetBuildInfo()
		kvMigrator.SetBackupPath(fmt.Sprintf(backupPattern, opts.BoltPath, info.Version))
		sqlMigrator.SetBackupPath(fmt.Sprintf(backupPattern, opts.SqLitePath, info.Version))
	}
	if err := kvMigrator.Up(ctx); err != nil {
		m.log.Error("Failed to apply KV migrations", zap.Error(err))
		return "", err
	}
	if err := sqlMigrator.Up(ctx, sqliteMigrations.AllUp); err != nil {
		m.log.Error("Failed to apply SQL migrations", zap.Error(err))
		return "", err
	}

	m.kvStore = kvStore
	m.sqlStore = sqlStore
	return procID, nil
}
