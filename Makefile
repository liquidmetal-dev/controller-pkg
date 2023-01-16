VER?=0.0.1
MODULES=$(shell find . -mindepth 2 -maxdepth 4 -type f -name 'go.mod' | cut -c 3- | sed 's|/[^/]*$$||' | sort -u | tr / :)
targets=$(addprefix fmt-, $(MODULES))
root_dir=$(shell git rev-parse --show-toplevel)

# Use $GOBIN from the environment if set, otherwise use ./bin
ifeq (,$(shell go env GOBIN))
GOBIN=$(root_dir)/bin
else
GOBIN=$(shell go env GOBIN)
endif

PKG?=$*

##@ General

all: ## Tidy, fmt, and vet all modules
	$(MAKE) $(targets)

fmt-%: generate-% tidy-% vet-% mocks-%
	cd $(subst :,/,$*); go fmt ./...

tidy-%:
	cd $(subst :,/,$*); go mod tidy -compat=1.19

vet-%:
# "git:libgit2" is the wildcard that comes after "vet-"
# running make vet-git:libgit2 will cd into git/libgit2 and run make vet
	@if [ "$(PKG)" = "git:libgit2" ]; then \
		cd $(subst :,/,$*); make vet ;\
	else \
		cd $(subst :,/,$*); go vet ./... ;\
	fi

##@ Tools

generate-%: controller-gen ## Generate schema
	# Run schemapatch to validate all the kubebuilder markers before generation
	cd $(subst :,/,$*); CGO_ENABLED=0 $(CONTROLLER_GEN) schemapatch:manifests="./" paths="./..."
	cd $(subst :,/,$*); CGO_ENABLED=0 $(CONTROLLER_GEN) object:headerFile="$(root_dir)/hack/boilerplate.go.txt" paths="./..."

# Find or download controller-gen
CONTROLLER_GEN = $(GOBIN)/controller-gen
.PHONY: controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.10.0)

# go-install-tool will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-install-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

##@ Utility

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
