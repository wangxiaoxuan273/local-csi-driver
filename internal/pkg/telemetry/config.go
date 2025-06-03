// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// Exporter is an enum for the type of telemetry exporter.
type Exporter int

const (
	Console Exporter = iota
	OTLP
	Prometheus
)

// config contains configuration options for the telemetry package.
type config struct {
	serviceID       string
	endpoint        string
	traceSampleRate int32
	prometheus      metrics.RegistererGatherer
	exporters       map[Exporter]bool
}

// newConfig returns a config configured with options.
func newConfig(options []Option) config {
	conf := config{
		prometheus: prometheus.NewRegistry(),
		exporters:  map[Exporter]bool{},
	}

	for _, o := range options {
		conf = o.apply(conf)
	}
	return conf
}

// Option applies a configuration option value to the telemetry package.
type Option interface {
	apply(config) config
}

// optionFunc applies a set of options to a config.
type optionFunc func(config) config

// apply returns a config with option(s) applied.
func (o optionFunc) apply(conf config) config {
	return o(conf)
}

// WithServiceInstanceID sets the service instance ID for the telemetry package.
//
// This is expected to be the pod name in a Kubernetes environment.
func WithServiceInstanceID(id string) Option {
	return optionFunc(func(cfg config) config {
		cfg.serviceID = id
		return cfg
	})
}

// WithConsole sets the console exporter.
func WithConsole() Option {
	return optionFunc(func(cfg config) config {
		cfg.exporters[Console] = true
		return cfg
	})
}

// WithOTLP sets the OTLP exporter.
//
// If the endpoint is not set with WithEndpoint(), it will default to exporting
// to "localhost:4317".
func WithOTLP() Option {
	return optionFunc(func(cfg config) config {
		cfg.exporters[OTLP] = true
		return cfg
	})
}

// WithPrometheus sets the Prometheus exporter.
func WithPrometheus(prom metrics.RegistererGatherer) Option {
	return optionFunc(func(cfg config) config {
		cfg.exporters[Prometheus] = true
		cfg.prometheus = prom
		return cfg
	})
}

// WithEndpoint sets the endpoint for all OTLP exporters.
//
// We may want to split per type later, but for now we just set it globally.
func WithEndpoint(endpoint string) Option {
	return optionFunc(func(cfg config) config {
		cfg.endpoint = endpoint
		return cfg
	})
}

// WithTraceSampleRate sets the number of samples to collect per million spans.
func WithTraceSampleRate(rate int) Option {
	return optionFunc(func(cfg config) config {
		cfg.traceSampleRate = int32(rate)
		return cfg
	})
}
