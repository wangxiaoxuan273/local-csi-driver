// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package events

import (
	"context"
	"testing"
)

func TestIntoContextAndFromContext(t *testing.T) {
	t.Run("should store and retrieve recorder from context", func(t *testing.T) {
		t.Parallel()
		// Create a mock recorder
		mockRecorder := NewNoopObjectRecorder()

		// Store in context
		ctx := context.Background()
		ctxWithRecorder := IntoContext(ctx, mockRecorder)

		// Retrieve from context
		retrievedRecorder := FromContext(ctxWithRecorder)

		// Verify it's the same recorder
		if mockRecorder != retrievedRecorder {
			t.Errorf("Expected retrieved recorder to be the same as the original recorder")
		}
	})

	t.Run("should return NoopObjectRecorder when no recorder in context", func(t *testing.T) {
		t.Parallel()
		// Create empty context
		ctx := context.Background()

		// Retrieve from context with no recorder
		retrievedRecorder := FromContext(ctx)

		// Verify it's a NoopObjectRecorder
		_, ok := retrievedRecorder.(*noopObjectRecorder)
		if !ok {
			t.Errorf("Expected FromContext to return a NoopObjectRecorder when no recorder in context")
		}
	})

}
