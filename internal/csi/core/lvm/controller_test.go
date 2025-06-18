// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package lvm_test

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	lvmMgr "local-csi-driver/internal/pkg/lvm"
	"local-csi-driver/internal/pkg/probe"
)

const (
	testVolumeGroup = "test-vg"
)

func TestAvailableCapacity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		vgName      string
		expectLvm   func(*lvmMgr.MockManager)
		expectProbe func(*probe.Mock)
		expectedCap int64
		expectedErr error
	}{
		{
			name:   "no devices found",
			vgName: testVolumeGroup,
			expectLvm: func(m *lvmMgr.MockManager) {
				m.EXPECT().GetVolumeGroup(gomock.Any(), testVolumeGroup).Return(nil, lvmMgr.ErrNotFound)
			},
			expectProbe: func(p *probe.Mock) {
				p.EXPECT().ScanDevices(gomock.Any(), gomock.Any()).Return(nil, probe.ErrNoDevicesFound)
			},
			expectedCap: 0,
		},
		{
			name:   "no devices matching filter",
			vgName: testVolumeGroup,
			expectLvm: func(m *lvmMgr.MockManager) {
				m.EXPECT().GetVolumeGroup(gomock.Any(), testVolumeGroup).Return(nil, lvmMgr.ErrNotFound)
			},
			expectProbe: func(p *probe.Mock) {
				p.EXPECT().ScanDevices(gomock.Any(), gomock.Any()).Return(nil, probe.ErrNoDevicesMatchingFilter)
			},
			expectedCap: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l, p, m, err := initTestLVM(gomock.NewController(t))
			if err != nil {
				t.Fatalf("failed to initialize LVM: %v", err)
			}
			if tt.expectLvm != nil {
				tt.expectLvm(m)
			}
			if tt.expectProbe != nil {
				tt.expectProbe(p)
			}
			cap, err := l.AvailableCapacity(context.Background(), tt.vgName)
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("EnsureVolume() error = %v, expectErr %v", err, tt.expectedErr)
			}
			if cap != tt.expectedCap {
				t.Errorf("EnsureVolume() cap = %d, expectCap %d", cap, tt.expectedCap)
			}
		})
	}

}
