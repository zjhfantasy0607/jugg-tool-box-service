package common

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var Rdb *redis.Client
var Ctx = context.Background()

func InitRDB() {
	// 从Viper中读取地址和端口
	host := viper.GetString("redis.host")
	port := viper.GetString("redis.port")
	password := viper.GetString("redis.password")
	addr := fmt.Sprintf("%s:%s", host, port)

	// 创建一个Redis客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,     // Redis服务器地址
		Password: password, // Redis密码（如果没有设置密码则为空）
		DB:       0,        // 使用的数据库编号
	})

	// 连接测试
	_, err := rdb.Ping(Ctx).Result()
	if err != nil {
		panic(err.Error())
	}

	Rdb = rdb
}
