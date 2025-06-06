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

// NodeService is the interface that all CSI node drivers must implement.
type NodeService interface {
	GetIdentityServer() csi.IdentityServer
	GetNodeServer() csi.NodeServer
}

// NodeServer is used to configure, start and stop the CSI server.
type NodeServer struct {
	proto string
	addr  string

	driver NodeService
	mp     metric.MeterProvider
	tp     trace.TracerProvider
}

// NewNodeServer creates a new CSI server for running in the agent.
//
// The server will listen on the provided endpoint and use the provided driver
// to handle incoming requests.
//
// The tracer provider will be used to create tracers for the server.
func NewNodeServer(endpoint string, driver NodeService, t telemetry.Provider) (*NodeServer, error) {
	proto, addr, err := parseEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	if err := prepareEndpoint(proto, addr); err != nil {
		return nil, err
	}

	return &NodeServer{
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
func (s *NodeServer) Start(ctx context.Context) error {
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
	csi.RegisterNodeServer(server, s.driver.GetNodeServer())

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
