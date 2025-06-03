// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package sys provides utility functions for working with the operating system.
package sys

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"k8s.io/utils/exec"
)

const (
	BlkidCmd = "blkid"
)

// Utils is an interface for OS utility functions.
type Utils interface {
	IsBlockDevice(path string) (bool, error)
	IsBlkDevUnformatted(device string) (bool, error)
	StringToUInt64(sizeStr string) (uint64, error)
}

// sys is a concrete implementation of the Utils interface.
type sys struct {
	exec exec.Interface
}

var _ Utils = &sys{}

func New() *sys {
	return &sys{
		exec: exec.New(),
	}
}

// IsBlockDevice reports whether the given path is a block device.
func (s *sys) IsBlockDevice(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return false, fmt.Errorf("failed to get raw syscall.Stat_t data for %s", path)
	}

	// Check if the file mode is a block device
	if (stat.Mode & syscall.S_IFMT) == syscall.S_IFBLK {
		return true, nil
	}

	return false, nil
}

// IsBlkDevUnformatted reports whether the given block device is unformatted.
// blockDev is the path to the block device e.g. /dev/nvme0n1.
// Uses system utility blkid to get information about the device.
// blkid exit status is:
//   - 0, the device is present and responds to information i.e. it has a filesystem.
//   - 2, device Not present or does not have a filesystem.
//   - 4, usage or other errors.
//   - 8, ambivalent probing result was detected by low-level probing mode (-p).
//
// The function is mainly centered around the exit status as 2 of blkid.
// returns true if the device is unformatted, false otherwise.
func (s *sys) IsBlkDevUnformatted(blockDev string) (bool, error) {
	devPresent, err := s.IsBlockDevice(blockDev)
	if err != nil {
		return false, err
	}
	if devPresent {
		// Run the "blkid" command to get information about the device
		args := []string{"-p", blockDev}
		_, err := s.exec.Command(BlkidCmd, args...).CombinedOutput()
		if err != nil {
			if exit, ok := err.(exec.ExitError); ok {
				if exit.ExitStatus() == 2 {
					// Disk device is unformatted.
					// For `blkid`, if the specified token (TYPE/PTTYPE, etc) was
					// not found, or no (specified) devices could be identified, an
					// exit code of 2 is returned.
					return true, nil
				}
			}
			return false, fmt.Errorf("could not discover filesystem format for device path (%s): %w", blockDev, err)
		}
	}
	return false, nil
}

// parseSize parses a size string (e.g., "16.00g") and returns the size in bytes
// Not used at the moment, will be required during vgStatus implementation when
// we change the size to uint64.
func (s *sys) StringToUInt64(sizeStr string) (uint64, error) {
	sizeStr = strings.ToLower(sizeStr)
	var multiplier uint64 = 1
	switch {
	case strings.HasSuffix(sizeStr, "k"):
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "k")
	case strings.HasSuffix(sizeStr, "m"):
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "m")
	case strings.HasSuffix(sizeStr, "g"):
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "g")
	case strings.HasSuffix(sizeStr, "t"):
		multiplier = 1024 * 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "t")
	}
	size, err := strconv.ParseFloat(sizeStr, 64)
	if err != nil {
		return 0, err
	}
	return uint64(size * float64(multiplier)), nil
}

// parseSize parses a size string (e.g., "16.00g") and returns the size in bytes.
func (s *sys) StringToInt(sizeStr string) (uint64, error) {
	sizeStr = strings.ToLower(sizeStr)
	var multiplier uint64 = 1
	switch {
	case strings.HasSuffix(sizeStr, "k"):
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "k")
	case strings.HasSuffix(sizeStr, "m"):
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "m")
	case strings.HasSuffix(sizeStr, "g"):
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "g")
	case strings.HasSuffix(sizeStr, "t"):
		multiplier = 1024 * 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "t")
	}
	size, err := strconv.ParseFloat(sizeStr, 64)
	if err != nil {
		return 0, err
	}
	return uint64(size * float64(multiplier)), nil
}
