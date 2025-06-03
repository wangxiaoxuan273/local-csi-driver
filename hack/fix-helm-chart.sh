#!/usr/bin/env bash
# Copyright (c) Microsoft Corporation.
# Licensed under the MIT License.


# This script is used to patch the output of helmify to make it work with our deployment.
# Helmify adds annotations to the issuer and cert to install them after helm install.
# However, our deployment relies on a secret created from these CRs being applied. To
# fix this, we remove the annotation and install the cert-manager CRD as part of our helm
# chart.

set -eou pipefail
CHART=""
TAG=""

while [[ "$#" -gt 0 ]]; do
    case $1 in
        --chart) CHART="$2"; shift ;;
        --version) TAG="$2"; shift ;;
        *) echo "Unknown parameter passed: $1"; exit 1 ;;
    esac
    shift
done

if [[ -z "$CHART" || -z "$TAG" ]]; then
    echo "Usage: $0 --chart <chart-directory> --version <version>"
    exit 1
fi

find "$CHART" -name 'serving-cert.yaml' | while read -r file; do
    if grep -q 'csi-local-selfsigned-issuer' "$file"; then
        echo "Patching $file"
        sed -i 's/csi-local-selfsigned-issuer/{{ include "chart.fullname" . }}-selfsigned-issuer/g' "$file"
    fi
done


cat <<EOF > "$CHART/Chart.yaml"
# Default values for local-csi-driver.
apiVersion: v2
description: Local CSI Driver Helm chart for Kubernetes
name: local-csi-driver
type: application
version: $TAG
EOF
