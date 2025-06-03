// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package server

import (
	"fmt"
	"os"
	"strings"
)

var (
	// remove is a variable that holds the function to remove a file. It allows
	// tests to replace it with a mock function.
	remove = os.Remove
)

// parseEndpoint parses the endpoint into protocol and address.
func parseEndpoint(ep string) (string, string, error) {
	if strings.HasPrefix(strings.ToLower(ep), "unix://") || strings.HasPrefix(strings.ToLower(ep), "tcp://") {
		s := strings.SplitN(ep, "://", 2)
		if s[1] != "" {
			return s[0], s[1], nil
		}
	}
	return "", "", fmt.Errorf("invalid endpoint: %v", ep)
}

// prepareEndpoint makes sure the endpoint is ready to be used.
func prepareEndpoint(proto string, addr string) error {
	if proto == "unix" {
		addr = "/" + addr
		if err := remove(addr); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove %s: %w", addr, err)
		}
	}
	return nil
}
