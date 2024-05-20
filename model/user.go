package model

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          uint         `gorm:"primaryKey"`
	Name        string       `gorm:"type:varchar(255); charset:utf8mb4; not null;"`
	Telephone   string       `gorm:"type:varchar(255); charset:utf8mb4; not null;"`
	Password    string       `gorm:"type:varchar(255); charset:utf8mb4; not null;"`
	ActivatedAt sql.NullTime // Uses sql.NullTime for nullable time fields
	CreatedAt   time.Time    // 创建时间（由GORM自动管理）
	UpdatedAt   time.Time    // 最后一次更新时间（由GORM自动管理）
}

func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	// 检查 ActivatedAt 是否为空，如果为空，则设置为当前时间
	if !user.ActivatedAt.Valid {
		now := time.Now()
		user.ActivatedAt.Time = now
		user.ActivatedAt.Valid = true
	}
	return nil
}
