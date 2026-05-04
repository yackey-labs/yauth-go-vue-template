// Package telemetry initializes OpenTelemetry for the application:
// OTLP/HTTP exporter + W3C TraceContext propagator + a service-named
// resource. It deliberately bypasses yauth-go's telemetry.Init because
// that helper:
//
//   - merges resource.Default() (semconv 1.40+) with an explicit
//     semconv 1.26.0 attribute, which fails with a "conflicting Schema
//     URL" error on current SDK versions; and
//   - uses gRPC, while the otel-{local,}.yackey.cloud collectors only
//     speak OTLP/HTTP on port 443 (with /v1/traces appended).
//
// yauth-go's plugin handlers and our otelhttp wrapper both call
// otel.Tracer(...) on the global provider, so setting the global here
// is enough — no builder-side WithTelemetry call is needed.
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	noopt "go.opentelemetry.io/otel/trace/noop"
)

// Config controls how Init wires the SDK.
type Config struct {
	Enabled     bool
	ServiceName string
	// Endpoint follows OTLP semantics — empty falls back to
	// OTEL_EXPORTER_OTLP_ENDPOINT (the SDK's exporter handles this
	// itself when we omit WithEndpointURL).
	Endpoint string
}

// Init sets the global TracerProvider + propagator. The returned
// shutdown drains pending spans on graceful exit; it's safe to call
// when Enabled is false (no-op).
func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	if !cfg.Enabled {
		otel.SetTracerProvider(noopt.NewTracerProvider())
		return func(context.Context) error { return nil }, nil
	}

	opts := []otlptracehttp.Option{}
	if cfg.Endpoint != "" {
		// WithEndpointURL takes the full base URL. The HTTP exporter
		// appends "/v1/traces" automatically per OTel spec, so callers
		// pass e.g. https://otel-local.yackey.cloud (no path).
		opts = append(opts, otlptracehttp.WithEndpointURL(cfg.Endpoint))
	}
	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("telemetry: otlp exporter: %w", err)
	}

	serviceName := cfg.ServiceName
	if serviceName == "" {
		serviceName = "yauth-go-vue-template"
	}

	// Schemaless resource — explicitly avoids merging with
	// resource.Default() which pulls in a different semconv schema URL
	// from the SDK and clashes with yauth-go's pinned semconv 1.26.0.
	res := resource.NewSchemaless(
		attribute.String("service.name", serviceName),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Shutdown, nil
}
