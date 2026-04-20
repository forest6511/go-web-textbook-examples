package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/time/rate"

	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/auth"
	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/db"
	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/handler"
	mw "github.com/forest6511/go-web-textbook-examples/ch12-production/internal/middleware"
	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/observability"
	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/repository"
	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/router"
	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/storage"
	"github.com/forest6511/go-web-textbook-examples/ch12-production/internal/usecase"
	appval "github.com/forest6511/go-web-textbook-examples/ch12-production/internal/validator"
)

// version は GitHub Actions で ldflags 経由で埋める: -X main.version=$GITHUB_SHA
var version = "dev"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if _, err := appval.Setup(); err != nil {
		logger.Error("setup validator", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// Sentry は最初に初期化し、以降の panic / エラーを必ず捕捉する
	if err := observability.InitSentry(version, env); err != nil {
		logger.Error("init sentry", "error", err)
		os.Exit(1)
	}
	// 2 秒以内に pending events を送信完了まで待つ
	defer observability.FlushSentry(2 * time.Second)

	// OTel 系の初期化（tracer / meter / logger）。すべて Providers.Shutdown で束ねる
	providers, err := observability.Init(ctx)
	if err != nil {
		logger.Error("init observability", "error", err)
		os.Exit(1)
	}
	// observability は最後に閉じる（handler/DB が trace 送出中の可能性があるため）
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := providers.Shutdown(shutdownCtx); err != nil {
			logger.Error("observability shutdown", "error", err)
		}
	}()

	// SLI メトリクス（default registry に登録、/metrics で露出）
	observability.InitSLI()

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

	storageCfg := storage.LoadConfig()
	s3Client, err := storage.NewS3Client(ctx, storageCfg)
	if err != nil {
		logger.Error("init s3 client", "error", err)
		os.Exit(1)
	}
	objStorage := storage.NewS3Storage(s3Client, storageCfg.Bucket)
	attachmentRepo := repository.NewPgAttachmentRepo(pool)
	attachmentHandler := handler.NewAttachmentHandler(
		objStorage, attachmentRepo)

	healthHandler := handler.NewHealthHandler(pool)

	r := router.New(router.Deps{
		Logger:            logger,
		RateLimiter:       limiter,
		TaskHandler:       th,
		AuthHandler:       authHandler,
		AttachmentHandler: attachmentHandler,
		HealthHandler:     healthHandler,
		AuthCfg:           authCfg,
		Production:        env == "production",
		SentryEnabled:     os.Getenv("SENTRY_DSN") != "",
	})
	r.MaxMultipartMemory = 8 << 20

	srv := &http.Server{
		Addr:              ":" + port(),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		logger.Info("server listening", "addr", srv.Addr, "version", version)
		if err := srv.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			stop() // 致命的エラーで shutdown をトリガ
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received, draining")

	// Cloud Run の SIGKILL (10 秒) より短い 8 秒でドレイン
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}

// port は PORT 環境変数を尊重する（Cloud Run / Fly.io 必須）。
// 未設定時は 8080 にフォールバック。Ch 11 と同じ idiom。
func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}
