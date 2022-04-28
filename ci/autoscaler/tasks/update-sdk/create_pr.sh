#!/usr/bin/env bash

set -exuo pipefail
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
autoscaler_dir="${script_dir}/../../../../app-autoscaler-release"
java_dir="${script_dir}/../../../../java-release"

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
  echo "${github_token}" | gh auth login --with-token
  gh pr create --base origin/main --title "${pr_title}" --body "${pr_description}"
popd > /dev/null
