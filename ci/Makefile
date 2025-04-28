.ONESHELL:
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

repository-root := $(shell git rev-parse --show-toplevel)

# All files that together define the development-environment:
development-environment-definition := \
	${repository-root}/devbox.json ${repository-root}/devbox.lock \
	${repository-root}/flake.nix ${repository-root}/flake.lock \
	${repository-root}/.devbox/virtenv/mysql/flake/flake.nix \
	${repository-root}/.devbox/virtenv/mysql/flake/flake.lock
# 🚸 Please be aware that devbox automatically creates
# '${repository-root}/.devbox/virtenv/mysql/flake' when it loads the development-environment.

# The generated Dockerfile just depends on the content of the used “devbox.json” and its lockfile.
# It is marked secondary to not auto-clean it.
.SECONDARY: Dockerfile
Dockerfile: ./Makefile ${repository-root}/devbox.json ${repository-root}/devbox.lock
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

# 🚧 To-do: Come with appropriate Dockerignore!
.PHONY: docker-image
docker-image: ./Makefile Dockerfile ${development-environment-definition}
	docker build --file='./Dockerfile' '${repository-root}' \
		--tag='${DOCKER_REGISTRY}/${DOCKER_REPOSITORY}:latest'

# 🚧 To-do: Include workflow to produce the image and push it from the CI/CD-pipeline, see:
# <https://github.com/cloudfoundry/app-autoscaler-release/pull/2943/files#diff-b4df0a4f0d80f73138c476afbd7aefdac9df339642ddfba323d27c8cbabb92e2>
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
