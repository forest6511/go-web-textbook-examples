package db

import (
	"context"
	"fmt"
	"time"

	"github.com/amirsalarsafaei/sqlc-pgx-monitoring/dbtracer"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
)

// NewPool は Cloud Run 向けに調整したコネクションプールを返す
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	cfg.MaxConns = 25
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnLifetimeJitter = 5 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute

	tracer, err := dbtracer.NewDBTracer("tasks_db",
		dbtracer.WithTraceProvider(otel.GetTracerProvider()),
		dbtracer.WithMeterProvider(otel.GetMeterProvider()),
		dbtracer.WithLogArgs(false),
	)
	if err != nil {
		return nil, fmt.Errorf("new db tracer: %w", err)
	}
	cfg.ConnConfig.Tracer = tracer

	return newWithConfig(ctx, cfg)
}

func newWithConfig(
	ctx context.Context, cfg *pgxpool.Config,
) (*pgxpool.Pool, error) {
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	return pool, nil
}
