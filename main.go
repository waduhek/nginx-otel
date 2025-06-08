package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/waduhek/nginxotel/telemetry"
)

func main() {
	ctx := context.Background()

	traceExporter, err := telemetry.NewOTLPTraceExporter(ctx)
	if err != nil {
		slog.Error("error while creating exporter", "err", err)
		os.Exit(1)
	}

	tracerProvider := telemetry.NewTracerProvider(traceExporter)
	defer tracerProvider.Shutdown(ctx)

	propagator := telemetry.NewTextMapPropagator()

	loggerExporter, err := telemetry.NewOTLPLoggerExporter(ctx)
	if err != nil {
		slog.Error("error while creating log exporter", "err", err)
		os.Exit(1)
	}

	loggerProvider, err := telemetry.NewLoggerProvider(loggerExporter)
	if err != nil {
		slog.Error("error while creating logger provider", "err", err)
		os.Exit(1)
	}
	defer loggerProvider.Shutdown(ctx)

	metricsExporter, err := telemetry.NewOTLPMetricsExporter(ctx)
	if err != nil {
		slog.Error("error while creating metric exporter", "err", err)
		os.Exit(1)
	}
	defer metricsExporter.Shutdown(ctx)

	meterProvider := telemetry.NewMeterProvider(metricsExporter)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagator)
	otel.SetMeterProvider(meterProvider)
	global.SetLoggerProvider(loggerProvider)

	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(":8080", nil)
}

func getAPICounter() (metric.Int64Counter, error) {
	return telemetry.GetMeter().Int64Counter(
		"api.calls",
		metric.WithDescription("Number of API calls"),
		metric.WithUnit("{call}"),
	)
}

func incrementAPICallCount(ctx context.Context, logger telemetry.Logger) {
	counter, err := getAPICounter()
	if err != nil {
		logger.ErrorContext(
			ctx,
			"error while getting api call counter",
			"err", err,
		)
	} else {
		counter.Add(
			ctx,
			1,
			metric.WithAttributes(semconv.HTTPResponseStatusCode(200)),
		)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	methodName := "github.com/waduhek/nginxotel.handleRequest"

	propagator := otel.GetTextMapPropagator()
	logger := otelslog.NewLogger(methodName)

	extractedCtx := propagator.Extract(
		r.Context(),
		propagation.HeaderCarrier(r.Header),
	)

	newTraceCtx, span := telemetry.NewSpan(extractedCtx, "GET /")
	defer span.End()

	incrementAPICallCount(newTraceCtx, logger)

	carrier := make(propagation.HeaderCarrier)
	propagator.Inject(newTraceCtx, &carrier)

	span.SetStatus(codes.Ok, "request completed")

	for key, values := range carrier {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	logger.InfoContext(newTraceCtx, "request completed")

	w.WriteHeader(200)
	fmt.Fprintf(w, "200 OK")
}
