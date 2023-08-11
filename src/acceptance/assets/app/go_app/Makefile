SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c ${SHELLFLAGS}
MAKEFLAGS=
GO_VERSION := $(shell go version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
GO_DEPENDENCIES := $(shell find . -type f -name '*.go')
PACKAGE_DIRS := $(shell go list ./... | grep -v /vendor/ | grep -v e2e)
CGO_ENABLED = 0
export GOWORK=off

binaries=$(shell find . -name "main.go" -exec dirname {} \; |  cut -d/ -f2 | sort | uniq)
test_dirs=$(shell   find . -name "*_test.go" -exec dirname {} \; |  cut -d/ -f2 | sort | uniq)

GINKGO_OPTS=-r --race --require-suite --randomize-all --cover ${OPTS}
export CONFIG?=../../../../acceptance_config.json

.PHONY: build
build:
	echo "# building  test-app"
	rm -rf build/* || true
	mkdir -p build/
	CGO_ENABLED='${CGO_ENABLED}' GOOS='linux' GOARCH='amd64' go build -o build/app
	cp app_manifest.yml build/manifest.yml


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

check: lint build test
test:
	@echo "Running tests"
	ginkgo run ${GINKGO_OPTS} ./...

lint:
	@cd ../../../../../; make lint_acceptance OPTS=${OPTS}

.PHONY: start
start: build
	docker  run -it --name go_app -v $$PWD/build/:/cf/bin -p 8080:8080 --rm cloudfoundry/cflinuxfs4 /cf/bin/app

.PHONY: deploy
deploy: build
	./deploy.sh

generate: go.mod ../../../../../api/custom-metrics-api.yaml
	go generate ./...

.PHONY: clean
clean:
	@echo "# cleaning autoscaler"
	@go clean -cache -testcache
	@rm -rf build
