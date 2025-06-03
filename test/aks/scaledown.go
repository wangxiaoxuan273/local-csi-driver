// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aks

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	"local-csi-driver/test/pkg/utils"
)

const (
	// maxScaleBy is the number of nodes to scale down to.
	maxScaleBy = 2
	// minimumNodeCount is the minimum number of nodes to scale down to.
	minimumNodeCount = 2
	// postScaleDownWaitTime is the time to wait after scaling down.
	postScaleDownWaitTime = time.Minute
)

// testClusterScaledown tests that application pods have writable mounts after cluster scaledown and up.
func testClusterScaledown(tc *testContext) {
	It("application pods should have writable mounts after cluster scaledown and up", Label("scaledown"), func(ctx context.Context) {
		var nodes corev1.NodeList
		Eventually(tc.List).WithArguments(ctx, &nodes).Should(Succeed(), "failed to list nodes")
		Expect(nodes.Items).NotTo(BeEmpty(), "no nodes found")
		if len(nodes.Items) < minimumNodeCount {
			Skip(fmt.Sprintf("We need at least %d nodes to test scale down", minimumNodeCount))
		}

		initialNodeCount := len(nodes.Items)

		// Note: this runs after the afterEach block. So the scale down will have afterEach verifications run on it
		DeferCleanup(func(ctx context.Context) {
			handleAzureError(scaleCluster(ctx, tc, initialNodeCount), "failed to scale cluster back up")
			Eventually(func(g Gomega, ctx context.Context) {
				var nodes corev1.NodeList
				g.Expect(tc.List(ctx, &nodes)).To(Succeed(), "failed to list nodes")
				g.Expect(nodes.Items).To(HaveLen(initialNodeCount), "failed to scale cluster back up")
			}).WithContext(ctx).Should(Succeed(), "failed to wait for cluster to scale back up")
		})

		for scaleBy := 1; scaleBy <= maxScaleBy; scaleBy++ {
			newNodeCount := initialNodeCount - scaleBy
			if newNodeCount < 1 {
				break
			}
			handleAzureError(scaleCluster(ctx, tc, newNodeCount), "failed to scale down cluster")

			Expect(utils.Sleep(ctx, postScaleDownWaitTime)).To(Succeed(), "context timed out during sleep")

			Eventually(func(g Gomega, ctx context.Context) {
				var nodes corev1.NodeList
				g.Expect(tc.List(ctx, &nodes)).To(Succeed(), "failed to list nodes")
				g.Expect(nodes.Items).To(HaveLen(newNodeCount), "failed to scale down cluster")
			}).WithContext(ctx).Should(Succeed(), "failed to wait for cluster to scale down")
		}
	})
}

func scaleCluster(ctx context.Context, tc *testContext, i int) error {
	if i < 1 {
		return fmt.Errorf("invalid node count: %d", i)
	}
	By("Scaling the cluster to " + strconv.Itoa(i))
	cmd := exec.CommandContext(ctx, "az", "aks", "scale", "--name", tc.managerCluster, "--resource-group", tc.resourceGroupName, "--node-count", strconv.Itoa(i))
	_, err := utils.Run(cmd)
	return err
}
