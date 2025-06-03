// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sanity

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/kubernetes-csi/csi-test/v5/pkg/sanity"
)

// Verify VGIDGenerator implements IDGenerator interface.
var _ sanity.IDGenerator = &LvmIDGenerator{}

// LvmIDGenerator implements IDGenerator for LVM volumes
// It generates volume IDs in the format "vgName#volumeName".
type LvmIDGenerator struct {
	sanity.DefaultIDGenerator
	// VolumeGroupName is the volume group name to use in generated IDs
	VolumeGroupName string
}

// NewLVMGenerator creates a new VGIDGenerator with the specified volume group name.
func NewLVMGenerator(volumeGroupName string) (*LvmIDGenerator, error) {
	if len(volumeGroupName) == 0 {
		return nil, fmt.Errorf("volume group name cannot be empty")
	}
	return &LvmIDGenerator{
		VolumeGroupName: volumeGroupName,
	}, nil
}

// GenerateUniqueValidVolumeID generates a unique volume ID in the format "vgName#volumeName".
func (g LvmIDGenerator) GenerateUniqueValidVolumeID() string {
	return fmt.Sprintf("%s#%s", g.VolumeGroupName, uuid.New().String()[:10])
}
