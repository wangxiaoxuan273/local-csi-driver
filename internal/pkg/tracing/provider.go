// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tracing

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"k8s.io/component-base/tracing"
	tracingv1 "k8s.io/component-base/tracing/api/v1"
	"k8s.io/klog/v2"
)

var (
	tp   tracing.TracerProvider
	once sync.Once
)

// New creates a new tracing provider. It should be called before traces are
// used.
//
// The service name should be set to the high-level service name, such as
// "cns".
//
// The ID should be a unique identifier for the instance of the service.
//
// The endpoint should be the address of the trace collector. It only supports
// OTLP over gRPC. If empty, tracing is disabled.
//
// The sample rate is the number of samples taken per million. If 0, tracing is
// disabled. If 1000000, there is no sampling and all traces are sent.
//
// The provider is stored globally and can be retrieved with GetTracerProvider.
func New(ctx context.Context, service string, id string, endpoint string, sampleRate int) (tracing.TracerProvider, error) {
	if endpoint == "" {
		klog.InfoS("--trace-address not set, tracing disabled")
		tp = tracing.NewNoopTracerProvider()
		return tp, nil
	}
	klog.InfoS("setting up trace exporter", "endpoint", endpoint, "rate", sampleRate)

	opts := []otlptracegrpc.Option{}
	rate := int32(sampleRate)
	cfg := &tracingv1.TracingConfiguration{
		Endpoint:               &endpoint,
		SamplingRatePerMillion: &rate,
	}

	resourceOpts := []resource.Option{
		resource.WithAttributes(
			semconv.ServiceNameKey.String(service),
			semconv.ServiceInstanceIDKey.String(id),
		),
	}

	// Use the trace provider from component-base so it can also be used with
	// Kubernetes components that support tracing.
	tp, err := tracing.NewProvider(ctx, cfg, opts, resourceOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer provider: %w", err)
	}

	// Register the global trace provider.
	SetTracerProvider(tp)

	return tp, nil
}

// NewNoopTracerProvider creates a new no-op tracing provider.
func NewNoopTracerProvider() tracing.TracerProvider {
	return tracing.NewNoopTracerProvider()
}

// GetTracerProvider returns the global tracing provider, if set.
func GetTracerProvider() tracing.TracerProvider {
	if tp == nil {
		return tracing.NewNoopTracerProvider()
	}
	return tp
}

// SetTracerProvider sets the global tracing provider.
func SetTracerProvider(p tracing.TracerProvider) {
	once.Do(func() {
		otel.SetTracerProvider(tp)

		// Set global propagator to tracecontext (the default is no-op).
		otel.SetTextMapPropagator(propagation.TraceContext{})
		tp = p
	})
}
