// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package node

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientcache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AddMonitor creates an event monitor that logs read-only filesystem events.
func AddMonitor(ctx context.Context, client client.Client, cache cache.Cache, namespaces []string) error {
	informer, err := cache.GetInformer(ctx, &corev1.Node{})
	if err != nil {
		return err
	}
	namespacesMap := make(map[string]struct{})
	for _, ns := range namespaces {
		namespacesMap[ns] = struct{}{}
	}
	_, err = informer.AddEventHandler(clientcache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			node, ok := obj.(*corev1.Node)
			if !ok {
				return
			}
			_, _ = fmt.Fprintf(GinkgoWriter, "Node %s added\n", node.Name)
		},
		UpdateFunc: func(oldObj, newObj any) {
			newNode, ok := newObj.(*corev1.Node)
			if !ok {
				return
			}
			oldNode, ok := oldObj.(*corev1.Node)
			if !ok {
				return
			}
			if oldNode.Spec.Unschedulable != newNode.Spec.Unschedulable {
				_, _ = fmt.Fprintf(GinkgoWriter, "Node marked unschedulable to %t: %s\n", newNode.Spec.Unschedulable, newNode.Name)
				logPodStatus(ctx, client, newNode.Name, namespacesMap)
			}

			oldReadyStatus, _ := getReadyConditionStatus(oldNode.Status.Conditions)
			newReadyStatus, newReason := getReadyConditionStatus(newNode.Status.Conditions)
			if oldReadyStatus != newReadyStatus {
				_, _ = fmt.Fprintf(GinkgoWriter, "Node %s ready status changed from %s to %s: %s\n", newNode.Name, oldReadyStatus, newReadyStatus, newReason)
			}
		},
		DeleteFunc: func(obj any) {
			node, ok := obj.(*corev1.Node)
			if !ok {
				return
			}
			_, _ = fmt.Fprintf(GinkgoWriter, "Node %s deleted\n", node.Name)
			logPodStatus(ctx, client, node.Name, namespacesMap)
		},
	})
	return err
}

// getReadyConditionStatus returns the status of the condition type.
func getReadyConditionStatus(conditions []corev1.NodeCondition) (corev1.ConditionStatus, string) {
	for _, condition := range conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status, condition.Reason
		}
	}
	return corev1.ConditionUnknown, ""
}

// logPodStatus logs the status of pods on the node.
func logPodStatus(ctx context.Context, c client.Client, nodeName string, namespaces map[string]struct{}) {
	pods := &corev1.PodList{}
	err := c.List(ctx, pods, client.InNamespace(metav1.NamespaceAll), client.MatchingFields{"spec.nodeName": nodeName})
	if err != nil {
		_, _ = fmt.Fprintf(GinkgoWriter, "Failed to list pods on node %s: %v\n", nodeName, err)
		return
	}

	podStatuses := make(map[string]string)
	for _, pod := range pods.Items {
		if _, ok := namespaces[pod.Namespace]; !ok {
			continue
		}
		podStatuses[pod.Name] = getPodStatus(pod)

	}
	_, _ = fmt.Fprintf(GinkgoWriter, "Pods on node %s:\n", nodeName)
	for podName, status := range podStatuses {
		_, _ = fmt.Fprintf(GinkgoWriter, "- %s: %s\n", podName, status)
	}
}

func getPodStatus(pod corev1.Pod) string {
	if pod.DeletionTimestamp != nil {
		return "Terminating"
	}
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.State.Waiting != nil {
			if containerStatus.State.Waiting.Reason == "CrashLoopBackOff" {
				return "CrashLoopBackOff"
			}
			if containerStatus.State.Waiting.Reason == "ContainerCreating" {
				return "ContainerCreating"
			}
		}
	}
	return string(pod.Status.Phase)
}
