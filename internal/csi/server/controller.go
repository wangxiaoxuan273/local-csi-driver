// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package server

import (
	"context"
	"fmt"
	"net"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"local-csi-driver/internal/pkg/telemetry"
)

// ControllerService is the interface that all CSI controller drivers must
// implement.
type ControllerService interface {
	GetIdentityServer() csi.IdentityServer
	GetControllerServer() csi.ControllerServer
	Register(ctx context.Context) error
}

// ControllerServer is used to configure, start and stop the CSI server.
type ControllerServer struct {
	proto string
	addr  string

	driver ControllerService
	mp     metric.MeterProvider
	tp     trace.TracerProvider
}

// +kubebuilder:rbac:groups=storage.k8s.io,resources=csidrivers,verbs=get;list;watch;create;update;patch;delete

// NewControllerServer creates a new CSI server for running in the manager.
//
// The server will listen on the provided endpoint and use the provided driver
// to handle incoming requests.
//
// The tracer provider will be used to create tracers for the server.
func NewControllerServer(endpoint string, driver ControllerService, t telemetry.Provider) (*ControllerServer, error) {
	proto, addr, err := parseEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	if err := prepareEndpoint(proto, addr); err != nil {
		return nil, err
	}

	return &ControllerServer{
		proto:  proto,
		addr:   addr,
		driver: driver,
		mp:     t.MeterProvider(),
		tp:     t.TraceProvider(),
	}, nil
}

// Start starts the CSI server.
//
// The server will listen on the configured endpoint and serve incoming
// requests.
//
// The server will stop when the provided context is canceled.
func (s *ControllerServer) Start(ctx context.Context) error {
	log := log.FromContext(ctx)

	listener, err := net.Listen(s.proto, s.addr)
	if err != nil {
		return fmt.Errorf("failed to create listener on %s://%s: %w", s.proto, s.addr, err)
	}

	opts := []otelgrpc.Option{
		otelgrpc.WithMeterProvider(s.mp),
		otelgrpc.WithTracerProvider(s.tp),
	}
	server := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler(opts...)))

	csi.RegisterIdentityServer(server, s.driver.GetIdentityServer())
	csi.RegisterControllerServer(server, s.driver.GetControllerServer())

	// Ensure the CSIDriver object exists in the cluster.
	if err := s.driver.Register(ctx); err != nil {
		return fmt.Errorf("failed to register driver: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		log.V(2).Info("listening for connections", "endpoint", listener.Addr())
		errCh <- server.Serve(listener)
		close(errCh)
	}()

	<-ctx.Done()
	server.GracefulStop()

	return <-errCh
}

// NeedLeaderElection returns false since the CSI controllers should run on each
// node.
func (s *ControllerServer) NeedLeaderElection() bool {
	return false
}
