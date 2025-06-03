// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package block

import (
	"context"
	"encoding/json"
	"fmt"

	utilexec "k8s.io/utils/exec"
)

const (
	// lsblkCommand is the command to list block devices.
	lsblkCommand = "lsblk"
)

// Interface defines the methods that block should implement
//
//go:generate mockgen -copyright_file ../../../hack/mockgen_copyright.txt -destination=mock_block.go -mock_names=Interface=Mock -package=block -source=block.go Interface
type Interface interface {
	GetDevices(ctx context.Context) (*DeviceList, error)
}

// block implements the Interface.
type block struct {
	exec utilexec.Interface
}

var _ Interface = &block{}

// New returns a new block instance.
func New(exec utilexec.Interface) Interface {
	return &block{
		exec: exec,
	}
}

// GetDevices runs the lsblk --json command and parses the output.
func (l *block) GetDevices(ctx context.Context) (*DeviceList, error) {
	_, err := l.exec.LookPath(lsblkCommand)
	if err != nil {
		return nil, fmt.Errorf("unable to find %s in PATH: %w", lsblkCommand, err)
	}

	cmd := l.exec.CommandContext(ctx, lsblkCommand, "--bytes", "--json", "--output-all")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command failed: %w, output: %s", err, string(output))
	}

	return parseLsblkOutput(output)
}

// parseLsblkOutput parses the JSON output of the lsblk command.
func parseLsblkOutput(output []byte) (*DeviceList, error) {
	var lsblkOutput DeviceList
	err := json.Unmarshal(output, &lsblkOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse lsblk output: %w", err)
	}
	return &lsblkOutput, nil
}
