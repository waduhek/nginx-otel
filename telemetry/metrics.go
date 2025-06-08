package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func NewOTLPMetricsExporter(ctx context.Context) (*otlpmetricgrpc.Exporter, error) {
	otlpEndpoint, err := getOTLPEndpoint()
	if err != nil {
		return nil, err
	}

	endpointOpt := otlpmetricgrpc.WithEndpoint(otlpEndpoint)
	insecureOpt := otlpmetricgrpc.WithInsecure()

	return otlpmetricgrpc.New(ctx, endpointOpt, insecureOpt)
}

func NewMeterProvider(exporter sdkmetric.Exporter) *sdkmetric.MeterProvider {
	res, err := newResource()
	if err != nil {
		fmt.Printf("error while initialising resource: %v\n", err)
		os.Exit(1)
	}

	return sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				exporter,
				sdkmetric.WithInterval(5*time.Second),
			),
		),
	)
}

func GetMeter() metric.Meter {
	return otel.Meter("github.com/waduhek/nginxotel")
}
