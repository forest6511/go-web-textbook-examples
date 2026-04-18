package observability

import (
	"io"
	"log/slog"
	"os"
)

func NewBaseLogger(w io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	if os.Getenv("APP_ENV") == "production" {
		return slog.New(slog.NewJSONHandler(w, opts))
	}
	opts.Level = slog.LevelDebug
	opts.AddSource = true
	return slog.New(slog.NewTextHandler(w, opts))
}
