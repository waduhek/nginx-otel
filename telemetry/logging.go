package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

type Logger interface {
	DebugContext(ctx context.Context, msg string, args ...any)

	InfoContext(ctx context.Context, msg string, args ...any)

	WarnContext(ctx context.Context, msg string, args ...any)

	ErrorContext(ctx context.Context, msg string, args ...any)
}

func NewOTLPLoggerExporter(ctx context.Context) (*otlploggrpc.Exporter, error) {
	otlpEndpoint, err := getOTLPEndpoint()
	if err != nil {
		return nil, err
	}

	endpointOpt := otlploggrpc.WithEndpoint(otlpEndpoint)
	insecureOpt := otlploggrpc.WithInsecure()

	return otlploggrpc.New(ctx, endpointOpt, insecureOpt)
}

func NewLoggerProvider(exporter sdklog.Exporter) (*sdklog.LoggerProvider, error) {
	resource, err := newResource()
	if err != nil {
		return nil, err
	}

	processor := sdklog.NewBatchProcessor(exporter)

	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(resource),
		sdklog.WithProcessor(processor),
	)

	return provider, nil
}
