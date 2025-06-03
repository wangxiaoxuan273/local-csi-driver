// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package probe

import (
	"strings"

	"local-csi-driver/internal/pkg/block"
)

// EphemeralDiskFilter is a filter for ephemeral disks.
var EphemeralDiskFilter = &Filter{
	Filters: []FilterPredicate{
		&PathFilter{Path: "/dev/nvme"},
		NewModelFilter("Microsoft NVMe Direct Disk", "Microsoft NVMe Direct Disk v2"),
		&TypeFilter{Type: "disk"},
	},
}

// FilterPredicate defines a predicate for filtering devices.
type FilterPredicate interface {
	Match(device block.Device) bool
}

// Filter holds multiple filters and matches if all contained filters match.
type Filter struct {
	Filters []FilterPredicate
}

func (f *Filter) Match(device block.Device) bool {
	for _, filter := range f.Filters {
		if !filter.Match(device) {
			return false
		}
	}
	return true
}

// PathFilter matches devices by path prefix.
type PathFilter struct {
	Path string
}

func (f *PathFilter) Match(device block.Device) bool {
	return strings.HasPrefix(device.Path, f.Path)
}

// TypeFilter matches devices by type.
type TypeFilter struct {
	Type string
}

func (f *TypeFilter) Match(device block.Device) bool {
	return strings.EqualFold(device.Type, f.Type)
}

func NewModelFilter(models ...string) *ModelFilter {
	modelsMap := make(map[string]struct{}, len(models))
	for _, model := range models {
		model = strings.TrimSpace(model)
		model = strings.ToLower(model)
		modelsMap[model] = struct{}{}
	}
	return &ModelFilter{models: modelsMap}
}

// ModelFilter matches devices by model.
type ModelFilter struct {
	models map[string]struct{}
}

func (f *ModelFilter) Match(device block.Device) bool {
	deviceModel := strings.TrimSpace(device.Model)
	deviceModel = strings.ToLower(deviceModel)
	_, exists := f.models[deviceModel]
	return exists
}
