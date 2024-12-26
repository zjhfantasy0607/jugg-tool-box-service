package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Point struct {
	ID           uint   `gorm:"primaryKey"`
	TaskId       string `gorm:"type:varchar(255);Index;charset:utf8mb4;"`
	UID          string `gorm:"type:varchar(255);Index;charset:utf8mb4"`
	BeforePoints int    `gorm:"type:int;default:0;"`
	Points       int    `gorm:"type:int;default:0;"`
	Tool         string `gorm:"type:char(10);default:'';comment:'resize,'"`
	Remark       string `gorm:"type:varchar(255);default:'';"`

	CreatedAt time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type PointDto struct {
	ID           uint   `json:"id"`
	TaskId       string `json:"task_id"`
	UID          string `json:"uid"`
	BeforePoints int    `json:"before_points"`
	Points       int    `json:"points"`
	Tool         string `json:"tool"`
	Remark       string `json:"remark"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type PointSlice []Point

// 默认自带事务版
func (p *Point) AddRecord(db *gorm.DB, user *User) error {
	user.Points += p.Points

	if (user.Points) < 0 {
		return errors.New("don't have enough points")
	}

	// 开始事务
	tx := db.Begin()

	// 修改用户当前积分
	if err := tx.Save(user).Error; err != nil {
		tx.Rollback() // 发生错误时回滚
		return err
	}

	// 存储积分记录
	if err := tx.Create(p).Error; err != nil {
		tx.Rollback() // 发生错误时回滚
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// UpdatePoints 外传事务版
func (p *Point) T_AddRecord(tx *gorm.DB, user *User) error {
	user.Points += p.Points

	if (user.Points) < 0 {
		return errors.New("don't have enough points")
	}

	// 修改用户当前积分
	if err := tx.Save(user).Error; err != nil {
		tx.Rollback() // 发生错误时回滚
		return err
	}

	// 存储积分记录
	if err := tx.Create(p).Error; err != nil {
		tx.Rollback() // 发生错误时回滚
		return err
	}

	return nil
}

func (p *Point) ConvertDto() PointDto {
	return PointDto{
		ID:           p.ID,
		TaskId:       p.TaskId,
		UID:          p.UID,
		BeforePoints: p.BeforePoints,
		Points:       p.Points,
		Tool:         p.Tool,
		Remark:       p.Remark,
		CreatedAt:    p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (p PointSlice) ConvertDto() []PointDto {
	pointsDto := make([]PointDto, 0)
	for _, points := range p {
		pointsDto = append(pointsDto, points.ConvertDto())
	}
	return pointsDto
}
