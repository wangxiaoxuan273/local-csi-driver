// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package version

import (
	"context"
	"fmt"
	"runtime"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

// These are set during build time via -ldflags.
var (
	buildId   = "N/A"
	version   = "N/A"
	gitCommit = "N/A"
	buildDate = "N/A"
)

// Info holds the version information.
type Info struct {
	BuildId   string
	Version   string
	GitCommit string
	BuildDate string
	GoVersion string
	Compiler  string
	Platform  string
}

// GetInfo returns the version information.
func GetInfo() Info {
	return Info{
		BuildId:   buildId,
		Version:   version,
		GitCommit: gitCommit,
		BuildDate: buildDate,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

func Log(log logr.Logger) {
	info := GetInfo()
	log.Info("version info",
		"buildId", info.BuildId,
		"version", info.Version,
		"gitCommit", info.GitCommit,
		"buildDate", info.BuildDate,
		"goVersion", info.GoVersion,
		"compiler", info.Compiler,
		"platform", info.Platform,
	)
}

// Detect is used by OTel to decorate resources with the version info.
func (i Info) Detect(context.Context) (*resource.Resource, error) {
	return resource.NewWithAttributes(
		"",
		attribute.String("service.version", i.Version),
		attribute.String("service.gitCommit", i.GitCommit),
		attribute.String("service.buildId", i.BuildId),
		attribute.String("service.buildDate", i.BuildDate),
	), nil
}
