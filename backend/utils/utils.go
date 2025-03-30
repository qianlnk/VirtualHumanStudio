package utils

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
	if err != nil || size < 1 || size > 100 {
		size = 10
	}

	return page, size
}

// GetFilePath 获取文件路径
func GetFilePath(baseDir, fileName string) string {
	return fmt.Sprintf("%s/%s", baseDir, fileName)
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
