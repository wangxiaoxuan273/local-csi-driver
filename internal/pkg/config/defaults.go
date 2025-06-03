// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package config

import (
	configv1alpha1 "k8s.io/component-base/config/v1alpha1"

	lvmv1alpha1 "local-csi-driver/api/v1alpha1"
)

// Defaults are un-exported so that they can't be used directly. Since they are
// settable via config, the config values must be used instead.
const (
	defaultVersion            = "0.0.0-latest"
	defaultLeaderElectionName = "lvm-leader-election"
)

// DefaultClusterConfig returns the default configuration for the capacity provisioner.
func DefaultClusterConfig() *lvmv1alpha1.ClusterConfig {
	return &lvmv1alpha1.ClusterConfig{
		LeaderElection: &configv1alpha1.LeaderElectionConfiguration{
			LeaderElect:  new(bool),
			ResourceName: defaultLeaderElectionName,
		},
		Version:                          defaultVersion,
		EnforceEphemeralPVC:              new(bool),
		EnforceHyperconvergedWithWebhook: new(bool),
	}
}
