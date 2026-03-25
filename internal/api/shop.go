package api

import (
	"errors"
	"review-platform/internal/service"
	"review-platform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ShopHandler struct {
	svc *service.ShopService
}

func NewShopHandler(svc *service.ShopService) *ShopHandler {
	return &ShopHandler{svc: svc}
}

func (h *ShopHandler) List(c *gin.Context) {
	categoryIDStr := c.Query("category_id")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	var categoryID int64
	if categoryIDStr != "" {
		id, err := strconv.ParseInt(categoryIDStr, 10, 64)
		if err != nil || id < 0 {
			response.Error(c, 40001, "invalid category_id")
			return
		}
		categoryID = id
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		response.Error(c, 40001, "invalid page")
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 {
		response.Error(c, 40001, "invalid page_size")
		return
	}

	if pageSize > 50 {
		pageSize = 50
	}

	shops, total, err := h.svc.ListShops(categoryID, page, pageSize)
	if err != nil {
		response.Error(c, 50001, "failed to list shops")
		return
	}

	response.Success(c, response.PageData{
		List:     shops,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *ShopHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, 40001, "invalid shop id")
		return
	}

	shop, err := h.svc.GetShopByID(id)
	if err != nil {
		if errors.Is(err, service.ErrShopNotFound) {
			response.Error(c, 40401, "shop not found")
			return
		}
		response.Error(c, 50001, "failed to get shop")
		return
	}

	response.Success(c, shop)
}