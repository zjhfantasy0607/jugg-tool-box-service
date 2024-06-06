package main

import (
	"log"
	"os"

	"jugg-tool-box-service/common"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	/***** 初始化配置信息 *****/
	InitConfig()

	/***** 初始化数据库链接 *****/
	db := common.InitDB()

	// 获取底层的 *sql.DB 对象
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("failed to get *sql.DB: ", err)
	}

	// 延迟关闭 Mysql 数据库链接
	defer sqlDB.Close()

	/***** 初始化路由 *****/
	r := CollectRoute(gin.Default())

	/***** 启动http服务 *****/
	port := viper.GetString("server.port")
	if port != "" {
		panic(r.Run(":" + port))
	} else {
		// 默认监听 8080 端口
		panic(r.Run())
	}
}

func InitConfig() {
	workDir, _ := os.Getwd()
	viper.SetConfigName("application")
	viper.SetConfigType("yml")
	viper.AddConfigPath(workDir + "/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic("config file load fail")
	}
}
