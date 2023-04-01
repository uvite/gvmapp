package services

import (
	"context"
	"github.com/influxdata/influxdb/v2/kit/signals"
	"github.com/uvite/gvmapp/backend/pkg/launcher"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"time"
)

type LauncherService struct {
	Ctx      context.Context
	Launcher *launcher.Launcher
}

func NewLauncherService( ) *LauncherService {
	launcherService:=&LauncherService{
		Launcher: launcher.NewLauncher(),

	}

	return launcherService
}
func (l *LauncherService) RunLauncher() {
	o := launcher.NewOpts()
	if err := l.Launcher.Run(signals.WithStandardSignals(l.Ctx), o); err != nil {

	}
	<-l.Launcher.Done()

	shutdownCtx, cancel := context.WithTimeout(l.Ctx, 2*time.Second)
	defer cancel()
	l.Launcher.Shutdown(shutdownCtx)
}

func (l *LauncherService) Event() {
	runtime.EventsOn(l.Ctx,"service.alert.create",func(option ...interface{}) {


	})
}

