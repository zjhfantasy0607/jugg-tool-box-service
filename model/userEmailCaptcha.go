package model

import (
	"time"

	"gorm.io/gorm"
)

type UserEmailCaptcha struct {
	ID        uint   `gorm:"primaryKey"`
	Email     string `gorm:"type:varchar(255);charset:utf8mb4;not null;"`
	RandStr   string `gorm:"type:varchar(255);charset:utf8mb4;not null;default:''"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
