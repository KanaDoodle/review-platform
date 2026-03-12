package router

import (
	"review-platform/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type App struct {
	Config *config.Config
	Log    *zap.Logger
	DB     *gorm.DB
	RDB    *redis.Client
}