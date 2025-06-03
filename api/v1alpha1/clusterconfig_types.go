// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	configv1alpha1 "k8s.io/component-base/config/v1alpha1"
)

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

// ClusterConfig is the Schema for the ClusterConfig API.
type ClusterConfig struct {
	metav1.TypeMeta `json:",inline"`
	// HealthProbeBindAddress is the TCP address that the controller should bind to
	// for serving health probes
	// +optional
	HealthProbeBindAddress string `json:"healthProbeBindAddress,omitempty"`
	// MetricsBindAddress is the TCP address that the controller should bind to
	// for serving prometheus metrics.
	// +optional
	MetricsBindAddress string `json:"metricsBindAddress,omitempty"`
	// LeaderElection defines the configuration of leader election clients for
	// components that can run with leader election enabled.
	LeaderElection *configv1alpha1.LeaderElectionConfiguration `json:"leaderElection,omitempty"`
	// NodeSelector is the node selector used to select nodes for the diskpool
	// workers.
	NodeSelector map[string]string `json:"nodeselector,omitempty"`
	// Version is the semver version of the Helm chart that installed this
	// configuration.
	Version string `json:"version,omitempty"`
	// Enable the PVC webhook, to enforce generic ephemeral volumes only unless
	// pvc annotations for persistent storage are set.
	EnforceEphemeralPVC *bool `json:"enforceEphemeralPVC,omitempty"`
	// EnforceHyperconvergedWithWebhook enables the webhook that enforces
	// hyperconverged mode for the Local CSI Driver.
	EnforceHyperconvergedWithWebhook *bool `json:"enforceHyperconvergedWithWebhook,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ClusterConfig{})
}
