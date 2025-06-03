// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package convert

const (
	// MiB represents the size of a mebibyte in bytes.
	MiB = 1024 * 1024

	// GiB represents the size of a gibibyte in bytes.
	GiB = 1024 * 1024 * 1024
)

// RoundUpMiB rounds up bytes to mebibytes.
func RoundUpMiB(bytes int64) int64 {
	return roundUpSize(bytes, MiB)
}

// RoundUpBytes rounds up the volume size in bytes upto multiplications of GiB
// in the unit of Bytes.
func RoundUpBytes(volumeSizeBytes int64) int64 {
	return roundUpSize(volumeSizeBytes, GiB) * GiB
}

// RoundUpGiB rounds up the volume size in bytes upto multiplications of GiB
// in the unit of GiB.
func RoundUpGiB(volumeSizeBytes int64) int64 {
	return roundUpSize(volumeSizeBytes, GiB)
}

// BytesToGiB conversts Bytes to GiB.
func BytesToGiB(volumeSizeBytes int64) int64 {
	return volumeSizeBytes / GiB
}

func BytesToMiB(volumeSizeBytes int64) int64 {
	return volumeSizeBytes / MiB
}

// GiBToBytes converts GiB to Bytes.
func GiBToBytes(volumeSizeGiB int64) int64 {
	return volumeSizeGiB * GiB
}

func MiBToBytes(volumeSizeMiB int64) int64 {
	return volumeSizeMiB * MiB
}

// roundUpSize calculates how many allocation units are needed to accommodate
// a volume of given size. E.g. when user wants 1500MiB volume, while Azure File
// allocates volumes in gibibyte-sized chunks,
// RoundUpSize(1500 * 1024*1024, 1024*1024*1024) returns '2'
// (2 GiB is the smallest allocatable volume that can hold 1500MiB).
func roundUpSize(volumeSizeBytes int64, allocationUnitBytes int64) int64 {
	roundedUp := volumeSizeBytes / allocationUnitBytes
	if volumeSizeBytes%allocationUnitBytes > 0 {
		roundedUp++
	}
	return roundedUp
}
