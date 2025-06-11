package main

import (
	"context"
	"fmt"

	"github.com/qianlnk/VirtualHumanStudio/backend/config"
	"github.com/qianlnk/VirtualHumanStudio/backend/storages"
)

// 同步COS上的文件到OBS

func main() {
	configPath := "../../config.json"
	config.LoadConfig(configPath)

	cos, err := storages.NewCOS(config.AppConfig.CosStorage)
	if err != nil {
		panic(err)
	}
	obs, err := storages.NewOBS(config.AppConfig.ObsStorage)
	if err != nil {
		panic(err)
	}

	cosFiles, err := cos.List(context.Background(), "", 0, 1000)
	if err != nil {
		panic(err)
	}

	fmt.Println(len(cosFiles))

	for _, cf := range cosFiles {
		fmt.Println("syncing: ", cf.Name)
		f, err := cos.OpenFile(context.Background(), cf.Name)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		err = obs.SaveFile(context.Background(), cf.Name, f)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("synced")
}
