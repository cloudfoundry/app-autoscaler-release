SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
MAKEFLAGS := -s

GO := GO111MODULE=on GO15VENDOREXPERIMENT=1 go
GO_NOMOD := GO111MODULE=off go

GO_VERSION = $(shell go version | sed -e s/^[^0-9.]*\([0-9.]*\).*/\1/')
GO_DEPENDENCIES = $(shell find . -type f -name '*.go')
PACKAGE_DIRS = $(shell go list './...' | grep --invert-match --regexp='/vendor/' \
								 | grep --invert-match --regexp='e2e')

CGO_ENABLED := 0
export GOWORK := off

GINKGO_VERSION = v$(shell cat '../../.tool-versions' | grep --regexp='ginkgo' \
										| cut --delimiter=' ' --fields='2')
GOLANGCI_LINT_VERSION = v$(shell cat '../../.tool-versions' | grep --regexp='golangci-lint' \
													 | cut --delimiter=' ' --fields='2' )



.PHONY: clean
clean:
	rm --recursive --force './vendor'


.PHONY: go-mod-tidy
go-mod-tidy:
	go mod tidy



go-vendoring-folder := ./vendor
go-vendored-files = $(shell find '${go-vendoring-folder}' -type f -name '*.go' 2> '/dev/null')
## This does not work: go-vendored-files = $(wildcard ${go-vendoring-folder}/**/*.go)

.PHONY: go-mod-vendor
go-mod-vendor: ${go-vendoring-folder} ${go-vendored-files}
${go-vendoring-folder} ${go-vendored-files} &:
	go mod vendor



test_dirs = $(shell find . -name "*_test.go" -exec dirname {} \; | cut --delimiter='/' --fields='2' \
						| sort | uniq)
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
	go run github.com/onsi/ginkgo/v2/ginkgo@${GINKGO_VERSION} --race run helpers config

test:
	bin/test $(filter-out $@,$(MAKECMDGOALS))

lint:
	make -C ../.. lint_acceptance OPTS=${OPTS}
