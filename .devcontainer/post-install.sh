#!/bin/bash
# Copyright (c) Microsoft Corporation.
# Licensed under the MIT License.

set -x

# Install bash-completion
apt-get update
apt-get install -y bash-completion

pre-commit install

curl -Lo ./kind "https://kind.sigs.k8s.io/dl/latest/kind-$(go env GOOS)-$(go env GOARCH)"
chmod +x ./kind
mv ./kind /usr/local/bin/kind

curl -L -o kubebuilder.tar.gz "https://storage.googleapis.com/kubebuilder-release/kubebuilder_master_$(go env GOOS)_$(go env GOARCH).tar.gz"
tar -zxvf kubebuilder.tar.gz && chmod +x kubebuilder && mv kubebuilder /usr/local/bin/ && rm kubebuilder.tar.gz

KUBECTL_VERSION=$(curl -L -s https://dl.k8s.io/release/stable.txt)
curl -LO "https://dl.k8s.io/release/$KUBECTL_VERSION/bin/$(go env GOOS)/$(go env GOARCH)/kubectl"
chmod +x kubectl
mv kubectl /usr/local/bin/kubectl

docker network create -d=bridge --subnet=172.19.0.0/24 kind

kind version
kubebuilder version
docker --version
go version
kubectl version --client

cat << 'EOF' >> ~/.bashrc
# enable programmable completion features (you don't need to enable
# this, if it's already enabled in /etc/bash.bashrc and /etc/profile
# sources /etc/bash.bashrc).
if ! shopt -oq posix; then
    if [ -f /usr/share/bash-completion/bash_completion ]; then
        . /usr/share/bash-completion/bash_completion
    elif [ -f /etc/bash_completion ]; then
        . /etc/bash_completion
    fi
fi
source <(kubectl completion bash)
source <(kind completion bash)

echo "complete -F __start_kubectl k" >> ~/.bashrc
echo "alias k='kubectl'" >> ~/.bashrc
EOF

# skip logging all the contents of the bashrc
set +x
# shellcheck source=/dev/null
source ~/.bashrc
