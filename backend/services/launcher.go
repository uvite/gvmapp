package services

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	taskmodel "github.com/uvite/gvmapp/backend/pkg/model"
	"github.com/uvite/gvmapp/backend/util"
	"time"

	"github.com/influxdata/influxdb/v2/kit/signals"
	"github.com/uvite/gvmapp/backend/pkg/launcher"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type LauncherService struct {
	Ctx      context.Context
	Launcher *launcher.Launcher
	//*ExchangeService
}

func NewLauncherService() *LauncherService {
	launcherService := &LauncherService{
		Launcher: launcher.NewLauncher(),
	}

	return launcherService
}
func (l *LauncherService) Init() {
	o := launcher.NewOpts()
	if err := l.Launcher.Run(signals.WithStandardSignals(l.Ctx), o); err != nil {

		log.Error(err)

	}
	log.Info("launcher staring")

}

func (l *LauncherService) Event() {
	runtime.EventsOn(l.Ctx, "service.alert.create", func(data ...interface{}) {
		task := data[0].(*taskmodel.Task)
		l.RunTask(task)
	})
}
func (l *LauncherService) ShutDown() {
	<-l.Launcher.Done()

	shutdownCtx, cancel := context.WithTimeout(l.Ctx, 2*time.Second)
	defer cancel()
	l.Launcher.Shutdown(shutdownCtx)
}

func (l *LauncherService) RunTask(task *taskmodel.Task) *util.Resp {

	promise, err := l.Launcher.Executor.PromisedExecute(l.Ctx, task.ID)
	if err != nil {
		return util.Error(fmt.Sprintf("启动失败 %s", task.ID))

	} else {
		l.ChangeStatus(task, taskmodel.TaskStatusActive)
		return util.Success(promise)
	}

}
func (l *LauncherService) CloseTask(task *taskmodel.Task) *util.Resp {
	err := l.Launcher.Executor.Close(l.Ctx, task.ID)
	if err != nil {
		return util.Error(fmt.Sprintf("启动失败 %s", task.ID))
	} else {
		l.ChangeStatus(task, taskmodel.TaskStatusInactive)
		return util.Success("关闭成功")
	}

}

func (l *LauncherService) ChangeStatus(task *taskmodel.Task, status string) *util.Resp {

	_, err := l.Launcher.KvService.UpdateTask(l.Ctx, task.ID, taskmodel.TaskUpdate{Status: &status})

	if err != nil {
		return util.Error(fmt.Sprintf("启动失败 %s", task.ID))
	} else {
		return util.Success("关闭成功")
	}

}
