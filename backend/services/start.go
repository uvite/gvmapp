package services

import (
	"context"
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"github.com/studio-b12/gowebdav"
	"github.com/uvite/gvmapp/backend/lib"
	"github.com/uvite/gvmapp/backend/util"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"
)

type StartService struct {
	Ctx context.Context
}

func NewStartService() *StartService {
	start := &StartService{}
	start.SetLog()
	return start
}

var client = gowebdav.NewClient("https://dav.jianguoyun.com/dav/",
	"airwms@126.com", "anpjd37an6vg65qv")

// 同步到本地
func (f *StartService) SetLog() {
	logDir := "log"
	if err := os.MkdirAll(logDir, 0777); err != nil {
		log.Panic(err)
	}
	writer, err := rotatelogs.New(
		path.Join(logDir, "access_log.%Y%m%d"),
		rotatelogs.WithLinkName("access_log"),
		// rotatelogs.WithMaxAge(24 * time.Hour),
		rotatelogs.WithRotationTime(time.Duration(24)*time.Hour),
	)
	if err != nil {
		log.Panic(err)
	}
	logger := log.StandardLogger()
	logger.AddHook(
		lfshook.NewHook(
			lfshook.WriterMap{
				log.DebugLevel: writer,
				log.InfoLevel:  writer,
				log.WarnLevel:  writer,
				log.ErrorLevel: writer,
				log.FatalLevel: writer,
			},
			&log.JSONFormatter{},
		),
	)
	log.Infof("debug mode is enabled")
	log.SetLevel(log.DebugLevel)
}
func (f *StartService) DownToLocal() util.RespDate {
	folders := make([]string, 0)
	localData := util.GetLocalDir()
	fs, err := client.ReadDir("bbgo")
	if err != nil {
		util.Message(f.Ctx, fmt.Sprintf("获取策略模板出错%s", err))
	}
	// 获取云端文件夹
	for _, f := range fs {
		if f.IsDir() {
			folders = append(folders, f.Name())
		}
	}

	for _, folderName := range folders {
		localFolderPath := filepath.Join(localData, folderName)
		// fmt.Println("文件夹:" + localFolderPath)
		// 云端文件夹路径
		webdavFolderPath := fmt.Sprintf("bbgo/%s", folderName)
		// fmt.Println("云端文件夹:" + webdavFolderPath)
		// 判断本地笔记本文件夹是否存在
		_, err2 := os.Stat(localFolderPath)
		if os.IsNotExist(err2) { // 本地文件夹不存在
			// 创建本地文件夹
			err3 := os.Mkdir(localFolderPath, os.ModePerm)
			if err3 != nil {
				util.Message(f.Ctx, fmt.Sprintf("获取策略模板出错%s", err))

			}
			// fmt.Println("创建本地文件夹成功")
		}

		// 遍历 笔记/文章 md文件并下载到本地
		files, _ := client.ReadDir(webdavFolderPath)
		for _, file := range files {
			fileName := file.Name()
			// fmt.Println("同步到本地时遍历的云端文件：" + fileName)
			// 读取云端文件
			bytes, _ := client.Read(fmt.Sprintf("%s/%s", webdavFolderPath, fileName))
			// 下载文件到本地（不存在则创建，存在则覆盖）
			fPath := filepath.Join(localFolderPath, fileName)
			// fmt.Println("同步保存到得本地文件路径：" + fPath)
			ioutil.WriteFile(fPath, bytes, os.ModePerm)
			// 授权
			os.Chmod(fPath, os.ModePerm)
		}
	}
	data := map[string]interface{}{}
	data["type"] = "syscdir"

	runtime.EventsEmit(f.Ctx, "debug", data)

	return f.GetDirs()
}

func (f *StartService) GetDirs() util.RespDate {
	dirs := make([]string, 0)
	files := make([]string, 0)

	localData := util.GetLocalDir()

	fs, err := ioutil.ReadDir(localData)
	if err != nil {
		util.Message(f.Ctx, fmt.Sprintf("获取策略模板出错%s", err))
	}

	for _, f := range fs {
		if f.IsDir() {
			// fmt.Println(f.Name())
			dirs = append(dirs, f.Name())
			filepath.Walk(filepath.Join(localData, f.Name()),

				func(path string, info os.FileInfo, err error) error {

					if err != nil {
						return err
					}
					files = append(files, info.Name())

					//fmt.Println(path, info.Size())
					return err
				})
		}
	}
	ret := map[string]interface{}{
		"dirs":  dirs,
		"files": files,
	}
	return util.RespDate{Code: 200, Data: ret}
}

func (f *StartService) ReadFileContents(path string) string {
	contents, err := lib.ReadFileContents(path)

	if err != nil {
		fmt.Println(err)
		return ""
	}

	return contents
}

func (f *StartService) OpenFileContents() string {
	contents, _ := lib.OpenFileContents(f.Ctx)

	return contents
}
