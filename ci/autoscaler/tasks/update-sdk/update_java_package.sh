#! /usr/bin/env bash

# This script is used to update the java version used in the bosh packaging
# It downloads the given SAP Machine java distribution and updates the bosh packaging with the new version.
# usage: ./update_java_package 21.0.3

[ -n "${DEBUG}" ] && set -x
set -euo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source "${script_dir}/vars.source.sh"

JAVA_VERSION=${1:-"21.0.3"} # default java version
desired_major_version=${JAVA_VERSION%%.*}

# identify current version
current_java_version=$(find ./packages -type d -name "openjdk-*" -exec bash -c '
                          directory_name="$1"
                          version_number=${directory_name#*-}
                          echo "${version_number}"
                          ' bash {} \;)
echo "Current Java Major version is ${current_java_version}"

# Step 1 --> Download java...
source "${script_dir}/download_java.sh" "${JAVA_VERSION}"

# Step 2 --> update bosh java package references
echo "updating bosh java packages..."
# shellcheck disable=SC2038
find . -type f ! -name "*.yml" ! -name "update_java_package.sh" ! -path '*/\.*' -exec grep -l "openjdk-${current_java_version}" {} \;| xargs sed -i "s/openjdk-${current_java_version}/openjdk-${desired_major_version}/g"

mv ./packages/openjdk-"${current_java_version}" ./packages/openjdk-"${desired_major_version}"

echo "  -- creating spec file"
binary_name="sapmachine-jdk-${JAVA_VERSION}_linux-x64_bin.tar.gz"
cat > "packages/openjdk-$desired_major_version/spec" <<EOF
---
name: openjdk-${desired_major_version}
dependencies: []
files:
- binaries/**/${binary_name}  # from https://github.com/SAP/SapMachine/releases/download/sapmachine-${JAVA_VERSION}/${binary_name}
EOF

echo "  -- creating packaging script "
cat > "packages/openjdk-$desired_major_version/packaging" <<EOF
set -ex
# compile and runtime bosh env vars
export JAVA_HOME=/var/vcap/packages/openjdk-$desired_major_version
export PATH=\$JAVA_HOME/bin:\$PATH

cd \${BOSH_INSTALL_TARGET}

# extract jdk from src/binaries/jdk
tar zxvf \${BOSH_COMPILE_TARGET}/binaries/jdk/*.tar.gz --strip 1
EOF

rm "packages/openjdk-$desired_major_version/spec.lock" || true

# creates pr
echo -n "${JAVA_VERSION}" > "${AUTOSCALER_DIR}/version"
echo -n "${JAVA_VERSION}" > "${AUTOSCALER_DIR}/vendored-commit"

echo -n "${JAVA_VERSION}" > "./packages/openjdk-${desired_major_version}/version"
