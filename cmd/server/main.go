package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"net/http"
	"os/signal"
	"review-platform/config"
	"review-platform/internal/router"
	"review-platform/pkg/logger"
	"review-platform/internal/repository"
	"syscall"
	"time"
)

func main() {
	log := logger.New()

	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatal("failed to load config", logger.Err(err))
	}

	db, err := repository.NewMySQL(cfg.MySQL.DSN)
	if err != nil {
		log.Fatal("failed to connect to MySQL", logger.Err(err))
	}

	rdb := repository.NewRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err := repository.PingRedis(rdb); err != nil {
		log.Fatal("failed to connect to Redis", logger.Err(err))
	}

	app := &router.App{
		Config: cfg,
		Log:    log,
		DB:     db,
		RDB:    rdb,
	}


	r, cleanup, err := router.NewRouter(app)
	if err != nil {
		log.Fatal("failed to init router", logger.Err(err))
	}
	defer cleanup()

	srv := router.NewHTTPServer(app,r)

	go func() {
		log.Info("server starting", logger.String("addr", fmt.Sprintf(":%d", cfg.App.Port)))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server failed", logger.Err(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server shutdown failed", logger.Err(err))
	} else {
		fmt.Println("server exited")
	}
}