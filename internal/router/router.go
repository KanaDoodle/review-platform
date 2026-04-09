package router

import (
	"fmt"
	"net/http"
	"review-platform/internal/api"
	"review-platform/internal/middleware"
	"review-platform/internal/repository"
	"review-platform/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

func NewRouter(app *App) (*gin.Engine, func(), error) {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		app.Log.Info("ping received")
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	categoryRepo := repository.NewShopCategoryRepository(app.DB)
	categorySvc := service.NewShopCategoryService(categoryRepo)
	categoryHandler := api.NewShopCategoryHandler(categorySvc)

	shopRepo := repository.NewShopRepository(app.DB)
	shopSvc := service.NewShopService(shopRepo, app.RDB)
	if err := shopSvc.LoadShopGeoData(); err != nil {
		return nil, nil, err
	}
	shopHandler := api.NewShopHandler(shopSvc)

	reviewRepo := repository.NewReviewRepository(app.DB)
	reviewSvc := service.NewReviewService(reviewRepo, shopRepo)
	reviewHandler := api.NewReviewHandler(reviewSvc)

	voucherRepo := repository.NewVoucherRepository(app.DB)
	voucherOrderRepo := repository.NewVoucherOrderRepository(app.DB)
	voucherSvc := service.NewVoucherService(app.DB, app.RDB, voucherRepo, voucherOrderRepo)
	if err := voucherSvc.LoadVoucherStockToRedis(); err != nil {
		return nil, nil, err
	}
	if err := voucherSvc.InitVoucherOrderStream(); err != nil {
		return nil, nil, err
	}
	voucherSvc.StartVoucherOrderConsumer()

	rateLimiter := service.NewRateLimiter(app.RDB)

	// 这里建议不要再把 rateLimiter 传给 VoucherHandler，避免重复限流
	voucherHandler := api.NewVoucherHandler(voucherSvc)

	userRepo := repository.NewUserRepository(app.DB)
	authSvc := service.NewAuthService(
		userRepo,
		app.RDB,
		app.Config.JWT.Secret,
		app.Config.JWT.ExpireHours,
	)
	authHandler := api.NewAuthHandler(authSvc)

	v1 := r.Group("/api/v1")
	{
		// 登录验证码：按 IP 限流
		sendCodeGroup := v1.Group("/auth")
		sendCodeGroup.Use(middleware.RateLimit(
			rateLimiter,
			func(c *gin.Context) string {
				return "send_code:" + c.ClientIP()
			},
			5,
			time.Minute,
		))
		{
			sendCodeGroup.POST("/send-code", authHandler.SendCode)
		}

		v1.POST("/auth/login", authHandler.Login)

		v1.GET("/categories", categoryHandler.List)
		v1.GET("/shops", shopHandler.List)
		v1.GET("/shops/nearby", shopHandler.Nearby)
		v1.GET("/shops/:id", shopHandler.GetByID)
		v1.GET("/shops/:id/reviews", reviewHandler.ListByShopID)
		v1.POST("/shops/update", shopHandler.Update)

		// 发布点评：先鉴权，再按用户维度限流
		reviewGroup := v1.Group("")
		reviewGroup.Use(middleware.Auth(app.Config))
		reviewGroup.Use(middleware.RateLimit(
			rateLimiter,
			func(c *gin.Context) string {
				userIDVal, ok := c.Get(middleware.CurrentUserIDKey)
				if !ok {
					return "review:anonymous"
				}
				return "review:user:" + toString(userIDVal)
			},
			10,
			time.Minute,
		))
		{
			reviewGroup.POST("/reviews", reviewHandler.Create)
		}

		// 秒杀：先鉴权，再按用户维度限流
		seckillGroup := v1.Group("")
		seckillGroup.Use(middleware.Auth(app.Config))
		seckillGroup.Use(middleware.RateLimit(
			rateLimiter,
			func(c *gin.Context) string {
				userIDVal, ok := c.Get(middleware.CurrentUserIDKey)
				if !ok {
					return "seckill:anonymous"
				}
				return "seckill:user:" + toString(userIDVal)
			},
			5,
			time.Second,
		))
		{
			seckillGroup.POST("/vouchers/seckill/:id", voucherHandler.Seckill)
		}
	}

	cleanup := func() {}

	return r, cleanup, nil
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case int64:
		return fmt.Sprintf("%d", val)
	case int:
		return fmt.Sprintf("%d", val)
	case string:
		return val
	default:
		return "unknown"
	}
}