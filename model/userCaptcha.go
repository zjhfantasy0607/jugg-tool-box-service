package model

import (
	"time"

	"gorm.io/gorm"
)

type UserCaptcha struct {
	ID        uint           `gorm:"primaryKey"`
	Email     string         `gorm:"type:varchar(255);uniqueIndex;charset:utf8mb4; not null;"`
	PuzzleX   int            `gorm:"default:-1"` // 默认拼图数据为 -1
	IsChecked int            `gorm:"type:tinyint;default:0;not null;"`
	CreatedAt time.Time      // 创建时间（由GORM自动管理）
	UpdatedAt time.Time      // 最后一次更新时间（由GORM自动管理）
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
