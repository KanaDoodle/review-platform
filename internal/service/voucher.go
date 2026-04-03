package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"strconv"
	"strings"

	"review-platform/internal/model"
	"review-platform/internal/repository"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	voucherOrderStreamKey = "stream.orders"
	voucherOrderGroupName = "g1"
	voucherOrderConsumer  = "c1"
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

	// Redis 预扣成功后，异步投递订单消息
	if err := s.enqueueVoucherOrder(ctx, userID, voucherID); err != nil {
		return err
	}

	return nil
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

func (s *VoucherService) enqueueVoucherOrder(ctx context.Context, userID, voucherID int64) error {
	_, err := s.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: voucherOrderStreamKey,
		Values: map[string]interface{}{
			"user_id":    userID,
			"voucher_id": voucherID,
		},
	}).Result()
	return err
}

func (s *VoucherService) StartVoucherOrderConsumer() {
	go func() {
		ctx := context.Background()

		for {
			streams, err := s.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    voucherOrderGroupName,
				Consumer: voucherOrderConsumer,
				Streams:  []string{voucherOrderStreamKey, ">"},
				Count:    1,
				Block:    2 * time.Second,
			}).Result()
			if err != nil {
				if err == redis.Nil {
					s.handlePendingList(ctx)
					continue
				}
				time.Sleep(1 * time.Second)
				continue
			}

			for _, stream := range streams {
				for _, msg := range stream.Messages {
					if err := s.handleVoucherOrderMessage(ctx, msg); err != nil {
						continue
					}
				}
			}
		}
	}()
}

func parseVoucherOrderMessage(values map[string]interface{}) (int64, int64, error) {
	userIDRaw, ok := values["user_id"]
	if !ok {
		return 0, 0, errors.New("missing user_id")
	}

	voucherIDRaw, ok := values["voucher_id"]
	if !ok {
		return 0, 0, errors.New("missing voucher_id")
	}

	userID, err := parseStreamInt64(userIDRaw)
	if err != nil {
		return 0, 0, err
	}

	voucherID, err := parseStreamInt64(voucherIDRaw)
	if err != nil {
		return 0, 0, err
	}

	return userID, voucherID, nil
}

func parseStreamInt64(v interface{}) (int64, error) {
	switch val := v.(type) {
	case string:
		return strconv.ParseInt(val, 10, 64)
	case []byte:
		return strconv.ParseInt(string(val), 10, 64)
	case int64:
		return val, nil
	case int:
		return int64(val), nil
	default:
		return 0, errors.New("invalid int64 value")
	}
}

func (s *VoucherService) InitVoucherOrderStream() error {
	ctx := context.Background()

	err := s.rdb.XGroupCreateMkStream(ctx, voucherOrderStreamKey, voucherOrderGroupName, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}

	return nil
}

func (s *VoucherService) handleVoucherOrderMessage(ctx context.Context, msg redis.XMessage) error {
	fmt.Println("Processing message:",msg.ID, msg.Values) //TempLog
	
	userID, voucherID, err := parseVoucherOrderMessage(msg.Values)
	if err != nil {
		return err
	}

	if err := s.createVoucherOrder(userID, voucherID); err != nil {
		return err
	}

	fmt.Println("Order created successfully for user:", userID, "voucher:", voucherID) //TempLog
	return s.rdb.XAck(ctx, voucherOrderStreamKey, voucherOrderGroupName, msg.ID).Err()
}

func (s *VoucherService) handlePendingList(ctx context.Context) {
	for {
		streams, err := s.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    voucherOrderGroupName,
			Consumer: voucherOrderConsumer,
			Streams:  []string{voucherOrderStreamKey, "0"},
			Count:    1,
			Block:    1 * time.Second,
		}).Result()
		if err != nil {
			return
		}

		if len(streams) == 0 || len(streams[0].Messages) == 0 {
			return
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				if err := s.handleVoucherOrderMessage(ctx, msg); err != nil {
					time.Sleep(500 * time.Millisecond)
					continue
				}
			}
		}
	}
}