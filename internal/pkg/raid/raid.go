// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package raid

import (
	"fmt"

	utilexec "k8s.io/utils/exec"
)

// modules is a list of kernel modules that must be loaded on the host to create
// a RAID0 device.
var modules = []string{"dm-raid", "raid0"}

// Initialize loads the required kernel modules for RAID0 device creation.
func Initialize(exec utilexec.Interface) error {
	for _, module := range modules {
		cmdStr := fmt.Sprintf("nsenter --target 1 --mount --uts --ipc --net --pid modprobe %s", module)
		cmd := exec.Command("sh", "-c", cmdStr)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to load module '%s': %v, output: %s", module, err, string(output))
		}
	}
	return nil
}
