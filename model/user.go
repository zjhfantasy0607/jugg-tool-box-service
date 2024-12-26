package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	UID       string `gorm:"type:varchar(255);uniqueIndex;default:'';not null;"` // 添加 UID 字段，并设置唯一索引和默认值
	Name      string `gorm:"type:varchar(255); charset:utf8mb4; not null;"`
	Telephone string `gorm:"type:varchar(255); charset:utf8mb4; not null;"`
	Email     string `gorm:"type:varchar(500); charset:utf8mb4; not null;"`
	Password  string `gorm:"type:varchar(255); charset:utf8mb4; not null;"`
	Points    int    `gorm:"type:int;default:0"`

	CreatedAt time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type UserDto struct {
	UID    string `json:"uid"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Points int    `json:"points"`
}

func (user User) ConvertDto() UserDto {
	return UserDto{
		UID:    user.UID,
		Name:   user.Name,
		Email:  user.Email,
		Points: user.Points,
	}
}
