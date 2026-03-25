package api

import (
	"errors"
	"review-platform/internal/service"
	"review-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) SendCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 40001, "invalid request body")
		return
	}

	if err := h.svc.SendCode(req.Phone); err != nil {
		if errors.Is(err, service.ErrInvalidPhone) {
			response.Error(c, 40001, "invalid phone")
			return
		}
		response.Error(c, 50001, "failed to send code")
		return
	}

	response.Success(c, gin.H{
		"message": "verification code sent",
		"code":    "123456", // 仅开发阶段返回，后续应删除
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 40001, "invalid request body")
		return
	}

	token, err := h.svc.Login(req.Phone, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPhone):
			response.Error(c, 40001, "invalid phone")
			return
		case errors.Is(err, service.ErrInvalidCode):
			response.Error(c, 40001, "invalid verification code")
			return
		default:
			response.Error(c, 50001, "failed to login")
			return
		}
	}

	response.Success(c, gin.H{
		"token": token,
	})
}