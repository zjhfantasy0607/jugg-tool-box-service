package common

import (
	"fmt"
	"jugg-tool-box-service/model"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDB() *gorm.DB {
	host := viper.GetString("database.host")
	port := viper.GetString("database.port")
	database := viper.GetString("database.database")
	username := viper.GetString("database.username")
	password := viper.GetString("database.password")
	charset := viper.GetString("database.charset")

	args := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True",
		username,
		password,
		host,
		port,
		database,
		charset,
	)

	var err error
	db, err = gorm.Open(mysql.Open(args), &gorm.Config{})

	if err != nil {
		panic("failed to connect database, err: " + err.Error())
	}

	// 建表
	db.AutoMigrate(
		&model.User{},
		&model.UserCaptcha{},
	)

	return db
}

func GetDB() *gorm.DB {
	return db
}
