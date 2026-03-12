package service

import (
	"review-platform/internal/model"
	"review-platform/internal/repository"
)

type ShopCategoryService struct {
	repo *repository.ShopCategoryRepository
}

func NewShopCategoryService(repo *repository.ShopCategoryRepository) *ShopCategoryService {
	return &ShopCategoryService{repo: repo}
}

func (s *ShopCategoryService) ListCategories() ([]model.ShopCategory, error) {
	return s.repo.List()
}