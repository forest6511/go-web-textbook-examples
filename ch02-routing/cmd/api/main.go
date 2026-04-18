package main

import (
	"log/slog"
	"os"

	"github.com/forest6511/go-web-textbook-examples/ch02-routing/internal/handler"
	"github.com/forest6511/go-web-textbook-examples/ch02-routing/internal/repository"
	"github.com/forest6511/go-web-textbook-examples/ch02-routing/internal/router"
	"github.com/forest6511/go-web-textbook-examples/ch02-routing/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	repo := repository.NewInMemoryTaskRepo()
	taskUsecase := usecase.New(repo)
	taskHandler := handler.NewTaskHandler(taskUsecase)

	r := router.New(taskHandler)
	if err := r.Run(":8080"); err != nil {
		logger.Error("server exited", "error", err)
		os.Exit(1)
	}
}
