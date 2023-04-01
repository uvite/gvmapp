package gvmapp

import (
	"context"
	"github.com/uvite/gvmapp/backend/services"

	"encoding/base64"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/sirupsen/logrus"
	"github.com/uvite/gvmapp/backend/configs"
	"github.com/uvite/gvmapp/backend/internal"
	"github.com/uvite/gvmapp/backend/pkg/bot"
	"github.com/uvite/gvmapp/backend/pkg/launcher"
	"github.com/uvite/gvmapp/backend/util"

	"github.com/uvite/gvmapp/backend/gvmbot"
	"image"
	"os"
	"path/filepath"
	"xorm.io/xorm"

	r "runtime"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	libdir, _ = os.UserConfigDir()
	basedir   = filepath.Join(libdir, "gvmapp")
	docsdir   = filepath.Join(basedir, "Documents")
)

// App struct
type App struct {
	Ctx             context.Context
	Log             *logrus.Logger
	Db              *xorm.Engine
	LogFile         string
	DBFile          string
	ConfigMap       map[string]map[string]string
	AliOSS          *oss.Client
	Exchange        *gvmbot.Exchange
	Launcher        *launcher.Launcher
	Qvm             *bot.Qvm
	StartService    *services.StartService
	LauncherService *services.LauncherService
	ExchangeService *services.ExchangeService
	AlertService    *services.AlertService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		StartService:    services.NewStartService(),
		LauncherService: services.NewLauncherService(),
		AlertService:    services.NewAlertService(),
		ExchangeService: services.NewExchangeService(),
	}
}

// Startup is called at application Startup
func (a *App) Startup(ctx context.Context) {
	a.Ctx = ctx
	a.StartService.Ctx = ctx
	a.LauncherService.Ctx = ctx
	a.ExchangeService.Ctx = ctx
	a.AlertService.Ctx = ctx

	a.AlertService.SetLauncher(a.LauncherService)

	confDir := util.GetConfigDir()
	util.GetEnvDir()
	//	a.Ctx = ctx
	a.InitLauncher()
	// 获取gvmapp数据文件夹路径

	// 初始化logrus
	a.LogFile = fmt.Sprintf(configs.LogFile, confDir)
	a.Log = internal.NewLogger(a.LogFile)
	//
	//// 初始化xorm
	//a.DBFile = fmt.Sprintf(configs.DBFile, confDir)
	//a.Db = internal.NewXormEngine(a.DBFile)

	//a.InitExchange()
	//app.FileSystemService.Ctx = ctx
	//app.CollectionService.Ctx = ctx

	m := menu.NewMenu()

	if r.GOOS == "darwin" {
		m.Items = append(m.Items, menu.AppMenu())
	}

	m.Items = append(m.Items, menu.SubMenu("File", menu.NewMenuFromItems(
		menu.Text("Refresh", keys.CmdOrCtrl("r"), func(cd *menu.CallbackData) {
			runtime.EventsEmit(ctx, "shortcut.view.refresh")
		}),
		menu.Text("Hard Refresh", keys.Combo("r", keys.CmdOrCtrlKey, keys.ShiftKey), func(cd *menu.CallbackData) {
			runtime.EventsEmit(ctx, "shortcut.view.hard-refresh")
		}),
		menu.Separator(),
		menu.Text("Open Collection", keys.CmdOrCtrl("o"), func(cd *menu.CallbackData) {
			runtime.EventsEmit(ctx, "shortcut.collection.open")
		}),
		menu.Text("Save Collection", keys.CmdOrCtrl("s"), func(cd *menu.CallbackData) {
			runtime.EventsEmit(ctx, "shortcut.collection.save")
		}),
		menu.Separator(),
		menu.Text("Print...", keys.CmdOrCtrl("p"), func(cd *menu.CallbackData) {
			runtime.EventsEmit(ctx, "shortcut.collection.print")
		}),
	)))

	if r.GOOS == "darwin" {
		m.Items = append(m.Items, menu.EditMenu())
	}

	m.Items = append(m.Items, menu.SubMenu("Language", menu.NewMenuFromItems(
		menu.Text("🇺🇸 English", nil, func(cd *menu.CallbackData) {
			runtime.EventsEmit(ctx, "shortcut.language.english")
		}),
		menu.Text("🇪🇸 Español", nil, func(cd *menu.CallbackData) {
			runtime.EventsEmit(ctx, "shortcut.language.spanish")
		}),
	)))

	runtime.MenuSetApplicationMenu(ctx, m)
}

// DomReady is called after the front-end dom has been loaded
func (app *App) DomReady(ctx context.Context) {
	// Add your action here
}

// Shutdown is called at application termination
func (app *App) Shutdown(ctx context.Context) {
	// Perform your teardown here
}

func (app *App) Title() string {
	if r.GOOS == "darwin" {
		return "🦄 Gvmapp"
	}

	return "Gvmapp"
}

func (app *App) OpenDirectoryDialog(title string) string {
	path, _ := runtime.OpenDirectoryDialog(app.Ctx, runtime.OpenDialogOptions{
		Title:                      title,
		CanCreateDirectories:       true,
		TreatPackagesAsDirectories: true,
	})

	return path
}

// OnBeforeClose
func (a *App) OnBeforeClose(ctx context.Context) bool {
	// 关闭xorm连接
	//a.Db.Close()
	// 返回 true 将阻止程序关闭
	return false
}

// OnDOMReady
func (a *App) OnDOMReady(ctx context.Context) {
	// 启动一个监听事件
	runtime.EventsOn(a.Ctx, "test", func(optionalData ...interface{}) {
		a.Log.Info(optionalData...)
	})
	//runtime.EventsEmit(a.Ctx, "alertList", a.GetAlertList())

}
func (a *App) OnDOMContentLoaded(arg1 string) string {
	return arg1
}
func (app *App) OpenFileDialog() string {
	path, _ := runtime.OpenFileDialog(app.Ctx, runtime.OpenDialogOptions{})

	return path
}

func (app *App) SaveFileDialog() string {
	path, _ := runtime.SaveFileDialog(app.Ctx, runtime.SaveDialogOptions{})

	return path
}

func (app *App) EncodeImage(path string) string {
	image, err := os.ReadFile(path)

	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	encoded := base64.StdEncoding.EncodeToString(image)
	encoded = fmt.Sprintf("data:image/png;base64,%s", encoded)

	return encoded
}

func (app *App) SaveFile(file string, data string) bool {
	path := fmt.Sprintf("%s%s", docsdir, file)

	err := os.WriteFile(path, []byte(data), os.ModePerm)

	return err == nil
}

func (app *App) GetImageStats(path string) image.Config {
	reader, err := os.Open(path)
	if err != nil {
		return image.Config{}
	}
	defer reader.Close()
	img, _, err := image.DecodeConfig(reader)
	if err != nil {
		return image.Config{}
	}

	return img
}

func (app *App) MessageDialog(options runtime.MessageDialogOptions) string {
	res, _ := runtime.MessageDialog(app.Ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         options.Title,
		Message:       options.Message,
		Buttons:       options.Buttons,
		DefaultButton: options.DefaultButton,
	})

	return res
}
