package model

import (
	"time"

	"gorm.io/gorm"
)

type Tool struct {
	ID          uint `gorm:"primaryKey"`
	Pid         uint
	Title       string `gorm:"type:varchar(255);Index;charset:utf8mb4;default:'';not null;"`
	Description string `gorm:"type:varchar(500);charset:utf8mb4;default:'';not null;"`
	Icon        string `gorm:"type:text;charset:utf8mb4;"`
	Url         string `gorm:"type:varchar(500);charset:utf8mb4;default:'';not null;"`
	Orders      int    `gorm:"default:0;not null;"`
	Tool        string `gorm:"type:char(10);default:'';"`

	CreatedAt time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Children  ToolSlice      `gorm:"-"`
}

type ToolSlice []*Tool

type ToolDto struct {
	ID          uint      `json:"id"`
	Pid         uint      `json:"pid"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Url         string    `json:"url"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
	Children    []ToolDto `json:"children,omitempty"` // 使用指针以便在JSON中省略空值
	Orders      int
}

func (tools ToolSlice) BuildTree(basePid uint) ToolSlice {
	toolMap := make(map[uint]*Tool)
	for _, tool := range tools {
		toolMap[tool.ID] = tool
	}

	var tree ToolSlice
	for _, tool := range tools {
		if tool.Pid != basePid {
			if parent, ok := toolMap[tool.Pid]; ok {
				parent.Children = append(parent.Children, tool)
			}
		} else {
			tree = append(tree, tool)
		}
	}

	return tree
}

func (tools ToolSlice) ConvertDto() []ToolDto {
	toolsDto := make([]ToolDto, 0)

	for _, tool := range tools {
		toolsDto = append(toolsDto, ToolDto{
			ID:          tool.ID,
			Pid:         tool.Pid,
			Title:       tool.Title,
			Description: tool.Description,
			Icon:        tool.Icon,
			Url:         tool.Url,
			CreatedAt:   tool.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   tool.UpdatedAt.Format("2006-01-02 15:04:05"),
			Children:    tool.Children.ConvertDto(),
			Orders:      tool.Orders,
		})
	}

	return toolsDto
}
