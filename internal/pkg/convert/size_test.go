// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package convert

import (
	"testing"
)

func TestRoundUpMiB(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected int64
	}{
		{
			name:     "zero bytes",
			input:    0,
			expected: 0,
		},
		{
			name:     "one byte",
			input:    1,
			expected: 1,
		},
		{
			name:     "one MiB",
			input:    1024 * 1024,
			expected: 1,
		},
		{
			name:     "one MiB + 1 byte",
			input:    1024*1024 + 1,
			expected: 2,
		},
		{
			name:     "one GiB",
			input:    1024 * 1024 * 1024,
			expected: 1024,
		},
		{
			name:     "one TiB",
			input:    1024 * 1024 * 1024 * 1024,
			expected: 1024 * 1024,
		},
		{
			name:     "non-exact MiB value",
			input:    1500 * 1024 * 1024,
			expected: 1500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RoundUpMiB(tt.input)
			if result != tt.expected {
				t.Errorf("ToMiB(%d) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}
