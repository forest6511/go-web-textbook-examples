package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/time/rate"

	"github.com/forest6511/go-web-textbook-examples/ch07-jwt-impl/internal/auth"
	"github.com/forest6511/go-web-textbook-examples/ch07-jwt-impl/internal/db"
	"github.com/forest6511/go-web-textbook-examples/ch07-jwt-impl/internal/handler"
	mw "github.com/forest6511/go-web-textbook-examples/ch07-jwt-impl/internal/middleware"
	"github.com/forest6511/go-web-textbook-examples/ch07-jwt-impl/internal/repository"
	"github.com/forest6511/go-web-textbook-examples/ch07-jwt-impl/internal/router"
	"github.com/forest6511/go-web-textbook-examples/ch07-jwt-impl/internal/usecase"
	appval "github.com/forest6511/go-web-textbook-examples/ch07-jwt-impl/internal/validator"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if _, err := appval.Setup(); err != nil {
		logger.Error("setup validator", "error", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://app:app@localhost:5432/app?sslmode=disable"
	}

	pool, err := db.NewPool(ctx, dsn)
	if err != nil {
		logger.Error("init pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()
	if err := pool.Ping(pingCtx); err != nil {
		logger.Error("ping db", "error", err)
		os.Exit(1)
	}

	// 開発環境向け: 起動時にマイグレーションを流す
	if os.Getenv("RUN_MIGRATIONS") == "true" {
		if err := db.RunMigrations(dsn); err != nil {
			logger.Error("run migrations", "error", err)
			os.Exit(1)
		}
	}

	authCfg, err := auth.LoadConfig()
	if err != nil {
		logger.Error("load auth config", "error", err)
		os.Exit(1)
	}

	repo := repository.NewPostgresTaskRepo(pool)
	tx := repository.NewTxRunner(pool)
	uc := usecase.New(repo, tx)
	th := handler.NewTaskHandler(uc)

	userRepo := repository.NewUserRepo(pool)
	refreshRepo := repository.NewRefreshTokenRepo(pool)
	authUc := usecase.NewAuthUsecase(userRepo, refreshRepo, authCfg)
	authHandler := handler.NewAuthHandler(authUc, authCfg)

	limiter := mw.NewIPRateLimiter(rate.Limit(10), 20)
	go limiter.StartGC(ctx, 5*time.Minute, 1*time.Hour)

	r := router.New(router.Deps{
		Logger:      logger,
		RateLimiter: limiter,
		TaskHandler: th,
		AuthHandler: authHandler,
		AuthCfg:     authCfg,
		Production:  os.Getenv("APP_ENV") == "production",
	})
	if err := r.Run(":8080"); err != nil {
		logger.Error("server exited", "error", err)
		os.Exit(1)
	}
}
