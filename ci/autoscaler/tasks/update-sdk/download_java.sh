#! /usr/bin/env bash

# This script is used to download java version used in the bosh packaging
# It fetch the given SAP Machine java distribution
# usage: ./download_java.sh 21.0.3

set -euo pipefail

JAVA_VERSION=${1:-"22.0.1"} # default java version

# Step 1 --> Download java from https://github.com/SAP/SapMachine/releases/download/sapmachine-21.0.3/sapmachine-jdk-21.0.3_linux-x64_bin.tar.gz
SAP_MACHINE_BASE_URL="https://github.com/SAP/SapMachine"
binary_name="sapmachine-jdk-${JAVA_VERSION}_linux-x64_bin.tar.gz"

echo "Fetching latest SAP Machine Java ${JAVA_VERSION} JDK for Linux"
printf "\n"

jdk_download_url="${SAP_MACHINE_BASE_URL}/releases/download/sapmachine-${JAVA_VERSION}/${binary_name}"

echo "Fetching ${jdk_download_url}"
mkdir -p src/binaries/jdk && pushd src/binaries/jdk > /dev/null
  curl -JLO "${jdk_download_url}"
popd > /dev/null

# Step 2 --> Build java
jdk_version="$(find . -name "sapmachine*.tar.gz" |   cut -d'-' -f3 | cut -d'_' -f1)"
echo "Downloaded JDK version: $jdk_version"
