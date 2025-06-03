// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

@description('This is the built-in Contributor role. See https://docs.microsoft.com/azure/role-based-access-control/built-in-roles#contributor')
resource contributorRoleDefinition 'Microsoft.Authorization/roleDefinitions@2022-04-01' existing = {
  scope: subscription()
  name: 'b24988ac-6180-42a0-ab88-20f7382dd24c'
}

// XStore Container Storage Cleanup Service object ID
var principalId = 'b92a9070-d7b5-4898-bbde-504acc872644'

resource roleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(resourceGroup().id, principalId, contributorRoleDefinition.id)
  scope: resourceGroup()
  properties: {
    roleDefinitionId: contributorRoleDefinition.id
    principalId: principalId
    principalType: 'ServicePrincipal'
  }
}

output roleAssignmentName string = roleAssignment.name
output roleAssignmentId string = roleAssignment.id
