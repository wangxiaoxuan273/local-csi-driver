// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package telemetry

import (
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"local-csi-driver/internal/pkg/version"
)

const (
	service = "local-csi-driver"
)

// NewResourceOptions creates a set of default resource options for
// OpenTelemetry. These are added to every span and metric, which may prove to
// be too verbose.
//
// ServiceID is expected to be set to the pod name.
func NewResourceOptions(cfg config) []resource.Option {
	return []resource.Option{
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(service),
			semconv.ServiceInstanceIDKey.String(cfg.serviceID),
		),
		resource.WithDetectors(version.GetInfo()), // Version info.
	}
}
