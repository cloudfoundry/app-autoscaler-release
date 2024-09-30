#! /usr/bin/env bash
#
set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

DEPLOYMENT=foo
export DEBUG=true
export PREV_VERSION=12.2.1
export DEST="${script_dir}/../build"
export VERSION="12.3.0"
export BUILD_OPTS="--force"
export AUTOSCALER_CI_BOT_NAME="foo"
export AUTOSCALER_CI_BOT_EMAIL="foo@bar.baz"
export PREV_VERSION=$(cat ${script_dir}/../VERSION)
export VERSION=$(cat ${script_dir}/../VERSION)-pre


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
		make clean generate-fakes generate-openapi-generated-clients-and-servers go-mod-tidy go-mod-vendor db scheduler
	popd > /dev/null
}

delete_dev_releases() {
	rm -rf dev_releases
}


release_autoscaler() {
	export AUTOSCALER_CI_BOT_SIGNING_KEY_PUBLIC=$(cat ~/.ssh/id_ed25519.pub)
	export AUTOSCALER_CI_BOT_SIGNING_KEY_PRIVATE=$(cat ~/.ssh/id_ed25519)
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

