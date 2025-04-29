.ONESHELL:
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

repository-root := $(shell realpath --relative-to $$(pwd) $$(git rev-parse --show-toplevel))

# All files that together define the development-environment:
#
# 🚧 To-do: Check if each of the last two lines should be substituted by something like:
# `$(wildcard ./.devbox/**/flake.{nix,lock})`
development-environment-definition := \
	${repository-root}/nix/flake.nix \
	${repository-root}/nix/flake.lock \
	${repository-root}/devbox.json \
	${repository-root}/devbox.lock \
	${repository-root}/.devbox/virtenv/mysql/flake/flake.nix \
	${repository-root}/.devbox/virtenv/mysql/flake/flake.lock
# 🚸 Please be aware that devbox automatically creates
# '${repository-root}/.devbox/virtenv/mysql/flake' when it loads the development-environment.

# The generated Dockerfile just depends on the content of the used “devbox.json” and its lockfile.
# It is marked secondary to not auto-clean it.
#
# ⚠️ As the Dockerfile that is auto-generated via its target “Dockerfile” is not of high quality and
# may contain errors, its dependencies are commented out to prevent re-generation by gnu-make.
.SECONDARY: Dockerfile
Dockerfile: # ./Makefile ${repository-root}/devbox.json ${repository-root}/devbox.lock
	devbox generate dockerfile --config='${repository-root}'
	mv '${repository-root}/Dockerfile' .

# 🚸 Overwrite the following parameters to work with a different repository on a different registry:
DOCKER_REGISTRY ?= ghcr.io
DOCKER_REPOSITORY ?= cloudfoundry/app-autoscaler-release-tools

.PHONY: docker-login
docker-login:
	@logged_in=$$(jq '.auths | has("${DOCKER_REGISTRY}")' ~/.docker/config.json)
	readonly logged_in
	if [[ "$${logged_in}" != 'true' ]]
	then
		docker login '${DOCKER_REGISTRY}'
	else
		echo 'Already logged in into ${DOCKER_REGISTRY}, skipping login.'
	fi

oci-image-location := ${repository-root}/build/app-autoscaler-release-tools.tar.zst
.PHONY: docker-image
docker-image: ${oci-image-location}

${oci-image-location}: ./Makefile Dockerfile ${development-environment-definition}
	@readonly image_tag='${DOCKER_REGISTRY}/${DOCKER_REPOSITORY}:latest'
	docker build --file='./Dockerfile' '${repository-root}' --tag="$${image_tag}"
	docker image save "$${image_tag}" | zstd > '${oci-image-location}'

.PHONY: docker-image-push
docker-image-push: docker-login docker-image
	@ARTIFACT='${DOCKER_REGISTRY}/${DOCKER_REPOSITORY}:$*'
	echo "Uploading image" "$${ARTIFACT}"
	docker push "$${ARTIFACT}"



.PHONY: set-autoscaler-pipeline
set-autoscaler-pipeline:
	@./autoscaler/set-pipeline.sh

.PHONY: set-infrastructure-pipeline
set-infrastructure-pipeline:
	@./infrastructure/set-pipeline.sh

.PHONY: unpause-pipeline
unpause-pipeline:
	@./scripts/unpause-pipeline.sh

.PHONY: delete-pipeline
delete-pipeline:
	@./scripts/delete-pipeline.sh

.PHONY: lint
lint:
	@shellcheck --external-sources --shell='bash' **/scripts/*.sh

.PHONY: fix-lint
fix-lint:
	@shellcheck --external-sources --shell='bash' -f diff **/scripts/*.sh

.PHONY: browse
browse:
	@open https://concourse.app-runtime-interfaces.ci.cloudfoundry.org
