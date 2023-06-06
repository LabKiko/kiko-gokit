package trace

import (
	"context"
	"fmt"

	"github.com/LabKiko/kiko-gokit/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

type Tracing struct {
	op       *Config
	provider *sdktrace.TracerProvider
}

func New(opts ...Option) (*Tracing, error) {
	op := &Config{
		Endpoint: "http://127.0.0.1:14268/api/traces",
		Sampler:  1.0,
		Batcher:  "jaeger",
	}
	for _, opt := range opts {
		opt.apply(op)
	}

	o := &Tracing{op: op}

	r, err := resource.New(context.Background(),
		resource.WithTelemetrySDK(),
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithHost(),
		resource.WithFromEnv(), // pull attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables
		resource.WithProcess(), // This option configures a set of Detectors that discover process information
		resource.WithAttributes(op.Attributes...),
	)
	if err != nil {
		return nil, err
	}

	options := []sdktrace.TracerProviderOption{
		// Set the sampling rate based on the parent span to 100%
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(op.Sampler))),
		// Record information about this application in an Resource.
		sdktrace.WithResource(r),
	}

	var exp sdktrace.SpanExporter
	exp, err = o.createExporter()
	if err != nil {
		return nil, err
	}
	// Always be sure to batch in production.
	options = append(options, sdktrace.WithBatcher(exp))
	o.provider = sdktrace.NewTracerProvider(options...)
	otel.SetTracerProvider(o.provider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.Errorf("[otel] error: %v", err)
	}))

	return o, nil
}

func (t *Tracing) createExporter() (sdktrace.SpanExporter, error) {
	// Just support jaeger and zipkin now, more for later
	switch t.op.Batcher {
	case kindJaeger:
		return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(t.op.Endpoint)))
	case kindZipkin:
		return zipkin.New(t.op.Endpoint)
	default:
		return nil, fmt.Errorf("trace: unknown exporter: %s", t.op.Batcher)
	}
}

// Shutdown shuts down the span processors in the order they were registered.
func (t *Tracing) Shutdown(ctx context.Context) error {
	return t.provider.Shutdown(ctx)
}
