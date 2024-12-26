package model

import (
	"time"

	"gorm.io/gorm"
)

type Captcha struct {
	ID        uint   `gorm:"primaryKey"`
	BrowserId string `gorm:"type:char(32);uniqueIndex;charset:utf8mb4;not null;"`
	PuzzleX   int    `gorm:"default:-1"` // 默认拼图数据为 -1
	IsChecked int    `gorm:"type:tinyint;default:0;not null;"`

	CreatedAt time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
