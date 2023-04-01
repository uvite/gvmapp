package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/influxdata/influxdb/v2"
	"github.com/uvite/gvmapp/backend/pkg/launcher"
	taskmodel "github.com/uvite/gvmapp/backend/pkg/model"
	"github.com/uvite/gvmapp/backend/pkg/platform"
	"github.com/uvite/gvmapp/backend/util"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"io/ioutil"
	"os"
	"path/filepath"
)

type AlertService struct {
	Ctx      context.Context
	Launcher *launcher.Launcher
}

func NewAlertService() *AlertService {
	return &AlertService{}
}
func (l *AlertService) SetLauncher(service *LauncherService) {
	l.Launcher = service.Launcher
}
func (a *AlertService) GetAlertList() *util.Resp {

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

// 创建警报
func (a *AlertService) CreateAlert(item taskmodel.Task) *util.Resp {

	Org := influxdb.Organization{Name: "gvm", ID: (1)}

	task, err := a.Launcher.KvService.CreateTask(a.Ctx, taskmodel.TaskCreate{
		OrganizationID: platform.ID(Org.ID),
		OwnerID:        platform.ID(Org.ID),
		Status:         string(taskmodel.TaskActive),
		Flux:           `1`,
		Symbol:         item.Symbol,
		Interval:       item.Interval,
		Path:           item.Path,
		Content:        item.Content,
		Metadata:       item.Metadata,
	})
	err = createJs(*task)
	fmt.Println(task, err)
	runtime.EventsEmit(a.Ctx, "service.alert.create", task)
	return util.Success(task)

}
func (a *AlertService) DelAlertItem(id string) *util.Resp {

	pid, _ := platform.IDFromString(id)

	err := a.Launcher.KvService.DeleteTask(a.Ctx, *pid)

	if err != nil {
		return util.Error(err.Error())
	}

	return util.Success("success")
}

func createJs(task taskmodel.Task) error {
	configData := util.GetConfigDir()

	destFilePath := filepath.Join(configData, "js", (task.ID.String() + ".js"))
	setupData, err := json.Marshal(task.Metadata)
	if err != nil {
		return err
	}

	jsonStr := string(setupData)
	content := fmt.Sprintf(`exports.setup = function() {
        return %s;
		}
		
		%s
		`, jsonStr, task.Content)

	err = ioutil.WriteFile(destFilePath, []byte(content), os.ModePerm)
	os.Chmod(destFilePath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

//
//// 获取全部

//

//
//func (a *App) GetAlertItem(id string) *util.Resp {
//	pid, _ := platform.IDFromString(id)
//	task, err := a.Launcher.KvService.FindTaskByID(a.Ctx, *pid)
//
//	if err != nil {
//		return util.Error(err.Error())
//	}
//	return util.Success(task)
//
//}
//
//func (a *App) UpdateAlertItem(item internal.AlertItem) *util.Resp {
//	_, err := a.Db.ID(item.Id).Update(&item)
//	if err != nil {
//		return util.Error(err.Error())
//	}
//	return a.GetAlertList()
//}
//
//func (a *App) DelAlertItemById(item internal.AlertItem) *util.Resp {
//	_, err := a.Db.ID(item.Id).Delete(&item)
//	if err != nil {
//		return util.Error(err.Error())
//	}
//	return a.GetAlertList()
//}
//
//// 启动一个任务
//func (a *App) RunAlert(id string) *util.Resp {
//	pid, _ := platform.IDFromString(id)
//	promise, err := a.Launcher.Executor.PromisedExecute(a.Ctx, *pid)
//
//	if err != nil {
//		a.Log.Error(fmt.Sprintf(configs.DelAlertItemErr, promise.ID(), err.Error()))
//		return util.Error(err.Error())
//	}
//	v := taskmodel.TaskStatusActive
//	updatedTask, err := a.Launcher.KvService.UpdateTask(a.Ctx, *pid, taskmodel.TaskUpdate{Status: &v})
//	if err != nil {
//		a.Log.Error(fmt.Sprintf(configs.DelAlertItemErr, promise.ID(), err.Error()))
//		return util.Error(err.Error())
//	}
//	return util.Success(updatedTask)
//}
//
//func (a *App) CloseAlert(id string) *util.Resp {
//
//	pid, _ := platform.IDFromString(id)
//	err := a.Launcher.Executor.Close(a.Ctx, *pid)
//
//	if err != nil {
//		a.Log.Error(fmt.Sprintf(configs.DelAlertItemErr, pid, err.Error()))
//		return util.Error(err.Error())
//	}
//	v := taskmodel.TaskStatusInactive
//	updatedTask, err := a.Launcher.KvService.UpdateTask(a.Ctx, *pid, taskmodel.TaskUpdate{Status: &v})
//	if err != nil {
//		return util.Error(err.Error())
//	}
//	return util.Success(updatedTask)
//}
//func (a *App) SetAlertStatus(id int, status bool) *util.Resp {
//	item := internal.AlertItem{}
//	has, err := a.Db.ID(id).Get(&item)
//	if err != nil {
//		a.Log.Error(fmt.Sprintf(configs.DelAlertItemErr, item.Title, err.Error()))
//		return util.Error(err.Error())
//	}
//	if has {
//		a.AddSymbolInterval(item.Symbol, item.Interval)
//		go a.RunTestFile(item)
//		return util.Success(item)
//	}
//	return nil
//}
