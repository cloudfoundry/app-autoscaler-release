#!/usr/bin/env bash

set -euo pipefail
autoscaler_dir="app-autoscaler-release"

function golang_version {
  version=$(<${autoscaler_dir}/packages/golang-1-linux/version)
  echo $version | sed s/\\./-/g
}

function java_version {
  cat "${autoscaler_dir}/packages/openjdk-11/spec" | grep -e "- jdk-" | sed -E 's/- jdk-(.*)\.tar\.gz/\1/g'
}

pushd app-autoscaler-release > /dev/null
  version=$(${type}_version)
  dashed_version=$(echo $version | sed s/\\./-/g )

  update_branch="${type}-version-bump-${dashed_version}"
  pr_title="Update ${type} version to ${version}"
  pr_description="Automatic version bump of ${type} to ${version}"

  git checkout -b ${update_branch}
  git commit -a -m "${pr_title}"
  gh auth login --with-token "${github_token}"
  gh pr create --base origin/main --title "${pr_title}" --body "${pr_description}"
popd > /dev/null
