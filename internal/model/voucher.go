package model

import "time"

type Voucher struct {
	ID        int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ShopID    int64     `json:"shop_id" gorm:"column:shop_id"`
	Stock     int       `json:"stock" gorm:"column:stock"`
	BeginTime time.Time `json:"begin_time" gorm:"column:begin_time"`
	EndTime   time.Time `json:"end_time" gorm:"column:end_time"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (Voucher) TableName() string {
	return "voucher"
}