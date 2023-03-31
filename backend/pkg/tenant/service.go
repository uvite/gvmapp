package tenant

import (
	"context"

	"github.com/influxdata/influxdb/v2"
	"go.uber.org/zap"
)

type contextKey string

const (
	ctxInternal contextKey = "influx/tenant/internal"
)

func internalCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxInternal, true)
}

func isInternal(ctx context.Context) bool {
	_, ok := ctx.Value(ctxInternal).(bool)
	return ok
}

type Service struct {
	store *Store
	influxdb.UserService
	influxdb.PasswordsService
	influxdb.UserResourceMappingService
	influxdb.OrganizationService
	influxdb.BucketService
}

func (s *Service) RLock() {
	s.store.RLock()
}

func (s *Service) RUnlock() {
	s.store.RUnlock()
}

// NewService creates a new base tenant service.
func NewService(st *Store) *Service {
	svc := &Service{store: st}

	return svc
}

// creates a new Service with logging and metrics middleware wrappers.
func NewSystem(store *Store, log *zap.Logger) *Service {
	ts := NewService(store)
	//ts.UserService = NewUserLogger(log, NewUserMetrics(reg, ts.UserService, metricOpts...))
	//ts.PasswordsService = NewPasswordLogger(log, NewPasswordMetrics(reg, ts.PasswordsService, metricOpts...))
	//ts.UserResourceMappingService = NewURMLogger(log, NewUrmMetrics(reg, ts.UserResourceMappingService, metricOpts...))
	//ts.OrganizationService = NewOrgLogger(log, NewOrgMetrics(reg, ts.OrganizationService, metricOpts...))
	//ts.BucketService = NewBucketLogger(log, NewBucketMetrics(reg, ts.BucketService, metricOpts...))

	return ts
}
