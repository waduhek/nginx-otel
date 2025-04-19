package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func getOTLPEndpoint() (string, error) {
	endpoint := os.Getenv("OTLP_ENDPOINT")
	if endpoint == "" {
		return "", errors.New("Value of OTLP_ENDPOINT is required")
	}

	return endpoint, nil
}

func newOTLPExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	otlpEndpoint, err := getOTLPEndpoint()
	if err != nil {
		return nil, err
	}

	endpointOpt := otlptracegrpc.WithEndpoint(otlpEndpoint)
	insecureOpt := otlptracegrpc.WithInsecure()

	return otlptracegrpc.New(ctx, endpointOpt, insecureOpt)
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

func newTracerProvider(exporter sdktrace.SpanExporter) *sdktrace.TracerProvider {
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

func newTextMapPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.Baggage{},
		propagation.TraceContext{},
	)
}
