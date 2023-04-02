package launcher

import (
	"github.com/uvite/gvmdesk/pkg/kv"
	taskmodel "github.com/uvite/gvmdesk/pkg/model"
	"go.uber.org/zap"
)

// NewAnalyticalStorage creates a new analytical store with access to the necessary systems for storing data and to act as a middleware (deprecated)
func NewAnalyticalStorage(log *zap.Logger, ts taskmodel.TaskService) *AnalyticalStorage {
	return &AnalyticalStorage{
		log:                log,
		TaskService:        ts,

	}
}
type AnalyticalStorage struct {
	taskmodel.TaskService
	Kvstore *kv.Service
	log     *zap.Logger
}
