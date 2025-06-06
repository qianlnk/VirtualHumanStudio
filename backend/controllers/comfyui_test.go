package controllers

import (
	"fmt"
	"testing"

	"github.com/qianlnk/VirtualHumanStudio/backend/config"
)

func TestQueryComfyUITaskStatus(t *testing.T) {
	config.AppConfig.ComfyUIConf.ServerURL = "https://aigc-ops-test.skyengine.com.cn/v1/model/comfyui/workflow-comfyui"
	config.AppConfig.ComfyUIConf.HistoryAPI = "/api/history"
	status, data, err := QueryComfyUITaskStatus("39d86343-af9d-44f2-bf2c-ca7bf813f88f")
	if err != nil {
		t.Errorf("QueryComfyUITaskStatus failed: %v", err)
	}
	// 将 data 写入文件

	fmt.Println(status, data)
}
