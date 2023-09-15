.ONESHELL:
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c ${SHELLFLAGS}
aes_terminal_font_yellow := \e[38;2;255;255;0m
aes_terminal_reset := \e[0m

# TODO: Do we need the next line?
MAKEFLAGS=

GO_VERSION = $(shell go version | sed --expression='s/^[^0-9.]*\([0-9.]*\).*/\1/')
GO_DEPENDENCIES = $(shell find . -type f -name '*.go')
PACKAGE_DIRS = $(shell go list ./... | grep --invert-match /vendor/ | grep --invert-match e2e)
CGO_ENABLED := 0
export GOWORK := off

binaries = $(shell find . -name "main.go" -exec dirname {} \; \
									 | cut --delimiter='/' --fields='2' | sort | uniq)
test_dirs = $(shell find . -name "*_test.go" -exec dirname {} \; \
										| cut --delimiter='/' --fields='2' | sort | uniq)

GINKGO_OPTS :=-r --race --require-suite --randomize-all --cover ${OPTS}
export CONFIG ?= ../../../../acceptance_config.json

GINKGO_VERSION = v$(shell cat ../../../../../.tool-versions | grep --regexp='ginkgo' | cut --delimiter=' ' --fields='2')
GOLANGCI_LINT_VERSION = v$(shell cat ../../../../../.tool-versions | grep --regexp='golangci-lint' | cut --delimiter=' ' --fields='2')



openapi-spec-path := ../../../../../api
openapi-specs-list := $(wildcard ${openapi-spec-path}/*.openapi.yaml)
appfakes-path := ./internal/app/appfakes

.PHONY: generate-fakes
generate-fakes: ${appfakes-path} $(wildcard ${appfakes-path}/*.go)
${appfakes-path} $(wildcard ${appfakes-path}/*.go) &: ./internal/generate.go ${openapi-specs-list}
	@echo -ne '${aes_terminal_font_yellow}'
	@echo -e '⚠️ The client-fakes generated from the openapi-specification depend on\n' \
					 'the files ./go.mod and ./go.sum. This has not been reflected in this\n' \
					 'make-target to avoid cyclic dependencies because `go mod tidy`, which\n' \
					 'modifies both files, depends itself on the client-fakes.'
	@echo -ne '${aes_terminal_reset}'
	go generate ./...



.PHONY: go-mod-tidy
go-mod-tidy: ${appfakes-path} $(wildcard ${appfakes-path}/*.go)
	go mod tidy



.PHONY: build
build: ./build/app ./build/manifest.yml
./build/app ./build/manifest.yml: ./go.mod ${appfakes-path} $(wildcard ${appfakes-path}/*.go)
	echo '# building  test-app'
	mkdir -p build
	CGO_ENABLED='${CGO_ENABLED}' GOOS='linux' GOARCH='amd64' go build -o './build/app'
	cp './app_manifest.yml' './build/manifest.yml'



.PHONY: build_tests
build_tests: $(addprefix build_test-,$(test_dirs))

build_test-%: ${appfakes-path} $(wildcard ${appfakes-path}/*.go)
	@echo " - building '$*' tests"
	@export build_folder='${PWD}/build/tests/$*'
	@mkdir -p '$${build_folder}'
	cd $*
	for package in $$(go list './...' | sed 's|.*/autoscaler/$*|.|' | awk '{ print length, $$0 }' | \
										sort --numeric-sort --reverse | cut --delimiter=' ' --fields='2-')
	do
		export test_file="$${build_folder}/$${package}.test"
		echo "   - compiling $${package} to $${test_file}"
		go test -c -o "$${test_file}" "$${package}"
	done


.PHONY: check lint lint-fix test
check: lint build test

test: generate-fakes
	@echo 'Running tests'
	go run 'github.com/onsi/ginkgo/v2/ginkgo@${GINKGO_VERSION}' run ${GINKGO_OPTS} './...'

lint: generate-fakes
	@go run 'github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}' run

lint-fix: generate-fakes
	go run 'github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}' run --fix



.PHONY: start
start: build
	docker run --interactive --tty --name go_app --volume="$${PWD}/build/:/cf/bin" \
						 --publish '8080:8080' --rm 'cloudfoundry/cflinuxfs4' '/cf/bin/app'



.PHONY: deploy
deploy: build
	./deploy.sh



.PHONY: clean
clean:
	@echo "# cleaning autoscaler"
	@go clean -cache -testcache
	@rm --force --recursive './build'
	@rm --force --recursive './internal/app/appfakes'
