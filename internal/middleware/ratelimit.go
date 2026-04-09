package middleware

import (
	"context"
	"time"

	"review-platform/internal/service"
	"review-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

func RateLimit(
	limiter *service.RateLimiter,
	keyFunc func(*gin.Context) string,
	limit int,
	window time.Duration,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)

		allowed, err := limiter.Allow(context.Background(), key, limit, window)
		if err != nil {
			response.Error(c, 50001, "rate limiter error")
			c.Abort()
			return
		}

		if !allowed {
			response.Error(c, 42901, "too many requests")
			c.Abort()
			return
		}

		c.Next()
	}
}