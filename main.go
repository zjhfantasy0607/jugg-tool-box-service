package main

import (
	"log"

	"jugg-tool-box-service/common"

	"github.com/gin-gonic/gin"
)

func main() {
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
	panic(r.Run())
}
