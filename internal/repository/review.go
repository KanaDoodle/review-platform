package repository

import (
	"review-platform/internal/model"

	"gorm.io/gorm"
)

type ReviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) ListByShopID(shopID int64) ([]model.Review, error) {
	var reviews []model.Review
	err := r.db.
		Where("shop_id = ?", shopID).
		Order("created_at DESC, id DESC").
		Find(&reviews).Error
	return reviews, err
}

func (r *ReviewRepository) Create(review *model.Review) error {
	return r.db.Create(review).Error
}