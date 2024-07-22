#! /usr/bin/env bash

# This script is used to update the java version used in the bosh packaging
# It downloads the given SAP Machine java distribution and updates the bosh packaging with the new version.
# usage: ./update_java_package 21.0.3

[ -n "${DEBUG}" ] && set -x
set -euox pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source "${script_dir}/vars.source.sh"

JAVA_VERSION=${1:-"22.0.1"} # default java version
desired_major_version=${JAVA_VERSION%%.*}


# identify current version
# shellcheck disable=SC2154
ls -lah "${autoscaler_dir}"
current_major_version=$(find "${autoscaler_dir}/packages" -type d -name "openjdk-*" -exec bash -c '
                          directory_name="$1"
                          version_number=${directory_name##*-}
                          echo "${version_number}"
                          ' bash {} \;)
current_java_version=$(cat "${autoscaler_dir}/packages/openjdk-${current_major_version}/version")

echo "Current Java Versions"
echo "-  Major version: ${current_major_version}"
echo "-  Full version: ${current_java_version}"

if [ "${JAVA_VERSION}" == "${current_java_version}" ]; then
  echo "Already on java version ${JAVA_VERSION}. No need to update the java version"
  # exit 0
fi

# Step 1 --> Download java...
source "${script_dir}/download_java.sh" "${JAVA_VERSION}"

# Step 2 --> update bosh java package references
echo "updating bosh java packages..."
# shellcheck disable=SC2038
find "${autoscaler_dir}/packages/" -type f ! -name "*.yml" ! -name "update_java_package.sh" ! -path '*/\.*' -exec grep -l "openjdk-${current_major_version}" {} \;| xargs sed -i "s/openjdk-${current_major_version}/openjdk-${desired_major_version}/g"

mv "${autoscaler_dir}/packages/openjdk-${current_major_version}" "${autoscaler_dir}/packages/openjdk-${desired_major_version}"

echo "  -- creating spec file"
binary_name="sapmachine-jdk-${JAVA_VERSION}_linux-x64_bin.tar.gz"
cat > "${autoscaler_dir}/packages/openjdk-$desired_major_version/spec" <<EOF
---
name: openjdk-${desired_major_version}
dependencies: []
files:
- binaries/**/${binary_name}  # from https://github.com/SAP/SapMachine/releases/download/sapmachine-${JAVA_VERSION}/${binary_name}
EOF

echo "  -- creating packaging script "
cat > "${autoscaler_dir}/packages/openjdk-$desired_major_version/packaging" <<EOF
set -ex
# compile and runtime bosh env vars
export JAVA_HOME=/var/vcap/packages/openjdk-$desired_major_version
export PATH=\$JAVA_HOME/bin:\$PATH

cd \${BOSH_INSTALL_TARGET}

# extract jdk from src/binaries/jdk
tar zxvf \${BOSH_COMPILE_TARGET}/binaries/jdk/*.tar.gz --strip 1
EOF

#required for PR creating
echo -n "${JAVA_VERSION}" > "${autoscaler_dir}/version"
echo -n "${JAVA_VERSION}" > "${autoscaler_dir}/vendored-commit"
echo -n "${JAVA_VERSION}" > "${autoscaler_dir}/packages/openjdk-${desired_major_version}/version"
