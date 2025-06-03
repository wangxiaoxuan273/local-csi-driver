// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package hostutil

import (
	"fmt"
	"os"
)

type FakeHostUtil struct {
	pathIsDeviceResult map[string]struct {
		isDevice bool
		err      error
	}
}

// NewFakeHostUtil returns a FakeHostUtil object suitable for use in unit tests.
func NewFakeHostUtil() *FakeHostUtil {
	return &FakeHostUtil{
		pathIsDeviceResult: make(map[string]struct {
			isDevice bool
			err      error
		}),
	}
}

// PathIsDevice return whether the path references a block device.
func (f *FakeHostUtil) PathIsDevice(path string) (bool, error) {
	if result, ok := f.pathIsDeviceResult[path]; ok {
		return result.isDevice, result.err
	}

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, fmt.Errorf("path %q does not exist", path)
	}

	return false, err
}

// SetPathIsDeviceResult set the result of calling IsBlockDevicePath for the specified path.
func (f *FakeHostUtil) SetPathIsDeviceResult(path string, isDevice bool, err error) {
	result := struct {
		isDevice bool
		err      error
	}{isDevice, err}

	f.pathIsDeviceResult[path] = result
}
