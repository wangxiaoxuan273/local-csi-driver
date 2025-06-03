// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package docker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"local-csi-driver/test/pkg/utils"
)

// PushImage pushes a local docker image to the registry.
func PushImage(ctx context.Context, name string) error {
	dockerOptions := []string{"push", name}
	cmd := exec.CommandContext(ctx, "docker", dockerOptions...)
	output, err := utils.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to push image: %s, error: %w", output, err)
	}
	return nil
}

// PullImage pulls a docker image from the registry.
func PullImage(ctx context.Context, name string) error {
	dockerOptions := []string{"pull", name}
	cmd := exec.CommandContext(ctx, "docker", dockerOptions...)
	output, err := utils.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to pull image: %s, error: %w", output, err)
	}
	return nil
}

// CheckImageExists checks if a docker image exists in the local registry.
func CheckImageExists(ctx context.Context, name string) (bool, error) {
	dockerOptions := []string{"image", "inspect", name}
	cmd := exec.CommandContext(ctx, "docker", dockerOptions...)
	output, err := utils.Run(cmd)
	if err != nil && strings.Contains(output, "No such image") {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
