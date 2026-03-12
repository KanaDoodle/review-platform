package model

import "time"

type ShopCategory struct {
	ID        int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"column:name"`
	Icon      string    `json:"icon" gorm:"column:icon"`
	Sort      int       `json:"sort" gorm:"column:sort"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (ShopCategory) TableName() string {
	return "shop_category"
}