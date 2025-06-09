# Copyright (c) Microsoft Corporation.
# Licensed under the MIT License.

# Image URL to use for all building/pushing image targets.
REGISTRY ?= docker.io

# REPO_BASE is used in the Dockerfile and will be set to empty when run in the
# pipelines. Set default for empty string, not just undefined.
REPO_BASE := $(if $(REPO_BASE),$(REPO_BASE),local-csi-driver)
REPO ?= $(REPO_BASE)/driver

COMMIT_HASH ?= $(shell git describe --always --dirty)
TAG ?= 0.0.0-$(COMMIT_HASH)
IMG ?= $(REGISTRY)/$(REPO):$(TAG)
CHART_IMG ?= $(REGISTRY)/$(REPO_BASE)/local-csi-driver

SKIP_CREATE_CLUSTER ?= false
ADDITIONAL_GINKGO_FLAGS ?=
SUPPORT_BUNDLE_OUTPUT_DIR ?= $(shell pwd)/support-bundles
SUPPORT_BUNDLE_SINCE_TIME ?=

# Build info
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS += -X local-csi-driver/internal/pkg/version.gitCommit=$(COMMIT_HASH)
LDFLAGS += -X local-csi-driver/internal/pkg/version.buildDate=$(BUILD_DATE)

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.31.0

TEST_OUTPUT ?= $(shell pwd)/test.xml
TEST_COVER ?= $(shell pwd)/coverage.out
SUMMARY_OUTPUT ?= $(shell pwd)/summary.md
NO_COLOR ?= false
FOCUS ?=
LABEL_FILTER ?=
SUPPORT_BUNDLE_OUTPUT ?= $(SUPPORT_BUNDLE_OUTPUT_DIR)/support-bundle-$(shell date -u +"%Y-%m-%dT%H:%M:%SZ").tar.gz
TEST_TIMEOUT ?= 50m
SCALE ?= 150
# Flag determines how many times to repeat aks integration test
REPEAT ?= 1

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	$(CONTROLLER_GEN) rbac:roleName=node-role \
		paths="./..." \
		output:dir=config/rbac

.PHONY: generate
generate: mocks controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations and mock interfaces.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: mocks
mocks: mockgen ## Generate mocks for the interfaces.
	@PATH="$(LOCALBIN):$(PATH)" go generate ./...

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate envtest go-junit-report gocov gocov-xml ## Run tests and generate coverage report
	$(eval TMP := $(shell mktemp -d))
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" \
		go test -race -v -p 8 $$(go list ./... | grep -v -e /test/) -coverprofile $(TMP)/cover.out 2>&1 | \
		$(GO_JUNIT_REPORT) -set-exit-code -iocopy -out $(TEST_OUTPUT)
	@cat $(TMP)/cover.out | grep -v -E -f .covignore > $(TMP)/cover.clean
	@$(GOCOV) convert $(TMP)/cover.clean | $(GOCOV_XML) > $(TEST_COVER)
	@rm $(TMP)/cover.out $(TMP)/cover.clean && rmdir $(TMP)


# The default setup assumes Kind is pre-installed and builds/loads the Manager Docker image locally.
# Prometheus and CertManager and Jaeger are installed by default if they not
# already installed. Skip install with:
# - SKIP_INSTALL_PROMETHEUS=true
# - SKIP_INSTALL_CERT_MANAGER=true
# To avoid uninstalling everything after the tests, use:
# - SKIP_UNINSTALL=true
.PHONY: e2e
e2e: ginkgo manifests generate ## Run the e2e tests (developers).
	SKIP_UNINSTALL=true \
	SKIP_SANITY=true \
	SKIP_SCALE=true \
	SKIP_METRICS=true \
	SKIP_CREATE_CLUSTER=$(SKIP_CREATE_CLUSTER) \
	$(call run_tests,e2e,./test/e2e)

.PHONY: test-e2e
test-e2e: ginkgo manifests generate ## Run the e2e tests.
	$(eval ARGS := $(ADDITIONAL_GINKGO_FLAGS) --fail-fast)
	$(call run_tests,e2e && !aks,./test/e2e,$(ARGS),)

.PHONY: test-e2e-aks
test-e2e-aks: ginkgo manifests generate ## Run the e2e tests on AKS.
	$(eval ARGS := $(ADDITIONAL_GINKGO_FLAGS) --fail-fast)
	$(call run_tests,e2e,./test/e2e,$(ARGS),)

.PHONY: test-sanity
test-sanity: ginkgo manifests generate ## Run the sanity tests.
	$(eval ARGS := $(ADDITIONAL_GINKGO_FLAGS))
	$(if $(findstring --dry-run,$(ADDITIONAL_GINKGO_FLAGS)), , $(eval ARGS := $(ARGS) --procs=16))
	$(call run_tests,sanity,./test/sanity,$(ARGS),)

.PHONY: test-external-e2e
test-external-e2e: ginkgo manifests generate ## Run the external e2e tests.
	$(eval ARGS := $(ADDITIONAL_GINKGO_FLAGS))
	$(if $(findstring --dry-run,$(ADDITIONAL_GINKGO_FLAGS)), , $(eval ARGS := $(ARGS) --procs=16))
	$(call run_tests,external-e2e && !Disruptive,./test/external,$(ARGS),)

.PHONY: test-scale
test-scale: ginkgo manifests generate ## Run the scale tests.
	$(call run_tests,scale && aks,./test/e2e,$(ADDITIONAL_GINKGO_FLAGS),--scale=$(SCALE))

.PHONY: test-restart
test-restart: ginkgo manifests generate ## Run the restart tests.
	$(call run_tests,restart && aks,./test/e2e,$(ADDITIONAL_GINKGO_FLAGS),--repeat=$(REPEAT))

.PHONY: test-upgrade
test-upgrade: ginkgo manifests generate ## Run the upgrade tests.
	$(call run_tests,upgrade && aks,./test/e2e,$(ADDITIONAL_GINKGO_FLAGS),--repeat=$(REPEAT))

.PHONY: test-scaledown
test-scaledown: ginkgo manifests generate ## Run the scale down tests.
	$(call run_tests,scaledown && aks,./test/e2e,$(ADDITIONAL_GINKGO_FLAGS),--repeat=$(REPEAT))

define run_tests
TAG=$(TAG) \
REGISTRY=$(REGISTRY) \
IMG=$(IMG) \
$(GINKGO) -v -r $(3) --label-filter="$(1)$(if $(LABEL_FILTER), && ($(LABEL_FILTER)))" --focus="$(FOCUS)" --no-color="$(NO_COLOR)" --timeout="$(TEST_TIMEOUT)" "$(2)" -- \
	--cloudtest-report=$(TEST_OUTPUT) --summary-report=$(SUMMARY_OUTPUT) --support-bundle-dir=$(SUPPORT_BUNDLE_OUTPUT_DIR) "$(4)"
endef

.PHONY: test-container-structure
test-container-structure: container-structure-test ## Run the container structure tests.
	$(CONTAINER_STRUCTURE_TEST) test --image $(IMG) --config test/container-structure/local-csi-driver.yaml

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter.
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes.
	$(GOLANGCI_LINT) run --fix

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build binary.
	go build -ldflags "$(LDFLAGS)" -o bin/local-csi-driver cmd/main.go

.PHONY: run
run: manifests generate fmt vet ## Run the local CSI driver from your host.
	go run ./cmd/main.go

.PHONY: docker-build
docker-build: docker-buildx controller-gen ## Build the docker image.
	$(call docker-build,Dockerfile,${IMG})

# buildx builder arguments
BUILDX_BUILDER_NAME ?= img-builder
OUTPUT_TYPE ?= type=registry
QEMU_VERSION ?= 7.2.0-1
ARCH ?= amd64,arm64
BUILDKIT_VERSION ?= v0.18.1

.PHONY: docker-buildx
docker-buildx: ## Install and configure docker buildx for multiple architectures.
	@if ! docker buildx ls | grep $(BUILDX_BUILDER_NAME); then \
		docker run --rm --privileged mcr.microsoft.com/mirror/docker/multiarch/qemu-user-static:$(QEMU_VERSION) --reset -p yes; \
		docker buildx create --name $(BUILDX_BUILDER_NAME) --driver-opt image=mcr.microsoft.com/oss/v2/moby/buildkit:$(BUILDKIT_VERSION) --use; \
		docker buildx inspect $(BUILDX_BUILDER_NAME) --bootstrap; \
	fi

# Define a function to build docker images. Skips the BUILD_DATE arg
# intentionally to avoid retriggering builds when the date changes.
define docker-build
	DOCKER_BUILDKIT=1 $(CONTAINER_TOOL) buildx build \
		--build-arg GIT_COMMIT=$(COMMIT_HASH) \
		--build-arg BUILD_ID=$(TAG) \
		--output=$(OUTPUT_TYPE) \
		--platform linux/$(ARCH) \
		--pull \
		--tag $(2) \
		-f $(1) .
endef

.PHONY: docker-load
docker-load: ## Load the docker image into the kind cluster.
	$(call docker-load,${IMG})

define docker-load
	kind load docker-image $(1) --name kind
endef

.PHONY: docker-lint
docker-lint: hadolint
	$(HADOLINT) Dockerfile

.PHONY: helm-build
helm-build: build-driver helm helmify manifests generate ## Generate a consolidated Helm chart with CRDs and deployment.
	cat dist/install.yaml | \
		$(HELMIFY) -v \
		-original-name \
		dist/chart && \
		./hack/fix-helm-chart.sh --chart dist/chart --version $(TAG)
	$(HELM) package --dependency-update dist/chart -d dist --version $(TAG)


# Helm login
.PHONY: helm-login
helm-login: helm ## Log in to the ACR Helm registry.
	@if echo $(REGISTRY) | grep -q "azurecr.io"; then \
		echo "Logging in to Azure Container Registry $(REGISTRY)..."; \
		TOKEN=$$(az acr login --name $(REGISTRY) --expose-token --output tsv --query accessToken); \
		echo $$TOKEN | $(HELM) registry login $(REGISTRY) -u 00000000-0000-0000-0000-000000000000 --password-stdin; \
		echo "Logged in to Azure Container Registry $(REGISTRY)."; \
	else \
		echo "Registry is not an Azure Container Registry."; \
	fi

.PHONY: helm-push
helm-push: helm helm-build ## Push the Helm chart to the Helm repository.
	$(HELM) push dist/local-csi-driver-$(TAG).tgz oci://$(REGISTRY)/$(REPO_BASE)

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = true
endif

.PHONY: arc-install
arc-install: ## Install the Azure Kubernetes Service extension.
	@if [ -z "${RELEASE_TRAIN}" ]; then \
		echo "Error: No release-train provided."; \
		echo "Usage: make arc-install RELEASE_TRAIN=<release-train>"; \
		exit 1; \
	fi; \
	chmod +x ./deploy/scripts/arc-install.sh && ./deploy/scripts/arc-install.sh $(RELEASE_TRAIN)

.PHONY: arc-uninstall
arc-uninstall: ## Uninstall the Azure Kubernetes Service extension.
	chmod +x ./deploy/scripts/arc-uninstall.sh && ./deploy/scripts/arc-uninstall.sh


HELM_ARGS ?=
.PHONY: helm-install-local
helm-install-local: helm helm-build ## Install the Helm chart from the local build into the K8s cluster specified in ~/.kube/config.
	$(HELM) install local-csi-driver dist/chart --namespace cns-system --create-namespace --debug --wait --atomic $(HELM_ARGS)

.PHONY: helm-install-local
helm-install: helm ## Install the Helm chart from REGISTRY into the K8s cluster specified in ~/.kube/config.
	$(HELM) install local-csi-driver oci://$(CHART_IMG) --version $(TAG) --namespace cns-system --create-namespace --debug --wait --atomic $(HELM_ARGS)

.PHONY: helm-show-values
helm-show-values: helm ## Show the default values of the Helm chart.
	@$(HELM) show values oci://$(CHART_IMG) --version $(TAG)

.PHONY: uninstall-helm
uninstall-helm: helm ## Uninstall the Helm chart from the K8s cluster specified in ~/.kube/config.
	$(HELM) uninstall local-csi-driver --namespace cns-system --debug --wait --ignore-not-found

.PHONY: deploy
deploy: manifests helm ## Deploy to the K8s cluster specified in ~/.kube/config.
	$(HELM) install local-csi-driver charts/latest \
	--namespace cns-system \
	--version $(TAG) \
	--set node.driver.image.repository=$(REGISTRY)/$(REPO) \
	--set node.driver.image.tag=$(TAG) \
	--create-namespace \
	--debug --wait --atomic $(HELM_ARGS)

.PHONY: build-driver
build-driver: manifests generate kustomize ## Generate a consolidated install YAML.
	mkdir -p dist
	(cd config/csi-driver && $(KUSTOMIZE) edit set image driver=${IMG})
	$(KUSTOMIZE) build config/default > dist/install.yaml

.PHONY: undeploy
undeploy: uninstall-helm ## Undeploy from the K8s cluster specified in ~/.kube/config.

##@ Infrastructure

AKS_TEMPLATE ?= nvme
AKS_LOCATION ?= uksouth
AKS_RESOURCE_GROUP ?= ${USER}-local-csi-driver-${AKS_TEMPLATE}
AKS_IS_TEST ?= false
.PHONY: aks
aks: bicep ## Create an AKS cluster using Bicep.
	@if [ ! -f deploy/parameters/${AKS_TEMPLATE}.json ]; then echo "deploy/parameters/${AKS_TEMPLATE}.json does not exist" && exit 1; fi
	@if [ -z "$(AKS_RESOURCE_GROUP)" ]; then echo "AKS_RESOURCE_GROUP is not set"; exit 1; fi
	@if [ -z "$(AKS_LOCATION)" ]; then echo "AKS_LOCATION is not set"; exit 1; fi
	chmod +x ./deploy/scripts/create.sh && ./deploy/scripts/create.sh --resource-group $(AKS_RESOURCE_GROUP) --location $(AKS_LOCATION) --template $(AKS_TEMPLATE) --is-test $(AKS_IS_TEST)

.PHONY: aks-cluster-requirements
aks-cluster-requirements: bicep ## print out the requirements for the AKS cluster.
	@if [ ! -f deploy/parameters/${AKS_TEMPLATE}.json ]; then echo "deploy/parameters/${AKS_TEMPLATE}.json does not exist" && exit 1; fi
	@if [ -z "$(AKS_RESOURCE_GROUP)" ]; then echo "AKS_RESOURCE_GROUP is not set"; exit 1; fi
	@if [ -z "$(AKS_LOCATION)" ]; then echo "AKS_LOCATION is not set"; exit 1; fi
	@chmod +x ./deploy/scripts/create.sh && ./deploy/scripts/create.sh --resource-group $(AKS_RESOURCE_GROUP) --location $(AKS_LOCATION) --template $(AKS_TEMPLATE) --is-test $(AKS_IS_TEST) --what-if

.PHONY: aks-clean
aks-clean: ## Delete the AKS cluster.
	@if [ -z "$(AKS_RESOURCE_GROUP)" ]; then echo "AKS_RESOURCE_GROUP is not set"; exit 1; fi
	chmod +x ./deploy/scripts/clean.sh && ./deploy/scripts/clean.sh --resource-group $(AKS_RESOURCE_GROUP)

.PHONY: bicep-lint
bicep-lint: bicep ## Lint all Bicep templates.
	find . -name '*.bicep' | xargs -P 4 -I {} sh -c 'echo "Linting {}"; $(BICEP) lint "{}"'

.PHONY: bicep-format
bicep-format: bicep ## Format all Bicep templates.
	find . -name '*.bicep' | xargs -P 4 -I {} sh -c 'echo "Formatting {}"; $(BICEP) format "{}"'

.PHONY: run-markdownlint
run-markdownlint: $(NODE)
	npx markdownlint-cli@$(MARKDOWNLINT_CLI_VERSION) -c .markdownlint.yaml --fix .

.PHONY: add-copyright
add-copyright:
	./hack/add_copyright.sh

.PHONY: npm-login
npm-login: $(NPM) ## Login to the Azure DevOps npm registry.
	@echo "Logging in to Azure DevOps npm registry..." && \
	echo "This will create a ~/.npmrc file with the authentication token." && \
	echo "Please make sure to set the correct scopes for the token." && \
	echo "You can generate a new token at https://msazure.visualstudio.com/_usersSettings/tokens" && \
	echo "Make a PAT with with Packaging read & write scopes." && \
	read -s -p "Enter PAT: " PAT && echo && \
	PAT_BASE64=$$(echo -n "$$PAT" | base64 | tr -d '\n') && \
	echo "; begin auth token" > ~/.npmrc && \
	echo "//msazure.pkgs.visualstudio.com/One/_packaging/One_PublicPackages/npm/registry/:username=msazure" >> ~/.npmrc && \
	echo "//msazure.pkgs.visualstudio.com/One/_packaging/One_PublicPackages/npm/registry/:_password=$$PAT_BASE64" >> ~/.npmrc && \
	echo "//msazure.pkgs.visualstudio.com/One/_packaging/One_PublicPackages/npm/registry/:email=npm requires email to be set but doesn't use the value" >> ~/.npmrc && \
	echo "//msazure.pkgs.visualstudio.com/One/_packaging/One_PublicPackages/npm/:username=msazure" >> ~/.npmrc && \
	echo "//msazure.pkgs.visualstudio.com/One/_packaging/One_PublicPackages/npm/:_password=$$PAT_BASE64" >> ~/.npmrc && \
	echo "//msazure.pkgs.visualstudio.com/One/_packaging/One_PublicPackages/npm/:email=npm requires email to be set but doesn't use the value" >> ~/.npmrc && \
	echo "; end auth token" >> ~/.npmrc

.PHONY: single
single: manifests kind ## Create a single node kind cluster.
	if $(KIND) get clusters | grep -q kind; then \
		echo "Cluster 'kind' already exists"; \
	else \
		$(KIND) create cluster --config=./test/config/kind-1-node.yaml; \
	fi

.PHONY: multi
multi: manifests kind ## Create a multi node kind cluster.
	if $(KIND) get clusters | grep -q kind; then \
		echo "Cluster 'kind' already exists"; \
	else \
		$(KIND) create cluster --config=./test/config/kind-3-node.yaml; \
	fi

.PHONY: clean
clean: kind ## Deletes the kind cluster.
	$(KIND) delete cluster --name kind

.PHONY: kubeconfig
kubeconfig: kind ## Sets the kubectl context to the kind cluster.
	$(KIND) export kubeconfig --name kind

.PHONY: get-support-bundle
get-support-bundle: support-bundle ## Download support-bundle locally if necessary.
	@mkdir -p $(dir $(SUPPORT_BUNDLE_OUTPUT))
	@$(SUPPORT_BUNDLE) \
		--output=$(SUPPORT_BUNDLE_OUTPUT) \
		--interactive=false \
		$(if $(SUPPORT_BUNDLE_SINCE_TIME),--since-time=$(SUPPORT_BUNDLE_SINCE_TIME)) \
		test/config/support-bundle.yaml
	@tar -xvf $(SUPPORT_BUNDLE_OUTPUT) -C $(dir $(SUPPORT_BUNDLE_OUTPUT))
	@rm $(SUPPORT_BUNDLE_OUTPUT)
	$(KUBECTL) get events --all-namespaces > $(dir $(SUPPORT_BUNDLE_OUTPUT))/events.txt

##@ Dependencies
## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
MOCK_GEN ?= $(LOCALBIN)/mockgen
ENVTEST ?= $(LOCALBIN)/setup-envtest
KIND ?= $(LOCALBIN)/kind
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
GO_JUNIT_REPORT ?= $(LOCALBIN)/go-junit-report
GINKGO ?= $(LOCALBIN)/ginkgo
GOCOV ?= $(LOCALBIN)/gocov
GOCOV_XML ?= $(LOCALBIN)/gocov-xml
HADOLINT ?= $(LOCALBIN)/hadolint
CONTAINER_STRUCTURE_TEST ?= $(LOCALBIN)/container-structure-test
SUPPORT_BUNDLE ?= $(LOCALBIN)/support-bundle
BICEP ?= $(LOCALBIN)/bicep
HELMIFY ?= $(LOCALBIN)/helmify
HELM ?= $(LOCALBIN)/helm

## Tool Versions
KUSTOMIZE_VERSION ?= v5.5.0
CONTROLLER_TOOLS_VERSION ?= v0.16.5
GOMOCK_VERSION ?= v0.5.2
ENVTEST_VERSION ?= release-0.19
KIND_VERSION ?= v0.25.0
GOLANGCI_LINT_VERSION ?= v2.1.6
GO_JUNIT_REPORT_VERSION ?= v2.1.0
CERT_MANAGER_VERSION ?= 1.12.12-5
PROMETHEUS_VERSION ?= v0.77.1
JAEGER_VERSION ?= v1.62.0
GINKGO_VERSION ?= v2.22.2
GOCOV_VERSION ?= v1.2.1
GOCOV_XML_VERSION ?= v1.1.0
HADOLINT_VERSION ?= v2.12.0
CONTAINER_STRUCTURE_TEST_VERSION ?= v1.19.3
GIT_SEMVER_VERSION ?= v6.9.0
SUPPORT_BUNDLE_VERSION ?= v0.114.0
BICEP_VERSION ?= v0.32.4
HELMIFY_VERSION ?= v0.4.17
HELM_VERSION ?= v3.16.4
MARKDOWNLINT_CLI_VERSION ?= v0.44.0

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: mockgen
mockgen: $(MOCK_GEN) ## Installs mockgen locally if necessary.
$(MOCK_GEN): $(LOCALBIN)
	$(call go-install-tool,$(MOCK_GEN),go.uber.org/mock/mockgen,$(GOMOCK_VERSION))

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: kind
kind: $(KIND) ## Download kustomize locally if necessary.
$(KIND): $(LOCALBIN)
	$(call go-install-tool,$(KIND),sigs.k8s.io/kind,$(KIND_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

.PHONY: go-junit-report
go-junit-report: $(GO_JUNIT_REPORT) ## Download go-junit-report locally if necessary.
$(GO_JUNIT_REPORT): $(LOCALBIN)
	$(call go-install-tool,$(GO_JUNIT_REPORT),github.com/jstemmer/go-junit-report/v2,$(GO_JUNIT_REPORT_VERSION))

.PHONY: gocov
gocov: $(GOCOV) ## Download gocov locally if necessary.
$(GOCOV): $(LOCALBIN)
	$(call go-install-tool,$(GOCOV),github.com/axw/gocov/gocov,$(GOCOV_VERSION))

.PHONY: gocov-xml
gocov-xml: $(GOCOV_XML) ## Download gocov-xml locally if necessary.
$(GOCOV_XML): $(LOCALBIN)
	$(call go-install-tool,$(GOCOV_XML),github.com/AlekSi/gocov-xml,$(GOCOV_XML_VERSION))

.PHONY: helmify
helmify: $(HELMIFY) ## Download helmify locally if necessary.
$(HELMIFY): $(LOCALBIN)
	$(call go-install-tool,$(HELMIFY),github.com/arttor/helmify/cmd/helmify,$(HELMIFY_VERSION))


##@ Utilities

set-docker-pipeline-variables: # Echos variables used to pass information to docker build pipeline.
	@echo "##vso[task.setvariable variable=build_date;isOutput=true]$$( date -u +%Y-%m-%dT%H:%M:%SZ )"
	@echo "##vso[task.setvariable variable=git_revision;isOutput=true]$$( git rev-parse HEAD )"


CERT_MANAGER_CHART="oci://mcr.microsoft.com/azurelinux/helm/cert-manager"
.PHONY: cert-manager
cert-manager: helm ## Install cert-manager into the current cluster.
	$(HELM) install cert-manager $(CERT_MANAGER_CHART) --version $(CERT_MANAGER_VERSION) \
		--set cert-manager.installCRDs=true --namespace cert-manager --create-namespace --wait --debug --atomic

.PHONY: cert-manager-uninstall
cert-manager-uninstall: helm ## Uninstall cert-manager from the current cluster.
	$(HELM) uninstall cert-manager --namespace cert-manager --wait --debug --ignore-not-found


.PHONY: kyverno
kyverno: helm ## Install kyverno into the current cluster.
	$(HELM) repo add kyverno https://kyverno.github.io/kyverno/
	$(HELM) repo update
	$(HELM) install kyverno kyverno/kyverno --namespace kyverno --create-namespace --wait --debug --atomic \
		--values deploy/values/kyverno.yaml

.PHONY: kyverno-uninstall
kyverno-uninstall: helm ## Uninstall kyverno from the current cluster.
	$(HELM) uninstall kyverno --namespace kyverno --wait --debug --ignore-not-found

PROMETHEUS_CRD_YAML="https://github.com/prometheus-operator/prometheus-operator/releases/download/$(PROMETHEUS_VERSION)/stripped-down-crds.yaml"
.PHONY: prometheus-crds
prometheus-crds: ## Install prometheus crds into the current cluster.
	$(KUBECTL) apply -f $(PROMETHEUS_CRD_YAML)

.PHONY: prometheus-crds-uninstall
prometheus-crds-uninstall: ## Uninstall prometheus crds from the current cluster.
	$(KUBECTL) delete -f $(PROMETHEUS_CRD_YAML) --ignore-not-found

.PHONY: jaeger
jaeger: jaeger-install jaeger-port-forward ## Install Jaeger into the current cluster and create port-forward.

.PHONY: jaeger-install
jaeger-install:
	$(KUBECTL) create namespace observability || true
	$(KUBECTL) apply -f https://github.com/jaegertracing/jaeger-operator/releases/download/$(JAEGER_VERSION)/jaeger-operator.yaml
	$(KUBECTL) wait deployment.apps/jaeger-operator --for condition=Available --namespace observability --timeout 5m
	$(KUBECTL) apply -n observability -f ./test/config/jaeger.yaml

.PHONY: jaeger-port-forward
jaeger-port-forward: ## Port forward the jaeger pod to localhost as a background process.
	$(KUBECTL) -n observability port-forward svc/jaeger-query 16686 &

.PHONY: ginkgo
ginkgo: $(GINKGO)
$(GINKGO): $(LOCALBIN)
	$(call go-install-tool,$(GINKGO),github.com/onsi/ginkgo/v2/ginkgo,$(GINKGO_VERSION))

hadolint: $(HADOLINT)
HADOLINT ?= $(LOCALBIN)/hadolint

.PHONY: hadolint
hadolint: $(HADOLINT) ## Download hadolint locally if necessary.
$(HADOLINT): $(LOCALBIN)
	@[ -f "$(HADOLINT)" ] || { \
	curl --fail --retry 5 --retry-delay 10 --retry-connrefused -sSLo $(HADOLINT) https://github.com/hadolint/hadolint/releases/download/$(HADOLINT_VERSION)/hadolint-Linux-x86_64; \
	chmod +x $(HADOLINT); \
	}

.PHONY: container-structure-test
container-structure-test: $(CONTAINER_STRUCTURE_TEST)
$(CONTAINER_STRUCTURE_TEST): $(LOCALBIN)
	@[ -f "$(CONTAINER_STRUCTURE_TEST)" ] || { \
	curl --fail --retry 5 --retry-delay 10 --retry-connrefused -LO https://github.com/GoogleContainerTools/container-structure-test/releases/download/$(CONTAINER_STRUCTURE_TEST_VERSION)/container-structure-test-linux-amd64; \
	chmod +x container-structure-test-linux-amd64; \
	mkdir -p $(LOCALBIN); \
	mv container-structure-test-linux-amd64 $(CONTAINER_STRUCTURE_TEST); \
	}

.PHONY: support-bundle
support-bundle: $(SUPPORT_BUNDLE)
$(SUPPORT_BUNDLE): $(LOCALBIN)
	@[ -f "$(SUPPORT_BUNDLE)" ] || { \
	curl --fail --retry 5 --retry-delay 10 --retry-connrefused -L https://github.com/replicatedhq/troubleshoot/releases/download/$(SUPPORT_BUNDLE_VERSION)/support-bundle_linux_amd64.tar.gz | \
	tar -xvzf - support-bundle; \
	chmod +x support-bundle; \
	mkdir -p $(LOCALBIN); \
	mv support-bundle $(SUPPORT_BUNDLE); \
	}

.PHONY: bicep
bicep: $(BICEP)
$(BICEP): $(LOCALBIN)
	@[ -f "$(BICEP)" ] || { \
	curl --fail --retry 5 --retry-delay 10 --retry-connrefused -Lo bicep https://github.com/Azure/bicep/releases/download/$(BICEP_VERSION)/bicep-linux-x64; \
	chmod +x bicep; \
	mkdir -p $(LOCALBIN); \
	mv bicep $(BICEP); \
	}

.PHONY: helm
helm: $(HELM)
$(HELM): $(LOCALBIN)
	@[ -f "$(HELM)" ] || { \
	curl --fail --retry 5 --retry-delay 10 --retry-connrefused -L https://get.helm.sh/helm-$(HELM_VERSION)-linux-amd64.tar.gz | \
	tar -xvzf - linux-amd64/helm; \
	chmod +x linux-amd64/helm; \
	mkdir -p $(LOCALBIN); \
	mv linux-amd64/helm $(HELM); \
	rmdir linux-amd64; \
	}



$(NODE):
	@command -v node > /dev/null || { \
		echo "Node.js not found. Please install Node.js version $(NODE_VERSION)"; \
		exit 1; \
	}
	@node -v | grep -qE '^v(1[8-9]|[2-9][0-9])\.' || { \
		echo "Node.js version must be at least 18. Please update your Node.js installation."; \
		exit 1; \
	}

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -srf $(1)-$(3) $(1)
endef
