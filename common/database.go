package common

import (
	"fmt"

	"log"
	"net/url"
	"os"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func InitDB() *gorm.DB {
	host := viper.GetString("database.host")
	port := viper.GetString("database.port")
	database := viper.GetString("database.database")
	username := viper.GetString("database.username")
	password := viper.GetString("database.password")
	charset := viper.GetString("database.charset")
	timezone := viper.GetString("database.timezone")

	// 对时区字符串进行URL编码
	encodedTimezone := url.QueryEscape(timezone)

	args := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=%s",
		username,
		password,
		host,
		port,
		database,
		charset,
		encodedTimezone,
	)

	// 创建 mysql日志文件
	logPath := "./log/gorm.log"
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("failed to open gorm log file: " + err.Error())
	}

	// 创建自定义的日志记录器
	newLogger := logger.New(
		log.New(logFile, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,  // 慢 SQL 阈值
			LogLevel:                  logger.Error, // 日志级别
			IgnoreRecordNotFoundError: true,         // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  false,        // 彩色打印
		},
	)

	// 开启连接
	db, err = gorm.Open(mysql.Open(args), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		panic("failed to connect database, err: " + err.Error())
	}

	// 数据库迁移
	autoMigrate()

	return db
}

func GetDB() *gorm.DB {
	return db
}
