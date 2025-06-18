// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

@description('The number of nodes for the agentpool.')
@minValue(1)
param agentCount int = 3

@description('Enable autoscaling for agent pools.')
param enableAutoScaling bool = false

@description('Minimum number of nodes for autoscaling.')
@minValue(1)
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

@description('The resource ID of the security Log Analytics workspace.')
param securityLogAnalyticsWorkspaceId string = resourceId(subscription().subscriptionId, 'DefaultResourceGroup-EUS', 'Microsoft.OperationalInsights/workspaces', 'DefaultWorkspace-${subscription().subscriptionId}-EUS')

// The system pool is used for system pods and the storage pool is used for storage pods.
var systemPool = {
  name: 'systempool'
  count: 3
  vmSize: agentVMSize
  osDiskSizeGB: 128
  type: 'VirtualMachineScaleSets'
  osType: 'Linux'
  osSKU: osSKU
  mode: 'System'
  maxPods: 125
  availabilityZones: zones
  nodeTaints: ['CriticalAddonsOnly=true:NoSchedule']
  enableFIPS: enableFIPS
}

// The storage pool is used for storage pods. If useSeparateUserPool is enabled, the storage pool will be a separate user pool.
var storagePool = {
  name: 'storagepool'
  vmSize: agentVMSize
  osDiskSizeGB: 128
  type: 'VirtualMachineScaleSets'
  osType: 'Linux'
  osSKU: osSKU
  mode: useSeparateUserPool ? 'User' : 'System'
  maxPods: 125
  availabilityZones: zones
  enableAutoScaling: enableAutoScaling
  minCount: enableAutoScaling ? minCount : null
  maxCount: enableAutoScaling ? maxCount : null
  count: agentCount
  enableFIPS: enableFIPS
}

resource aks 'Microsoft.ContainerService/managedClusters@2023-08-01' = {
  name: '${resourceGroup().name}-cluster'
  location: resourceGroup().location
  identity: {
    type: 'SystemAssigned'
  }
  sku: {
    name: 'Base'
    tier: 'Standard'
  }
  properties: {
    nodeResourceGroup: '${resourceGroup().name}-node-rg'
    dnsPrefix: '${resourceGroup().name}-dns'
    addonProfiles: {
      azurePolicy: {
        enabled: true
      }
    }
    agentPoolProfiles: useSeparateUserPool ? [ systemPool, storagePool ] : [ storagePool ]
    networkProfile: {
      networkPlugin: 'azure'
    }
    securityProfile: {
      defender: {
        logAnalyticsWorkspaceResourceId: securityLogAnalyticsWorkspaceId
        securityMonitoring: {
          enabled: true
        }
      }
    }
    storageProfile: {
      fileCSIDriver: {
        enabled: false
      }
      diskCSIDriver: {
        enabled: false
      }
      snapshotController: {
        enabled: false
      }
    }
  }
}

output managedClusterName string = aks.name
