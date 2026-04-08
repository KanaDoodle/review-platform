package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"review-platform/internal/model"
	"review-platform/internal/repository"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	shopCacheTTL     = 10 * time.Minute
	shopNullCacheTTL = 2 * time.Minute
	shopGeoAllKey    = "geo:shop:All"
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

	allLocations := make([]*redis.GeoLocation, 0, len(shops))
	categoryLocations := make(map[int64][]*redis.GeoLocation)

	for _, shop := range shops {
		loc := &redis.GeoLocation{
			Name:      fmt.Sprintf("%d", shop.ID),
			Longitude: shop.Lng,
			Latitude:  shop.Lat,
		}

		allLocations = append(allLocations, loc)
		categoryLocations[shop.CategoryID] = append(categoryLocations[shop.CategoryID], loc)
	}

	// 先清理旧 GEO 数据，避免重复加载
	keys := []string{shopGeoAllKey}
	for categoryID := range categoryLocations {
		keys = append(keys, shopGeoCategoryKey(categoryID))
	}
	if len(keys) > 0 {
		_ = s.rdb.Del(ctx, keys...).Err()
	}

	if err := s.rdb.GeoAdd(ctx, shopGeoAllKey, allLocations...).Err(); err != nil {
		return err
	}

	for categoryID, locations := range categoryLocations {
		if len(locations) == 0 {
			continue
		}
		if err := s.rdb.GeoAdd(ctx, shopGeoCategoryKey(categoryID), locations...).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (s *ShopService) NearbyShops(categoryID int64, lng, lat float64, radius float64, page, pageSize int) ([]NearbyShop, int, error) {
	ctx := context.Background()

	geoKey := shopGeoAllKey
	if categoryID > 0 {
		geoKey = shopGeoCategoryKey(categoryID)
	}

	end := page * pageSize
	results, err := s.rdb.GeoSearchLocation(ctx, geoKey, &redis.GeoSearchLocationQuery{
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

func shopGeoCategoryKey(categoryID int64) string {
	return fmt.Sprintf("geo:shop:category:%d", categoryID)
}

func (s *ShopService) UpdateShop(id int64, name, address string) error {
	ctx := context.Background()

	cacheKey := fmt.Sprintf("cache:shop:%d", id)

	// 第一次删除缓存
	_ = s.rdb.Del(ctx, cacheKey).Err()

	// 更新数据库
	err := s.repo.UpdateByID(id, map[string]interface{}{
		"name":    name,
		"address": address,
	})

	if err != nil {
		return err
	}

	// 延迟双删
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = s.rdb.Del(context.Background(), cacheKey).Err()
	}()

	return nil
}
