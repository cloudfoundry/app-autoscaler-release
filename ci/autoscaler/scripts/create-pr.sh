#!/usr/bin/env bash

set -eu

: "${GITHUB_TOKEN:?}"

pushd app-autoscaler-release-main
  if git show --pretty="" --name-only | grep -q packages/golang-1-linux/spec.lock; then
    version=$(<packages/golang-1-linux/version)
    dashed_version=$(echo $version | sed s/\\./-/g )

    UPDATE_BRANCH="golang-version-bump-${dashed_version}"
    PR_TITLE="Update golang version to ${version}"
    PR_DESCRIPTION="Automatic version bump of golang to ${version}"

  elif git show --pretty="" --name-only | grep -q packages/packages/openjdk-11/spec.lock; then
    UPDATE_BRANCH="java-version-bump"
    PR_TITLE="Update java version"
    PR_DESCRIPTION="Automatic version bump of java"

  else
    exit 1
  fi


  BODY=("\"head\": \"${UPDATE_BRANCH}\"," "\"base\": \"main\"," "\"title\": \"${PR_TITLE}\"," "\"body\": \"${PR_DESCRIPTION}\"")
  BODY_JSON="{${BODY[*]}}"
  API_URL="https://api.github.com/repos/cloudfoundry/app-autoscaler-release/pulls"

  echo "Creating PR on ${UPDATE_BRANCH} ..."
  RESPONSE=$(curl -i -s -o - -X POST -H "Authorization: token ${GITHUB_TOKEN}" -H "Accept: application/vnd.github.v3+json" -d "${BODY_JSON}" ${API_URL})
  RESPONSE_CODE=$(echo "${RESPONSE}" | grep HTTP | awk '{print $2}')

  if [ "${RESPONSE_CODE}" == "422" ]; then
    EXISTING_PR="$( curl -s -o - -H "Authorization: token ${GITHUB_TOKEN}" "${API_URL}?head=cloudfoundry:${UPDATE_BRANCH}" )"
    URL=$(echo "${EXISTING_PR}" | jq .[0].html_url)
    echo "PR has already been created: ${URL}"
    exit 0
  else
    if [ "${RESPONSE_CODE}" != "201" ]; then
      echo "Error while creating PR. HTTP return code was ${RESPONSE_CODE}. Exiting."
      exit 1
    fi
  fi
popd

URL=$(echo "${RESPONSE}" | grep -Eo "\"html_url\": \"(.*?\/pull\/\\d+)\"," | sed -E "s/.*\"(https.*)\",/\\1/") || true
echo "PR created: ${URL}"
