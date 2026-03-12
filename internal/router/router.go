package router

import (
	"net/http"
	"review-platform/internal/api"
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
	shopSvc := service.NewShopService(shopRepo)
	shopHandler := api.NewShopHandler(shopSvc)

	reviewRepo := repository.NewReviewRepository(app.DB)
	reviewSvc := service.NewReviewService(reviewRepo, shopRepo)
	reviewHandler := api.NewReviewHandler(reviewSvc)


	v1 := r.Group("/api/v1")
	{
		v1.GET("/categories", categoryHandler.List)
		v1.GET("/shops", shopHandler.List)
		v1.GET("/shops/:id", shopHandler.GetByID)
		v1.GET("/shops/:id/reviews", reviewHandler.ListByShopID)

		v1.POST("/reviews", reviewHandler.Create)
	}

	cleanup := func() {}

	return r, cleanup, nil
}