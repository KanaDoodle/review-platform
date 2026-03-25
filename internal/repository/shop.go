package repository

import (
	"fmt"
	"review-platform/internal/model"

	"gorm.io/gorm"
)

type ShopRepository struct {
	db *gorm.DB
}

func NewShopRepository(db *gorm.DB) *ShopRepository {
	return &ShopRepository{db: db}
}

func (r *ShopRepository) List(categoryID int64) ([]model.Shop, error) {
	var shops []model.Shop

	tx := r.db.Order("score DESC, id ASC")
	if categoryID > 0 {
		tx = tx.Where("category_id = ?", categoryID)
	}

	err := tx.Find(&shops).Error
	return shops, err
}

func (r *ShopRepository) GetByID(id int64) (*model.Shop, error) {
	fmt.Println("query shop from mysql, id =", id)

	var shop model.Shop
	err := r.db.First(&shop, id).Error
	if err != nil {
		return nil, err
	}
	return &shop, nil
}