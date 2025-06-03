// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package filesystem

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"

	corev1 "k8s.io/api/core/v1"
	clientcache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

const (
	reason = "FilesystemIsReadOnly"
)

// AddMonitor creates an event monitor that logs read-only filesystem events.
func AddMonitor(ctx context.Context, cache cache.Cache) error {
	informer, err := cache.GetInformer(ctx, &corev1.Event{})
	if err != nil {
		return err
	}
	startTime := time.Now()
	_, err = informer.AddEventHandler(clientcache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			if event, ok := obj.(*corev1.Event); ok && event.Reason == reason && event.CreationTimestamp.After(startTime) {
				_, _ = fmt.Fprintf(GinkgoWriter, "read-only filesystem event: type=%s, kind=%s, name=%s, message=%s", event.Type, event.InvolvedObject.Kind, event.InvolvedObject.Name, event.Message)
			}
		},
	})
	return err
}
