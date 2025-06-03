// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package e2e

import (
	"context"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"local-csi-driver/test/pkg/common"
	"local-csi-driver/test/pkg/utils"
)

var testLvmStoragePool = func() {
	Context("lvm controllers", Label("aks", "lvm"), func() {
		// BeforeAll(func(ctx context.Context) {

		// 	By("Creating storagepool")
		// 	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", common.LvmStoragePoolFixture)
		// 	_, err := utils.Run(cmd)
		// 	Expect(err).NotTo(HaveOccurred(), "Failed to apply storagepool")

		// 	DeferCleanup(func(ctx context.Context) {
		// 		By("Deleting storagepool")
		// 		cmd := exec.CommandContext(ctx, "kubectl", "delete", "--wait", "--ignore-not-found", "-f", common.LvmStoragePoolFixture)
		// 		_, err := utils.Run(cmd)
		// 		Expect(err).NotTo(HaveOccurred(), "Failed to delete storagepool")
		// 		By("Waiting for diskpools to be Available")
		// 		EventuallyDiskpoolPhase("Available").WithContext(ctx).Should(Succeed(), "Diskpools should eventually be Available again")
		// 	})

		// 	By("validating that diskpools are created for every l-series node")
		// 	Eventually(func(g Gomega, ctx context.Context) {
		// 		cmd := exec.Command("kubectl", "get",
		// 			"node",
		// 			"-l", "node.kubernetes.io/instance-type in (standard_l8s_v3, standard_l16s_v3, standard_l32s_v3, "+
		// 				"standard_l48s_v3, standard_l64s_v3, standard_l80s_v3)",
		// 			"-o", "go-template={{ range .items }}"+
		// 				"{{ .metadata.name }}"+
		// 				"{{ \"\\n\" }}{{ end }}",
		// 		)
		// 		nodeOutput, err := utils.Run(cmd)
		// 		g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve node information")
		// 		nodeNames := utils.GetNonEmptyLines(nodeOutput)
		// 		cmd = exec.CommandContext(ctx, "kubectl", "get",
		// 			"diskpools.local.storage.azure.com",
		// 			"-o", "go-template={{ range .items }}"+
		// 				"{{ .spec.nodeName }}"+
		// 				"{{ \"\\n\" }}{{ end }}",
		// 			"-n", namespace,
		// 		)
		// 		diskpoolOutput, err := utils.Run(cmd)
		// 		g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve diskpool information")
		// 		diskpoolNodesNames := utils.GetNonEmptyLines(diskpoolOutput)
		// 		g.Expect(diskpoolNodesNames).To(ConsistOf(nodeNames), "All nodes should be in diskpool")
		// 		g.Expect(diskpoolNodesNames).To(HaveLen(len(nodeNames)), "All nodes should be in diskpool")
		// 	}).WithContext(ctx).Should(Succeed(), "Diskpools should eventually be created for every l-series node")

		// 	By("Waiting for storagepool to be ready")
		// 	cmd = exec.CommandContext(ctx, "kubectl", "wait", "--for=condition=Ready", "--timeout=5m", "-f", common.LvmStoragePoolFixture)
		// 	_, err = utils.Run(cmd)
		// 	Expect(err).NotTo(HaveOccurred(), "Failed to wait for storagepool")

		// 	By("Waiting for diskpools to be Bound")
		// 	EventuallyDiskpoolPhase("Bound").WithContext(ctx).Should(Succeed(), "Diskpools should eventually be Bound")

		// 	EventuallyUsedAmount("0").WithContext(ctx).Should(Succeed(), "Failed to verify storagepool used capacity")
		// })

		It("should create storageclass", func(ctx context.Context) {
			cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", common.LvmStorageClassFixture)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred(), "Failed to apply storageclass")

			Eventually(func(g Gomega, ctx context.Context) {
				cmd := exec.CommandContext(ctx, "kubectl", "get", "storageclass", "local")
				_, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred(), "local storageclass does not exist")
			}).WithContext(ctx).Should(Succeed(), "Failed to create storageclass")
		})

		lvmStatefulsetTest("should create statefulset with local storagepool", common.LvmStatefulSetFixture)
		lvmStatefulsetTest("should create annotation statefulset with local storagepool", common.LvmAnnotationStatefulSetFixture)
	})

}

func lvmStatefulsetTest(name, statefulsetFixture string) {
	It(name, func(ctx context.Context) {
		cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", statefulsetFixture)
		_, err := utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to apply statefulset")

		DeferCleanup(func(ctx context.Context) {
			By("Deleting statefulset")
			cmd := exec.CommandContext(ctx, "kubectl", "delete", "--wait", "--ignore-not-found", "-f", statefulsetFixture)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred(), "Failed to delete statefulset")

			cmd = exec.CommandContext(ctx, "kubectl", "delete", "--wait", "--ignore-not-found", "pvc", "-l", "part-of=e2e-test")
			_, err = utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred(), "Failed to delete pvc")
		})

		By("Waiting for statefulset to be ready")
		cmd = exec.CommandContext(ctx, "kubectl", "rollout", "status", "--timeout=5m", "-f", statefulsetFixture)
		_, err = utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to wait for statefulset")
	})
}
