package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/time/rate"

	"github.com/forest6511/go-web-textbook-examples/ch03-middleware/internal/handler"
	mw "github.com/forest6511/go-web-textbook-examples/ch03-middleware/internal/middleware"
	"github.com/forest6511/go-web-textbook-examples/ch03-middleware/internal/repository"
	"github.com/forest6511/go-web-textbook-examples/ch03-middleware/internal/router"
	"github.com/forest6511/go-web-textbook-examples/ch03-middleware/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	repo := repository.NewInMemoryTaskRepo()
	uc := usecase.New(repo)
	th := handler.NewTaskHandler(uc)

	limiter := mw.NewIPRateLimiter(rate.Limit(10), 20)
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	go limiter.StartGC(ctx, 5*time.Minute, 1*time.Hour)

	r := router.New(router.Deps{
		Logger:      logger,
		RateLimiter: limiter,
		TaskHandler: th,
		Production:  os.Getenv("APP_ENV") == "production",
	})
	if err := r.Run(":8080"); err != nil {
		logger.Error("server exited", "error", err)
		os.Exit(1)
	}
}
