package controllers

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/qianlnk/VirtualHumanStudio/backend/config"
)

func UploadFileToServer(userID uint, typ string, fullFilePath string) (string, error) {
	filename := filepath.Base(fullFilePath)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileHandle, err := os.Open(fullFilePath)
	if err != nil {
		return "", err
	}
	defer fileHandle.Close()

	part, err := writer.CreateFormFile("attachment", filename)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, fileHandle)
	if err != nil {
		return "", err
	}

	fileUploadPath := fmt.Sprintf("vhs/%d/%ss", userID, typ)
	writer.WriteField("path", fileUploadPath)
	writer.Close()

	uploadResp, err := http.Post(config.AppConfig.FileUploadAPI, writer.FormDataContentType(), body)
	if err != nil {
		return "", err
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("上传文件失败，状态码: %d", uploadResp.StatusCode)
	}

	return path.Join(fileUploadPath, filename), nil
}
