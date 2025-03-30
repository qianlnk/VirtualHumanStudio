package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 应用配置结构
type Config struct {
	// 服务器配置
	ServerPort int    `json:"server_port"`
	ServerHost string `json:"server_host"`
	JWTSecret  string `json:"jwt_secret"`

	// 数据库配置
	DBType     string `json:"db_type"` // sqlite 或 mysql
	SQLitePath string `json:"sqlite_path"`

	// MySQL配置
	MySQLHost string `json:"mysql_host"`
	MySQLPort int    `json:"mysql_port"`
	MySQLUser string `json:"mysql_user"`
	MySQLPass string `json:"mysql_pass"`
	MySQLDB   string `json:"mysql_db"`

	// Redis配置
	RedisHost     string `json:"redis_host"`
	RedisPort     int    `json:"redis_port"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`

	// 文件存储配置
	UploadDir string `json:"upload_dir"`

	// API配置
	VoiceCloneAPI      string `json:"voice_clone_api"`
	TTSAPI             string `json:"tts_api"`
	DigitalHumanAPI    string `json:"digital_human_api"`
	DigitalHumanQuery  string `json:"digital_human_query"`
	DigitalHumanResult string `json:"digital_human_result"`
}

// AppConfig 全局配置实例
var AppConfig Config

// LoadConfig 从配置文件加载配置
func LoadConfig(configPath string) error {
	// 设置默认配置
	AppConfig = Config{
		ServerPort: 8080,
		ServerHost: "0.0.0.0",
		JWTSecret:  "virtual_human_studio_secret_key",

		DBType:     "sqlite",
		SQLitePath: "./data/vhs.db",

		MySQLHost: "localhost",
		MySQLPort: 3306,
		MySQLUser: "root",
		MySQLPass: "",
		MySQLDB:   "virtual_human_studio",

		RedisHost:     "localhost",
		RedisPort:     6379,
		RedisPassword: "",
		RedisDB:       0,

		UploadDir: "./uploads",

		VoiceCloneAPI:      "http://10.64.24.23:8082/api/voice/clone",
		TTSAPI:             "http://10.64.24.23:8082/api/tts", // 假设的TTS API
		DigitalHumanAPI:    "http://10.64.24.249:8383/easy/submit",
		DigitalHumanQuery:  "http://10.64.24.249:8383/easy/query",
		DigitalHumanResult: "http://10.64.24.249:8383/easy/result", // 假设的结果获取API
	}

	// 如果配置文件存在，则从文件加载配置
	if configPath != "" {
		// 确保配置文件目录存在
		configDir := filepath.Dir(configPath)
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			err = os.MkdirAll(configDir, 0755)
			if err != nil {
				return fmt.Errorf("无法创建配置目录: %v", err)
			}
		}

		// 检查配置文件是否存在
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// 配置文件不存在，创建默认配置文件
			configJSON, err := json.MarshalIndent(AppConfig, "", "  ")
			if err != nil {
				return fmt.Errorf("无法序列化默认配置: %v", err)
			}

			err = os.WriteFile(configPath, configJSON, 0644)
			if err != nil {
				return fmt.Errorf("无法写入默认配置文件: %v", err)
			}

			fmt.Printf("已创建默认配置文件: %s\n", configPath)
		} else {
			// 配置文件存在，从文件加载配置
			configData, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("无法读取配置文件: %v", err)
			}

			err = json.Unmarshal(configData, &AppConfig)
			if err != nil {
				return fmt.Errorf("无法解析配置文件: %v", err)
			}

			fmt.Printf("已从 %s 加载配置\n", configPath)
		}
	} else {
		fmt.Println("使用默认配置")
	}

	// 确保上传目录存在
	if _, err := os.Stat(AppConfig.UploadDir); os.IsNotExist(err) {
		err = os.MkdirAll(AppConfig.UploadDir, 0755)
		if err != nil {
			return fmt.Errorf("无法创建上传目录: %v", err)
		}
	}

	return nil
}
