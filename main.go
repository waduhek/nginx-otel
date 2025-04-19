package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

func main() {
	ctx := context.Background()

	exporter, err := newOTLPExporter(ctx)
	if err != nil {
		slog.Error("error while creating exporter", "err", err)
		os.Exit(1)
	}

	tracerProvider := newTracerProvider(exporter)
	defer tracerProvider.Shutdown(ctx)

	propagator := newTextMapPropagator();

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagator)

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		extractedCtx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		newTraceCtx, span := tracerProvider.Tracer("api_tracer").Start(extractedCtx, "GET /")
		span.End()

		carrier := propagation.HeaderCarrier{}
		propagator.Inject(newTraceCtx, &carrier)

		span.SetStatus(codes.Error, "request completed")

		for key, value := range carrier {
			w.Header().Add(key, value[0])
		}

		w.WriteHeader(200)
		fmt.Fprintf(w, "200 OK")
    })
    http.ListenAndServe(":8080", nil)
}
