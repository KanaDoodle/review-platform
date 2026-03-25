package middleware

import (
	"strings"

	"review-platform/config"
	jwtutil "review-platform/pkg/jwt"
	"review-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

const CurrentUserIDKey = "current_user_id"

func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, 40101, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
			response.Error(c, 40101, "invalid authorization header")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwtutil.ParseToken(cfg.JWT.Secret, tokenString)
		if err != nil {
			response.Error(c, 40102, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(CurrentUserIDKey, claims.UserID)
		c.Next()
	}
}