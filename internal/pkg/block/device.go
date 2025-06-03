// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package block

// Device represents a block device in the lsblk output.
type Device struct {
	Name        string   `json:"name"`
	Path        string   `json:"path,omitempty"`
	MajMin      string   `json:"maj:min,omitempty"`
	Removable   bool     `json:"rm,omitempty"`
	Type        string   `json:"type,omitempty"`
	Mountpoint  string   `json:"mountpoint,omitempty"`
	Mountpoints []string `json:"mountpoints,omitempty"`
	Model       string   `json:"model,omitempty"`
	Serial      string   `json:"serial,omitempty"`
	Size        int64    `json:"size,omitempty"`
}

// DeviceList represents the output of the lsblk command.
type DeviceList struct {
	Devices []Device `json:"blockdevices"`
}
