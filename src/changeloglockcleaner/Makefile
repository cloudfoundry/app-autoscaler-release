SHELL := /bin/bash
GO := GO111MODULE=on GO15VENDOREXPERIMENT=1 go
GO_NOMOD := GO111MODULE=off go
GO_VERSION := $(shell $(GO) version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
GO_DEPENDENCIES := $(shell find . -type f -name '*.go')
PACKAGE_DIRS := $(shell $(GO) list ./... | grep -v /vendor/ | grep -v e2e)

CGO_ENABLED = 0
BUILDTAGS :=

build:
	@echo "# building changeloglockcleaner"
	@CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(BUILDTAGS) $(BUILDFLAGS) -o changeloglockcleaner main.go

build_tests:
	@for package in $$( find . -name "*.go" -exec dirname {} \; | sort | uniq  ); do echo " - compiling changeloglockcleaner/$${package}"; go test -c -o /dev/null $${package}; done;

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
	golangci-lint run

lint-fix:
	golangci-lint run --fix
