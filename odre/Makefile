include ../Makefile.defs
include ../Makefile.golang

build:
	$(GO_BUILD) main.go

test:
	$(GO_TEST) | $(GOTEST_FORMATTER)

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