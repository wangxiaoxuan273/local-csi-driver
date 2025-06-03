// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aks

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"local-csi-driver/test/pkg/utils"
)

const (
	timeToSleepBeforeStop = time.Minute
)

// testClusterRestart stops and starts the cluster.
func testClusterRestart(tc *testContext) {
	It("application pods should have writable mounts after cluster restart", Label("restart"), func(ctx context.Context) {
		By(fmt.Sprintf("Sleeping for %s prior to the so logs get pushed to Kusto", timeToSleepBeforeStop))

		Expect(utils.Sleep(ctx, timeToSleepBeforeStop)).To(Succeed(), "context timed out during sleep")

		By("Stopping the cluster")
		cmd := exec.CommandContext(ctx, "az", "aks", "stop", "--name", tc.managerCluster, "--resource-group", tc.resourceGroupName)
		_, err := utils.Run(cmd)
		handleAzureError(err, "failed to stop cluster")

		By("Starting the cluster")
		cmd = exec.CommandContext(ctx, "az", "aks", "start", "--name", tc.managerCluster, "--resource-group", tc.resourceGroupName)
		_, err = utils.Run(cmd)
		handleAzureError(err, "failed to start cluster")
	})
}
