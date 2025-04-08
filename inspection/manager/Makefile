# SPDX-License-Identifier: Apache-2.0
# Copyright Authors of InfraHI

##@ Default
all: pre-check post-check ##默认命令，包括预检查-构建-部署-后处理
	@echo "Build finished"

pre-check: ##前处理
post-check: ##后处理

ROOT_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
include $(ROOT_DIR)/Makefile.defs

##@ CodeGen

OUTPUT_DIR=output

##@ Build

IMAGE_TAG=$(GIT_VERSION)
ifeq ($(ENV), DEBUG)
    IMAGE_TAG=$(GIT_VERSION)
else ifeq ($(ENV), TEST)
	IMAGE_TAG=$(VERSION)
else ifeq ($(ENV), PROD)
	IMAGE_TAG=$(VERSION)
endif

self-command:
#	apt-get update
#	apt-get install -y python3
#	apt-get install -y python3-pip
#	python3 -m pip install setuptools
#	python3 setup.py sdist bdist_wheel

build-manager-image: ##构建manager镜像
	$(CONTAINER_ENGINE) build -f $(ROOT_DIR)/Dockerfile . -t $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/inspection_manager:$(IMAGE_TAG)

push-manager-image: ##推送manager镜像到镜像仓库
	$(CONTAINER_ENGINE) push $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/inspection_manager:$(IMAGE_TAG)

rm-manager-image:
	$(CONTAINER_ENGINE) rmi $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/inspection_manager:$(IMAGE_TAG)

ci: self-command build-manager-image push-manager-image rm-manager-image ##构建并推送镜像

##@ Develop

CURRENT_BRANCH=
ifeq ($(CI_COMMIT_REF_NAME),)
CURRENT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
else
CURRENT_BRANCH=$(CI_COMMIT_REF_NAME)
endif

ifeq ($(ENV), PROD)
ensure_version: ##修改正式环境version
	@echo "CD inspect-manager with version $(IMAGE_TAG)"
else
ensure_version: incress-patch-version GIT_VERSION ##修改测试环境version并gitpush
	@echo "CD inspect-manager with version $(IMAGE_TAG)"
endif

ifeq ($(ENV), PROD)
update_version: ##修改正式环境version
	@echo "Update inspect-manager with version $(IMAGE_TAG)"
else
update_version: ##修改测试环境version并gitpush
	@echo "Update inspect-manager with version $(IMAGE_TAG) in test environment"
	@# 检查远程分支是否存在
	@if ! git ls-remote --heads origin $(CURRENT_BRANCH) >/dev/null 2>&1; then \
		echo "ERROR: Remote branch $(CURRENT_BRANCH) does not exist. Please create it first."; \
		exit 1; \
	fi
	@git add .
	@git commit -am '[skip ci] $(ENV) version update'
	@# 推送更改
	@git push origin $(CURRENT_BRANCH) --force
endif

incress-major-version: ##修改VERSION文件，将major version增加1
	@echo "$$(($(VERSION_MAJOR)+1)).$(VERSION_MINOR).$(VERSION_PATCH)-$(VERSION_IDENTIFIER)" > $(ROOT_DIR)/VERSION

incress-minor-version: ##修改VERSION文件，将minor version增加1
	@echo "$(VERSION_MAJOR).$$(($(VERSION_MINOR)+1)).$(VERSION_PATCH)-$(VERSION_IDENTIFIER)" > $(ROOT_DIR)/VERSION

incress-patch-version: ##修改VERSION文件，将patch version增加1
	@echo "$(VERSION_MAJOR).$(VERSION_MINOR).$$(($(VERSION_PATCH) + 1))-$(VERSION_IDENTIFIER)" > $(ROOT_DIR)/VERSION

clear-patch-version: ##修改VERSION文件，将patch version置为0
	@echo "$(VERSION_MAJOR).$(VERSION_MINOR).0-$(VERSION_IDENTIFIER)" > $(ROOT_DIR)/VERSION

new-sprint: incress-minor-version clear-patch-version ##开始新的sprint迭代，将会修改版本号
	VER_ = $(shell cat $(ROOT_DIR)/VERSION)
	$(INFO)"Sprint started with Version [$(VER_)]"

GIT_VERSION: FORCE ##update GIT_VERSION
	@if [ "$(GIT_VERSION)" != "`cat 2>/dev/null GIT_VERSION`" ] ; then echo "$(GIT_VERSION)" >GIT_VERSION; fi
	$(INFO)"Succeed write $(GIT_VERSION) to GIT_VERSION"

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
