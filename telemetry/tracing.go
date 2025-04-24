package telemetry

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const packageName string = "github.com/waduhek/nginxotel"

func NewOTLPTraceExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	otlpEndpoint, err := getOTLPEndpoint()
	if err != nil {
		return nil, err
	}

	endpointOpt := otlptracegrpc.WithEndpoint(otlpEndpoint)
	insecureOpt := otlptracegrpc.WithInsecure()

	return otlptracegrpc.New(ctx, endpointOpt, insecureOpt)
}

func NewTracerProvider(exporter sdktrace.SpanExporter) *sdktrace.TracerProvider {
	res, err := newResource()
	if err != nil {
		fmt.Printf("error while initialising resource: %v\n", err)
		os.Exit(1)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
}

func NewTextMapPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.Baggage{},
		propagation.TraceContext{},
	)
}

func GetTracer() trace.Tracer {
	return otel.GetTracerProvider().Tracer(packageName)
}

func NewSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	hostName := os.Getenv("HOSTNAME")
	tracer := GetTracer()

	return tracer.Start(
		ctx,
		spanName,
		trace.WithAttributes(semconv.ContainerID(hostName)),
	)
}
