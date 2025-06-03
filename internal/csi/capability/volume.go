// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package capability

import (
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var (
	ErrVolCapabilityMissing               = fmt.Errorf("volume capabilities missing in request")
	ErrVolCapabilityAccessModeInvalid     = fmt.Errorf("invalid access mode")
	ErrVolCapabilityAccessModeUnsupported = fmt.Errorf("unsupported access mode")
	ErrVolCapabilityAccessTypeNotSet      = fmt.Errorf("no access type set")
)

var (
	// validCaps represents how the volume could be accessed. Supported caps are
	// defined in the driver.
	validCaps = []*csi.VolumeCapability_AccessMode{
		{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		},
		{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY,
		},
		{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_SINGLE_WRITER,
		},
		{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_MULTI_WRITER,
		},
		{
			Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
		},
		{
			Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER,
		},
		{
			Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
		},
	}
)

// ValidateVolume checks if the requested volume capabilities are valid and
// supported.
func ValidateVolume(requested []*csi.VolumeCapability, supported []*csi.VolumeCapability_AccessMode) error {
	if err := IsValidVolume(requested); err != nil {
		return err
	}
	if !IsSupportedVolume(requested, supported) {
		return ErrVolCapabilityAccessModeUnsupported
	}
	return nil
}

// IsSupportedVolume checks if the requested volume capabilities are supported.
func IsSupportedVolume(requested []*csi.VolumeCapability, supported []*csi.VolumeCapability_AccessMode) bool {
	if len(requested) == 0 {
		return false
	}
	for _, r := range requested {
		found := false
		for _, s := range supported {
			if r.GetAccessMode().GetMode() == s.GetMode() {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// IsValidVolume checks if the volume capabilities are valid.
func IsValidVolume(requested []*csi.VolumeCapability) error {
	if len(requested) == 0 {
		return ErrVolCapabilityMissing
	}
	if ok := IsValidAccessModes(requested); !ok {
		return ErrVolCapabilityAccessModeInvalid
	}
	for _, c := range requested {
		if c.GetBlock() == nil && c.GetMount() == nil {
			return ErrVolCapabilityAccessTypeNotSet
		}
	}
	return nil
}

// IsValidAccessModes checks if the requested volume capabilities are valid.
func IsValidAccessModes(requested []*csi.VolumeCapability) bool {
	hasSupport := func(cap *csi.VolumeCapability) bool {
		for _, c := range validCaps {
			if c.GetMode() == cap.AccessMode.GetMode() {
				return true
			}
		}
		return false
	}

	foundAll := true
	for _, c := range requested {
		if !hasSupport(c) {
			foundAll = false
		}
	}
	return foundAll
}

// IsValidVolumeCapabilities checks whether the volume capabilities are valid.
func IsValidVolumeCapabilities(volCaps []*csi.VolumeCapability) error {
	if ok := IsValidAccessModes(volCaps); !ok {
		return fmt.Errorf("invalid access mode: %v", volCaps)
	}
	for _, c := range volCaps {
		blockVolume := c.GetBlock()
		mountVolume := c.GetMount()
		accessMode := c.GetAccessMode().GetMode()

		if blockVolume == nil && mountVolume == nil {
			return fmt.Errorf("blockVolume and mountVolume are both nil")
		}

		if blockVolume != nil && mountVolume != nil {
			return fmt.Errorf("blockVolume and mountVolume are both not nil")
		}
		if mountVolume != nil && (accessMode == csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER ||
			accessMode == csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY ||
			accessMode == csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER) {
			return fmt.Errorf("mountVolume is not supported for access mode: %s", accessMode.String())
		}
	}
	return nil
}

func NewNodeServiceCapability(cap csi.NodeServiceCapability_RPC_Type) *csi.NodeServiceCapability {
	return &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}
