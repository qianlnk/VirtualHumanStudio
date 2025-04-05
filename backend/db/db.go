package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"VirtualHumanStudio/backend/config"
	"VirtualHumanStudio/backend/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库连接
var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() error {
	// 配置GORM日志
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel: logger.Info,
		},
	)

	var err error

	// 根据配置选择数据库类型
	switch config.AppConfig.DBType {
	case "sqlite":
		// 确保SQLite数据库目录存在
		dbDir := filepath.Dir(config.AppConfig.SQLitePath)
		if _, err := os.Stat(dbDir); os.IsNotExist(err) {
			err = os.MkdirAll(dbDir, 0755)
			if err != nil {
				return fmt.Errorf("无法创建SQLite数据库目录: %v", err)
			}
		}

		// 连接SQLite数据库
		DB, err = gorm.Open(sqlite.Open(config.AppConfig.SQLitePath), &gorm.Config{
			Logger: newLogger,
		})
		if err != nil {
			return fmt.Errorf("无法连接SQLite数据库: %v", err)
		}

		// 自动迁移数据库表结构
		err = DB.AutoMigrate(
			&models.User{},
			&models.VoiceClone{},
			&models.TTSTask{},
			&models.DigitalHuman{},
			&models.ASRTask{}, // 添加ASR任务表迁移
		)
		if err != nil {
			return fmt.Errorf("数据库迁移失败: %v", err)
		}

		fmt.Println("已连接SQLite数据库:", config.AppConfig.SQLitePath)

	case "mysql":
		// 构建MySQL连接DSN
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.AppConfig.MySQLUser,
			config.AppConfig.MySQLPass,
			config.AppConfig.MySQLHost,
			config.AppConfig.MySQLPort,
			config.AppConfig.MySQLDB,
		)

		// 连接MySQL数据库
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: newLogger,
		})
		if err != nil {
			return fmt.Errorf("无法连接MySQL数据库: %v", err)
		}

		fmt.Println("已连接MySQL数据库:", config.AppConfig.MySQLHost)

	default:
		return fmt.Errorf("不支持的数据库类型: %s", config.AppConfig.DBType)
	}

	return nil
}

// MigrateDB 执行数据库迁移
func MigrateDB(models ...interface{}) error {
	err := DB.AutoMigrate(models...)
	if err != nil {
		return fmt.Errorf("数据库迁移失败: %v", err)
	}

	fmt.Println("数据库迁移成功")
	return nil
}
