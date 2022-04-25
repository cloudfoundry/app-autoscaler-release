#!/usr/bin/env bash

set -euo pipefail
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
autoscaler_dir="${script_dir}/../../../app-autoscaler-release"

function golang_version {
  version=$(<${autoscaler_dir}/packages/golang-1-linux/version)
  echo $version | sed s/\\./-/g
}

function java_version {
  pushd "${autoscaler_dir}"; git status; popd
  ls "${autoscaler_dir}/packages/openjdk-11/"
  cat "${autoscaler_dir}/packages/openjdk-11/spec" | grep -e "- jdk-" | sed -E 's/- jdk-(.*)\.tar\.gz/\1/g'
}

function configure_git_credentials(){
 if [[ -z $(git config --global user.email) ]]; then
       git config --global user.email "${GIT_USER_EMAIL}"
     fi
  if [[ -z $(git config --global user.name) ]]; then
       git config --global user.name "${GIT_USER_NAME}"
  fi
}

pushd "${autoscaler_dir}" > /dev/null
  version=$(${type}_version)
  dashed_version=$(echo "$version" | sed s/\\./-/g )

  update_branch="${type}-version-bump-${dashed_version}"
  pr_title="Update ${type} version to ${version}"
  pr_description="Automatic version bump of ${type} to ${version}"

  configure_git_credentials

  git checkout -b "${update_branch}"
  git commit -a -m "${pr_title}"
  gh auth login --with-token "${github_token}"
  gh pr create --base origin/main --title "${pr_title}" --body "${pr_description}"
popd > /dev/null
