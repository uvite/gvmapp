package main

import (
	"embed"
	"fmt"
	gvmapp "github.com/uvite/gvmapp/backend/gvmapp"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/options/mac"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed frontend/dist
var assets embed.FS

// icon会默认使用 build/appicon.png 转换为byte数组
var icon []byte

type FileLoader struct {
	http.Handler
}

func NewFileLoader() *FileLoader {
	return &FileLoader{}
}

func (h *FileLoader) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	requestedFilename := strings.TrimPrefix(req.URL.Path, "/")
	fileData, err := os.ReadFile("/" + requestedFilename)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(fmt.Sprintf("Could not load file %s", requestedFilename)))
	}
	res.Write(fileData)
}

func main() {
	// Create an instance of the app structure
	app := gvmapp.NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title: app.Title(),

		Width:             1100,  // 启动宽度
		Height:            768,   // 启动高度
		MinWidth:          1100,  // 最小宽度
		MinHeight:         768,   // 最小高度
		HideWindowOnClose: false, // 关闭的时候隐藏窗口
		StartHidden:       false, // 启动的时候隐藏窗口 （建议生产环境关闭此项，开发环境开启此项，原因自己体会）
		AlwaysOnTop:       false, // 窗口固定在最顶层

		DisableResize: false,
		Fullscreen:    false,
		Frameless:     false,

		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: NewFileLoader(),
		},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 128},
		OnStartup:        app.Startup,
		OnDomReady:       app.DomReady,
		OnShutdown:       app.Shutdown,
		OnBeforeClose:    app.OnBeforeClose,
		CSSDragProperty:  "--wails-draggable",
		CSSDragValue:     "drag",

		LogLevel: logger.DEBUG,

		Bind: []interface{}{
			app,
			app.StartService,
			app.LauncherService,
			app.ExchangeService,
			app.AlertService,
			app.PoolService,
		},
		// Windows platform specific options
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: false,
			WebviewUserDataPath:               "",
		},

		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: false,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            true,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   app.Title(),
				Message: "A bot manange",
				Icon:    icon,
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
