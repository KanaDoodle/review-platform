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

func (h *ShopHandler) Nearby(c *gin.Context) {
	categoryIDStr := c.Query("category_id")
	lngStr := c.Query("lng")
	latStr := c.Query("lat")
	radiusStr := c.DefaultQuery("radius", "5000")
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

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		response.Error(c, 40001, "invalid lng")
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		response.Error(c, 40001, "invalid lat")
		return
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil || radius <= 0 {
		response.Error(c, 40001, "invalid radius")
		return
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

	items, total, err := h.svc.NearbyShops(categoryID, lng, lat, radius, page, pageSize)
	if err != nil {
		response.Error(c, 50001, "failed to query nearby shops")
		return
	}

	respItems := make([]NearbyShopItem, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, NearbyShopItem{
			ID:          item.Shop.ID,
			Name:        item.Shop.Name,
			CategoryID:  item.Shop.CategoryID,
			Address:     item.Shop.Address,
			Lng:         item.Shop.Lng,
			Lat:         item.Shop.Lat,
			Score:       item.Shop.Score,
			AvgPrice:    item.Shop.AvgPrice,
			Description: item.Shop.Description,
			Distance:    item.Distance,
		})
	}

	response.Success(c, response.PageData{
		List:     respItems,
		Total:    int64(total),
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *ShopHandler) Update(c *gin.Context) {
	var req struct {
		ID      int64  `json:"id"`
		Name    string `json:"name"`
		Address string `json:"address"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 40001, "invalid params")
		return
	}

	if req.ID <= 0 {
		response.Error(c, 40001, "invalid id")
		return
	}

	err := h.svc.UpdateShop(req.ID, req.Name, req.Address)
	if err != nil {
		response.Error(c, 50001, "update failed")
		return
	}

	response.Success(c, gin.H{
		"message": "update success",
	})
}