#! /usr/bin/env bash

# shellcheck disable=SC2154

[ -n "${DEBUG}" ] && set -x
set -euo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
autoscaler_dir="${script_dir}/../../../../../app-autoscaler-release"
github_access_token=${GITHUB_ACCESS_TOKEN:-)}
github_private_key=${GITHUB_PRIVATE_KEY:-}

function add_private_key(){
  if [ -n "${github_private_key}" ]; then
    mkdir -p "$HOME/.ssh"
    chmod 700 "$HOME/.ssh"
    printenv GITHUB_PRIVATE_KEY > "$HOME/.ssh/id_rsa"
    chmod 600 "$HOME/.ssh/id_rsa"
    eval "$(ssh-agent -s)"
    ssh-add "$HOME/.ssh/id_rsa"
    ssh-keyscan -t rsa,dsa github.com > "$HOME/.ssh/known_hosts" 2>&1
  fi
}

function login_gh(){
  if [ -n "${github_access_token}" ]; then
    printenv github_access_token | gh auth login --with-token -h github.com
  fi
}

function configure_git_credentials(){
 if [[ -z $(git config --global user.email) ]]; then
       git config --global user.email "${GIT_USER_EMAIL}"
     fi
  if [[ -z $(git config --global user.name) ]]; then
       git config --global user.name "${GIT_USER_NAME}"
  fi
}

if [ "$( git status -s | wc -l)" -eq 0 ]; then
  echo " - Nothing changed !! "
  exit 0
fi

package_version=$(cat ./version)
package_sha=$(cat ./vendored-commit)

dashed_version=$(echo "${package_version}" | sed -E 's/[._]/-/g' )
update_branch="${type}-version-bump-${dashed_version}_${package_sha}"
pr_title="Update ${type} version to ${package_version}"
pr_description="Automatic version bump of ${type} to \`${package_version}\`<br/>Package commit sha: [${package_sha}](https://github.com/bosh-packages/${type}-release/commit/${package_sha})"
login_gh
add_private_key
configure_git_credentials

pushd "${autoscaler_dir}" > /dev/null
  git checkout -b "${update_branch}"
  git commit -a -m "${pr_title}"
  git push --set-upstream origin "${update_branch}"
  gh pr create --base main --head "${update_branch}" --title "${pr_title}" --body "${pr_description}" --label 'dependencies'
popd > /dev/null
