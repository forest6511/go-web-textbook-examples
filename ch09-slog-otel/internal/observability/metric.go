package observability

import (
	"fmt"

	"go.opentelemetry.io/otel"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func NewMeterProvider(
	res *resource.Resource,
) (*metric.MeterProvider, error) {
	exp, err := otelprom.New()
	if err != nil {
		return nil, fmt.Errorf("new prometheus exporter: %w", err)
	}
	mp := metric.NewMeterProvider(
		metric.WithReader(exp),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(mp)
	return mp, nil
}
