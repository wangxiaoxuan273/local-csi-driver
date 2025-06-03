// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	"github.com/gotidy/ptr"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/textlogger"
	"k8s.io/utils/exec"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	driver "local-csi-driver/internal/csi"
	"local-csi-driver/internal/csi/core/lvm"
	"local-csi-driver/internal/csi/server"
	"local-csi-driver/internal/pkg/block"
	"local-csi-driver/internal/pkg/config"
	"local-csi-driver/internal/pkg/events"
	lvmMgr "local-csi-driver/internal/pkg/lvm"
	"local-csi-driver/internal/pkg/probe"
	"local-csi-driver/internal/pkg/raid"

	"local-csi-driver/internal/pkg/telemetry"
	"local-csi-driver/internal/pkg/version"
	"local-csi-driver/internal/webhook/hyperconverged"
	"local-csi-driver/internal/webhook/pvc"
	// +kubebuilder:scaffold:imports
)

const (
	// ServiceName is the name of the service used in traces.
	ServiceName = "local-csi-driver"

	// terminationMessagePath is the path to the termination message file for the
	// Kubernetes pod. This file is used to store the last error message.
	terminationMessagePath = "/tmp/termination-log"
)

var (
	scheme = runtime.NewScheme()
	log    = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	// +kubebuilder:scaffold:scheme
}

func main() {
	var nodeName string
	var podName string
	var namespace string
	var configFile string
	var csiAddr string
	var metricsAddr string
	var probeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	var workers int
	var apiQPS int
	var apiBurst int
	var traceAddr string
	var traceSampleRate int
	var traceServiceID string
	var tlsOpts []func(*tls.Config)
	var printVersionAndExit bool
	var eventRecorderEnabled bool
	flag.StringVar(&nodeName, "nodeName", "",
		"The name of the node this agent is running on.")
	flag.StringVar(&podName, "podName", "",
		"The name of the pod this agent is running on.")
	flag.StringVar(&namespace, "namespace", "default",
		"The namespace to use for creating objects.")
	flag.StringVar(&configFile, "config", "",
		"The controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values.")
	flag.StringVar(&csiAddr, "csi-bind-address", "unix:///tmp/csi.sock",
		"The address the CSI endpoint binds to. Format: <proto>://<address>")
	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
		"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&secureMetrics, "metrics-secure", true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	flag.IntVar(&workers, "worker-threads", 10,
		"Number of worker threads per controller, in other words nr. of simultaneous CSI calls.")
	flag.IntVar(&apiQPS, "kube-api-qps", 20,
		"QPS to use while communicating with the kubernetes apiserver. Defaults to 20.")
	flag.IntVar(&apiBurst, "kube-api-burst", 30,
		"Burst to use while communicating with the kubernetes apiserver. Defaults to 30.")
	flag.StringVar(&traceAddr, "trace-address", "",
		"The address to send traces to. localhost:4317 if not set.")
	flag.IntVar(&traceSampleRate, "trace-sample-rate", 0,
		"Sample rate per million. 0 to disable tracing, 1000000 to trace everything.")
	flag.StringVar(&traceServiceID, "trace-service-id", "",
		"The service id to set in traces that identifies this service instance.")
	flag.BoolVar(&printVersionAndExit, "version", false, "Print version and exit")
	flag.BoolVar(&eventRecorderEnabled, "event-recorder-enabled", true, "If enabled, the driver will use the event recorder to record events. This is useful for debugging and monitoring purposes.")
	// Initialize klog flags
	klog.InitFlags(flag.CommandLine)

	// Parse flags
	flag.Parse()

	ctrl.SetLogger(textlogger.NewLogger(textlogger.NewConfig()))

	version.Log(log)
	if printVersionAndExit {
		return
	}

	// Parent context will be closed on interrupt or sigterm. From this point,
	// context should be closed before exiting.
	ctx, cancel := context.WithCancel(ctrl.SetupSignalHandler())
	defer cancel()

	// Add telemetry.
	t, err := telemetry.New(ctx,
		// telemetry.WithOTLP(),	// Needs testing.
		telemetry.WithServiceInstanceID(traceServiceID),
		telemetry.WithPrometheus(metrics.Registry),
		telemetry.WithEndpoint(traceAddr),
		telemetry.WithTraceSampleRate(traceSampleRate),
	)
	if err != nil {
		logAndExit(err, "failed to initialize telemetry")
	}

	// TraceProvider is passed into controllers and other components that need
	// to create spans.
	tp := t.TraceProvider()

	ctx, span := tp.Tracer("main").Start(ctx, "setup")
	defer span.End()

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		log.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOpts,
	})

	// Metrics endpoint is enabled in 'config/default/kustomization.yaml'. The Metrics options configure the server.
	// More info:
	// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/metrics/server
	// - https://book.kubebuilder.io/reference/metrics.html
	metricsServerOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
		TLSOpts:       tlsOpts,
	}

	if secureMetrics {
		// FilterProvider is used to protect the metrics endpoint with authn/authz.
		// These configurations ensure that only authorized users and service accounts
		// can access the metrics endpoint. The RBAC are configured in 'config/rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/metrics/filters#WithAuthenticationAndAuthorization
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization

		// TODO(user): If CertDir, CertName, and KeyName are not specified, controller-runtime will automatically
		// generate self-signed certificates for the metrics server. While convenient for development and testing,
		// this setup is not recommended for production.

		// TODO(user): If cert-manager is enabled in config/default/kustomization.yaml,
		// you can uncomment the following lines to use the certificate managed by cert-manager.
		// metricsServerOptions.CertDir = "/tmp/k8s-metrics-server/metrics-certs"
		// metricsServerOptions.CertName = "tls.crt"
		// metricsServerOptions.KeyName = "tls.key"

	}

	// Override the default QPS and Burst settings for the Kubernetes client.
	restCfg, err := ctrl.GetConfig()
	if err != nil {
		logAndExit(err, "unable to get rest config for api server")
	}
	restCfg.QPS = float32(apiQPS)
	restCfg.Burst = apiBurst

	mgr, err := ctrl.NewManager(restCfg, ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
	})
	if err != nil {
		logAndExit(err, "unable to start manager")
	}

	// Add telemetry to manager.
	if err := mgr.Add(t); err != nil {
		logAndExit(err, "unable to add telemetry to internal manager")
	}

	// Setup all Controllers.
	cfg, err := config.GetClusterConfig(configFile)
	if err != nil || cfg == nil {
		logAndExit(err, "unable to get cluster config")
	}

	if err := raid.Initialize(exec.New()); err != nil {
		logAndExit(err, "failed to initialize raid")
	}

	// TODO(sc): move filter to controller so we can read filters from
	// storageclass params. Hardcoded for now.
	deviceProbe := probe.New(block.New(exec.New()), probe.EphemeralDiskFilter)

	// Create the LVM manager.
	// LVM manager is an abstraction that understands how to create and
	// manage LVM resources like PV, VG, and LV.
	lvmMgr := lvmMgr.NewClient(lvmMgr.WithTracerProvider(tp))
	if !lvmMgr.IsSupported() {
		logAndExit(fmt.Errorf("lvm is not supported on this node"), "lvm is not supported")
	}

	// Create the LVM CSI server.
	//
	// Volume client is an abstraction that understands csi requests and
	// responses and how to implement them for a storage type.
	volumeClient, err := lvm.New(mgr.GetClient(), podName, nodeName, namespace, deviceProbe, lvmMgr, tp)
	if err != nil {
		logAndExit(err, "unable to create lvm volume client")
	}

	// setup the volume client with the manager for running volume client
	// cleanup tasks when the manager is stopped.
	err = mgr.Add(volumeClient)
	if err != nil {
		logAndExit(err, "unable to setup volume client with manager")
	}

	// +kubebuilder:scaffold:builder

	recorder := events.NewNoopRecorder()
	if eventRecorderEnabled {
		log.Info("event recorder enabled")
		recorder = mgr.GetEventRecorderFor("local-csi-driver")
	}

	// Create the CSI server.
	csiServer, err := server.NewCombined(csiAddr, driver.NewCombined(nodeName, volumeClient, mgr.GetClient(), ptr.ToBool(cfg.EnforceHyperconvergedWithWebhook), recorder, tp), t)
	if err != nil {
		logAndExit(err, "unable to create csi server")
	}
	if err := mgr.Add(csiServer); err != nil {
		logAndExit(err, "unable to add csi server to internal manager")
	}

	// Register webhooks.
	if ptr.ToBool(cfg.EnforceEphemeralPVC) {
		pvcHandler, err := pvc.NewHandler(volumeClient.GetDriverName(), mgr.GetClient(), mgr.GetScheme(), recorder)
		if err != nil {
			logAndExit(err, "unable to create pvc handler")
		}
		mgr.GetWebhookServer().Register("/pvc-create", &webhook.Admission{Handler: pvcHandler})
	}

	// When node affinity is removed from PVs, we ensure that the PV is bound to
	// the correct node through the hyperconverged webhook.
	if ptr.ToBool(cfg.EnforceHyperconvergedWithWebhook) {
		hyperconvergedHandler, err := hyperconverged.NewHandler(namespace, mgr.GetClient(), mgr.GetScheme())
		if err != nil {
			logAndExit(err, "unable to create hyperconverged handler")
		}
		mgr.GetWebhookServer().Register("/mutate-hyperconverged-pods", &webhook.Admission{Handler: hyperconvergedHandler})
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		logAndExit(err, "unable to set up health check")
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		logAndExit(err, "unable to set up ready check")
	}

	log.Info("starting manager")
	span.AddEvent("starting manager")
	if err := mgr.Start(ctx); err != nil {
		logAndExit(err, "problem running manager")
	}
}

// logAndExit logs the error and exits the program with a non-zero status code.
// It also writes the error message to the termination message file, if possible.
// This is useful for debugging and monitoring purposes.
// The termination message file is used by Kubernetes to display the last error message.
func logAndExit(err error, msg string) {
	logError(err, msg)
	os.Exit(1)
}

func logError(err error, msg string) {
	log.Error(err, msg)
	errMsg := fmt.Sprintf("%s: %v", msg, err)
	parentDir := filepath.Dir(terminationMessagePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		log.Error(err, "failed to create directory for termination message")
		return
	}
	if err := os.WriteFile(terminationMessagePath, []byte(errMsg), 0600); err != nil {
		log.Error(err, "failed to write termination message")
	}
}
