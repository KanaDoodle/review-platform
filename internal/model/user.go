package model

import "time"

type User struct {
	ID           int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Phone        string    `json:"phone" gorm:"column:phone"`
	Nickname     string    `json:"nickname" gorm:"column:nickname"`
	PasswordHash string    `json:"-" gorm:"column:password_hash"`//-不把值传递给前端
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "user"
}