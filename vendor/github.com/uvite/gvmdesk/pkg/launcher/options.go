package launcher

import (
	"fmt"
	"github.com/influxdata/influxdb/v2/bolt"
	"github.com/influxdata/influxdb/v2/sqlite"
	"github.com/influxdata/influxdb/v2/storage"
	"github.com/influxdata/influxdb/v2/vault"
	"github.com/spf13/viper"
	"path/filepath"
	"time"
)

// InfluxdOpts captures all arguments for running the InfluxDB server.
type InfluxdOpts struct {
	AssetsPath  string
	BoltPath    string
	SqLitePath  string
	EnginePath  string
	TracingType string
	StoreType   string
	SecretStore string
	VaultConfig vault.Config

	// Storage options.
	StorageConfig storage.Config

	Viper *viper.Viper

	HardeningEnabled bool

	HttpBindAddress       string
	HttpReadHeaderTimeout time.Duration
	HttpReadTimeout       time.Duration
	HttpWriteTimeout      time.Duration
	HttpIdleTimeout       time.Duration
}

// NewOpts constructs options with default values.
func NewOpts() *InfluxdOpts {
	dir, err := InfluxDir()
	if err != nil {
		panic(fmt.Errorf("failed to determine influx directory: %v", err))
	}

	return &InfluxdOpts{

		StorageConfig: storage.NewConfig(),

		BoltPath:   filepath.Join(dir, bolt.DefaultFilename),
		SqLitePath: filepath.Join(dir, sqlite.DefaultFilename),
		EnginePath: filepath.Join(dir, "engine"),

		StoreType:        DiskStore,
		SecretStore:      BoltStore,
		HardeningEnabled: false,
		HttpBindAddress:  ":8888",
	}
}
