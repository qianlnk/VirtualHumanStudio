package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/qianlnk/VirtualHumanStudio/backend/config"

	"github.com/gin-gonic/gin"
)

// GetFilePath 获取文件路径（相对于数据目录）
func GetFilePath(baseDir, fileName string) string {
	return filepath.Join(baseDir, fileName)
}

// GetUserFilePath 获取用户文件路径（相对于数据目录）
func GetUserFilePath(userID uint, fileType, fileName string) string {
	// 构建相对路径
	relativePath := filepath.Join(strconv.Itoa(int(userID)), fileType, fileName)
	// 确保目录存在
	baseDir := filepath.Join(config.AppConfig.DataDir, filepath.Dir(relativePath))
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		fmt.Printf("创建目录失败: %v\n", err)
	}
	return relativePath
}

// GetFileExtension 获取文件扩展名
func GetFileExtension(fileName string) string {
	for i := len(fileName) - 1; i >= 0; i-- {
		if fileName[i] == '.' {
			return fileName[i+1:]
		}
	}
	return ""
}

// GetFileURL 生成文件访问URL
func GetFileURL(filePath string) string {
	if filePath == "" {
		return ""
	}

	// 移除数据目录前缀，只保留相对路径
	relativePath := filePath
	if config.AppConfig.DataDir != "" {
		relativePath = strings.TrimPrefix(filePath, config.AppConfig.DataDir)
	}

	// 构建完整的URL
	return fmt.Sprintf("%s/api/file/view?path=%s", config.AppConfig.Domain, relativePath)
}

// GetPaginationParams 获取分页参数
func GetPaginationParams(c *gin.Context) (page, size int) {
	// 默认值
	page = 1
	size = 10

	// 获取查询参数
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")

	// 转换为整数
	var err error
	page, err = strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size, err = strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 1000 {
		size = 10
	}

	return page, size
}

// Base64ToFile 将Base64编码的数据解码并保存为文件
func Base64ToFile(base64Data, filePath string) ([]byte, error) {
	// 解码Base64数据
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("解码Base64数据失败: %v", err)
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %v", err)
	}

	return data, nil
}

// DecodeJSON 解析HTTP响应体为JSON
func DecodeJSON(body io.Reader, v interface{}) error {
	return json.NewDecoder(body).Decode(v)
}
