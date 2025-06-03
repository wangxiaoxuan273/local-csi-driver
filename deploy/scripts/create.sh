#!/usr/bin/env bash
# Copyright (c) Microsoft Corporation.
# Licensed under the MIT License.


set -euo pipefail

# Function to display usage information
usage() {
    echo "Usage: $0 --resource-group <AKS_RESOURCE_GROUP> --location <AKS_LOCATION> --template <AKS_TEMPLATE> [--is-test <IS_TEST>] [--what-if]"
    echo "  --resource-group  AKS resource group name"
    echo "  --location        AKS location"
    echo "  --template        AKS template name"
    echo "  --is-test         (Optional) Flag to indicate if the deployment is a test (default: false)"
    echo "  --what-if         (Optional) Flag to indicate if the deployment should run in 'what-if' mode (default: false)"
    exit 1
}

# Parse command-line arguments
IS_TEST=false
WHAT_IF=false
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --resource-group) AKS_RESOURCE_GROUP="$2"; shift ;;
        --location) AKS_LOCATION="$2"; shift ;;
        --template) AKS_TEMPLATE="$2"; shift ;;
        --is-test) IS_TEST="$2"; shift ;;
        --what-if) WHAT_IF=true ;;
        *) usage ;;
    esac
    shift
done

echo "AKS_RESOURCE_GROUP: $AKS_RESOURCE_GROUP"
echo "AKS_LOCATION: $AKS_LOCATION"
echo "AKS_TEMPLATE: $AKS_TEMPLATE"
echo "IS_TEST: $IS_TEST"
echo "WHAT_IF: $WHAT_IF"

# Ensure required arguments are provided and not just whitespace
if [[ -z "${AKS_RESOURCE_GROUP:-}" || "$AKS_RESOURCE_GROUP" =~ ^[[:space:]]*$ ]] || \
   [[ -z "${AKS_LOCATION:-}" || "$AKS_LOCATION" =~ ^[[:space:]]*$ ]] || \
   [[ -z "${AKS_TEMPLATE:-}" || "$AKS_TEMPLATE" =~ ^[[:space:]]*$ ]]; then
    echo "Error: Arguments cannot be empty or just whitespace."
    usage
fi

if [[ "$WHAT_IF" == "true" ]]; then
    # Run a what-if deployment and extract node configuration
    WHAT_IF_OUTPUT=$(az deployment sub what-if \
        --name "$AKS_RESOURCE_GROUP-deployment" \
        --location "$AKS_LOCATION" \
        --template-file deploy/template/main.bicep \
        --parameters "deploy/parameters/$AKS_TEMPLATE.json" \
        --parameters resourceGroupName="$AKS_RESOURCE_GROUP" \
        --parameters isTest="$IS_TEST" \
        --no-pretty-print | \
    jq '
      # Find the AKS cluster resource
      .changes[] |
      select(.after.type == "Microsoft.ContainerService/managedClusters") |

      # Extract configuration from the first agent pool
      .after.properties.agentPoolProfiles[0] |

      # Return a clean object with the values we need
      {
        nodeCount: .count,
        vmSku: .vmSize,
        availabilityZones: (.availabilityZones // [])  # Use empty array if null
      }
    ')
    echo "Requirements: $WHAT_IF_OUTPUT"
    exit 0
fi

# Check if the AKS cluster already exists
if az aks show --resource-group "$AKS_RESOURCE_GROUP" --name "$AKS_RESOURCE_GROUP-cluster" > /dev/null 2>&1; then
    echo "AKS cluster already exists in resource group $AKS_RESOURCE_GROUP"
else
    # Create the AKS cluster using Bicep template
    az deployment sub create \
        --name "$AKS_RESOURCE_GROUP-deployment" \
        --location "$AKS_LOCATION" \
        --template-file deploy/template/main.bicep \
        --parameters "deploy/parameters/$AKS_TEMPLATE.json" \
        --parameters resourceGroupName="$AKS_RESOURCE_GROUP" \
        --parameters isTest="$IS_TEST" \
        --what-if \
        --verbose

    az deployment sub create \
        --name "$AKS_RESOURCE_GROUP-deployment" \
        --location "$AKS_LOCATION" \
        --template-file deploy/template/main.bicep \
        --parameters "deploy/parameters/$AKS_TEMPLATE.json" \
        --parameters resourceGroupName="$AKS_RESOURCE_GROUP" \
        --parameters isTest="$IS_TEST" \
        --verbose
fi

# Get AKS cluster credentials
az aks get-credentials \
    --resource-group "$AKS_RESOURCE_GROUP" \
    --name "$AKS_RESOURCE_GROUP-cluster" \
    --overwrite-existing
