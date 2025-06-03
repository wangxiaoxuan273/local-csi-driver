// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package probe

import (
	"context"
	"fmt"
	"sort"

	"github.com/go-logr/logr"

	"local-csi-driver/internal/pkg/block"
)

var (
	// ErrNoDevicesFound is returned when no devices are found.
	ErrNoDevicesFound = fmt.Errorf("no devices found")

	// ErrNoDevicesMatchingFilter is returned when no devices match the filter.
	ErrNoDevicesMatchingFilter = fmt.Errorf("no devices matching filter found")
)

//go:generate mockgen -copyright_file ../../../hack/mockgen_copyright.txt -destination=mock_probe.go -mock_names=Interface=Mock -package=probe -source=probe.go Interface
type Interface interface {
	ScanDevices(ctx context.Context, log logr.Logger) ([]string, error)
	GetDevices(ctx context.Context) (*block.DeviceList, error)
}

var _ Interface = &deviceScanner{}

// deviceScanner is a struct that implements the DeviceScanner interface.
type deviceScanner struct {
	block.Interface
	filter *Filter
}

func New(b block.Interface, f *Filter) Interface {
	return &deviceScanner{b, f}
}

func (m *deviceScanner) ScanDevices(ctx context.Context, log logr.Logger) ([]string, error) {
	devices, err := m.GetDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}
	if len(devices.Devices) == 0 {
		return nil, ErrNoDevicesFound
	}

	// we do not know the size of the slice, it is likely to be small
	// so we do not preallocate it
	var paths []string //nolint:prealloc
	for _, device := range devices.Devices {
		if !m.filter.Match(device) {
			log.V(2).Info("device filtered out", "device", device)
			continue
		}
		paths = append(paths, device.Path)
		log.V(1).Info("device found", "device", device)
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		return nil, ErrNoDevicesMatchingFilter
	}
	return paths, nil
}
