// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package block

import (
	"reflect"
	"testing"
)

func TestParseLsblkOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    *DeviceList
		wantErr bool
	}{
		{
			name: "Valid lsblk output",
			output: `{
                "blockdevices": [
                    {
                        "name": "sda",
                        "path": "/dev/sda",
                        "maj:min": "8:0",
                        "rm": false,
                        "type": "disk",
                        "mountpoint": "/mnt/data",
                        "mountpoints": ["/mnt/data"],
                        "model": "Samsung SSD",
                        "serial": "123456789",
                        "size": 536870912000
                    },
                    {
                        "name": "sdb",
                        "path": "/dev/sdb",
                        "maj:min": "8:16",
                        "rm": false,
                        "type": "disk",
                        "mountpoint": "/mnt/backup",
                        "mountpoints": ["/mnt/backup"],
                        "model": "WD HDD",
                        "serial": "987654321",
                        "size": 1099511627776
                    },
                    {
                        "name": "sdc",
                        "path": "/dev/sdc",
                        "maj:min": "8:32",
                        "rm": true,
                        "type": "disk",
                        "mountpoint": "/mnt/usb",
                        "mountpoints": ["/mnt/usb"],
                        "model": "SanDisk USB",
                        "serial": "1122334455",
                        "size": 68719476736
                    }
                ]
            }`,
			want: &DeviceList{
				Devices: []Device{
					{
						Name:        "sda",
						Path:        "/dev/sda",
						MajMin:      "8:0",
						Removable:   false,
						Type:        "disk",
						Mountpoint:  "/mnt/data",
						Mountpoints: []string{"/mnt/data"},
						Model:       "Samsung SSD",
						Serial:      "123456789",
						Size:        536870912000,
					},
					{
						Name:        "sdb",
						Path:        "/dev/sdb",
						MajMin:      "8:16",
						Removable:   false,
						Type:        "disk",
						Mountpoint:  "/mnt/backup",
						Mountpoints: []string{"/mnt/backup"},
						Model:       "WD HDD",
						Serial:      "987654321",
						Size:        1099511627776,
					},
					{
						Name:        "sdc",
						Path:        "/dev/sdc",
						MajMin:      "8:32",
						Removable:   true,
						Type:        "disk",
						Mountpoint:  "/mnt/usb",
						Mountpoints: []string{"/mnt/usb"},
						Model:       "SanDisk USB",
						Serial:      "1122334455",
						Size:        68719476736,
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Empty output",
			output:  `{}`,
			want:    &DeviceList{},
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			output:  `{invalid json}`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLsblkOutput([]byte(tt.output))
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLsblkOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseLsblkOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}
