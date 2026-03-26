package repository

import (
	"review-platform/internal/model"

	"gorm.io/gorm"
)

type VoucherRepository struct {
	db *gorm.DB
}

func NewVoucherRepository(db *gorm.DB) *VoucherRepository {
	return &VoucherRepository{db: db}
}

func (r *VoucherRepository) GetByID(id int64) (*model.Voucher, error) {
	var voucher model.Voucher
	err := r.db.First(&voucher, id).Error
	if err != nil {
		return nil, err
	}
	return &voucher, nil
}

func (r *VoucherRepository) DecreaseStock(tx *gorm.DB, voucherID int64) error {
	return tx.Model(&model.Voucher{}).Where("id = ? AND stock > 0", voucherID).Update("stock", gorm.Expr("stock - 1")).Error
}
