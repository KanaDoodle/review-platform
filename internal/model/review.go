package model

import "time"

type Review struct {
	ID        int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64     `json:"user_id" gorm:"column:user_id"`
	ShopID    int64     `json:"shop_id" gorm:"column:shop_id"`
	Content   string    `json:"content" gorm:"column:content"`
	LikeCount int       `json:"like_count" gorm:"column:like_count"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (Review) TableName() string {
	return "review"
}