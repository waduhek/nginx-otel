package telemetry

import (
	"errors"
	"os"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func getOTLPEndpoint() (string, error) {
	endpoint := os.Getenv("OTLP_ENDPOINT")
	if endpoint == "" {
		return "", errors.New("Value of OTLP_ENDPOINT is required")
	}

	return endpoint, nil
}

func newResource() (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("nginx-otel"),
		),
	)
}
