https://sap.github.io/SapMachine/latest/21/linux-x64/jdk/#! /usr/bin/env bash

[ -n "${DEBUG}" ] && set -x
set -euo pipefail

# Step 1 --> Download java

# sample jdk download url
# https://github.com/SAP/SapMachine/releases/download/sapmachine-21.0.3/sapmachine-jdk-21.0.3_linux-x64_bin.tar.gz
SAP_MACHINE_BASE_URL="https://github.com/SAP/SapMachine"
JAVA_VERSION="21.0.3"
BINARY_NAME="sapmachine-jdk-${JAVA_VERSION}_linux-x64_bin.tar.gz"

echo "Fetching latest SAP Machine Java ${JAVA_VERSION} JDK for Linux"
printf "\n"

jdk_download_url="${SAP_MACHINE_BASE_URL}/releases/download/sapmachine-${JAVA_VERSION}/${BINARY_NAME}"

echo "Fetching ${jdk_download_url}"
mkdir -p src/binaries/jdk && cd src/binaries/jdk

curl -JLO "${jdk_download_url}"

# Step 2 --> Build java
jdk_version="$(find . -name "sapmachine*.tar.gz" |   cut -d'-' -f3 | cut -d'_' -f1)"
echo "Downloaded JDK version: $jdk_version"




