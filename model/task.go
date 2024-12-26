package model

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	ID         uint      `gorm:"primaryKey"`
	TaskId     string    `gorm:"type:varchar(255);uniqueIndex;charset:utf8mb4;not null;"`
	UID        string    `gorm:"type:varchar(255);Index;not null;"` // 添加 UID 字段，并设置唯一索引和默认值
	UsedPoints int       `gorm:"type:int;default:0"`
	Tool       string    `gorm:"type:char(10);default:'';comment:'resize,'"`
	Status     string    `gorm:"type:char(10);default:'pending';comment:'pending, producing, success or failed'"`
	Params     string    `gorm:"type:json;not null"`
	Source     string    `gorm:"type:json;not null"`
	Output     string    `gorm:"type:json;not null"`
	StartTime  time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	EndTime    time.Time `gorm:"type:datetime;default:null"`
	UsedTime   int64     `gorm:"type:int;default:0;comment:任务执行消耗的时间，使用单位ms"`

	CreatedAt time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type TaskSlice []Task

type TaskDto struct {
	ID         uint   `json:"id"`
	TaskId     string `json:"task_id"`
	UID        string `json:"uid"`
	UsedPoints int    `json:"used_points"`
	Tool       string `json:"tool"`
	ToolTitle  string `json:"tool_title"`
	ToolUrl    string `json:"tool_url"`
	Status     string `json:"status"`
	Params     string `json:"params"`
	Source     string `json:"source"`
	Output     string `json:"output"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	UsedTime   int64  `json:"used_time"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func (t Task) ConvertDto() TaskDto {
	return TaskDto{
		ID:         t.ID,
		TaskId:     t.TaskId,
		UID:        t.UID,
		UsedPoints: t.UsedPoints,
		Tool:       t.Tool,
		Status:     t.Status,
		Params:     t.Params,
		Source:     t.Source,
		Output:     t.Output,
		StartTime:  t.StartTime.Format("2006-01-02 15:04:05"),
		EndTime:    t.EndTime.Format("2006-01-02 15:04:05"),
		UsedTime:   t.UsedTime,
		CreatedAt:  t.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  t.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (t TaskSlice) ConvertDto() []TaskDto {
	tasksDto := make([]TaskDto, 0)
	for _, task := range t {
		tasksDto = append(tasksDto, task.ConvertDto())
	}

	return tasksDto
}
