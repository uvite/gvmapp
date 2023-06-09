package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/uvite/gvmapp/backend/pkg/kv"
	"os"
	"time"

	"github.com/influxdata/influxdb/v2/kit/migration"
	"github.com/influxdata/influxdb/v2/kit/platform"
	"go.uber.org/zap"
)

var migrationBucket = []byte("migrationsv1")

type Store = kv.SchemaStore

// MigrationState is a type for describing the state of a migration.
type MigrationState uint

const (
	// DownMigrationState is for a migration not yet applied.
	DownMigrationState MigrationState = iota
	// UpMigration State is for a migration which has been applied.
	UpMigrationState
)

// String returns a string representation for a migration state.
func (s MigrationState) String() string {
	switch s {
	case DownMigrationState:
		return "down"
	case UpMigrationState:
		return "up"
	default:
		return "unknown"
	}
}

// Migration is a record of a particular migration.
type Migration struct {
	ID         platform.ID    `json:"id"`
	Name       string         `json:"name"`
	State      MigrationState `json:"-"`
	StartedAt  *time.Time     `json:"started_at"`
	FinishedAt *time.Time     `json:"finished_at,omitempty"`
}

// Spec is a specification for a particular migration.
// It describes the name of the migration and up and down operations
// needed to fulfill the migration.
type Spec interface {
	MigrationName() string
	Up(ctx context.Context, store kv.SchemaStore) error
	Down(ctx context.Context, store kv.SchemaStore) error
}

// Migrator is a type which manages migrations.
// It takes a list of migration specifications and undo (down) all or apply (up) outstanding migrations.
// It records the state of the world in store under the migrations bucket.
type Migrator struct {
	logger *zap.Logger
	store  Store

	Specs []Spec

	now        func() time.Time
	backupPath string
}

// NewMigrator constructs and configures a new Migrator.
func NewMigrator(logger *zap.Logger, store Store, ms ...Spec) (*Migrator, error) {
	m := &Migrator{
		logger: logger,
		store:  store,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}

	// create migration bucket if it does not exist
	if err := store.CreateBucket(context.Background(), migrationBucket); err != nil {
		return nil, err
	}

	m.AddMigrations(ms...)

	return m, nil
}

// AddMigrations appends the provided migration specs onto the Migrator.
func (m *Migrator) AddMigrations(ms ...Spec) {
	m.Specs = append(m.Specs, ms...)
}

// SetBackupPath records the filepath where pre-migration state should be written prior to running migrations.
func (m *Migrator) SetBackupPath(path string) {
	m.backupPath = path
}

// List returns a list of migrations and their states within the provided store.
func (m *Migrator) List(ctx context.Context) (migrations []Migration, _ error) {
	if err := m.walk(ctx, m.store, func(id platform.ID, m Migration) {
		migrations = append(migrations, m)
	}); err != nil {
		return nil, err
	}

	migrationsLen := len(migrations)
	for idx, spec := range m.Specs[migrationsLen:] {
		migration := Migration{
			ID:   platform.ID(migrationsLen + idx + 1),
			Name: spec.MigrationName(),
		}

		migrations = append(migrations, migration)
	}

	return
}

// Up applies each outstanding migration in order.
// Migrations are applied in order from the lowest indexed migration in a down state.
//
// For example, given:
// 0001 add bucket foo         | (up)
// 0002 add bucket bar         | (down)
// 0003 add index "foo on baz" | (down)
//
// Up would apply migration 0002 and then 0003.
func (m *Migrator) Up(ctx context.Context) error {
	wrapErr := func(err error) error {
		if err == nil {
			return nil
		}

		return fmt.Errorf("up: %w", err)
	}

	var lastMigration int
	if err := m.walk(ctx, m.store, func(id platform.ID, mig Migration) {
		// we're interested in the last up migration
		if mig.State == UpMigrationState {
			lastMigration = int(id)
		}
	}); err != nil {
		return wrapErr(err)
	}

	migrationsToDo := len(m.Specs[lastMigration:])
	if migrationsToDo == 0 {
		return nil
	}

	if m.backupPath != "" && lastMigration != 0 {
		m.logger.Info("Backing up pre-migration metadata", zap.String("backup_path", m.backupPath))
		if err := func() error {
			out, err := os.Create(m.backupPath)
			if err != nil {
				return err
			}
			defer out.Close()

			if err := m.store.Backup(ctx, out); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return fmt.Errorf("failed to back up pre-migration metadata: %w", err)
		}
	}

	m.logger.Info("Bringing up metadata migrations", zap.Int("migration_count", migrationsToDo))
	for idx, spec := range m.Specs[lastMigration:] {
		startedAt := m.now()
		migration := Migration{
			ID:        platform.ID(lastMigration + idx + 1),
			Name:      spec.MigrationName(),
			StartedAt: &startedAt,
		}

		m.logMigrationEvent(UpMigrationState, migration, "started")

		if err := m.putMigration(ctx, m.store, migration); err != nil {
			return wrapErr(err)
		}

		if err := spec.Up(ctx, m.store); err != nil {
			return wrapErr(err)
		}

		finishedAt := m.now()
		migration.FinishedAt = &finishedAt
		migration.State = UpMigrationState

		if err := m.putMigration(ctx, m.store, migration); err != nil {
			return wrapErr(err)
		}

		m.logMigrationEvent(UpMigrationState, migration, "completed")
	}

	return nil
}

// Down applies the down operation of each currently applied migration.
// Migrations are applied in reverse order from the highest indexed migration in a down state.
//
// For example, given:
// 0001 add bucket foo         | (up)
// 0002 add bucket bar         | (up)
// 0003 add index "foo on baz" | (down)
//
// Down would call down() on 0002 and then on 0001.
func (m *Migrator) Down(ctx context.Context, untilMigration int) (err error) {
	wrapErr := func(err error) error {
		if err == nil {
			return nil
		}

		return fmt.Errorf("down: %w", err)
	}

	var migrations []struct {
		Spec
		Migration
	}

	if err := m.walk(ctx, m.store, func(id platform.ID, mig Migration) {
		migrations = append(
			migrations,
			struct {
				Spec
				Migration
			}{
				m.Specs[int(id)-1],
				mig,
			},
		)
	}); err != nil {
		return wrapErr(err)
	}

	migrationsToDo := len(migrations) - untilMigration
	if migrationsToDo == 0 {
		return nil
	}
	if migrationsToDo < 0 {
		m.logger.Warn("KV metadata is already on a schema older than target, nothing to do")
		return nil
	}

	if m.backupPath != "" {
		m.logger.Info("Backing up pre-migration metadata", zap.String("backup_path", m.backupPath))
		if err := func() error {
			out, err := os.Create(m.backupPath)
			if err != nil {
				return err
			}
			defer out.Close()

			if err := m.store.Backup(ctx, out); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return fmt.Errorf("failed to back up pre-migration metadata: %w", err)
		}
	}

	m.logger.Info("Tearing down metadata migrations", zap.Int("migration_count", migrationsToDo))
	for i := len(migrations) - 1; i >= untilMigration; i-- {
		migration := migrations[i]

		m.logMigrationEvent(DownMigrationState, migration.Migration, "started")

		if err := migration.Spec.Down(ctx, m.store); err != nil {
			return wrapErr(err)
		}

		if err := m.deleteMigration(ctx, m.store, migration.Migration); err != nil {
			return wrapErr(err)
		}

		m.logMigrationEvent(DownMigrationState, migration.Migration, "completed")
	}

	return nil
}

func (m *Migrator) logMigrationEvent(state MigrationState, mig Migration, event string) {
	m.logger.Debug(
		"Executing metadata migration",
		zap.String("migration_name", mig.Name),
		zap.String("target_state", state.String()),
		zap.String("migration_event", event),
	)
}

func (m *Migrator) walk(ctx context.Context, store kv.Store, fn func(id platform.ID, m Migration)) error {
	if err := store.View(ctx, func(tx kv.Tx) error {
		bkt, err := tx.Bucket(migrationBucket)
		if err != nil {
			return err
		}

		cursor, err := bkt.ForwardCursor(nil)
		if err != nil {
			return err
		}

		return kv.WalkCursor(ctx, cursor, func(k, v []byte) (bool, error) {
			var id platform.ID
			if err := id.Decode(k); err != nil {
				return false, fmt.Errorf("decoding migration id: %w", err)
			}

			var mig Migration
			if err := json.Unmarshal(v, &mig); err != nil {
				return false, err
			}

			idx := int(id) - 1
			if idx >= len(m.Specs) {
				return false, migration.ErrInvalidMigration(mig.Name)
			}

			if spec := m.Specs[idx]; spec.MigrationName() != mig.Name {
				return false, fmt.Errorf("expected migration %q, found %q", spec.MigrationName(), mig.Name)
			}

			if mig.FinishedAt != nil {
				mig.State = UpMigrationState
			}

			fn(id, mig)

			return true, nil
		})
	}); err != nil {
		return fmt.Errorf("reading migrations: %w", err)
	}

	return nil
}

func (m *Migrator) putMigration(ctx context.Context, store kv.Store, migration Migration) error {
	return store.Update(ctx, func(tx kv.Tx) error {
		bkt, err := tx.Bucket(migrationBucket)
		if err != nil {
			return err
		}

		data, err := json.Marshal(migration)
		if err != nil {
			return err
		}

		id, _ := migration.ID.Encode()
		return bkt.Put(id, data)
	})
}

func (m *Migrator) deleteMigration(ctx context.Context, store kv.Store, migration Migration) error {
	return store.Update(ctx, func(tx kv.Tx) error {
		bkt, err := tx.Bucket(migrationBucket)
		if err != nil {
			return err
		}

		id, _ := migration.ID.Encode()
		return bkt.Delete(id)
	})
}
