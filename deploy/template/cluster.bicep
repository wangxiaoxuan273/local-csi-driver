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

@description('The streams to collect and send to Log Analytics.')
param streams array = [
  'Microsoft-ContainerInsights-Group-Default'
]

@description('The resource ID of the Log Analytics workspace.')
param logAnalyticsWorkspaceId string = resourceId('d64ddb0c-7399-4529-a2b6-037b33265372', 'xstore-log-analytics-rg', 'Microsoft.OperationalInsights/workspaces', 'xstore-log-analytics-workspace')

@description('The location of the Log Analytics workspace.')
param logAnalyticsWorkspaceLocation string = 'eastus'

@description('The resource ID of the user-assigned managed identity.')
param userAssignedManagedIdentity string = resourceId('d64ddb0c-7399-4529-a2b6-037b33265372', 'azdiskdrivertest-rg', 'Microsoft.ManagedIdentity/userAssignedIdentities', 'azdiskdrivertest-id')

@description('The resource ID of the security Log Analytics workspace.')
param securityLogAnalyticsWorkspaceId string = resourceId('d64ddb0c-7399-4529-a2b6-037b33265372', 'DefaultResourceGroup-EUS', 'Microsoft.OperationalInsights/workspaces', 'DefaultWorkspace-d64ddb0c-7399-4529-a2b6-037b33265372-EUS')

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
  nodeLabels: {
    'acstor.azure.com/io-engine': 'acstor'
  }
  enableAutoScaling: enableAutoScaling
  minCount: enableAutoScaling ? minCount : null
  maxCount: enableAutoScaling ? maxCount : null
  count: agentCount
  enableFIPS: enableFIPS
}

// Create the data collection rule and associate it with the AKS cluster.
// The data collection rule collects the specified streams and sends them to the Log Analytics workspace.
resource dataCollectionRule 'Microsoft.Insights/dataCollectionRules@2022-06-01' = {
  name: '${resourceGroup().name}-dcr'
  location: logAnalyticsWorkspaceLocation
  tags: resourceGroup().tags
  kind: 'Linux'
  properties: {
    dataSources: {
      extensions: [
        {
          name: 'ContainerInsightsExtension'
          streams: streams
          extensionSettings: {
            dataCollectionSettings: {
              enableContainerLogV2: true
            }
          }
          extensionName: 'ContainerInsights'
        }
      ]
    }
    destinations: {
      logAnalytics: [
        {
          workspaceResourceId: logAnalyticsWorkspaceId
          name: 'la-workspace'
        }
      ]
    }
    dataFlows: [
      {
        streams: streams
        destinations: [
          'la-workspace'
        ]
      }

    ]
  }
}

// Associate the data collection rule with the AKS cluster.
resource dataCollectionRuleAssociation 'Microsoft.Insights/dataCollectionRuleAssociations@2022-06-01' = {
  name: 'ContainerInsightsExtension'
  scope: aks
  properties: {
    description: 'Association of data collection rule. Deleting this association will break the data collection for this AKS Cluster.'
    dataCollectionRuleId: dataCollectionRule.id
  }
}

resource aks 'Microsoft.ContainerService/managedClusters@2023-08-01' = {
  name: '${resourceGroup().name}-cluster'
  location: resourceGroup().location
  sku: {
    name: 'Base'
    tier: 'Standard'
  }
  tags: resourceGroup().tags
  identity: {
    type: 'UserAssigned'
    userAssignedIdentities: {
      '${userAssignedManagedIdentity}': {}
    }
  }
  properties: {
    nodeResourceGroup: '${resourceGroup().name}-node-rg'
    dnsPrefix: '${resourceGroup().name}-dns'
    addonProfiles: {
      azurePolicy: {
        enabled: true
      }
      omsagent: {
        enabled: true
        config: {
          logAnalyticsWorkspaceResourceID: logAnalyticsWorkspaceId
          useAADAuth: 'true'
        }
      }
    }
    agentPoolProfiles: useSeparateUserPool ? [ systemPool, storagePool ] : [ storagePool ]
    servicePrincipalProfile: {
      clientId: 'msi'
    }
    networkProfile: {
      networkPlugin: 'azure'
    }
    identityProfile: {
      kubeletidentity: {
        resourceId: userAssignedManagedIdentity
      }
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
