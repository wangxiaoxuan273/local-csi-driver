// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package config

import (
	"context"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	lvmv1alpha1 "local-csi-driver/api/v1alpha1"
)

const (
	// ConfigLabelName is the label name used to identify ConfigMaps containing
	// configuration.
	ConfigLabelName = "local.csi.azure.com/config"

	// ConfigLabelValueCluster is the label value used to identify ConfigMaps
	// containing cluster configuration.
	ConfigLabelValueCluster = "cluster"
)

var (
	// ErrNotFound is returned when a requested config is not found.
	ErrNotFound = fmt.Errorf("config not found")
)

func GetClusterConfig(configFile string) (*lvmv1alpha1.ClusterConfig, error) {
	ctrlConfig := DefaultClusterConfig()
	if configFile != "" {
		// read yaml file
		yamlFile, err := os.ReadFile(configFile)
		if err != nil {
			return nil, err
		}

		// convert yaml to a ClusterConfig object
		err = yaml.Unmarshal(yamlFile, ctrlConfig)
		if err != nil {
			return nil, err
		}
	}

	return ctrlConfig, nil
}

// GetClusterConfigMapName returns the name of the ConfigMap containing the
// cluster configuration in the given namespace.
//
// If multiple ConfigMaps are found, the first one that is not being deleted is
// returned. If all ConfigMaps are being deleted, the last one found is
// returned. If no ConfigMaps are found, an error is returned.
func GetClusterConfigMapName(ctx context.Context, c client.Client, namespace string) (string, error) {
	// Only match ConfigMaps for "cluster" type.
	labelSelector := labels.SelectorFromSet(labels.Set{ConfigLabelName: ConfigLabelValueCluster})

	cms := &corev1.ConfigMapList{}
	err := c.List(ctx, cms, client.InNamespace(namespace), client.MatchingLabelsSelector{Selector: labelSelector})
	if err != nil {
		return "", err
	}

	switch len(cms.Items) {
	case 0:
		return "", fmt.Errorf("cluster config not found in namespace %s: %w", namespace, ErrNotFound)
	case 1:
		return cms.Items[0].Name, nil
	default:
		var last string
		for _, cm := range cms.Items {
			if cm.DeletionTimestamp == nil {
				return cm.Name, nil
			}
			last = cm.Name
		}
		return last, nil
	}
}
