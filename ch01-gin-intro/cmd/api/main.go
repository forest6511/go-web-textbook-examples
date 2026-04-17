package main

import (
	"log/slog"
	"os"

	"github.com/forest6511/go-web-textbook-examples/ch01-gin-intro/internal/router"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	r := router.New()
	if err := r.Run(":8080"); err != nil {
		logger.Error("server exited", "error", err)
		os.Exit(1)
	}
}
