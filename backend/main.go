package main

import (
	"fmt"
	"log"
	"os"

	"VirtualHumanStudio/backend/config"
	"VirtualHumanStudio/backend/controllers"
	"VirtualHumanStudio/backend/db"
	"VirtualHumanStudio/backend/middleware"
	"VirtualHumanStudio/backend/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	configPath := "./config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	err = db.InitDB()
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 数据库迁移
	err = db.MigrateDB(
		&models.User{},
		&models.VoiceClone{},
		&models.TTSTask{},
		&models.DigitalHuman{},
		&models.VoiceLibrary{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 初始化Redis
	middleware.InitRedis()

	// 创建Gin引擎
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// 配置CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 静态文件服务
	router.Static("/uploads", config.AppConfig.UploadDir)
	router.Static("/data", config.AppConfig.DataDir)

	// 注册路由
	registerRoutes(router)

	// 启动服务器
	serverAddr := fmt.Sprintf("%s:%d", config.AppConfig.ServerHost, config.AppConfig.ServerPort)
	fmt.Printf("服务器启动在 %s\n", serverAddr)
	err = router.Run(serverAddr)
	if err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

func registerRoutes(router *gin.Engine) {
	// API路由组
	api := router.Group("/api")

	// 公开路由
	public := api.Group("/")
	{
		// 用户认证
		public.POST("/register", controllers.Register)
		public.POST("/login", controllers.Login)
	}

	// 需要认证的路由
	protected := api.Group("/")
	protected.Use(middleware.JWTAuth())
	{
		// 用户相关
		protected.GET("/user", controllers.GetUserInfo)
		protected.PUT("/user", controllers.UpdateUserInfo)
		protected.POST("/logout", controllers.Logout)
		protected.PUT("/user/password", controllers.ChangePassword)

		// 音色克隆
		protected.POST("/voice/clone", controllers.CreateVoiceClone)
		protected.GET("/voice/clone/:id", controllers.GetVoiceClone)
		protected.GET("/voice/clones", controllers.ListVoiceClones)
		protected.DELETE("/voice/clone/:id", controllers.DeleteVoiceClone)
		protected.POST("/voice/clone/:id/retry", controllers.RetryVoiceClone)

		// 音色库
		protected.POST("/voice/upload", controllers.UploadVoice)
		protected.GET("/voices", controllers.ListVoices)
		protected.DELETE("/voice/:id", controllers.DeleteVoice)
		protected.GET("/voice/:id/download", controllers.DownloadVoice)

		// TTS
		protected.POST("/tts", controllers.CreateTTSTask)
		protected.GET("/tts/:id", controllers.GetTTSTask)
		protected.GET("/tts", controllers.ListTTSTasks)
		protected.DELETE("/tts/:id", controllers.DeleteTTSTask)
		protected.GET("/tts/:id/download", controllers.DownloadTTSOutput)

		// 数字人
		protected.POST("/digital-human", controllers.CreateDigitalHuman)
		protected.GET("/digital-human/:id", controllers.GetDigitalHuman)
		protected.GET("/digital-human/:id/progress", controllers.QueryDigitalHumanProgress)
		protected.GET("/digital-human", controllers.ListDigitalHumans)
		protected.DELETE("/digital-human/:id", controllers.DeleteDigitalHuman)
		protected.GET("/digital-human/:id/download", controllers.DownloadDigitalHumanResult)

		// 文件处理
		protected.GET("/file/view", controllers.FileView)
	}

	// 管理员路由
	admin := api.Group("/admin")
	admin.Use(middleware.AdminAuth())
	{
		// 用户管理
		admin.GET("/users", controllers.ListUsers)
		admin.PUT("/user/:id/status", controllers.UpdateUserStatus)
	}
}
