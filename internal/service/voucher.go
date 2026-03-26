package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"review-platform/internal/model"
	"review-platform/internal/repository"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var seckillLuaScript = redis.NewScript(`
local stockKey = KEYS[1]
local orderKey = KEYS[2]
local userId = ARGV[1]

local stock = tonumber(redis.call("GET", stockKey))
if not stock or stock <= 0 then
    return 1
end

if redis.call("SISMEMBER", orderKey, userId) == 1 then
    return 2
end

redis.call("DECR", stockKey)
redis.call("SADD", orderKey, userId)
return 0
`)

type VoucherService struct {
	db        *gorm.DB
	rdb       *redis.Client
	voucherRepo      *repository.VoucherRepository
	voucherOrderRepo *repository.VoucherOrderRepository
}

func NewVoucherService(
	db *gorm.DB,
	rdb *redis.Client,
	voucherRepo *repository.VoucherRepository,
	voucherOrderRepo *repository.VoucherOrderRepository,
) *VoucherService {
	return &VoucherService{
		db:               db,
		rdb:              rdb,
		voucherRepo:      voucherRepo,
		voucherOrderRepo: voucherOrderRepo,
	}
}

func (s *VoucherService) LoadVoucherStockToRedis() error {
	ctx := context.Background()

	var vouchers []model.Voucher
	if err := s.db.Find(&vouchers).Error; err != nil {
		return err
	}

	for _, voucher := range vouchers {
		stockKey := voucherStockKey(voucher.ID)
		if err := s.rdb.Set(ctx, stockKey, voucher.Stock, 0).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (s *VoucherService) Seckill(userID, voucherID int64) error {
	ctx := context.Background()

	voucher, err := s.voucherRepo.GetByID(voucherID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVoucherNotFound
		}
		return err
	}

	now := time.Now()
	if now.Before(voucher.BeginTime) {
		return ErrVoucherNotStarted
	}
	if now.After(voucher.EndTime) {
		return ErrVoucherEnded
	}

	stockKey := voucherStockKey(voucherID)
	orderKey := voucherOrderKey(voucherID)

	res, err := seckillLuaScript.Run(ctx, s.rdb, []string{stockKey, orderKey}, userID).Int()
	if err != nil {
		return err
	}

	switch res {
	case 1:
		return ErrVoucherSoldOut
	case 2:
		return ErrDuplicateVoucherOrder
	}

	return s.createVoucherOrder(userID, voucherID)
}

func (s *VoucherService) createVoucherOrder(userID, voucherID int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		count, err := s.voucherOrderRepo.CountByUserAndVoucher(userID, voucherID)
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrDuplicateVoucherOrder
		}

		result := tx.Model(&model.Voucher{}).
			Where("id = ? AND stock > 0", voucherID).
			Update("stock", gorm.Expr("stock - 1"))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrVoucherSoldOut
		}

		order := &model.VoucherOrder{
			UserID:    userID,
			VoucherID: voucherID,
			Status:    1,
		}

		if err := s.voucherOrderRepo.Create(tx, order); err != nil {
			return err
		}

		return nil
	})
}

func voucherStockKey(voucherID int64) string {
	return fmt.Sprintf("seckill:stock:%d", voucherID)
}

func voucherOrderKey(voucherID int64) string {
	return fmt.Sprintf("seckill:order:%d", voucherID)
}