package api

import (
	"errors"
	"review-platform/internal/middleware"
	"review-platform/internal/service"
	"review-platform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type VoucherHandler struct {
	svc *service.VoucherService
}

func NewVoucherHandler(svc *service.VoucherService) *VoucherHandler {
	return &VoucherHandler{svc: svc}
}

func (h *VoucherHandler) Seckill(c *gin.Context) {
	idStr := c.Param("id")
	voucherID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || voucherID <= 0 {
		response.Error(c, 40001, "invalid voucher id")
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

	err = h.svc.Seckill(userID, voucherID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrVoucherNotFound):
			response.Error(c, 40401, "voucher not found")
		case errors.Is(err, service.ErrVoucherNotStarted):
			response.Error(c, 40001, "voucher not started")
		case errors.Is(err, service.ErrVoucherEnded):
			response.Error(c, 40001, "voucher ended")
		case errors.Is(err, service.ErrVoucherSoldOut):
			response.Error(c, 40001, "voucher sold out")
		case errors.Is(err, service.ErrDuplicateVoucherOrder):
			response.Error(c, 40001, "duplicate order is not allowed")
		default:
			response.Error(c, 50001, "failed to seckill voucher")
		}
		return
	}

	response.Success(c, gin.H{
		"message": "seckill request accepted.",
	})
}
