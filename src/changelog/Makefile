SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
MAKEFLAGS := -s

GO := GO111MODULE=on GO15VENDOREXPERIMENT=1 go
GO_NOMOD := GO111MODULE=off go

GO_VERSION := $(shell ${GO} version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
GO_DEPENDENCIES = $(shell find . -type f -name '*.go')
PACKAGE_DIRS = $(shell ${GO} list './...' | grep --invert-match --regexp='/vendor/' \
								 | grep --invert-match --regexp='e2e')

CGO_ENABLED := 0
export GOWORK := off
BUILDTAGS :=

.PHONY: go-mod-tidy go-mod-vendor
go-mod-tidy:
	go mod tidy
go-mod-vendor:
	go mod vendor



build:
	@echo "# building changelog"
	@CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(BUILDTAGS) $(BUILDFLAGS) -o changelog main.go

test_dirs=$(shell   find . -name "*_test.go" -exec dirname {} \; |  cut -d/ -f2 | sort | uniq)
build_tests: $(addprefix build_test-,$(test_dirs))
build_test-%:
	@echo " - building '$*' tests"
	@export build_folder=${PWD}/build/tests/$* &&\
	 mkdir -p $${build_folder} &&\
	 cd $* &&\
	 for package in $$(  go list ./... | sed 's|.*/autoscaler/$*|.|' | awk '{ print length, $$0 }' | sort -n -r | cut -d" " -f2- );\
	 do\
		 export test_file=$${build_folder}/$${package}.test;\
		 echo "   - compiling $${package} to $${test_file}";\
		 go test -c -o $${test_file} $${package};\
	 done;
check: fmt lint build test

test:
	ginkgo -r --race --randomize-all --keep-going

get-fmt-deps: ## Install goimports
	$(GO_NOMOD) get golang.org/x/tools/cmd/goimports

importfmt: get-fmt-deps
	@echo "Formatting the imports..."
	goimports -w $(GO_DEPENDENCIES)

fmt: importfmt
	@FORMATTED=`$(GO) fmt $(PACKAGE_DIRS)`
	@([[ ! -z "$(FORMATTED)" ]] && printf "Fixed unformatted files:\n$(FORMATTED)") || true

lint:
	@golangci-lint run

lint-fix:
	golangci-lint run --fix
