// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"local-csi-driver/test/pkg/utils"
)

// serviceAccountName created for the project.
const serviceAccountName = "cns-cluster-manager" //nolint:unused

// metricsServiceName is the name of the metrics service of the project.
const metricsServiceName = "cns-cluster-manager-metrics-service" //nolint:unused

// metricsRoleBindingName is the name of the RBAC that will be created to allow get the metrics data.
const metricsRoleBindingName = "cns-metrics-binding" //nolint:unused

var testMetrics = func() {
	Context("Metrics", func() {
		if os.Getenv("SKIP_METRICS") != "" {
			When("SKIP_METRICS is set", func() {
				It("should skip metrics tests", func() {
					Skip("Skipping metrics tests")
				})
			})
			return
		}

		// TODO: flaky test
		// It("should be properly installed", func() {
		// 	By("ensuring a ClusterRoleBinding exists for the service account to access metrics")
		// 	cmd := exec.Command("kubectl", "get", "clusterrolebinding", metricsRoleBindingName)
		// 	if _, err := utils.Run(cmd); err != nil {
		// 		cmd := exec.Command("kubectl", "create", "clusterrolebinding", metricsRoleBindingName,
		// 			"--clusterrole=cns-metrics-reader",
		// 			fmt.Sprintf("--serviceaccount=%s:%s", namespace, serviceAccountName),
		// 		)
		// 		_, err := utils.Run(cmd)
		// 		Expect(err).NotTo(HaveOccurred(), "Failed to create ClusterRoleBinding")
		// 	}

		// 	By("validating that the metrics service is available")
		// 	cmd = exec.Command("kubectl", "get", "service", metricsServiceName, "-n", namespace)
		// 	_, err := utils.Run(cmd)
		// 	Expect(err).NotTo(HaveOccurred(), "Metrics service should exist")

		// 	By("validating that the ServiceMonitor for Prometheus is applied in the namespace")
		// 	cmd = exec.Command("kubectl", "get", "ServiceMonitor", "-n", namespace)
		// 	_, err = utils.Run(cmd)
		// 	Expect(err).NotTo(HaveOccurred(), "ServiceMonitor should exist")

		// 	By("getting the service account token")
		// 	token, err := serviceAccountToken()
		// 	Expect(err).NotTo(HaveOccurred())
		// 	Expect(token).NotTo(BeEmpty())

		// 	By("waiting for the metrics endpoint to be ready")
		// 	verifyMetricsEndpointReady := func(g Gomega) {
		// 		cmd := exec.Command("kubectl", "get", "endpoints", metricsServiceName, "-n", namespace)
		// 		output, err := utils.Run(cmd)
		// 		g.Expect(err).NotTo(HaveOccurred())
		// 		g.Expect(output).To(ContainSubstring("8443"), "Metrics endpoint is not ready")
		// 	}
		// 	Eventually(verifyMetricsEndpointReady).Should(Succeed())

		// 	By("verifying that the controller manager is serving the metrics server")
		// 	Skip("Skipping flaky metrics test")
		// 	verifyMetricsServerStarted := func(g Gomega) {
		// 		cmd := exec.Command("kubectl", "logs", managerPodName, "-n", namespace)
		// 		output, err := utils.Run(cmd)
		// 		g.Expect(err).NotTo(HaveOccurred())
		// 		g.Expect(output).To(ContainSubstring("controller-runtime.metrics\tServing metrics server"),
		// 			"Metrics server not yet started")
		// 	}
		// 	Eventually(verifyMetricsServerStarted).Should(Succeed())

		// 	By("creating the curl-metrics pod to access the metrics endpoint")
		// 	Skip("Skipping flaky metrics test")
		// 	cmd = exec.Command("kubectl", "run", "curl-metrics", "--restart=Never",
		// 		"--namespace", namespace,
		// 		"--image=curlimages/curl:7.78.0",
		// 		"--", "/bin/sh", "-c", fmt.Sprintf(
		// 			"curl -v -k -H 'Authorization: Bearer %s' https://%s.%s.svc.cluster.local:8443/metrics",
		// 			token, metricsServiceName, namespace))
		// 	_, err = utils.Run(cmd)
		// 	Expect(err).NotTo(HaveOccurred(), "Failed to create curl-metrics pod")

		// 	By("waiting for the curl-metrics pod to complete.")
		// 	Skip("Skipping flaky metrics test")
		// 	verifyCurlUp := func(g Gomega) {
		// 		cmd := exec.Command("kubectl", "get", "pods", "curl-metrics",
		// 			"-o", "jsonpath={.status.phase}",
		// 			"-n", namespace)
		// 		output, err := utils.Run(cmd)
		// 		g.Expect(err).NotTo(HaveOccurred())
		// 		g.Expect(output).To(Equal("Succeeded"), "curl pod in wrong status")
		// 	}
		// 	Eventually(verifyCurlUp, 5*time.Minute).Should(Succeed())

		// 	By("getting the metrics by checking curl-metrics logs")
		// 	Skip("Skipping flaky metrics test")
		// 	metricsOutput := getMetricsOutput()
		// 	Expect(metricsOutput).To(ContainSubstring(
		// 		"controller_runtime_reconcile_total",
		// 	))
		// })
	})
}

// serviceAccountToken returns a token for the specified service account in the given namespace.
// It uses the Kubernetes TokenRequest API to generate a token by directly sending a request
// and parsing the resulting token from the API response.
func serviceAccountToken() (string, error) { //nolint:unused
	const tokenRequestRawString = `{
"apiVersion": "authentication.k8s.io/v1",
"kind": "TokenRequest"
}`

	// Temporary file to store the token request
	secretName := fmt.Sprintf("%s-token-request", serviceAccountName)
	tokenRequestFile := filepath.Join("/tmp", secretName)
	err := os.WriteFile(tokenRequestFile, []byte(tokenRequestRawString), os.FileMode(0o644))
	if err != nil {
		return "", err
	}

	var out string
	verifyTokenCreation := func(g Gomega) {
		// Execute kubectl command to create the token
		cmd := exec.Command("kubectl", "create", "--raw", fmt.Sprintf(
			"/api/v1/namespaces/%s/serviceaccounts/%s/token",
			namespace,
			serviceAccountName,
		), "-f", tokenRequestFile)

		output, err := cmd.CombinedOutput()
		g.Expect(err).NotTo(HaveOccurred())

		// Parse the JSON output to extract the token
		var token tokenRequest
		err = json.Unmarshal(output, &token)
		g.Expect(err).NotTo(HaveOccurred())

		out = token.Status.Token
	}
	Eventually(verifyTokenCreation).Should(Succeed())

	return out, err
}

// getMetricsOutput retrieves and returns the logs from the curl pod used to access the metrics endpoint.
func getMetricsOutput() string { //nolint:unused
	By("getting the curl-metrics logs")
	cmd := exec.Command("kubectl", "logs", "curl-metrics", "-n", namespace)
	metricsOutput, err := utils.Run(cmd)
	Expect(err).NotTo(HaveOccurred(), "Failed to retrieve logs from curl pod")
	Expect(metricsOutput).To(ContainSubstring("< HTTP/1.1 200 OK"))
	return metricsOutput
}

// tokenRequest is a simplified representation of the Kubernetes TokenRequest API response,
// containing only the token field that we need to extract.
type tokenRequest struct { //nolint:unused
	Status struct {
		Token string `json:"token"`
	} `json:"status"`
}
