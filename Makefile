# Image URL to use all building/pushing image targets
IMG ?= redis-operator:latest
# K8s version used by envtest
ENVTEST_K8S_VERSION = 1.26.1

# Set shell to bash
SHELL = /usr/bin/env bash
.SHELLFLAGS = -o pipefail -ec

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate CustomResourceDefinition objects.
	$(CONTROLLER_GEN) crd paths="./api/..." output:crd:artifacts:config=crds ;\
	test ! -d chart || test -e chart/crds || ln -s ../crds chart/crds

.PHONY: generate
generate: controller-gen client-gen informer-gen lister-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations, and generate typed client.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./api/..."
	./hack/genclient.sh

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

##@ Testing

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(LOCALBIN)/k8s/current" go test ./... -coverprofile cover.out

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go

# Build docker image in current architecture and tag it as ${IMG}.
.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	docker build -t ${IMG} .

# Push docker image to the target specified in ${IMG}.
.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

# Build and push docker image for all given platforms.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support.
	- docker buildx create --name project-v3-builder
	docker buildx use project-v3-builder
	- docker buildx build --push --platform=$(PLATFORMS) --tag ${IMG} .
	- docker buildx rm project-v3-builder

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CLIENT_GEN ?= $(LOCALBIN)/client-gen
INFORMER_GEN ?= $(LOCALBIN)/informer-gen
LISTER_GEN ?= $(LOCALBIN)/lister-gen
SETUP_ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
CONTROLLER_TOOLS_VERSION ?= v0.13.0
CODE_GENERATOR_VERSION ?= v0.28.1
SETUP_ENVTEST_VERSION ?= latest

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: client-gen
client-gen: $(CLIENT_GEN) ## Download client-gen locally if necessary.
$(CLIENT_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/client-gen || GOBIN=$(LOCALBIN) go install k8s.io/code-generator/cmd/client-gen@$(CODE_GENERATOR_VERSION)

.PHONY: informer-gen
informer-gen: $(INFORMER_GEN) ## Download informer-gen locally if necessary.
$(INFORMER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/informer-gen || GOBIN=$(LOCALBIN) go install k8s.io/code-generator/cmd/informer-gen@$(CODE_GENERATOR_VERSION)

.PHONY: lister-gen
lister-gen: $(LISTER_GEN) ## Download lister-gen locally if necessary.
$(LISTER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/lister-gen || GOBIN=$(LOCALBIN) go install k8s.io/code-generator/cmd/lister-gen@$(CODE_GENERATOR_VERSION)

.PHONY: setup-envtest
setup-envtest: $(SETUP_ENVTEST) ## Download setup-envtest locally if necessary.
$(SETUP_ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@$(SETUP_ENVTEST_VERSION)

.PHONY: envtest
envtest: setup-envtest
	ENVTESTDIR=$$($(SETUP_ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path) ;\
	chmod -R u+w $$ENVTESTDIR ;\
	rm -f $(LOCALBIN)/k8s/current ;\
	ln -s $$ENVTESTDIR $(LOCALBIN)/k8s/current
