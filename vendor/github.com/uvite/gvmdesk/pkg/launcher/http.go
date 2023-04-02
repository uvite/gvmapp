package launcher

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/influxdata/influxdb/v2"
	log "github.com/sirupsen/logrus"
	taskmodel "github.com/uvite/gvmdesk/pkg/model"
	"github.com/uvite/gvmdesk/pkg/platform"
	"net/http"
)

// GetById Find by id structure
type GetById struct {
	ID uint64 `json:"id" form:"id"`
}
type ResponseInfo struct {
	List  interface{} `json:"list"`
	Total int         `json:"total" form:"total"` // 页码

}

func (m *Launcher) runHTTP(opts *InfluxdOpts, router *gin.Engine) error {

	router.GET("/api/addtask", func(c *gin.Context) {

		//orgId,_:=platform.IDFromString("7654")
		Org := influxdb.Organization{Name: "genv-org", ID: (7654)}
		data := map[string]interface{}{}
		data["symbol"] = "ETHUSDT"
		data["interval"] = "1m"
		data["path"] = string("/hein/gvmdesk/js/4.js")
		task, err := m.KvService.CreateTask(c, taskmodel.TaskCreate{
			OrganizationID: platform.ID(Org.ID),
			OwnerID:        platform.ID(Org.ID),
			Status:         string(taskmodel.TaskActive),
			Flux:           `234234`,
			Metadata:       data,
		})
		if err != nil {
			fmt.Println(err)
		}

		if err != nil {
			FailWithMessage("获取失败", c)
		}
		OkWithDetailed(task, "获取成功", c)
	})
	router.GET("/api/task", func(c *gin.Context) {

		id := c.Query("id")

		pid, _ := platform.IDFromString(id)
		task, err := m.KvService.FindTaskByID(c, *pid)
		if err != nil {
			fmt.Println(err)
		}

		if err != nil {
			FailWithMessage("获取失败", c)
		}
		OkWithDetailed(task, "获取成功", c)
	})

	router.GET("/api/close", func(c *gin.Context) {

		id := c.Query("id")

		pid, _ := platform.IDFromString(id)
		err := m.Executor.Close(c, *pid)
		fmt.Println(err)
		if err != nil {
			FailWithMessage("获取失败", c)
		}
		OkWithDetailed("task", "获取成功", c)
	})

	router.GET("/api/cancel", func(c *gin.Context) {

		id := c.Query("id")

		pid, _ := platform.IDFromString(id)
		err := m.Executor.Cancel(c, *pid)
		fmt.Println(err)
		if err != nil {
			FailWithMessage("获取失败", c)
		}
		OkWithDetailed("task", "获取成功", c)
	})

	router.GET("/api/start", func(c *gin.Context) {
		symbol := c.Query("symbol")
		Org := influxdb.Organization{Name: "genv-org", ID: (7654)}
		data := map[string]interface{}{}
		data["symbol"] = symbol
		data["interval"] = "1m"
		data["path"] = string("/hein/gvmdesk/js/4.js")
		task, err := m.KvService.CreateTask(c, taskmodel.TaskCreate{
			OrganizationID: platform.ID(Org.ID),
			OwnerID:        platform.ID(Org.ID),
			Status:         string(taskmodel.TaskActive),
			Flux:           `234234`,
			Metadata:       data,
			Symbol:         symbol,
		})
		if err != nil {
			fmt.Println(err)
		}

		promise, err := m.Executor.PromisedExecute(c, task.ID)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(promise)
		if err != nil {
			FailWithMessage("获取失败", c)
		}
		OkWithDetailed(task, "获取成功", c)
	})

	router.GET("/api/all", func(c *gin.Context) {
		filter := taskmodel.TaskFilter{}
		task, total, err := m.KvService.FindTasks(c, filter)
		if err != nil {
			fmt.Println(err)
		}

		if err != nil {
			FailWithMessage("获取失败", c)
		}
		OkWithDetailed(ResponseInfo{List: task, Total: total}, "获取成功", c)
	})

	srv := &http.Server{
		Addr:    ":8888",
		Handler: router,
	}

	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	return nil
}

//
//// runHTTP configures and launches a listener for incoming HTTP(S) requests.
//// The listener is run in a separate goroutine. If it fails to start up, it
//// will cancel the launcher.
//func (m *Launcher) runHTTP(opts *InfluxdOpts, handler nethttp.Handler, httpLogger *zap.Logger) error {
//	log := m.log.With(zap.String("service", "tcp-listener"))
//	fmt.Println("[33]", opts)
//	httpServer := &nethttp.Server{
//		Addr:              opts.HttpBindAddress,
//		Handler:           handler,
//		ReadHeaderTimeout: opts.HttpReadHeaderTimeout,
//		ReadTimeout:       opts.HttpReadTimeout,
//		WriteTimeout:      opts.HttpWriteTimeout,
//		IdleTimeout:       opts.HttpIdleTimeout,
//		ErrorLog:          zap.NewStdLog(httpLogger),
//	}
//	m.closers = append(m.closers, labeledCloser{
//		label:  "HTTP server",
//		closer: httpServer.Shutdown,
//	})
//
//	ln, err := net.Listen("tcp", opts.HttpBindAddress)
//	if err != nil {
//		log.Error("Failed to set up TCP listener", zap.String("addr", opts.HttpBindAddress), zap.Error(err))
//		return err
//	}
//	if addr, ok := ln.Addr().(*net.TCPAddr); ok {
//		m.httpPort = addr.Port
//	}
//	fmt.Println("[33]", m.httpPort)
//
//	m.wg.Add(1)
//
//	return nil
//}
