// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package version

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

func TestGetVersion(t *testing.T) {
	version := GetInfo()

	expected := Info{
		BuildId:   "N/A",
		GitCommit: "N/A",
		BuildDate: "N/A",
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	if !reflect.DeepEqual(version, expected) {
		t.Errorf("Unexpected error. \n Expected: %#v \n Found: %#v", expected, version)
	}

}
