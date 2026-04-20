package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"golang.org/x/time/rate"

	"github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/auth"
	"github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/db"
	"github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/handler"
	mw "github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/middleware"
	"github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/observability"
	"github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/repository"
	"github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/router"
	"github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/storage"
	"github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/usecase"
	appval "github.com/forest6511/go-web-textbook-examples/ch09-slog-otel/internal/validator"
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

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	res, err := observability.NewResource(ctx, env)
	if err != nil {
		logger.Error("new resource", "error", err)
		os.Exit(1)
	}
	tp, err := observability.NewTracerProvider(ctx, res)
	if err != nil {
		logger.Error("new tracer provider", "error", err)
		os.Exit(1)
	}
	defer func() {
		sctx, scancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer scancel()
		_ = tp.Shutdown(sctx)
	}()
	mp, err := observability.NewMeterProvider(res)
	if err != nil {
		logger.Error("new meter provider", "error", err)
		os.Exit(1)
	}
	defer func() {
		sctx, scancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer scancel()
		_ = mp.Shutdown(sctx)
	}()
	lp, err := observability.NewLoggerProvider(ctx, res)
	if err != nil {
		logger.Error("new logger provider", "error", err)
		os.Exit(1)
	}
	defer func() {
		sctx, scancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer scancel()
		_ = lp.Shutdown(sctx)
	}()

	// stdout + OTel 両方にログを流す。OTel 経路で trace_id/span_id が自動付与される
	stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	otelHandler := otelslog.NewHandler("go-web-textbook",
		otelslog.WithLoggerProvider(lp))
	combined := observability.NewMultiHandler(stdoutHandler, otelHandler)
	logger = slog.New(combined)
	slog.SetDefault(logger)
	slog.InfoContext(ctx, "observability ready")

	if err := otelruntime.Start(
		otelruntime.WithMinimumReadMemStatsInterval(15 * time.Second),
	); err != nil {
		logger.Error("start runtime instrumentation", "error", err)
		os.Exit(1)
	}

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

	r := router.New(router.Deps{
		Logger:            logger,
		RateLimiter:       limiter,
		TaskHandler:       th,
		AuthHandler:       authHandler,
		AttachmentHandler: attachmentHandler,
		AuthCfg:           authCfg,
		Production:        os.Getenv("ENV") == "production",
	})
	r.MaxMultipartMemory = 8 << 20
	if err := r.Run(":8080"); err != nil {
		logger.Error("server exited", "error", err)
		os.Exit(1)
	}
}
