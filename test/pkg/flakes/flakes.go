// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package flakes

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
)

var (
	// these can happen when konnectivity agent is failing over during test.
	flakeyTestPatterns = []string{
		"error dialing backend",
		"an error on the server (\"unknown\") has prevented the request from succeeding",
		"error: Internal error occurred: error sending request",
		"connection reset by peer",
		"Internal error occurred: failed calling webhook \"mutate.kyverno.svc-fail\"",
		"unexpected EOF",
	}
)

// PatchReport patches the report to skip flakey tests.
func PatchReport(r Report) Report {
	for i, spec := range r.SpecReports {
		if spec.LeafNodeType != types.NodeTypeIt {
			continue
		}
		if spec.State == types.SpecStatePassed {
			continue
		}
		completeTimeline := reporters.RenderTimeline(spec, false)
		for _, pattern := range flakeyTestPatterns {
			if strings.Contains(completeTimeline, pattern) {
				_, _ = fmt.Fprintf(GinkgoWriter, "Skipping report generation for flakey test: %s\n", spec.FullText())
				r.SpecReports[i].State = types.SpecStateSkipped
				break
			}
		}
	}
	return r
}
