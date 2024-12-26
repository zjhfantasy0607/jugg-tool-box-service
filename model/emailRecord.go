package model

import (
	"time"

	"gorm.io/gorm"
)

type EmailRecord struct {
	ID      uint   `gorm:"primaryKey"`
	Email   string `gorm:"type:varchar(255);Index;charset:utf8mb4;not null;"`
	Subject string `gorm:"type:varchar(255);charset:utf8mb4;not null;default:''"`
	Body    string `gorm:"type:varchar(255);charset:utf8mb4;not null;default:''"`
	Code    string `gorm:"type:varchar(255);charset:utf8mb4;not null;default:''"`

	CreatedAt time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
