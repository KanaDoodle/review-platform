package model

import "time"

type VoucherOrder struct {
	ID        int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64     `json:"user_id" gorm:"column:user_id"`
	VoucherID int64     `json:"voucher_id" gorm:"column:voucher_id"`
	Status    int       `json:"status" gorm:"column:status"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (VoucherOrder) TableName() string {
	return "voucher_order"
}