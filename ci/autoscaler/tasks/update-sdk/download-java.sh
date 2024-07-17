#! /usr/bin/env bash

# This script is used to update the java version used in the bosh packaging
# It fetch the given SAP Machine java distribution and updates the bosh packaging with the new version.
# usage: ./get-java.sh 21.0.3

set -euo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source "${script_dir}/vars.source.sh"
create_pr=${CREATE_PR:-"false"}

JAVA_VERSION="$1"

# current version
current_java_version=$(find ./packages -type d -name "openjdk-*" -exec bash -c '
                          directory_name="$1"
                          version_number=${directory_name#*-}
                          echo "${version_number}"
                          ' bash {} \;)
echo " Current Java Major version is ${current_java_version}"

# Step 1 --> Download java from https://github.com/SAP/SapMachine/releases/download/sapmachine-21.0.3/sapmachine-jdk-21.0.3_linux-x64_bin.tar.gz
SAP_MACHINE_BASE_URL="https://github.com/SAP/SapMachine"
desired_major_version=${JAVA_VERSION%%.*}
binary_name="sapmachine-jdk-${JAVA_VERSION}_linux-x64_bin.tar.gz"

echo "Fetching latest SAP Machine Java ${JAVA_VERSION} JDK for Linux"
printf "\n"

jdk_download_url="${SAP_MACHINE_BASE_URL}/releases/download/sapmachine-${JAVA_VERSION}/${binary_name}"

echo "Fetching ${jdk_download_url}"
mkdir -p src/binaries/jdk && pushd src/binaries/jdk
  curl -JLO "${jdk_download_url}"
popd

# Step 2 --> Build java
jdk_version="$(find . -name "sapmachine*.tar.gz" |   cut -d'-' -f3 | cut -d'_' -f1)"
echo "Downloaded JDK version: $jdk_version"
echo -n "${jdk_version}" > "./packages/openjdk-${current_java_version}/version"
#
## Step 3 --> update bosh java package
#echo "updating bosh java packages..."
#
#mv ./packages/openjdk-"${current_java_version}" ./packages/openjdk-"${desired_major_version}"
## shellcheck disable=SC2038
#find . -type f ! -name "*.yml" ! -name "update_java_package.sh" ! -path '*/\.*' -exec grep -l "openjdk-${current_java_version}" exec {} \; | xargs sed -i "s/openjdk-${current_java_version}/openjdk-${desired_major_version}/g"
