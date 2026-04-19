package observability

import (
	"context"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func NewResource(ctx context.Context, env string) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("go-web-textbook"),
			semconv.ServiceVersion("1.0.0"),
			semconv.DeploymentEnvironment(env),
		),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithProcess(),
		resource.WithHost(),
	)
}
