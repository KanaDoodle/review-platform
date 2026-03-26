package repository

import (
	"review-platform/internal/model"

	"gorm.io/gorm"
)

type VoucherOrderRepository struct {
	db *gorm.DB
}

func NewVoucherOrderRepository(db *gorm.DB) *VoucherOrderRepository {
	return &VoucherOrderRepository{db: db}
}

func (r *VoucherOrderRepository) Create(tx *gorm.DB, order *model.VoucherOrder) error {
	return tx.Create(order).Error
}

func (r *VoucherOrderRepository) CountByUserAndVoucher(userID, voucherID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.VoucherOrder{}).Where("user_id = ? AND voucher_id = ?", userID, voucherID).Count(&count).Error
	return count, err
}
