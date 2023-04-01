package gvmapp
//
//import (
//	"encoding/json"
//	"fmt"
//	"github.com/influxdata/influxdb/v2"
//	"github.com/uvite/gvmapp/backend/configs"
//	"github.com/uvite/gvmapp/backend/internal"
//	taskmodel "github.com/uvite/gvmapp/backend/pkg/model"
//	"github.com/uvite/gvmapp/backend/pkg/platform"
//	"github.com/uvite/gvmapp/backend/util"
//	"io/ioutil"
//	"os"
//	"path/filepath"
//)
//
//func createJs(task taskmodel.Task) error {
//	configData := util.GetConfigDir()
//
//	destFilePath := filepath.Join(configData, "js", (task.ID.String() + ".js"))
//	setupData, err := json.Marshal(task.Metadata)
//	if err != nil {
//		return err
//	}
//
//	jsonStr := string(setupData)
//	content := fmt.Sprintf(`exports.setup = function() {
//        return %s;
//		}
//
//		%s
//		`, jsonStr, task.Content)
//
//	err = ioutil.WriteFile(destFilePath, []byte(content), os.ModePerm)
//	os.Chmod(destFilePath, os.ModePerm)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//// 创建警报
//func (a *App) AddAlertItem(item taskmodel.Task) *util.Resp {
//
//	Org := influxdb.Organization{Name: "gvm", ID: (1)}
//
//	task, err := a.Launcher.KvService.CreateTask(a.Ctx, taskmodel.TaskCreate{
//		OrganizationID: platform.ID(Org.ID),
//		OwnerID:        platform.ID(Org.ID),
//		Status:         string(taskmodel.TaskActive),
//		Flux:           `1`,
//		Symbol:         item.Symbol,
//		Interval:       item.Interval,
//		Path:           item.Path,
//		Content:        item.Content,
//		Metadata:       item.Metadata,
//	})
//	err = createJs(*task)
//	fmt.Println(task, err)
//	//_, err := a.Db.InsertOne(&item)
//	//if err != nil {
//	//	a.Log.Error(fmt.Sprintf(configs.AddAlertItemErr, item.Title, err.Error()))
//	//	return util.Error(err.Error())
//	//}
//	return a.GetAlertList()
//}
//
//// 获取全部
//func (a *App) GetAlertList() *util.Resp {
//
//	filter := taskmodel.TaskFilter{}
//	fmt.Println(111)
//	task, total, err := a.Launcher.KvService.FindTasks(a.Ctx, filter)
//	fmt.Println(2222, task, total, err)
//
//	if err != nil {
//		fmt.Println(err)
//		return util.Error(err.Error())
//	}
//	fmt.Println(total)
//
//	resultMap := make(map[string]interface{}, 0)
//	resultMap["list"] = task
//
//	return util.Success(resultMap)
//}
//
//func (a *App) DelAlertItem(id string) *util.Resp {
//
//	pid, _ := platform.IDFromString(id)
//
//	err := a.Launcher.KvService.DeleteTask(a.Ctx, *pid)
//
//	if err != nil {
//		return util.Error(err.Error())
//	}
//
//	return a.GetAlertList()
//}
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
