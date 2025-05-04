package controllers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// ImageProcessingModule 图像处理模块定义
type ImageProcessingModule struct {
	ID           string     `json:"id"`           // 模块ID，用于前端路由和后端处理
	Name         string     `json:"name"`         // 模块名称，显示在菜单中
	Description  string     `json:"description"`  // 模块描述
	Icon         string     `json:"icon"`         // 图标名称，使用Element UI的图标
	Route        string     `json:"route"`        // 前端路由路径
	InputParams  []ParamDef `json:"inputParams"`  // 输入参数定义
	OutputParams []ParamDef `json:"outputParams"` // 输出参数定义
}

// ParamDef 参数定义
type ParamDef struct {
	Key         string   `json:"key"`               // 参数键名
	Alias       string   `json:"alias"`             // 参数显示名称
	Type        string   `json:"type"`              // 参数类型：text, image, select, number等
	Required    bool     `json:"required"`          // 是否必填
	Description string   `json:"description"`       // 参数描述
	Default     string   `json:"default"`           // 默认值
	Options     []Option `json:"options,omitempty"` // 选项，用于select类型
}

// Option 选项定义，用于select类型参数
type Option struct {
	Label string `json:"label"` // 选项标签
	Value string `json:"value"` // 选项值
}

// ModuleConfig 模块配置文件结构
type ModuleConfig struct {
	Modules []ImageProcessingModule `json:"modules"`
}

func getModuleConfig() (*ModuleConfig, error) {
	// 读取模块配置文件
	configPath := filepath.Join("modules.json")
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	// 解析配置文件
	var moduleConfig ModuleConfig
	if err := json.Unmarshal(configFile, &moduleConfig); err != nil {
		return nil, err
	}
	return &moduleConfig, nil
}

func getMuduleConfigByID(moduleID string) (*ImageProcessingModule, error) {
	modules, err := getModuleConfig()
	if err != nil {
		return nil, err
	}
	for _, module := range modules.Modules {
		if module.ID == moduleID {
			return &module, nil
		}
	}

	return nil, fmt.Errorf("module not found")
}

// GetImageProcessingModules 获取图像处理模块列表
func GetImageProcessingModules(c *gin.Context) {
	// 读取模块配置文件
	moduleConfig, err := getModuleConfig()
	if err != nil {
		c.JSON(500, gin.H{"error": "读取模块配置失败"})
		return
	}

	// 返回模块列表
	c.JSON(200, gin.H{
		"success": true,
		"modules": moduleConfig.Modules,
	})
}
