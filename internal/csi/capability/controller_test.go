// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package capability

import (
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

func TestValidateController(t *testing.T) {
	tests := []struct {
		name      string
		c         csi.ControllerServiceCapability_RPC_Type
		supported []*csi.ControllerServiceCapability
		wantErr   bool
	}{
		{
			name: "valid capability",
			c:    csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
			supported: []*csi.ControllerServiceCapability{
				{
					Type: &csi.ControllerServiceCapability_Rpc{
						Rpc: &csi.ControllerServiceCapability_RPC{
							Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "unsupported capability",
			c:    csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
			supported: []*csi.ControllerServiceCapability{
				{
					Type: &csi.ControllerServiceCapability_Rpc{
						Rpc: &csi.ControllerServiceCapability_RPC{
							Type: csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name:      "empty supported capabilities",
			c:         csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
			supported: []*csi.ControllerServiceCapability{},
			wantErr:   true,
		},
		{
			name:      "nil supported capabilities",
			c:         csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
			supported: nil,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateController(tt.c, tt.supported); (err != nil) != tt.wantErr {
				t.Errorf("ValidateController() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
