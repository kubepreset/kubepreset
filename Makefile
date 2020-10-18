# Next KubePreset version.  It should be updated in the main branch and release
# branches immediately after the tag is published.
KUBEPRESET_VERSION ?= "0.1.0"

# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

KUBEBUILDER_VERSION ?= "2.3.1"
export KUBEBUILDER_ASSETS ?= $(CURDIR)/bin

# Other allowed values are DEBUG and INFO
export LOG_LEVEL ?= TRACE

all: manager

# Run tests
test: bin/kubebuilder generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# download golangci-lint
# binary will be $(go env GOPATH)/bin/golangci-lint
golangci-lint:
ifeq (, $(shell which golangci-lint))
	@{ \
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v1.30.0 ;\
	golangci-lint --version ;\
	}
GOLANGCI_LINT=$(GOBIN)/golangci-lint
else
GOLANGCI_LINT=$(shell which golangci-lint)
endif

# Run golangci-lint
lint: golangci-lint
	${GOLANGCI_LINT} run

bin/kubebuilder:
	mkdir -p $(CURDIR)/bin
	hack/download-kubebuilder.sh ${KUBEBUILDER_VERSION}

clean:
	rm -f $(CURDIR)/bin/*
