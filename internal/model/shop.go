package model

import "time"

type Shop struct {
	ID          int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"column:name"`
	CategoryID  int64     `json:"category_id" gorm:"column:category_id"`
	Address     string    `json:"address" gorm:"column:address"`
	Lng         float64   `json:"lng" gorm:"column:lng"`
	Lat         float64   `json:"lat" gorm:"column:lat"`
	Score       float64   `json:"score" gorm:"column:score"`
	AvgPrice    int       `json:"avg_price" gorm:"column:avg_price"`
	Description string    `json:"description" gorm:"column:description"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (Shop) TableName() string {
	return "shop"
}