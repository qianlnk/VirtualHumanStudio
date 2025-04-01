package controllers

import (
	"VirtualHumanStudio/backend/config"
	"VirtualHumanStudio/backend/utils"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

func ttsInvoke(context context.Context, req *APITTSRequest) (string, error) {
	userID := context.Value("user_id").(uint)
	req.SpeakerName = fmt.Sprintf("%d_%s", userID, req.SpeakerName)

	// 序列化请求
	reqData, err := json.Marshal(req)
	if err != nil {
		return "", errors.New("序列化请求失败" + err.Error())
	}
	fmt.Println("======", string(reqData))
	// 发送请求
	resp, err := http.Post(config.AppConfig.TTSAPI, "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		return "", errors.New("调用API失败" + err.Error())
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("读取API响应失败" + err.Error())
	}

	// 解析响应
	var apiResp map[string]interface{}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", errors.New("解析API响应失败" + err.Error())
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("API调用失败")
	}

	// 文本转语音，保存输出文件
	waveBase64, ok := apiResp["wave_base64"].(string)
	if !ok {
		return "", errors.New("API响应中未包含音频数据")
	}

	// 生成唯一的输出文件名
	outputFileName := fmt.Sprintf("%s.wav", uuid.New().String())
	outputFilePath := utils.GetUserFilePath(userID, config.AppConfig.AudioDir, outputFileName)
	fullOutputFilePath := utils.GetFilePath(config.AppConfig.DataDir, outputFilePath)
	// 将Base64音频数据解码并保存为文件
	_, err = utils.Base64ToFile(waveBase64, fullOutputFilePath)
	if err != nil {
		return "", errors.New("保存音频文件失败" + err.Error())
	}

	return outputFilePath, nil
}
