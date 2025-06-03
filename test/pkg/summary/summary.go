// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package summary

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
)

// GenerateReport generates a markdown report for the test suite.
func GenerateReport(r Report, clusterName, summaryPath string) error {
	var sb strings.Builder

	sb.WriteString("# Test Suite Report\n\n")
	sb.WriteString(fmt.Sprintf("Suite Start Time: %s\n", r.StartTime.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("Suite End Time: %s\n\n", r.EndTime.Format(time.RFC3339)))
	sb.WriteString("### Suite Links\n\n")
	sb.WriteString(generateLinks(r.StartTime, r.EndTime, clusterName))

	for _, spec := range r.SpecReports {
		if spec.Failed() && spec.LeafNodeType == types.NodeTypeIt {
			sb.WriteString(fmt.Sprintf("## Failure: %s\n", spec.FullText()))
			sb.WriteString("### Failure details\n")
			sb.WriteString(fmt.Sprintf("Failure Message\n\n```\n%s\n```\n\n", spec.Failure.Message))
			sb.WriteString(fmt.Sprintf("Failure Location\n\n```\n%s\n```\n\n", spec.Failure.Location.String()))
			sb.WriteString("### Failure Links\n")
			sb.WriteString(generateLinks(spec.StartTime, spec.EndTime, clusterName))
		}
	}

	return os.WriteFile(summaryPath, []byte(sb.String()), 0644)
}

func generateLinks(startTime, endTime time.Time, clusterName string) string {
	var sb strings.Builder
	asiLink := generateASILink(startTime, endTime, clusterName)
	kubeeventsLink := generateKubeEventsLink(startTime, endTime, clusterName)
	containerLogsLink := generateContainerLogsLink(startTime, endTime, clusterName)

	sb.WriteString(fmt.Sprintf("- [ASI Link](%s)\n", asiLink))
	sb.WriteString(fmt.Sprintf("- [KubeEvents Link](%s)\n", kubeeventsLink))
	sb.WriteString(fmt.Sprintf("- [Container Logs Link](%s)\n\n", containerLogsLink))
	return sb.String()
}

// generateASILink generates a link to Azure Service Insights for the given time range.
func generateASILink(startTime, endTime time.Time, clusterName string) string {
	clusterNameEncoded := url.QueryEscape(clusterName)
	startTimeEncoded := url.QueryEscape(startTime.Format(time.RFC3339))
	endTimeEncoded := url.QueryEscape(endTime.Format(time.RFC3339))
	return fmt.Sprintf("https://asi.azure.ms/search/services/AKS?searchText=%s&globalFrom=%s&globalTo=%s&resources=%%5B%%22Managed+Clusters%%22%%5D", clusterNameEncoded, startTimeEncoded, endTimeEncoded)
}

// generateKubeEventsLink generates a link to Kubernetes events for the given time range.
func generateKubeEventsLink(startTime, endTime time.Time, clusterName string) string {
	startTimeEncoded := startTime.Format(time.RFC3339)
	endTimeEncoded := endTime.Format(time.RFC3339)
	kubeEventsQuery := fmt.Sprintf("let _startTime = datetime(%s);\nlet _endTime = datetime(%s);\nlet _clusterName = \"%s\";\nKubeEvents\n| where TimeGenerated between (_startTime .. _endTime)\n| where _ResourceId endswith _clusterName\n| take 10000", startTimeEncoded, endTimeEncoded, clusterName)
	kubeEventsQueryEncoded := url.QueryEscape(kubeEventsQuery)
	return fmt.Sprintf("https://dataexplorer.azure.com/clusters/https%%3A%%2F%%2Fade.loganalytics.io%%2Fsubscriptions%%2Fd64ddb0c-7399-4529-a2b6-037b33265372%%2FresourceGroups%%2Fxstore-log-analytics-rg%%2Fproviders%%2FMicrosoft.OperationalInsights%%2Fworkspaces%%2Fxstore-log-analytics-workspace/databases/xstore-log-analytics-workspace?query=%s", kubeEventsQueryEncoded)
}

// generateContainerLogsLink generates a link to container logs for the given time range.
func generateContainerLogsLink(startTime, endTime time.Time, clusterName string) string {
	startTimeEncoded := startTime.Format(time.RFC3339)
	endTimeEncoded := endTime.Format(time.RFC3339)
	containerLogQuery := fmt.Sprintf("let _startTime = datetime(%s);\nlet _endTime = datetime(%s);\nlet _clusterName = \"%s\";\nContainerLogV2\n| where TimeGenerated between (_startTime .. _endTime)\n| where _ResourceId endswith _clusterName\n| take 10000", startTimeEncoded, endTimeEncoded, clusterName)
	return fmt.Sprintf("https://dataexplorer.azure.com/clusters/https%%3A%%2F%%2Fade.loganalytics.io%%2Fsubscriptions%%2Fd64ddb0c-7399-4529-a2b6-037b33265372%%2FresourceGroups%%2Fxstore-log-analytics-rg%%2Fproviders%%2FMicrosoft.OperationalInsights%%2Fworkspaces%%2Fxstore-log-analytics-workspace/databases/xstore-log-analytics-workspace?query=%s", url.QueryEscape(containerLogQuery))
}
