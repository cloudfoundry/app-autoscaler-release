.ONESHELL:
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

repository-root := $(shell realpath --relative-to $$(pwd) $$(git rev-parse --show-toplevel))

# All files that together define the development-environment:
development-environment-definition := \
	$(wildcard ${repository-root}/nix/**/*) \
	${repository-root}/devbox.json \
	${repository-root}/devbox.lock

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
OCI_REGISTRY ?= ghcr.io
OCI_REPOSITORY ?= cloudfoundry/app-autoscaler-release-tools
OCI_TAG ?= latest

.PHONY: docker-login
docker-login:
	@logged_in=$$(jq '.auths | has("${OCI_REGISTRY}")' ~/.docker/config.json)
	readonly logged_in
	if [[ "$${logged_in}" != 'true' ]]
	then
		docker login '${OCI_REGISTRY}'
	else
		echo 'Already logged in into ${OCI_REGISTRY}, skipping login.'
	fi

oci-image-location := ${repository-root}/build/app-autoscaler-release-tools.tar.zst
.PHONY: oci-image docker-image
oci-image: ${oci-image-location}
docker-image: ${oci-image-location}

${oci-image-location}: ./Makefile Dockerfile Dockerfile.dockerignore ${development-environment-definition}
	@readonly image_tag='${OCI_REGISTRY}/${OCI_REPOSITORY}:${OCI_TAG}'
	docker build --file='./Dockerfile' '${repository-root}' --tag="$${image_tag}"
	docker image save "$${image_tag}" | zstd > '${oci-image-location}'

.PHONY: oci-release-%
oci-release-%: docker-login oci-image
	@ARTIFACT='${OCI_REGISTRY}/${OCI_REPOSITORY}:$*'
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
