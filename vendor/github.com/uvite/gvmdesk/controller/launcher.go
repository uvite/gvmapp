package gtools

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb/v2/kit/signals"
	"github.com/uvite/gvmdesk/pkg/bot"
	"github.com/uvite/gvmdesk/pkg/launcher"
	"time"
)

func (a *App) InitLauncher() {
	l := launcher.NewLauncher()

	ctx := a.Ctx
	o := launcher.NewOpts()

	fmt.Println(o)
	eb := bot.NewExbot(ctx)

	eb.InitExchange()
	eb.Subscript()
	l.Exbot = eb
	a.Launcher = l

	// Start the launcher and wait for it to exit on SIGINT or SIGTERM.
	if err := l.Run(signals.WithStandardSignals(ctx), o); err != nil {

	}
	<-l.Done()

	// Tear down the launcher, allowing it a few seconds to finish any
	// in-progress requests.
	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	l.Shutdown(shutdownCtx)
}
