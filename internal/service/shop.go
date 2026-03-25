package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"strconv"

	"review-platform/internal/model"
	"review-platform/internal/repository"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	shopCacheTTL     = 10 * time.Minute
	shopNullCacheTTL = 2 * time.Minute
	shopGeoKey = "geo:shop"
)

type ShopService struct {
	repo *repository.ShopRepository
	rdb  *redis.Client
}

type NearbyShop struct {
	Shop     model.Shop
	Distance float64
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

func (s *ShopService) LoadShopGeoData() error {
	ctx := context.Background()

	shops, err := s.repo.ListAll()
	if err != nil {
		return err
	}

	if len(shops) == 0 {
		return nil
	}

	locations := make([]*redis.GeoLocation, 0, len(shops))
	for _, shop := range shops {
		locations = append(locations, &redis.GeoLocation{
			Name:      fmt.Sprintf("%d", shop.ID),
			Longitude: shop.Lng,
			Latitude:  shop.Lat,
		})
	}

	return s.rdb.GeoAdd(ctx, shopGeoKey, locations...).Err()
}

func (s *ShopService) NearbyShops(lng, lat float64, radius float64, page, pageSize int) ([]NearbyShop, int, error) {
	ctx := context.Background()

	end := page * pageSize
	results, err := s.rdb.GeoSearchLocation(ctx, shopGeoKey, &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Longitude:  lng,
			Latitude:   lat,
			Radius:     radius,
			RadiusUnit: "m",
			Sort:       "ASC",
			Count:      end,
		},
		WithDist: true,
	}).Result()
	if err != nil {
		return nil, 0, err
	}

	total := len(results)
	start := (page - 1) * pageSize
	if start >= total {
		return []NearbyShop{}, total, nil
	}

	pageResults := results[start:]
	if len(pageResults) > pageSize {
		pageResults = pageResults[:pageSize]
	}

	ids := make([]int64, 0, len(pageResults))
	distanceMap := make(map[int64]float64, len(pageResults))
	orderMap := make(map[int64]int, len(pageResults))

	for i, item := range pageResults {
		id, err := strconv.ParseInt(item.Name, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
		distanceMap[id] = item.Dist
		orderMap[id] = i
	}

	if len(ids) == 0 {
		return []NearbyShop{}, total, nil
	}

	shops, err := s.repo.ListByIDs(ids)
	if err != nil {
		return nil, 0, err
	}

	ordered := make([]NearbyShop, len(ids))
	for _, shop := range shops {
		idx, ok := orderMap[shop.ID]
		if !ok {
			continue
		}
		ordered[idx] = NearbyShop{
			Shop:     shop,
			Distance: distanceMap[shop.ID],
		}
	}

	return ordered, total, nil
}

func shopCacheKey(id int64) string {
	return fmt.Sprintf("cache:shop:%d", id)
}

func shopNullCacheKey(id int64) string {
	return fmt.Sprintf("cache:shop:null:%d", id)
}