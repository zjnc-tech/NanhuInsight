all: k8s-codegen grpcgen

include ../Makefile.defs
include ../Makefile.golang
#@ $(APIS_DIR)@apis Makefile所在目录绝对路径
APIS_DIR := $(ROOT_DIR)/$(RELATIVE_DIR)


##@ Bin
#@ $(CODEGENERATOR_VERSION)@client-gen工具版本
CODEGENERATOR_VERSION ?= v0.31.0
#@ $(CONTROLLER_TOOLS_VERSION)@controller-gen工具版本，要和CODEGENERATOR_VERSION对应，详见https://github.com/kubernetes-sigs/controller-tools
CONTROLLER_TOOLS_VERSION ?= v0.16.0
#@ $(YQ_VERSION)@yq工具版本
YQ_VERSION ?= v4.44.2
#@ $(KUSTOMIZE_VERSION)@kustomize工具版本
KUSTOMIZE_VERSION ?= v5.4.1

#@ $(LOCALBIN)@二进制工具目录
LOCALBIN=$(APIS_DIR)/bin
#@ #(CLIENT_GEN)@本地的client-gen
CLIENT_GEN = $(LOCALBIN)/client-gen-$(CODEGENERATOR_VERSION)
#@ #(LISTER_GEN)@本地的lister-gen
LISTER_GEN = $(LOCALBIN)/lister-gen-$(CODEGENERATOR_VERSION)
#@ #(INFORMER_GEN)@本地的informer-gen
INFORMER_GEN = $(LOCALBIN)/informer-gen-$(CODEGENERATOR_VERSION)
#@ #(CONTROLLER_GEN)@本地的controller-gen
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen-$(CONTROLLER_TOOLS_VERSION)
#@ #(KUSTOMIZE)@本地的kustomize
KUSTOMIZE ?= $(LOCALBIN)/kustomize-$(KUSTOMIZE_VERSION)
#@ #(YQ)@本地的yq
YQ = $(LOCALBIN)/yq-$(YQ_VERSION)

PHONY: $(LOCALBIN)
$(LOCALBIN): ##创建bin目录
	$(INFO) "Creating bin directory $(LOCALBIN)"
	$(QUIET)mkdir -p $(LOCALBIN)

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ##下载kustomize到本地
$(KUSTOMIZE): $(LOCALBIN)
	$(INFO) "Downloading kustomize $(KUSTOMIZE_VERSION)"
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))
	$(SUCCESS) "Kustomize downloaded"
PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ##下载controller-gen到本地
$(CONTROLLER_GEN): $(LOCALBIN)
	$(INFO) "Downloading controller-gen $(CONTROLLER_TOOLS_VERSION)"
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))
	$(SUCCESS) "Controller-gen downloaded"

.PHONY: client-gen
client-gen: $(CLIENT_GEN) ##下载client-gen到本地
$(CLIENT_GEN): $(LOCALBIN)
	$(INFO) "Downloading client-gen $(CODEGENERATOR_VERSION)"
	$(call go-install-tool,$(CLIENT_GEN),k8s.io/code-generator/cmd/client-gen,$(CODEGENERATOR_VERSION))
	$(SUCCESS) "Client-gen downloaded"

.PHONY: lister-gen
lister-gen: $(LISTER_GEN) ##下载lister-gen到本地
$(LISTER_GEN): $(LOCALBIN)
	$(INFO) "Downloading lister-gen $(CODEGENERATOR_VERSION)"
	$(call go-install-tool,$(LISTER_GEN),k8s.io/code-generator/cmd/lister-gen,$(CODEGENERATOR_VERSION))
	$(SUCCESS) "Lister-gen downloaded"

.PHONY: informer-gen
informer-gen: $(INFORMER_GEN) ##下载informer-gen到本地
$(INFORMER_GEN): $(LOCALBIN)
	$(INFO) "Downloading informer-gen $(CODEGENERATOR_VERSION)"
	$(call go-install-tool,$(INFORMER_GEN),k8s.io/code-generator/cmd/informer-gen,$(CODEGENERATOR_VERSION))
	$(SUCCESS) "Informer-gen downloaded"

.PHONY: yq
yq: $(YQ)  ##下载yq到本地
$(YQ): $(LOCALBIN)
	$(INFO) "Downloading yq $(YQ_VERSION)"
	$(call go-install-tool,$(YQ),github.com/mikefarah/yq/v4,$(YQ_VERSION))
	$(SUCCESS) "Yq downloaded"
##@ K8s Code Gen

k8s-codegen: generate api-gen manifests ##生成crd apis/**
	$(QUIET)$(GO) mod tidy

#@ $(CRD_DIR)@CRD生成目录
CRD_DIR=$(APIS_DIR)/../deploy/crds
CRD_DEFINITIONS_DIR=$(APIS_DIR)/k8s

crd-dir: ##创建crd目录
	$(QUIET)mkdir -p $(CRD_DIR)

.PHONY: manifests
manifests: crd-dir controller-gen ##生成CRD
	$(INFO) "Generating CRD manifests..."
	$(QUIET)rm -rf $(CRD_DIR)/*.yaml
	$(CONTROLLER_GEN) crd:generateEmbeddedObjectMeta=true,allowDangerousTypes=true paths=$(CRD_DEFINITIONS_DIR)/... output:crd:artifacts:config="$(CRD_DIR)"
	$(SUCCESS) "CRD manifests generated"

.PHONY: generate
generate: controller-gen ##生成deepcopy
	$(INFO) "Generating deepcopy..."
	$(QUIET)$(CONTROLLER_GEN) object:headerFile="boilerplate.go.txt" paths="$(CRD_DEFINITIONS_DIR)/..." 
	$(SUCCESS) "Deepcopy generated"
GOMOD := $(shell sed -n 's/^module //p' go.mod)

API_DIRS := $(shell find ./k8s -type d -links 2 | awk '{print $$0}')
API_DIRS_FULL := $(foreach dir, $(API_DIRS), $(GOMOD)/$(dir))
COMMA_API_DIRS := 
$(foreach dir, $(API_DIRS_FULL), $(eval COMMA_API_DIRS := $(COMMA_API_DIRS)$(dir),))
COMMA_API_DIRS := $(shell echo $(COMMA_API_DIRS) | sed 's/,$$//')

.PHONY: api-gen
api-gen: client-gen lister-gen informer-gen ##生成client-go lister和informer
	$(INFO) "Remove generated files."
	$(QUIET)rm -rf client-go
	$(INFO) "Generate client-go for $(CRD_DEFINITIONS_DIR)."
	$(INFO) ">> generating with client-gen"
	$(QUIET)$(CLIENT_GEN) \
		--clientset-name versioned \
		--input-base "" \
		--input $(COMMA_API_DIRS) \
		--output-pkg $(GOMOD)/client-go \
		--output-dir ./client-go \
		--go-header-file boilerplate.go.txt 
	$(INFO) ">> generating with lister-gen"
	$(QUIET)$(foreach input, $(API_DIRS_FULL), \
	echo "generating lister for $(input)"; \
	$(LISTER_GEN) $(input) --output-dir ./client-go/listers --output-pkg $(GOMOD)/client-go/listers --go-header-file boilerplate.go.txt;)
	$(INFO) ">> generating with informer-gen"
	$(QUIET)$(foreach input, $(API_DIRS_FULL), \
	echo "generating informer for $(input)"; \
	$(INFORMER_GEN) $(input) --versioned-clientset-package $(GOMOD)/client-go/versioned \
	--listers-package $(GOMOD)/client-go/listers --output-dir ./client-go/informers --output-pkg $(GOMOD)/client-go/informers --go-header-file boilerplate.go.txt;)
	$(SUCCESS) "Client-go generated"
##@ GRPC Code Gen

PROTO_DIR := $(APIS_DIR)/grpc

grpcgen: ##proto gen locally
	$(INFO) "Generating gRPC code..."
	$(QUIET)protoc --proto_path=$(PROTO_DIR) --go_out=$(PROTO_DIR) --go-grpc_out=$(PROTO_DIR) $(PROTO_DIR)/*.proto
	$(QUIET)$(GO) mod tidy
	$(SUCCESS) "gRPC code generation completed"

grpcgen-cross: ##proto gen with docker
	$(INFO) "Generating gRPC code with docker..."
	$(QUIET)docker run --rm --user `id -u ${USER}` -v `pwd`:`pwd` -w `pwd` cr.registry.res.cloud.zhejianglab.com/infrahi-installer/protoc:v1 \
	protoc --proto_path=$(PROTO_DIR) --go_out=$(PROTO_DIR) --go-grpc_out=$(PROTO_DIR) $(PROTO_DIR)/*.proto
	$(QUIET)$(GO) mod tidy
	$(SUCCESS) "gRPC code generation with docker completed"

##@ Clean
.PHONY: clean
clean: ##清理生成的文件
	$(INFO) "Cleaning generated files..."
	$(QUIET)rm -rf $(LOCALBIN)
	$(QUIET)rm -rf $(CRD_DIR)/*.yaml
	$(QUIET)rm -rf client-go
	$(SUCCESS) "Cleanup completed"

#@ $(call go-install-tool {target path with name of binary (ideally with version)} {package url which can be installed} {specific version of package})@通过go install安装二进制
define go-install-tool
$(QUIET)[ -f $(1) ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
GOBIN=$(LOCALBIN) $(GO) install $${package} ;\
mv "$$(echo "$(1)" | sed "s/-$(3)$$//")" $(1) ;\
}
endef

##@ Help
.PHONY: print_all_variables
print_all_variables: FORCE ##print all makefile variables
	$(foreach v, $(sort $(.VARIABLES)), \
		$(if $(filter file,$(origin $(v))), \
		$(info $(shell printf "%-20s" "$(v)")= $($(v)))) \
	)
.PHONY: help
help: FORCE ##print help
	$(call print_help_from_makefile)

.PHONY: params
params: FORCE ##print params
	$(call print_params_from_makefile)