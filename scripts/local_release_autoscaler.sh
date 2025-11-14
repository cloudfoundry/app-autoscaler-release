#! /usr/bin/env bash
#
# shellcheck disable=SC1091
#
set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

DEPLOYMENT=foo
DEBUG=true
DEST="${script_dir}/../build"
BUILD_OPTS="--force"
AUTOSCALER_CI_BOT_NAME="foo"
AUTOSCALER_CI_BOT_EMAIL="foo@bar.baz"
PREV_VERSION="$(yq  ".properties.\"autoscaler.apiserver.info.build\".default" jobs/golangapiserver/spec)"

VERSION="$(cat "${script_dir}/../VERSION")-pre"

export DEPLOYMENT
export DEBUG
export DEST
export BUILD_OPTS
export AUTOSCALER_CI_BOT_NAME
export AUTOSCALER_CI_BOT_EMAIL
export PREV_VERSION
export VERSION

# check for GITHUB_TOKEN
if [ -z "${GITHUB_TOKEN}" ]; then
	echo "GITHUB_TOKEN is not set"
	exit 1
fi

find_or_create_ssh_key() {
	if [ -f ~/.ssh/id_ed25519 ]; then
		echo "ssh key already exists"
		return
	fi

	ssh-keygen -t ed25519 -C "${AUTOSCALER_CI_BOT_EMAIL}" -f ~/.ssh/id_ed25519 -N ""
}

prerelease() {
	pushd "${script_dir}/.." > /dev/null
		make clean generate-fakes generate-openapi-generated-clients-and-servers go-mod-vendor db scheduler
	popd > /dev/null
}

delete_dev_releases() {
	rm -rf dev_releases
}


release_autoscaler() {
	AUTOSCALER_CI_BOT_SIGNING_KEY_PUBLIC=$(cat ~/.ssh/id_ed25519.pub)
	AUTOSCALER_CI_BOT_SIGNING_KEY_PRIVATE=$(cat ~/.ssh/id_ed25519)
	export AUTOSCALER_CI_BOT_SIGNING_KEY_PUBLIC
	export AUTOSCALER_CI_BOT_SIGNING_KEY_PRIVATE

	source "${script_dir}/../ci/autoscaler/scripts/release-autoscaler.sh"
	echo "beware that it adds a commit you need to drop each time also you need to remove dev_releases from root."
}

main() {
	find_or_create_ssh_key
	delete_dev_releases
	prerelease
	release_autoscaler
}

main

