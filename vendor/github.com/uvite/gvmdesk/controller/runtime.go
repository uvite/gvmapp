package gtools

import (
	"github.com/uvite/gvmdesk/configs"
	"github.com/uvite/gvmdesk/util"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) OpenMdSaveFileWindow() *util.Resp {
	option := runtime.SaveDialogOptions{
		DefaultFilename:      "new",
		Title:                "保存文件",
		Filters:              [](runtime.FileFilter){configs.MdFilter},
		CanCreateDirectories: true,
	}
	fpath, err := runtime.SaveFileDialog(a.Ctx, option)
	if err != nil {
		a.Log.Error("新markdown路径获取失败")
		return util.Error("新markdown路径获取失败")
	}
	return util.Success(fpath)
}

func (a *App) OpenMdFolderWindow() *util.Resp {
	options := runtime.OpenDialogOptions{
		Title:                "选择文件夹",
		CanCreateDirectories: true,
	}
	dirPath, err := runtime.OpenDirectoryDialog(a.Ctx, options)
	if err != nil {
		return util.Error(err.Error())
	}
	return util.Success(dirPath)
}

func (a *App) OpenHtmlSaveWindow() *util.Resp {
	options := runtime.SaveDialogOptions{
		Title:                "导出为HTML",
		DefaultFilename:      "new",
		Filters:              [](runtime.FileFilter){configs.HtmlFilter},
		CanCreateDirectories: true,
	}
	fpath, err := runtime.SaveFileDialog(a.Ctx, options)
	if err != nil {
		return util.Error(err.Error())
	}
	return util.Success(fpath)
}
