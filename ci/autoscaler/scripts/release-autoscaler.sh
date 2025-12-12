#! /usr/bin/env bash

[ -n "${DEBUG}" ] && set -x

set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/vars.source.sh"

mkdir -p "build"
build_path=$(realpath build)
build_opts=${BUILD_OPTS:-"--final"}
mkdir -p "keys"
keys_path="$(realpath keys)"
PERFORM_BOSH_RELEASE=${PERFORM_BOSH_RELEASE:-"true"}
export UPLOADER_KEY=${UPLOADER_KEY:-"NOT_SET"}
CI=${CI:-false}
SUM_FILE="${build_path}/artifacts/files.sum.sha256"

function create_release() {
   set -e
   mkdir -p "${build_path}/artifacts"
   local version=$1
   local build_path=$2
   local release_file=$3
   echo " - building new release from ${PWD} at revision $(git rev-parse HEAD)"
   echo " - creating release '${version}' in '${build_path}' as ${release_file}"

   yq eval -i ".properties.\"autoscaler.apiserver.info.build\".default = \"${version}\"" jobs/golangapiserver/spec

   git add jobs/golangapiserver/spec

   [ "${CI}" = "true" ] && git commit -S -m "Updated release version to ${version} in golangapiserver"

   # shellcheck disable=SC2086
   bosh create-release \
        ${build_opts} \
        --version "${version}" \
        --tarball="${build_path}/artifacts/${release_file}"
}



function commit_release(){
  git add -A
  git status
  git commit -S -m "created release v${VERSION}"
}

function create_bosh_config(){
   # generate the private.yml file with the credentials
   config_file="${autoscaler_dir}/config/private.yml"
    cat > "$config_file" <<EOF
---
blobstore:
  options:
    credentials_source: static
    json_key:
EOF
    echo ' - Generating private.yml...'
    yq eval -i '.blobstore.options.json_key = strenv(UPLOADER_KEY)' "$config_file"
}

function get_version_from_submodule(){
  echo " - Getting version from src/autoscaler submodule..."
  pushd "${autoscaler_dir}/src/autoscaler" > /dev/null
  local tag_version
  tag_version=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
  popd > /dev/null

  if [ -z "$tag_version" ]; then
    echo " - ERROR: No tag found in src/autoscaler submodule"
    exit 1
  fi

  # Remove 'v' prefix if present
  local version_number=${tag_version#v}

  echo " - Version from submodule: $version_number"
  echo "$version_number" > "${build_path}/name"
}

function generate_changelog(){
  [ -e "${build_path}/changelog.md" ] && return
	echo " - Generating changelog using github cli..."
	mkdir -p "${build_path}"
	# Check if release exists and is a draft before deleting
	if gh release view "${VERSION}" &>/dev/null; then
		local is_draft
		is_draft=$(gh release view "${VERSION}" --json isDraft --jq '.isDraft')
		if [ "$is_draft" = "true" ]; then
			echo " - Deleting existing draft release ${VERSION}"
			gh release delete "${VERSION}" --yes
		else
			echo " - ERROR: Release ${VERSION} already exists and is published (not a draft)"
			echo " - Refusing to delete published release. Please check version logic."
			exit 1
		fi
	fi
	gh release create "${VERSION}" --generate-notes --draft
	gh release view "${VERSION}" --json body --jq '.body' > "${build_path}/changelog.md"
}

function upload_assets_and_promote(){
  if [ "${PERFORM_BOSH_RELEASE}" != "true" ]; then
    echo " - Skipping asset upload and promotion (PERFORM_BOSH_RELEASE=${PERFORM_BOSH_RELEASE})"
    return
  fi

  echo " - Uploading artifacts to release ${VERSION}..."
  gh release upload "${VERSION}" "${build_path}/artifacts/"* --clobber

  echo " - Updating release notes with deployment information..."
  gh release edit "${VERSION}" --notes-file "${build_path}/changelog.md"

  echo " - Publishing release ${VERSION}..."
  gh release edit "${VERSION}" --draft=false --target="v${VERSION}"
  echo " - Release ${VERSION} published successfully!"
}

function setup_git(){
  if [[ -z $(git config --global user.email) ]]; then
    git config --global user.email "${AUTOSCALER_CI_BOT_EMAIL}"
  fi

  if [[ -z $(git config --global user.name) ]]; then
    git config --global user.name "${AUTOSCALER_CI_BOT_NAME}"
  fi

  # Add GitHub's SSH host keys to known_hosts to avoid interactive prompt
  mkdir -p ~/.ssh
  ssh-keyscan -t ed25519,rsa github.com >> ~/.ssh/known_hosts 2>/dev/null

  # Setup SSH authentication key for GitHub pushes
  if [ -n "${AUTOSCALER_CI_BOT_SSH_KEY:-}" ]; then
    ssh_auth_key_path="${keys_path}/autoscaler-ci-bot-ssh-key"
    echo "$AUTOSCALER_CI_BOT_SSH_KEY" > "${ssh_auth_key_path}"
    chmod 600 "${ssh_auth_key_path}"

    # Configure SSH to use this key for GitHub
    cat >> ~/.ssh/config <<EOF
Host github.com
  HostName github.com
  User git
  IdentityFile ${ssh_auth_key_path}
  IdentitiesOnly yes
EOF
    chmod 600 ~/.ssh/config
  fi

  public_key_path="${keys_path}/autoscaler-ci-bot-signing-key.pub"
  private_key_path="${keys_path}/autoscaler-ci-bot-signing-key"
  echo "$AUTOSCALER_CI_BOT_SIGNING_KEY_PUBLIC" > "${public_key_path}"
  echo "$AUTOSCALER_CI_BOT_SIGNING_KEY_PRIVATE" > "${private_key_path}"
  chmod 600 "${public_key_path}"
  chmod 600 "${private_key_path}"

  git config --global gpg.format ssh
  git config --global user.signingkey "${private_key_path}"
}


pushd "${autoscaler_dir}" > /dev/null
  setup_git
  create_bosh_config
  get_version_from_submodule

  VERSION=${VERSION:-$(cat "${build_path}/name")}
  generate_changelog

  echo " - Displaying diff..."
  export GIT_PAGER=cat
  git diff
  echo "v${VERSION}" > "${build_path}/tag"
  if [ "${PERFORM_BOSH_RELEASE}" == "true" ]; then
    RELEASE_TGZ="app-autoscaler-v${VERSION}.tgz"
    create_release "${VERSION}" "${build_path}" "${RELEASE_TGZ}"
    [ "${CI}" = "true" ] && commit_release

    sha256sum "${build_path}/artifacts/"* > "${build_path}/artifacts/files.sum.sha256"
    RELEASE_SHA256=$( grep "${RELEASE_TGZ}$" "${SUM_FILE}" | awk '{print $1}')
  else
    RELEASE_SHA256="dummy-sha"
  fi
  export RELEASE_SHA256

  # Push commits and tag before promoting the release
  if [ "${CI}" = "true" ]; then
    echo " - Creating and pushing tag v${VERSION}..."
    git tag -s -m "Release v${VERSION}" "v${VERSION}"

    echo " - Fetching latest changes from remote..."
    git fetch origin main

    echo " - Checking out main branch..."
    git checkout -B main

    echo " - Rebasing local changes on top of remote..."
    git rebase origin/main

    echo " - Pushing to main branch..."
    git push origin main

    echo " - Pushing tag..."
    git push origin "v${VERSION}"
  fi

  cat >> "${build_path}/changelog.md" <<EOF

## Deployment

\`\`\`yaml
releases:
- name: app-autoscaler
  version: ${VERSION}
  url: https://storage.googleapis.com/app-autoscaler-releases/releases/app-autoscaler-v${VERSION}.tgz
  sha1: sha256:${RELEASE_SHA256}
\`\`\`
EOF
  echo "---------- Changelog file ----------"
  cat "${build_path}/changelog.md"
  echo "---------- end file ----------"

  upload_assets_and_promote

popd > /dev/null
echo " - Completed"
