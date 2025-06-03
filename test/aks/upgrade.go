// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aks

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/mod/semver"

	"local-csi-driver/test/pkg/utils"
)

func testClusterUpgrade(tc *testContext) {
	It("application pods should have writable mounts after cluster upgrade", Label("upgrade"), func(ctx context.Context) {
		By("Getting the eligible upgrades")
		cmd := exec.CommandContext(ctx, "az", "aks", "get-upgrades", "-o", "tsv", "--query", "controlPlaneProfile.upgrades[*].kubernetesVersion", "--name", tc.managerCluster, "--resource-group", tc.resourceGroupName)
		output, err := utils.Run(cmd)
		handleAzureError(err, "failed to get eligible upgrades")

		upgrades := utils.GetNonEmptyLines(output)
		if len(upgrades) == 0 {
			Skip("No eligible upgrades found")
		}
		Expect(upgrades).NotTo(BeEmpty(), "No eligible upgrades found")

		version, err := getNextSemver(upgrades)
		Expect(err).NotTo(HaveOccurred(), "Failed to get next semver")

		cmd = exec.CommandContext(ctx, "az", "aks", "upgrade", "--name", tc.managerCluster, "--resource-group", tc.resourceGroupName, "--kubernetes-version", version, "--yes")
		_, err = utils.Run(cmd)
		handleAzureError(err, "failed to upgrade cluster")
	})
}

// the semver go module require "v" prefix on semvers, so we add it if it's missing
// and remove it at the end to match what the AKS RP returns.
func getNextSemver(versions []string) (string, error) {
	semvers := make([]string, 0, len(versions))
	for _, version := range versions {
		if !strings.HasPrefix(version, "v") {
			version = "v" + version
		}
		if ok := semver.IsValid(version); !ok {
			return "", fmt.Errorf("failed to parse version %s", version)
		}
		semvers = append(semvers, version)
	}
	semver.Sort(semvers)
	if len(semvers) == 0 {
		return "", fmt.Errorf("no semver versions found")
	}
	nextVersion, ok := strings.CutPrefix(semvers[0], "v")
	if !ok {
		return "", fmt.Errorf("failed to remove v prefix from version %s", semvers[0])
	}
	return nextVersion, nil
}
