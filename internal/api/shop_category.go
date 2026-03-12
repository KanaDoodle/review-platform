package api

import (
	"review-platform/internal/service"
	"review-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

type ShopCategoryHandler struct {
	svc *service.ShopCategoryService
}

func NewShopCategoryHandler(svc *service.ShopCategoryService) *ShopCategoryHandler {
	return &ShopCategoryHandler{svc: svc}
}

func (h *ShopCategoryHandler) List(c *gin.Context) {
	categories, err := h.svc.ListCategories()
	if err != nil {
		response.Error(c, 50001, "failed to list categories")
		return
	}

	response.Success(c, categories)
}