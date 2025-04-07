package controllers

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestQueryComfyUITaskStatus(t *testing.T) {

	body, err := os.ReadFile("/Users/xiezhenjia/go/src/github.com/qianlnk/VirtualHumanStudio/backend/data/workflow/a")

	// 解析响应
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		panic(err)
	}

	// 检查响应是否为空
	if response == nil || len(response) == 0 {
		panic(fmt.Errorf("响应数据为空"))
	}

	// 检查任务状态
	status := "processing"

	// 尝试获取任务数据
	data, ok := response["5fa75f2e-ab79-4203-ade0-611981ee6e36"].(map[string]interface{})
	if !ok {
		// 记录响应内容以便调试
		respBytes, _ := json.Marshal(response)
		fmt.Printf("响应数据结构异常: %s\n", string(respBytes))

		// 检查是否有其他可用的数据
		if len(response) > 0 {
			// 尝试使用第一个可用的键
			for key, value := range response {
				if mapData, mapOk := value.(map[string]interface{}); mapOk {
					fmt.Printf("使用替代键: %s\n", key)
					data = mapData
					ok = true
					break
				}
			}
		}

		// 如果仍然没有找到有效数据
		if !ok {
			panic(fmt.Errorf("任务数据格式错误: 响应中不包含有效的任务数据"))
		}
	}

	// 检查是否有输出节点的数据
	outputs, ok := data["outputs"].(map[string]interface{})
	if ok && len(outputs) > 0 {
		// 检查SaveImage节点的输出
		if nodeOutput, ok := outputs["9"].(map[string]interface{}); ok {
			if images, ok := nodeOutput["images"].([]interface{}); ok && len(images) > 0 {
				status = "completed"
			}
		} else {
			// 尝试查找任何包含images的节点
			for nodeID, nodeData := range outputs {
				if nodeMap, ok := nodeData.(map[string]interface{}); ok {
					if images, ok := nodeMap["images"].([]interface{}); ok && len(images) > 0 {
						fmt.Printf("找到替代输出节点: %s\n", nodeID)
						status = "completed"
						break
					}
				}
			}
		}
	} else if execInfo, ok := data["exec_info"].(map[string]interface{}); ok {
		// 检查执行信息
		if execStatus, ok := execInfo["status"].(string); ok {
			if execStatus == "error" || execStatus == "failed" {
				status = "failed"
			} else if execStatus == "success" || execStatus == "completed" {
				status = "completed"
			}
		}
	}

	fmt.Println(status, data)
}
