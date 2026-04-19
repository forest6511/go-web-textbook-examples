package observability

import (
	"context"
	"errors"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Providers は Init が返す、個別 shutdown 可能な SDK プロバイダ群。
type Providers struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *metric.MeterProvider
	LoggerProvider *log.LoggerProvider
	Resource       *resource.Resource
}

// Shutdown はすべてのプロバイダを逆順で Shutdown する。
// graceful shutdown の defer 順序に組み込みやすいよう 1 関数に集約している。
// 最初に起きたエラーを返しつつ、後続の Shutdown は必ず呼ぶ（リソースリーク防止）。
func (p *Providers) Shutdown(ctx context.Context) error {
	var errs []error
	if p.LoggerProvider != nil {
		if err := p.LoggerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("logger provider shutdown: %w", err))
		}
	}
	if p.MeterProvider != nil {
		if err := p.MeterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("meter provider shutdown: %w", err))
		}
	}
	if p.TracerProvider != nil {
		if err := p.TracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("tracer provider shutdown: %w", err))
		}
	}
	return errors.Join(errs...)
}

// Init は OTel の Tracer/Meter/Logger プロバイダを一括初期化する。
// Ch 09 で個別に書いた初期化コードを 1 エントリにまとめたヘルパ。
// 返り値の Providers.Shutdown を defer で呼ぶだけでクリーンアップが完結する。
func Init(ctx context.Context) (*Providers, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	res, err := NewResource(ctx, env)
	if err != nil {
		return nil, fmt.Errorf("new resource: %w", err)
	}

	tp, err := NewTracerProvider(ctx, res)
	if err != nil {
		return nil, fmt.Errorf("new tracer provider: %w", err)
	}
	mp, err := NewMeterProvider(res)
	if err != nil {
		// tracer は起動済みなのでここで Shutdown しておく
		_ = tp.Shutdown(ctx)
		return nil, fmt.Errorf("new meter provider: %w", err)
	}
	lp, err := NewLoggerProvider(ctx, res)
	if err != nil {
		_ = mp.Shutdown(ctx)
		_ = tp.Shutdown(ctx)
		return nil, fmt.Errorf("new logger provider: %w", err)
	}

	return &Providers{
		TracerProvider: tp,
		MeterProvider:  mp,
		LoggerProvider: lp,
		Resource:       res,
	}, nil
}
