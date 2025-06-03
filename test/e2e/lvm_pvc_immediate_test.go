// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package e2e

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"local-csi-driver/test/pkg/utils"
)

var testLVMImmediate = func() { //nolint:unused
	Context("LVM StorageClass with immediate binding", func() {
		It("should create pv", func() {
			By("creating a storage class")
			cmd := exec.Command("kubectl", "apply", "-f", "test/e2e/fixtures/lvm_storageclass_immediate.yaml")
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				cmd := exec.Command("kubectl", "delete", "-f", "test/e2e/fixtures/lvm_storageclass_immediate.yaml")
				_, err := utils.Run(cmd)
				Expect(err).NotTo(HaveOccurred())
			})

			By("creating a PVC")
			cmd = exec.Command("kubectl", "apply", "-f", "test/e2e/fixtures/lvm_pvc.yaml")
			_, err = utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				cmd := exec.Command("kubectl", "delete", "-f", "test/e2e/fixtures/lvm_pvc.yaml")
				_, err := utils.Run(cmd)
				Expect(err).NotTo(HaveOccurred())
			})

			Eventually(func(g Gomega) {
				cmd = exec.Command("kubectl", "get", "pvc", "lvm", "-o", "jsonpath={.status.phase}")
				status, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(status).To(Equal("Bound"), "PVC should be bound")
			}).Should(Succeed(), "PVC should eventually be bound")
		})

		It("should delete pv", func() {
			By("creating a PVC")
			cmd := exec.Command("kubectl", "apply", "-f", "test/e2e/fixtures/lvm_pvc.yaml")
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("deleting a PVC")
			cmd = exec.Command("kubectl", "delete", "-f", "test/e2e/fixtures/lvm_pvc.yaml")
			_, err = utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func(g Gomega) {
				cmd = exec.Command("kubectl", "get", "pvc", "lvm")
				_, err := utils.Run(cmd)
				g.Expect(err).To(HaveOccurred(), "PVC should be deleted")
			}).Should(Succeed(), "PVC should eventually be deleted")
		})
	})

}
