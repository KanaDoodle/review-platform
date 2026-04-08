package repository

import (
	"review-platform/internal/model"

	"gorm.io/gorm"
)

type ShopRepository struct {
	db *gorm.DB
}

func NewShopRepository(db *gorm.DB) *ShopRepository {
	return &ShopRepository{db: db}
}

func (r *ShopRepository) List(categoryID int64, page, pageSize int) ([]model.Shop, int64, error) {
	var (
		shops []model.Shop
		total int64
	)

	tx := r.db.Model(&model.Shop{})
	if categoryID > 0 {
		tx = tx.Where("category_id = ?", categoryID)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	err := tx.
		Order("score DESC, id ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&shops).Error
	if err != nil {
		return nil, 0, err
	}

	return shops, total, nil
}

func (r *ShopRepository) GetByID(id int64) (*model.Shop, error) {
	var shop model.Shop
	err := r.db.First(&shop, id).Error
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

func (r *ShopRepository) ListByIDs(ids []int64) ([]model.Shop, error) {
	var shops []model.Shop
	err := r.db.Where("id IN ?", ids).Find(&shops).Error
	return shops, err
}

func (r *ShopRepository) ListAll() ([]model.Shop, error) {
	var shops []model.Shop
	err := r.db.Order("id ASC").Find(&shops).Error
	return shops, err
}

func (r *ShopRepository) UpdateByID(id int64, updates map[string]interface{}) error {
	return r.db.Model(&model.Shop{}).
		Where("id = ?", id).
		Updates(updates).Error
}