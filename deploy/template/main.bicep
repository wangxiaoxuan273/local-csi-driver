// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

targetScope='subscription'

@description('The resource group of the Managed Cluster resource.')
param resourceGroupName string

@description('The number of nodes for the agentpool.')
@minValue(1)
param agentCount int = 3

@description('Enable autoscaling for agent pools.')
param enableAutoScaling bool = false

@description('Minimum number of nodes for autoscaling.')
@minValue(3)
param minCount int = 3

@description('Maximum number of nodes for autoscaling.')
@minValue(1)
param maxCount int = 10

@description('The size of the Virtual Machine.')
param agentVMSize string = 'standard_d4s_v3'

@description('The SKU of the nodepool.')
param osSKU string = 'AzureLinux'

@description('FIPS enabled.')
param enableFIPS bool = false

@description('The zones in which the cluster should be created.')
param zones array = []

@description('Enabled will use a separate user pool for the storage pool.')
param useSeparateUserPool bool = false

@description('Is test resource.')
param isTest bool = false

@description('The tags for the resource group.')
param tags object = {}

@description('The current time in ISO 8601 format.')
param currentTimeISO string = utcNow('o')

// Combine tags with test-specific tags if isTest is true
var finalTags = union(tags, isTest ? {
  createdBy: 'e2e-tests'
  createdAt: currentTimeISO
} : {})


resource resourceGroup 'Microsoft.Resources/resourceGroups@2024-03-01' = {
  name: resourceGroupName
  location: deployment().location
  tags: finalTags
}

module cluster 'cluster.bicep' = {
  name: 'clusterModule'
  scope: resourceGroup
  params: {
    zones: zones
    useSeparateUserPool: useSeparateUserPool
    enableAutoScaling: enableAutoScaling
    minCount: minCount
    maxCount: maxCount
    agentCount: agentCount
    agentVMSize: agentVMSize
    osSKU: osSKU
    enableFIPS: enableFIPS
  }
}

module cleanupAssignment 'cleanupRoleAssignment.bicep' = if (isTest) {
  name: 'cleanupAssignmentModule'
  scope: resourceGroup
}

output managedClusterName string = cluster.outputs.managedClusterName
