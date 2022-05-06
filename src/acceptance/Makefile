SHELL := /bin/bash
GO := GO111MODULE=on GO15VENDOREXPERIMENT=1 go
GO_NOMOD := GO111MODULE=off go
GO_VERSION := $(shell $(GO) version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
GO_DEPENDENCIES := $(shell find . -type f -name '*.go')
PACKAGE_DIRS := $(shell $(GO) list ./... | grep -v /vendor/ | grep -v e2e)

CGO_ENABLED = 0
BUILDTAGS :=

build_tests:
	@for package in $$( find . -name "*.go" -exec dirname {} \; | sort | uniq  ); do echo " - compiling acceptance/$${package}"; go test -c -o /dev/null $${package}; done;

check: lint build_tests

test-unit:
	ginkgo --race run helpers

test:
	bin/test $(filter-out $@,$(MAKECMDGOALS))

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix
