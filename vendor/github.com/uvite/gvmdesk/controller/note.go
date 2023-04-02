package gtools

import (
	"fmt"
	"github.com/studio-b12/gowebdav"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type RespDate struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

// 创建客户端
var client = gowebdav.NewClient("https://dav.jianguoyun.com/dav/",
	"airwms@126.com", "anpjd37an6vg65qv")

// 读取坚果云笔记本
func (a *App) GetJGTestDir() RespDate {
	fs, err := client.ReadDir("bbgo")
	if err != nil {
		// fmt.Println(err.Error())
		return RespDate{Code: 500, Msg: err.Error()}
	}

	dirs := make([]string, 0)
	for _, f := range fs {
		if f.IsDir() {
			// fmt.Println(f.Name())
			dirs = append(dirs, f.Name())
		}
	}

	return RespDate{Code: 200, Data: dirs}
}

// 读取坚果云笔记本文件夹下的笔记
func (a *App) GetJGTestDirFile(dirName string) RespDate {
	files, err := client.ReadDir("bbgo/" + dirName)
	if err != nil {
		return RespDate{Code: 500, Msg: err.Error()}
	}

	fs := make([]string, 0)
	for _, file := range files {
		// fmt.Println("坚果云文件夹下笔记：" + file.Name())
		fs = append(fs, file.Name())
	}

	return RespDate{Code: 200, Data: fs}
}

// 创建笔记本文件夹
func (a *App) CreateNotebook(dirName string) RespDate {
	pwd, _ := os.Getwd()
	localData := filepath.Join(pwd, "local-data")
	// fmt.Println(localData)
	folderPath := filepath.Join(localData, dirName)
	// fmt.Println(folderPath)
	if _, err := os.Stat(folderPath); os.IsNotExist(err) { // 文件夹不存在
		// 创建文件夹
		err := os.MkdirAll(folderPath, os.ModePerm)
		if err != nil {
			return RespDate{Code: 500, Msg: err.Error()}
		}

		// 授权
		os.Chmod(folderPath, os.ModePerm)

		//定义一个同步等待的组
		var wg sync.WaitGroup
		var inErr error
		go func() {
			wg.Add(1)
			path := "bbgo/" + dirName

			if _, err := client.Stat(path); err != nil { // 文件夹不存在
				err := client.MkdirAll(path, 0644)
				if err != nil {
					// fmt.Println("坚果云文件夹创建失败")
					inErr = err
				}
			}

			defer wg.Done()
		}()

		wg.Wait()

		if inErr != nil {
			return RespDate{Code: 500, Msg: "坚果云文件夹创建失败"}
		}

		return RespDate{Code: 200}
	}
	return RespDate{Code: 400, Msg: "文件夹已存在"}
}

// 遍历目录
func (a *App) GetDirs() RespDate {
	dirs := make([]string, 0)
	pwd, _ := os.Getwd()
	localData := filepath.Join(pwd, "local-data")
	fs, err := ioutil.ReadDir(localData)
	if err != nil {
		return RespDate{Code: 500, Msg: err.Error()}
	}

	for _, f := range fs {
		if f.IsDir() {
			// fmt.Println(f.Name())
			dirs = append(dirs, f.Name())
		}
	}

	return RespDate{Code: 200, Data: dirs}
}

// 创建笔记文件
func (a *App) CreateNoteFile(dirName string) RespDate {
	pwd, _ := os.Getwd()
	// 未命名.md文件路径
	defaultFile := filepath.Join(pwd, "local-data", "test.js")
	newFileName := time.Now().Local().Format("20060102150405")
	destFilePath := filepath.Join(pwd, "local-data", dirName, (newFileName))
	// 判断文件是否存在
	_, err := os.Stat(defaultFile)
	if err == nil { // 文件存在
		input, err := ioutil.ReadFile(defaultFile)
		if err != nil {
			// fmt.Println(err)
			return RespDate{Code: 500, Msg: "文件读取失败"}
		}

		err = ioutil.WriteFile(destFilePath, input, os.ModePerm)
		if err != nil {
			// fmt.Println("创建失败：", destinationFile)
			// fmt.Println(err)
			return RespDate{Code: 500, Msg: "文件创建失败"}
		}

		// 授权
		os.Chmod(destFilePath, os.ModePerm)
		return RespDate{Code: 200, Data: newFileName, Msg: "文件创建成功"}
	}

	return RespDate{Code: 400, Msg: "Markdown模板文件不存在"}
}

// 保存文章(笔记)
func (a *App) SaveNote(dirName, noteName, noteTitle, noteContent string) RespDate {
	pwd, _ := os.Getwd()
	// md文件路径
	oldFilePath := filepath.Join(pwd, "local-data", dirName, (noteName + ".md"))
	// 重命名文件路径
	newFilePath := filepath.Join(pwd, "local-data", dirName, (noteTitle + ".md"))

	// 写入内容到文件
	err := ioutil.WriteFile(oldFilePath, []byte(noteContent), os.ModePerm)
	if err != nil {
		fmt.Println(err.Error())
		return RespDate{Code: 500, Msg: err.Error()}
	}

	// 文件重命名
	err = os.Rename(oldFilePath, newFilePath)
	if err != nil {
		fmt.Println(err.Error())
		return RespDate{Code: 500, Msg: err.Error()}
	}

	// 上传到坚果云
	webdavFilePath := fmt.Sprintf("bbgo/%s/%s.md", dirName, noteTitle)
	bytes, _ := ioutil.ReadFile(newFilePath)
	err = client.Write(webdavFilePath, bytes, 0644)
	if err != nil {
		return RespDate{Code: 500, Msg: err.Error()}
	}

	return RespDate{Code: 200, Data: newFilePath}
}

// 读取文件
func (a *App) ReadNoteFile(dirName, fileName string) RespDate {
	pwd, _ := os.Getwd()
	filePath := filepath.Join(pwd, "local-data", dirName, (fileName))
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return RespDate{Code: 500, Msg: err.Error()}
	}
	return RespDate{Code: 200, Data: string(data)}
}

// 删除笔记
func (a *App) RemoveNote(dirName, fileName string) RespDate {
	var err1 error
	var err2 error
	var wg sync.WaitGroup

	pwd, _ := os.Getwd()
	localFilePath := filepath.Join(pwd, "local-data", dirName, (fileName + ".md"))
	go func() {
		wg.Add(1)
		// 删除本地端文件
		err1 = os.Remove(localFilePath)
		defer wg.Done()
	}()

	go func() {
		wg.Add(1)
		webdavFilePath := fmt.Sprintf("bbgo/%s/%s.md", dirName, fileName)
		_, err3 := client.Stat(webdavFilePath)
		if err3 == nil { // 文件存在
			// 删除云端文件
			err2 = client.Remove(webdavFilePath)
		}

		defer wg.Done()
	}()

	wg.Wait()

	if err1 != nil {
		return RespDate{Code: 500, Msg: err1.Error()}
	}
	if err2 != nil {
		return RespDate{Code: 500, Msg: err2.Error()}
	}

	return RespDate{Code: 200}
}

// 同步到云端
func (a *App) SyncToCloud() RespDate {
	folders := make([]string, 0)
	pwd, _ := os.Getwd()
	localData := filepath.Join(pwd, "local-data")
	fs, err := ioutil.ReadDir(localData)
	if err != nil {
		return RespDate{Code: 500, Msg: err.Error()}
	}
	// 遍历本地笔记本文件夹
	for _, f := range fs {
		if f.IsDir() {
			folders = append(folders, f.Name())
		}
	}

	for _, folderName := range folders {
		localFolderPath := filepath.Join(localData, folderName)
		webdavFolderPath := fmt.Sprintf("bbgo/%s", folderName)
		// 判断云端笔记本文件夹是否存在
		_, err2 := client.Stat(webdavFolderPath)
		if err2 != nil { // 云端文件夹不存在
			// 创建文件夹
			err3 := client.Mkdir(webdavFolderPath, 0644)
			if err3 != nil {
				return RespDate{Code: 500, Msg: err3.Error()}
			}
		}

		// 遍历 笔记/文章 md文件并上传到云
		files, _ := ioutil.ReadDir(localFolderPath)
		for _, file := range files {
			fileName := file.Name()
			// fmt.Println("同步到云时遍历的本地文件：" + fileName)
			// 读取文件
			bytes, _ := ioutil.ReadFile(filepath.Join(localFolderPath, fileName))
			// 上传到云端(不存在则创建，存在则覆盖)
			err4 := client.Write(fmt.Sprintf("%s/%s", webdavFolderPath, file.Name()), bytes, 0644)
			if err4 != nil {
				return RespDate{Code: 500, Msg: err4.Error()}
			}
		}
	}
	return RespDate{Code: 200}
}

// 同步到本地
func (a *App) DownToLocal() RespDate {
	folders := make([]string, 0)
	pwd, _ := os.Getwd()
	localData := filepath.Join(pwd, "local-data")

	fs, err := client.ReadDir("bbgo")
	if err != nil {
		return RespDate{Code: 500, Msg: err.Error()}
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
				return RespDate{Code: 500, Msg: err3.Error()}
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

	return RespDate{Code: 200}
}
