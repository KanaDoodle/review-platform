package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	rdb *redis.Client
}

func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{rdb: rdb}
}

// key: 限流维度（比如 user:123）
// limit: 时间窗口内最多请求数
// window: 时间窗口（秒）
func (l *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now().UnixMilli()
	windowStart := now - window.Milliseconds()

	redisKey := fmt.Sprintf("rate_limit:%s", key)

	pipe := l.rdb.TxPipeline()

	// 1. 删除窗口外的旧数据
	pipe.ZRemRangeByScore(ctx, redisKey, "0", fmt.Sprintf("%d", windowStart))

	// 2. 统计当前窗口内请求数
	countCmd := pipe.ZCard(ctx, redisKey)

	// 3. 添加当前请求
	pipe.ZAdd(ctx, redisKey, redis.Z{
		Score:  float64(now),
		Member: fmt.Sprintf("%d", now),
	})

	// 4. 设置过期时间
	pipe.Expire(ctx, redisKey, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count := countCmd.Val()

	if int(count) >= limit {
		return false, nil
	}

	return true, nil
}