# deploy

## Cluster Creation

This project deploys an aks cluster for ACStor testing.

### Parameters

| Parameter Name | Description                           | Default Value                                     |
| -------------- | ------------------------------------- | ------------------------------------------------- |
| LOCATION       | The location of the resources         | eastus                                            |
| SUFFIX         | The suffix to append to the resources | `${USER}-${TEMPLATE}-$(shell date +%Y%m%d%H%M%S)` |
| RESOURCE_GROUP | The name of the resource group        | `${SUFFIX}-rg`                                    |
| CLUSTER_NAME   | The name of the AKS cluster           | `${SUFFIX}-aks`                                   |
| TEMPLATE       | The template to use                   | azurelinux                                        |

### Templates

The following table lists all the templates and their descriptions:

| TEMPLATE values     | Description                                                                               |
| ------------------- | ----------------------------------------------------------------------------------------- |
| azurelinux          | Create AKS cluster with 3 Azure Linux VMs                                                 |
| azurelinux-zonal    | Create AKS cluster with 3 Azure Linux VMs in 3 zones                                      |
| azurelinux-userpool | Create AKS cluster with 3 Azure Linux VMs in systempool and 3 Azure Linux VMs in userpool |
| ubuntu              | Create AKS cluster with 3 Ubuntu VMs                                                      |
| ubuntu-fips         | Create AKS cluster with 3 Ubuntu VMs with FIPS enabled                                    |
| nvme                | Create AKS cluster with 3 standard_l8s_v3 VMs with NVMe disk                              |
| nvme-zonal          | Create AKS cluster with 3 standard_l8s_v3 VMs with NVMe disk in 3 zones                   |
| nvme-autoscaler     | Create AKS cluster with 3 standard_l8s_v3 VMs with NVMe disk and cluster autoscaler       |

### Usage

```bash
make aks TEMPLATE=azurelinux
```
