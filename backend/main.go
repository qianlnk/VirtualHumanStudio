package main

import (
	"fmt"
	"log"
	"os"

	"github.com/qianlnk/VirtualHumanStudio/backend/config"
	"github.com/qianlnk/VirtualHumanStudio/backend/controllers"
	"github.com/qianlnk/VirtualHumanStudio/backend/db"
	"github.com/qianlnk/VirtualHumanStudio/backend/middleware"
	"github.com/qianlnk/VirtualHumanStudio/backend/models"
	"github.com/qianlnk/VirtualHumanStudio/backend/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// 全局变量
var membershipController *controllers.MembershipController

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
		&models.ComfyUITask{},
		&models.Accessory{},
		&models.Message{},
		&models.UserLoginLog{},
		&models.ModuleUsageLog{},
		&models.DailyStatistics{},
		&models.Membership{},
		&models.MembershipPlan{},
		&models.MembershipOrder{},
		&controllers.UserUsage{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 初始化Redis
	middleware.InitRedis()

	// 初始化Promptt
	controllers.InitPromptt()

	// 初始化会员控制器
	membershipController = controllers.NewMembershipController(db.GetDB())

	// 初始化会员服务
	membershipService := services.NewMembershipService(db.GetDB())
	middleware.SetMembershipService(membershipService)

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
	api.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 公开路由
	public := api.Group("/")
	{
		// 用户认证
		public.POST("/register", controllers.Register)
		public.POST("/login", controllers.Login)
		public.GET("/file/view", controllers.FileView)
	}

	// 需要认证的路由
	protected := api.Group("/")
	protected.Use(middleware.JWTAuth())
	// 添加统计中间件
	protected.Use(middleware.StatisticsMiddleware())
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
		protected.POST("/voice/clone/:id/add_to_library", controllers.AddVoiceToLibrary) // 添加到音色库

		// 音色库
		protected.POST("/voice/upload", controllers.UploadVoice)
		protected.GET("/voices", controllers.ListVoices)
		protected.DELETE("/voice/:id", controllers.DeleteVoice)

		// TTS
		protected.POST("/tts", controllers.CreateTTSTask)
		protected.GET("/tts/:id", controllers.GetTTSTask)
		protected.GET("/tts", controllers.ListTTSTasks)
		protected.DELETE("/tts/:id", controllers.DeleteTTSTask)

		// ASR
		protected.POST("/asr", controllers.CreateASRTask)
		protected.GET("/asr/:id", controllers.GetASRTask)
		protected.GET("/asr", controllers.ListASRTasks)
		protected.DELETE("/asr/:id", controllers.DeleteASRTask)

		// 数字人
		protected.POST("/digital-human", controllers.CreateDigitalHuman)
		protected.GET("/digital-human/:id", controllers.GetDigitalHuman)
		protected.GET("/digital-human/:id/progress", controllers.QueryDigitalHumanProgress)
		protected.GET("/digital-human", controllers.ListDigitalHumans)
		protected.DELETE("/digital-human/:id", controllers.DeleteDigitalHuman)

		// 文件处理
		// protected.GET("/file/view", controllers.FileView)

		// comfyui相关功能
		// 饰品替换，输入白底图片和模特物品图片，蒙版图，输出替换后的图片
		protected.POST("/accessory", controllers.CreateAccessory)
		protected.GET("/accessory/:id", controllers.GetAccessory)
		protected.GET("/accessory", controllers.ListAccessories)
		protected.DELETE("/accessory/:id", controllers.DeleteAccessory)

		// 图像处理API
		protected.GET("/image-processing/modules", controllers.GetImageProcessingModules)
		protected.GET("/image-processing/tasks/:moduleId", controllers.GetImageProcessingTasks)
		protected.POST("/image-processing/tasks/:moduleId", controllers.CreateImageProcessingTask)
		protected.GET("/image-processing/tasks/:moduleId/:taskId", controllers.GetImageProcessingTask)
		protected.DELETE("/image-processing/tasks/:moduleId/:taskId", controllers.DeleteImageProcessingTask)
		protected.POST("/image-processing/tasks/:moduleId/:taskId/retry", controllers.RetryImageProcessingTask)

		// 留言
		protected.POST("/message", controllers.CreateMessage)
		protected.GET("/message/:id", controllers.GetMessage)
		protected.GET("/messages", controllers.ListMessages)
		protected.DELETE("/message/:id", controllers.DeleteMessage)
	}

	// 会员中心路由
	membershipAPI := api.Group("/membership")
	{
		membershipAPI.GET("/plans", membershipController.GetMembershipPlans)
		membershipAPI.GET("/user", middleware.JWTAuth(), membershipController.GetUserMembership)
		membershipAPI.POST("/purchase", middleware.JWTAuth(), membershipController.PurchaseMembership)
		membershipAPI.POST("/cancel", middleware.JWTAuth(), membershipController.CancelAutoRenew)
		membershipAPI.GET("/daily-usage", middleware.JWTAuth(), membershipController.GetDailyUsage)
		membershipAPI.GET("/pending-orders", middleware.JWTAuth(), membershipController.GetUserPendingOrders)
	}

	// 管理员路由
	admin := api.Group("/admin")
	admin.Use(middleware.AdminAuth())
	{
		// 用户管理
		admin.GET("/users", controllers.ListUsers)
		admin.PUT("/user/:id/status", controllers.UpdateUserStatus)

		// 留言管理
		admin.PUT("/message/:id/read", controllers.MarkMessageAsRead)
		admin.PUT("/message/:id/reply", controllers.ReplyMessage)

		// 统计数据API
		admin.GET("/statistics/users", controllers.GetUserStatistics)
		admin.GET("/statistics/modules", controllers.GetModuleUsageStatistics)
		admin.GET("/statistics/login-logs", controllers.GetUserLoginLogs)

		// 会员订单管理
		admin.GET("/membership/orders/pending", membershipController.GetAllPendingOrders)
		admin.POST("/membership/orders/:id/approve", membershipController.ApproveOrder)
		admin.POST("/membership/orders/:id/reject", membershipController.RejectOrder)
	}
}
