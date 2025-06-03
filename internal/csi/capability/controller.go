// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package capability

import (
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

// ValidateController checks if the requested controller capability is valid and supported.
func ValidateController(c csi.ControllerServiceCapability_RPC_Type, supported []*csi.ControllerServiceCapability) error {
	if c == csi.ControllerServiceCapability_RPC_UNKNOWN {
		return nil
	}

	for _, cap := range supported {
		if c == cap.GetRpc().GetType() {
			return nil
		}
	}
	return fmt.Errorf("unsupported capability %s", c)
}

func NewControllerServiceCapability(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
	return &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}
