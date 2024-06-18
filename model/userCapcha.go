package model

import (
	"time"

	"gorm.io/gorm"
)

type UserCapcha struct {
	ID        uint           `gorm:"primaryKey"`
	UID       string         `gorm:"type:varchar(255);uniqueIndex;default:'';not null;"` // 添加 UID 字段，并设置唯一索引和默认值
	UserName  string         `gorm:"type:varchar(255); charset:utf8mb4; not null;"`
	PuzzleX   int            `gorm:"default:0"`                                          // 默认给0的积分字段
	Token     string         `gorm:"type:varchar(255);uniqueIndex;default:'';not null;"` // 添加 UID 字段，并设置唯一索引和默认值                                    // 默认给0的积分字段
	CreatedAt time.Time      // 创建时间（由GORM自动管理）
	UpdatedAt time.Time      // 最后一次更新时间（由GORM自动管理）
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
