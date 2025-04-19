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
	"go.opentelemetry.io/otel/propagation"

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

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagator)
	global.SetLoggerProvider(loggerProvider)

	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(":8080", nil)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	methodName := "github.com/waduhek/nginxotel.handleRequest"

	tracer := telemetry.GetTracer()
	propagator := otel.GetTextMapPropagator()
	logger := otelslog.NewLogger(methodName)

	extractedCtx := propagator.Extract(
		r.Context(),
		propagation.HeaderCarrier(r.Header),
	)

	newTraceCtx, span := tracer.Start(extractedCtx, "GET /")
	span.End()

	carrier := make(propagation.HeaderCarrier)
	propagator.Inject(newTraceCtx, &carrier)

	span.SetStatus(codes.Error, "request completed")

	for key, values := range carrier {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	logger.InfoContext(newTraceCtx, "request completed")

	w.WriteHeader(200)
	fmt.Fprintf(w, "200 OK")
}
