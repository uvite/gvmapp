package gtools

import (
	"fmt"
	"github.com/influxdata/influxdb/v2"
	"github.com/uvite/gvmdesk/configs"
	"github.com/uvite/gvmdesk/internal"
	taskmodel "github.com/uvite/gvmdesk/pkg/model"
	"github.com/uvite/gvmdesk/pkg/platform"
	"github.com/uvite/gvmdesk/util"
)

// 创建警报
func (a *App) AddAlertItem(item taskmodel.Task) *util.Resp {

	Org := influxdb.Organization{Name: "genv-org", ID: (7654)}
	data := map[string]interface{}{}
	data["symbol"] = item.Symbol
	data["interval"] = item.Interval
	data["path"] = string("/hein/gvmdesk/js/4.js")
	task, err := a.Launcher.KvService.CreateTask(a.Ctx, taskmodel.TaskCreate{
		OrganizationID: platform.ID(Org.ID),
		OwnerID:        platform.ID(Org.ID),
		Status:         string(taskmodel.TaskActive),
		Flux:           `234234`,
		Metadata:       data,
	})
	fmt.Println(task, err)
	//_, err := a.Db.InsertOne(&item)
	//if err != nil {
	//	a.Log.Error(fmt.Sprintf(configs.AddAlertItemErr, item.Title, err.Error()))
	//	return util.Error(err.Error())
	//}
	return a.GetAlertList()
}

// 获取全部
func (a *App) GetAlertList() *util.Resp {

	filter := taskmodel.TaskFilter{}
	task, total, err := a.Launcher.KvService.FindTasks(a.Ctx, filter)
	if err != nil {
		return util.Error(err.Error())
	}
	fmt.Println(total)

	resultMap := make(map[string]interface{}, 0)
	resultMap["list"] = task

	return util.Success(resultMap)
}

func (a *App) DelAlertItem(id string) *util.Resp {

	pid, _ := platform.IDFromString(id)

	err := a.Launcher.KvService.DeleteTask(a.Ctx, *pid)

	if err != nil {
		return util.Error(err.Error())
	}

	return a.GetAlertList()
}

func (a *App) GetAlertItem(id string) *util.Resp {
	pid, _ := platform.IDFromString(id)
	task, err := a.Launcher.KvService.FindTaskByID(a.Ctx, *pid)

	if err != nil {
		return util.Error(err.Error())
	}
	return util.Success(task)

}

func (a *App) UpdateAlertItem(item internal.AlertItem) *util.Resp {
	_, err := a.Db.ID(item.Id).Update(&item)
	if err != nil {
		return util.Error(err.Error())
	}
	return a.GetAlertList()
}

func (a *App) DelAlertItemById(item internal.AlertItem) *util.Resp {
	_, err := a.Db.ID(item.Id).Delete(&item)
	if err != nil {
		return util.Error(err.Error())
	}
	return a.GetAlertList()
}

// 启动一个任务
func (a *App) RunAlert(id string) *util.Resp {
	pid, _ := platform.IDFromString(id)
	promise, err := a.Launcher.Executor.PromisedExecute(a.Ctx, *pid)
	if err != nil {
		fmt.Println(err)
	}
	if err != nil {
		a.Log.Error(fmt.Sprintf(configs.DelAlertItemErr, promise.ID(), err.Error()))
		return util.Error(err.Error())
	}
	return util.Success("运行成功")
}

func (a *App) CloseAlert(id string) *util.Resp {

	pid, _ := platform.IDFromString(id)
	err := a.Launcher.Executor.Close(a.Ctx, *pid)
	if err != nil {
		fmt.Println(err)
	}
	if err != nil {
		a.Log.Error(fmt.Sprintf(configs.DelAlertItemErr, pid, err.Error()))
		return util.Error(err.Error())
	}
	return util.Success("关闭成功")
}
func (a *App) SetAlertStatus(id int, status bool) *util.Resp {
	item := internal.AlertItem{}
	has, err := a.Db.ID(id).Get(&item)
	if err != nil {
		a.Log.Error(fmt.Sprintf(configs.DelAlertItemErr, item.Title, err.Error()))
		return util.Error(err.Error())
	}
	if has {
		a.AddSymbolInterval(item.Symbol, item.Interval)
		go a.RunTestFile(item)
		return util.Success(item)
	}
	return nil
}
