package controllers

import (
	"fmt"
	"os"
	"testing"

	"VirtualHumanStudio/backend/config"
)

func TestCallASRService(t *testing.T) {
	// 保存原始的ASRAPI配置，测试结束后恢复
	originalASRAPI := "https://aigc-ops-test.skyengine.com.cn/v1/model/proxy/aigc-ops-llm-asr:8080/recognition"

	config.AppConfig.ASRAPI = originalASRAPI

	audioFile, err := os.Open("../data/1/uploads/99751628-9863-40c5-bea1-2cd09aef3175.wav")
	if err != nil {
		t.Fatalf("Failed to open audio file: %v", err)
	}
	defer audioFile.Close()

	res, err := callASRService("llm-asr_1", audioFile, "b64")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res.Text)

}
