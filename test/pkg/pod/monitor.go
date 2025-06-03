// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pod

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"

	corev1 "k8s.io/api/core/v1"
	clientcache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

// AddMonitor creates an event monitor for pod restarts in the specified namespace
// and returns a map of pod names to their restart counts. This gets updated during
// the test run.
func AddMonitor(ctx context.Context, cache cache.Cache, namespace string) (map[string]int32, error) {
	podRestartCount := make(map[string]int32)

	informer, err := cache.GetInformer(ctx, &corev1.Pod{})
	if err != nil {
		return nil, err
	}

	_, err = informer.AddEventHandler(clientcache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj any) {
			newPod, ok := newObj.(*corev1.Pod)
			if !ok || newPod.Namespace != namespace {
				return
			}
			for _, newStatus := range newPod.Status.ContainerStatuses {
				key := newPod.Name + "/" + newStatus.Name
				if newStatus.RestartCount > 0 && podRestartCount[key] != newStatus.RestartCount {
					podRestartCount[key] = newStatus.RestartCount
					_, _ = fmt.Fprintf(GinkgoWriter, "Pod %s in namespace %s has a container with restart count %d", newPod.Name, newPod.Namespace, newStatus.RestartCount)
				}
			}
		},
	})
	if err != nil {
		return nil, err
	}
	return podRestartCount, nil
}
