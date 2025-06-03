// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package retry

import (
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

var (
	// OnError allows the caller to retry fn in case the error returned by fn is
	// retriable according to the provided function. backoff defines the maximum
	// retries and the wait interval between two retries.
	OnError = retry.OnError

	// RetryOnConflict can be used to retry only on conflicts.
	RetryOnConflict = retry.RetryOnConflict

	// WriteBackoff is the recommended retry for a conflict where multiple
	// clients are making changes to the same resource.
	WriteBackoff = wait.Backoff{
		Steps:    10,
		Duration: 10 * time.Millisecond,
		Factor:   1.0,
		Jitter:   0.1,
	}

	// ReadBackoff if the recommended retry for reads where a resource should
	// exist but the cache may not have been updated yet.
	ReadBackoff = wait.Backoff{
		Steps:    20,
		Duration: 10 * time.Millisecond,
		Factor:   1.4,
		Jitter:   0.1,
	}
)

// IsRetriableError returns true if the error is retriable.
func IsRetriableError(err error) bool {
	switch apierrors.ReasonForError(err) {
	case
		metav1.StatusReasonNotFound,
		metav1.StatusReasonUnauthorized,
		metav1.StatusReasonForbidden,
		metav1.StatusReasonAlreadyExists,
		metav1.StatusReasonGone,
		metav1.StatusReasonInvalid,
		metav1.StatusReasonBadRequest,
		metav1.StatusReasonMethodNotAllowed,
		metav1.StatusReasonNotAcceptable,
		metav1.StatusReasonRequestEntityTooLarge,
		metav1.StatusReasonUnsupportedMediaType,
		metav1.StatusReasonExpired:
		return false
	default:
		return true
	}
}

// Always returns true.
func Always(err error) bool {
	return true
}

// IsRetriableOrNotFoundError returns true if the error is retriable or NotFound.
func IsRetriableOrNotFoundError(err error) bool {
	return apierrors.IsNotFound(err) || IsRetriableError(err)
}
