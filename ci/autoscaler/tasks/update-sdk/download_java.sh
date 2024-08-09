#! /usr/bin/env bash

# This script is used to download java version used in the bosh packaging
# It fetch the given SAP Machine java distribution
# usage: ./download_java.sh 21.0.3

set -euo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source "${script_dir}/vars.source.sh"

JAVA_VERSION=${1:-"21.0.3"}

# Step 1 --> Download java from https://github.com/SAP/SapMachine/releases/download/sapmachine-21.0.3/sapmachine-jdk-21.0.3_linux-x64_bin.tar.gz
SAP_MACHINE_BASE_URL="https://github.com/SAP/SapMachine"
binary_name="sapmachine-jdk-${JAVA_VERSION}_linux-x64_bin.tar.gz"

printf "\n"
jdk_download_url="${SAP_MACHINE_BASE_URL}/releases/download/sapmachine-${JAVA_VERSION}/${binary_name}"

echo "Fetching SAP Machine JDK  ${JAVA_VERSION} from ${jdk_download_url}"
# shellcheck disable=SC2154
pushd "${autoscaler_dir}"
  mkdir -p src/binaries/jdk && pushd src/binaries/jdk > /dev/null
  curl -JLO "${jdk_download_url}"
  popd > /dev/null
popd > /dev/null

# Step 2 --> Build java
jdk_version="$(find . -name "sapmachine*.tar.gz" |   cut -d'-' -f3 | cut -d'_' -f1)"
echo "Downloaded JDK version: $jdk_version"
