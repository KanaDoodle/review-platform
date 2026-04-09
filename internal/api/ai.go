package api

import (
	"context"
	"strconv"
	"strings"
	"time"

	"review-platform/internal/service"
	"review-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	svc *service.AIService
}

func NewAIHandler(svc *service.AIService) *AIHandler {
	return &AIHandler{svc: svc}
}

func (h *AIHandler) SearchReviews(c *gin.Context) {
	query := strings.TrimSpace(c.Query("query"))
	if query == "" {
		response.Error(c, 40001, "query is required")
		return
	}

	topKStr := c.DefaultQuery("top_k", "5")
	topK, err := strconv.Atoi(topKStr)
	if err != nil || topK <= 0 {
		response.Error(c, 40001, "invalid top_k")
		return
	}
	if topK > 20 {
		topK = 20
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	results, err := h.svc.SearchReviews(ctx, query, topK)
	if err != nil {
		response.Error(c, 50001, err.Error())
		return
	}

	response.Success(c, gin.H{
		"query":   query,
		"top_k":   topK,
		"results": results,
	})
}

func (h *AIHandler) SummarizeShop(c *gin.Context) {
	idStr := c.Param("id")
	shopID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || shopID <= 0 {
		response.Error(c, 40001, "invalid shop id")
		return
	}

	query := strings.TrimSpace(c.Query("query"))
	if query == "" {
		response.Error(c, 40001, "query is required")
		return
	}

	topKStr := c.DefaultQuery("top_k", "3")
	topK, err := strconv.Atoi(topKStr)
	if err != nil || topK <= 0 {
		response.Error(c, 40001, "invalid top_k")
		return
	}
	if topK > 10 {
		topK = 10
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := h.svc.SummarizeShopReviews(ctx, shopID, query, topK)
	if err != nil {
		response.Error(c, 50001, err.Error())
		return
	}

	response.Success(c, result)
}