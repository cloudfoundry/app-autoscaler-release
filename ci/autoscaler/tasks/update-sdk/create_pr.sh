#!/usr/bin/env bash

set -exuo pipefail
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
autoscaler_dir="${script_dir}/../../../../app-autoscaler-release"
java_dir="${script_dir}/../../../../java-release"

function golang_version {
  cat ${autoscaler_dir}/packages/golang-1-linux/version
}

function java_version {
  cat "${java_dir}/packages/openjdk-11/spec" | grep -e "- jdk-" | sed -E 's/- jdk-(.*)\.tar\.gz/\1/g'
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
  dashed_version=$(echo "$version" | sed -E 's/[._]/-/g' )

  update_branch="${type}-version-bump-${dashed_version}"
  pr_title="Update ${type} version to ${version}"
  pr_description="Automatic version bump of ${type} to ${version}"

  configure_git_credentials

  git checkout -b "${update_branch}"
  git commit -a -m "${pr_title}"

  printenv GITHUB_ACCESS_TOKEN | gh auth login --with-token -h github.com

  mkdir -p "$HOME/.ssh"
  chmod 700 "$HOME/.ssh"
  printenv GITHUB_PRIVATE_KEY > "$HOME/.ssh/id_rsa"
  chmod 600 "$HOME/.ssh/id_rsa"
  eval "$(ssh-agent -s)"
  ssh-add "$HOME/.ssh/id_rsa"
  ssh-keyscan -t rsa,dsa github.com 2>&1 > "$HOME/.ssh/known_hosts"

  git push --set-upstream origin "${update_branch}"

  gh pr create --base origin/main --title "${pr_title}" --body "${pr_description}"
popd > /dev/null
