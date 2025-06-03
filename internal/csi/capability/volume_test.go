// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package capability

import (
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

func TestValidateVolume(t *testing.T) {
	tests := []struct {
		name      string
		requested []*csi.VolumeCapability
		supported []*csi.VolumeCapability_AccessMode
		wantErr   bool
	}{
		{
			name: "valid single node writer mount",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			supported: []*csi.VolumeCapability_AccessMode{
				{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
				},
			},
			wantErr: false,
		},
		{
			name: "unsupported multi node reader mount",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
					},
				},
			},
			supported: []*csi.VolumeCapability_AccessMode{
				{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
				},
			},
			wantErr: true,
		},
		{
			name: "supported multi node reader mount",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
					},
				},
			},
			supported: []*csi.VolumeCapability_AccessMode{
				{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
				},
				{
					Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
				},
			},
			wantErr: false,
		},
		{
			name: "valid single node writer block",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Block{
						Block: &csi.VolumeCapability_BlockVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			supported: []*csi.VolumeCapability_AccessMode{
				{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
				},
			},
			wantErr: false,
		},
		{
			name: "unsupported multi node reader block",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Block{
						Block: &csi.VolumeCapability_BlockVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
					},
				},
			},
			supported: []*csi.VolumeCapability_AccessMode{
				{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
				},
			},
			wantErr: true,
		},
		{
			name: "supported multi node reader block",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Block{
						Block: &csi.VolumeCapability_BlockVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
					},
				},
			},
			supported: []*csi.VolumeCapability_AccessMode{
				{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
				},
				{
					Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
				},
			},
			wantErr: false,
		},
		{
			name:      "empty requested capabilities",
			requested: []*csi.VolumeCapability{},
			supported: []*csi.VolumeCapability_AccessMode{
				{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
				},
			},
			wantErr: true,
		},
		{
			name: "empty supported capabilities",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			supported: []*csi.VolumeCapability_AccessMode{},
			wantErr:   true,
		},
		{
			name:      "nil requested capabilities",
			requested: nil,
			supported: []*csi.VolumeCapability_AccessMode{
				{
					Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
				},
			},
			wantErr: true,
		},
		{
			name: "nil supported capabilities",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			supported: nil,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateVolume(tt.requested, tt.supported); (err != nil) != tt.wantErr {
				t.Errorf("ValidateVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidVolume(t *testing.T) {
	tests := []struct {
		name      string
		requested []*csi.VolumeCapability
		wantErr   error
	}{
		{
			name: "valid mount capabilities",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid mount access mode",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: 10,
					},
				},
			},
			wantErr: ErrVolCapabilityAccessModeInvalid,
		},
		{
			name: "valid block capabilities",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Block{
						Block: &csi.VolumeCapability_BlockVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid block access mode",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Block{
						Block: &csi.VolumeCapability_BlockVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: 10,
					},
				},
			},
			wantErr: ErrVolCapabilityAccessModeInvalid,
		},
		{
			name:      "empty request",
			requested: []*csi.VolumeCapability{},
			wantErr:   ErrVolCapabilityMissing,
		},
		{
			name:    "nil request",
			wantErr: ErrVolCapabilityMissing,
		},
		{
			name: "empty volume capability",
			requested: []*csi.VolumeCapability{
				{
					AccessType: nil,
					AccessMode: nil,
				},
			},
			wantErr: ErrVolCapabilityAccessModeInvalid,
		},
		{
			name: "empty access type",
			requested: []*csi.VolumeCapability{
				{
					AccessType: nil,
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			wantErr: ErrVolCapabilityAccessTypeNotSet,
		},
		{
			name: "empty access mode",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: nil,
				},
			},
			wantErr: ErrVolCapabilityAccessModeInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := IsValidVolume(tt.requested); err != tt.wantErr {
				t.Errorf("IsValidVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidVolumeCapabilities(t *testing.T) {
	tests := []struct {
		name      string
		requested []*csi.VolumeCapability
		maxShares int
		wantErr   bool
	}{
		{
			name: "valid mount access mode",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "unsupported mount access mode",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid mount access mode",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: 10,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid block capabilities",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Block{
						Block: &csi.VolumeCapability_BlockVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid block access mode",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Block{
						Block: &csi.VolumeCapability_BlockVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid block access mode",
			requested: []*csi.VolumeCapability{
				{
					AccessType: &csi.VolumeCapability_Block{
						Block: &csi.VolumeCapability_BlockVolume{},
					},
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: 10,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty volume capability",
			requested: []*csi.VolumeCapability{
				{
					AccessType: nil,
					AccessMode: nil,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := IsValidVolumeCapabilities(tt.requested); (err != nil) != tt.wantErr {
				t.Errorf("IsValidVolumeCapabilities() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
