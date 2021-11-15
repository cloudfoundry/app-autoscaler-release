SHELL := /bin/bash
modules:= acceptance changelog changeloglockcleaner
# TODO add to the module list "apitester common app-autoscaler"

lint_config:=${PWD}/.golangci.yaml

.PHONY: golangci-lint lint $(addprefix lint_,$(modules))
lint: golangci-lint $(addprefix lint_,$(modules))

golangci-lint:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

$(addprefix lint_,$(modules)): lint_%:
	@echo " - linting: $(patsubst lint_%,%,$@)"
	@pushd src/$(patsubst lint_%,%,$@) >/dev/null && golangci-lint --config ${lint_config} run

spec-test:
	bundle exec rspec