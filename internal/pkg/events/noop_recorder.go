// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package events

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
)

var _ record.EventRecorder = &NoopRecorder{}

// NoopRecorder is a no-op implementation of record.EventRecorder.
type NoopRecorder struct{}

// Event logs an event for the given object.
func (n *NoopRecorder) Event(object runtime.Object, eventtype, reason, message string) {}

// Eventf logs an event with a formatted message for the given object.
func (n *NoopRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...any) {
}

// AnnotatedEventf logs an event with annotations and a formatted message for the given object.
func (n *NoopRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...any) {
}

// NewNoopRecorder creates a new NoopEventRecorder.
func NewNoopRecorder() record.EventRecorder {
	return &NoopRecorder{}
}
