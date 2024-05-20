package common

import (
	"fmt"
	"jugg-tool-box-service/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDB() *gorm.DB {
	host := "localhost"
	port := "3306"
	database := "jugg-tool-box"
	username := "root"
	password := "root"
	charset := "utf8mb4"

	args := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True",
		username,
		password,
		host,
		port,
		database,
		charset)

	var err error
	db, err = gorm.Open(mysql.Open(args), &gorm.Config{})

	if err != nil {
		panic("failed to connect database, err: " + err.Error())
	}

	// 建表
	db.AutoMigrate(&model.User{})

	return db
}

func GetDB() *gorm.DB {
	return db
}
