package model

import (
	"time"

	"gorm.io/gorm"
)

type Seo struct {
	ID          uint           `gorm:"primaryKey"`
	Url         string         `gorm:"type:varchar(500);uniqueIndex;charset:utf8mb4;default:'';not null;"` // 唯一索引
	Title       string         `gorm:"type:varchar(255);charset:utf8mb4;default:'';not null;"`
	Keywords    string         `gorm:"type:varchar(255);charset:utf8mb4;default:'';not null;"` // SEO关键字
	Description string         `gorm:"type:varchar(500);charset:utf8mb4;default:'';not null;"` // SEO描述
	CreatedAt   time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type SeoDto struct {
	ID          uint   `json:"id"`
	Url         string `json:"url"`
	Title       string `json:"title"`
	Keywords    string `json:"keywords"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (s Seo) ConvertDto() SeoDto {
	return SeoDto{
		ID:          s.ID,
		Url:         s.Url,
		Title:       s.Title,
		Keywords:    s.Keywords,
		Description: s.Description,
		CreatedAt:   s.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   s.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
