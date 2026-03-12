package repository

import (
	"review-platform/internal/model"

	"gorm.io/gorm"
)

type ShopCategoryRepository struct {
	db *gorm.DB
}

func NewShopCategoryRepository(db *gorm.DB) *ShopCategoryRepository {
	return &ShopCategoryRepository{db: db}
}

func (r *ShopCategoryRepository) List() ([]model.ShopCategory, error) {
	var categories []model.ShopCategory
	err := r.db.Order("sort ASC,  id ASC").Find(&categories).Error
	return categories, err
}

