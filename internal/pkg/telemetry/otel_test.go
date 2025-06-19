// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package telemetry

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"k8s.io/component-base/tracing"
)

func TestInitTracing(t *testing.T) {
	tests := []struct {
		name        string
		config      config
		expectedErr error
		validate    func(t *testing.T, tp tracing.TracerProvider)
	}{
		{
			name: "without endpoint creates noop provider",
			config: config{
				endpoint:        "",
				traceSampleRate: 1,
			},
			validate: func(t *testing.T, tp tracing.TracerProvider) {
				if !reflect.DeepEqual(tp, tracing.NewNoopTracerProvider()) {
					t.Fatalf("expected noop tracer provider, got %v", tp)
				}
			},
		},
		{
			name: "without traceSampleRate as zero creates noop provider",
			config: config{
				endpoint:        "jaeger-collector.observability.svc.cluster.local:4317",
				traceSampleRate: 0,
			},
			validate: func(t *testing.T, tp tracing.TracerProvider) {
				if !reflect.DeepEqual(tp, tracing.NewNoopTracerProvider()) {
					t.Fatalf("expected noop tracer provider, got %v", tp)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tp, err := InitTracing(context.Background(), tt.config)
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("InitTracing() error = %v, wantErr %v", err, tt.expectedErr)
			}
			if tt.validate != nil {
				tt.validate(t, tp)
			}
		})
	}
}
