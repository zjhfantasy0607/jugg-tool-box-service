package main

import (
	"fmt"
	"io"
	"os"

	"jugg-tool-box-service/common"
	"jugg-tool-box-service/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	// 初始化配置信息
	initConfig()

	// 初始化数据库链接
	db := common.InitDB()

	// 延迟关闭 Mysql 数据库链接
	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	} else {
		panic("failed to get *sql.DB: " + err.Error())
	}

	// 初始化 nsq producer
	common.InitProducer()
	defer common.Producer.Stop() // 延迟停止 producer

	// 初始化 redis
	common.InitRDB()

	runGin()
}

func runGin() {
	// 创建记录gin日志的文件
	logFile, _ := os.Create("./log/gin.log")
	gin.DefaultWriter = io.MultiWriter(logFile)

	r := gin.Default()

	domain := os.Getenv("MY_DOMAIN")
	// 配置 CORS 中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://127.0.0.1:3000",
			"http://localhost:3000",
			fmt.Sprintf("http://%s:3000", domain),
			fmt.Sprintf("http://%s", domain),
		},
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "DELETE", "PUT", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
	}))

	// 初始化路由
	r = routes.CollectRoute(r)

	// 启动 gin 服务
	port := viper.GetString("server.port")
	if port != "" {
		panic(r.Run(":" + port))
	} else {
		// 默认监听 8080 端口
		panic(r.Run())
	}
}

func initConfig() {
	// 获取环境变量的值
	env := os.Getenv("JUGG_TOOL_BOX_SERVICE_ENV")

	// 应用环境判断
	fileName := "application"
	if env == "production" {
		// 根据当前环境读取不同的配置文件
		fileName = fileName + ".production"
		// 设置 gin 为生产模式
		gin.SetMode(gin.ReleaseMode)
	}

	viper.AddConfigPath("./config")
	viper.SetConfigName(fileName)
	viper.SetConfigType("yml")
	err := viper.ReadInConfig()

	if err != nil {
		panic("config file load fail")
	}
}
