package api

import (
	"errors"
	"review-platform/internal/middleware"
	"review-platform/internal/service"
	"review-platform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	svc *service.ReviewService
}

func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{svc: svc}
}

func (h *ReviewHandler) ListByShopID(c *gin.Context) {
	idStr := c.Param("id")
	shopID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || shopID <= 0 {
		response.Error(c, 40001, "invalid shop id")
		return
	}

	reviews, err := h.svc.ListByShopID(shopID)
	if err != nil {
		if errors.Is(err, service.ErrShopNotFound) {
			response.Error(c, 40401, "shop not found")
			return
		}
		response.Error(c, 50001, "failed to list reviews")
		return
	}

	response.Success(c, reviews)
}

func (h *ReviewHandler) Create(c *gin.Context) {
	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 40001, "invalid request body")
		return
	}

	userIDVal, ok := c.Get(middleware.CurrentUserIDKey)
	if !ok {
		response.Error(c, 40101, "unauthorized")
		return
	}

	userID, ok := userIDVal.(int64)
	if !ok {
		response.Error(c, 40101, "invalid user context")
		return
	}

	err := h.svc.Create(userID, req.ShopID, req.Content)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidReviewData):
			response.Error(c, 40001, "invalid review data")
			return
		case errors.Is(err, service.ErrShopNotFound):
			response.Error(c, 40401, "shop not found")
			return
		default:
			response.Error(c, 50001, "failed to create review")
			return
		}
	}

	response.Success(c, gin.H{
		"message": "review created",
	})
}