package service

import (
	"errors"
	"review-platform/internal/model"
	"review-platform/internal/repository"

	"gorm.io/gorm"
)

type ShopService struct {
	repo *repository.ShopRepository
}

func NewShopService(repo *repository.ShopRepository) *ShopService {
	return &ShopService{repo: repo}
}

func (s *ShopService) ListShops(categoryID int64) ([]model.Shop, error) {
	return s.repo.List(categoryID)
}

func (s *ShopService) GetShopByID(id int64) (*model.Shop, error) {
	shop, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShopNotFound
		}
		return nil, err
	}
	return shop, nil
}