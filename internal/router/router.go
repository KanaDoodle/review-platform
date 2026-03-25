package router

import (
	"net/http"
	"review-platform/internal/api"
	"review-platform/internal/middleware"
	"review-platform/internal/repository"
	"review-platform/internal/service"

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
	shopHandler := api.NewShopHandler(shopSvc)

	reviewRepo := repository.NewReviewRepository(app.DB)
	reviewSvc := service.NewReviewService(reviewRepo, shopRepo)
	reviewHandler := api.NewReviewHandler(reviewSvc)

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
		v1.POST("/auth/send-code", authHandler.SendCode)
		v1.POST("/auth/login", authHandler.Login)

		v1.GET("/categories", categoryHandler.List)
		v1.GET("/shops", shopHandler.List)
		v1.GET("/shops/:id", shopHandler.GetByID)
		v1.GET("/shops/:id/reviews", reviewHandler.ListByShopID)

		authGroup := v1.Group("")
		authGroup.Use(middleware.Auth(app.Config))
		{
			authGroup.POST("/reviews", reviewHandler.Create)
		}
	}

	cleanup := func() {}

	return r, cleanup, nil
}