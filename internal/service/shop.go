package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"review-platform/internal/model"
	"review-platform/internal/repository"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	shopCacheTTL     = 10 * time.Minute
	shopNullCacheTTL = 2 * time.Minute
)

type ShopService struct {
	repo *repository.ShopRepository
	rdb  *redis.Client
}

func NewShopService(repo *repository.ShopRepository, rdb *redis.Client) *ShopService {
	return &ShopService{
		repo: repo,
		rdb:  rdb,
	}
}

func (s *ShopService) ListShops(categoryID int64, page, pageSize int) ([]model.Shop, int64, error) {
	return s.repo.List(categoryID, page, pageSize)
}

func (s *ShopService) GetShopByID(id int64) (*model.Shop, error) {
	ctx := context.Background()

	cacheKey := shopCacheKey(id)
	nullKey := shopNullCacheKey(id)

	// 1. 先查正常缓存
	cached, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var shop model.Shop
		if unmarshalErr := json.Unmarshal([]byte(cached), &shop); unmarshalErr == nil {
			return &shop, nil
		}
		// 如果缓存数据损坏，继续走数据库查询
	}

	// 2. 再查空值缓存
	exists, err := s.rdb.Exists(ctx, nullKey).Result()
	if err == nil && exists > 0 {
		return nil, ErrShopNotFound
	}

	// 3. 查数据库
	shop, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = s.rdb.Set(ctx, nullKey, "1", shopNullCacheTTL).Err()
			return nil, ErrShopNotFound
		}
		return nil, err
	}

	// 4. 回填缓存
	data, err := json.Marshal(shop)
	if err == nil {
		_ = s.rdb.Set(ctx, cacheKey, data, shopCacheTTL).Err()
	}

	return shop, nil
}

func shopCacheKey(id int64) string {
	return fmt.Sprintf("cache:shop:%d", id)
}

func shopNullCacheKey(id int64) string {
	return fmt.Sprintf("cache:shop:null:%d", id)
}