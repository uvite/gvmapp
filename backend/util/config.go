package util

import (
	"embed"
	"github.com/uvite/gvmapp/backend/lib"
	"os"
	"os/user"
)

//go:embed assets
var assets embed.FS

func GetEnvDir() string {
	var localDir = GetConfigDir() + "/genv"
	_, err := os.Stat(localDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(localDir, os.ModePerm)
			env, _ := assets.ReadFile("assets/env.local")

			err := lib.WriteFileContents(localDir+"/.env.local", env)
			bbgo, _ := assets.ReadFile("assets/bbgo.yaml")
			err = lib.WriteFileContents(localDir+"/bbgo.yaml", bbgo)

			if err != nil {
				panic("配置文件夹创建失败")
			}
		} else {
			panic("配置文件夹不存在--" + err.Error())
		}
	}
	return localDir
}

func GetLocalDir() string {
	var localDir = GetConfigDir() + "/local"
	_, err := os.Stat(localDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(localDir, os.ModePerm)
			err = os.Mkdir(GetConfigDir()+"/js", os.ModePerm)
			if err != nil {
				panic("配置文件夹创建失败")
			}
		} else {
			panic("配置文件夹不存在--" + err.Error())
		}
	}
	return localDir
}

func GetConfigDir() string {
	// 获取配置文件夹路径路径
	userInfo, err := user.Current()
	if err != nil {
		panic("gvmapp配置文件夹路径获取失败" + err.Error())
	}
	var homeDir = userInfo.HomeDir
	// 判断 homeDir/gvmapp 文件夹是否存在
	var gtDir = homeDir + "/.gvmapp"

	_, err = os.Stat(gtDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(gtDir, os.ModePerm)
			//err = os.Mkdir(gtDir, os.ModePerm)

			if err != nil {
				panic("配置文件夹创建失败")
			}
		} else {
			panic("配置文件夹不存在--" + err.Error())
		}
	}
	return gtDir
}
