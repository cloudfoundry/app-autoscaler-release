SHELL := /bin/bash
.SHELLFLAGS = -euo pipefail -c
MAKEFLAGS = -s
GO := GO111MODULE=on GO15VENDOREXPERIMENT=1 go
GO_NOMOD := GO111MODULE=off go
GO_VERSION := $(shell $(GO) version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
GO_DEPENDENCIES := $(shell find . -type f -name '*.go')
PACKAGE_DIRS := $(shell $(GO) list ./... | grep -v /vendor/ | grep -v e2e)

CGO_ENABLED = 0
export GOWORK=off
BUILDTAGS :=

GOLANGCI_LINT_VERSION=v$(shell cat ../../.tool-versions | grep golangci-lint  | cut -d " " -f 2 )

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

check: lint build_tests

test-unit:
	ginkgo --race run helpers

test:
	bin/test $(filter-out $@,$(MAKECMDGOALS))

lint:
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} run

lint-fix:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} run --fix
