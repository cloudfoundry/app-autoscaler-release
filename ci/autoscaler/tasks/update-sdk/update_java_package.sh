#! /usr/bin/env bash

# This script is used to update the java version used in the bosh packaging
# It downloads the given SAP Machine java distribution and updates the bosh packaging with the new version.

[ -n "${DEBUG}" ] && set -x
set -euox pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source "${script_dir}/vars.source.sh"
ls -lah
java_dir=${JAVA_DIR:-"${autoscaler_dir}/../SapMachine"}
java_dir=$(realpath --canonicalize-existing "${java_dir}")

pushd "${java_dir}"
	ls -lah
	cat version
	cat tag
	cat url

  #java_major_version=$(grep "DEFAULT_JDK_SOURCE_TARGET_VERSION" make/conf/version-numbers.conf | cut -d= -f2)
  #java_version_interim=$(grep "DEFAULT_VERSION_INTERIM" make/conf/version-numbers.conf | cut -d= -f2)
  #java_version_update=$(grep "DEFAULT_VERSION_UPDATE" make/conf/version-numbers.conf | cut -d= -f2)
  #java_version_prerelease=$(grep "DEFAULT_PROMOTED_VERSION_PRE" make/conf/version-numbers.conf | cut -d= -f2)
  #java_full_version=$(grep '^version=' .jcheck/conf | cut -d'=' -f2 | tr -d '[:space:]' | sed 's/.$//')
	java_full_version=$(cat version | sed 's/^sapmachine-//' | sed 's/+//')
popd

sapmachine_java_version="${java_full_version}"
echo "Desired Java Version: ${java_full_version}"

# consider only lts releases
if [ "${java_full_version}" != "${sapmachine_java_version}" ]; then
     echo "java version ${sapmachine_java_version} is not a LTS release. skipping update"
     exit 0
fi

JAVA_VERSION=${sapmachine_java_version:-"21.0.3"}
desired_major_version=${JAVA_VERSION%%.*}

# identify current version
# shellcheck disable=SC2154
current_major_version=$(find "${autoscaler_dir}/packages" -type d -name "openjdk-*" -exec bash -c '
                          directory_name="$1"
                          version_number=${directory_name##*-}
                          echo "${version_number}"
                          ' bash {} \;)
# shellcheck disable=SC2154
current_java_version=$(cat "${autoscaler_dir}/packages/openjdk-${current_major_version}/version")
echo "Checking current Java Versions"
echo " - Full version: ${current_java_version}"

if [ "${JAVA_VERSION}" == "${current_java_version}" ]; then
  echo "Already on java version ${JAVA_VERSION}. No need to update the java version"
  exit 0
fi

#Step 1 --> Download java...
binary_name="sapmachine-jdk-${JAVA_VERSION}_linux-x64_bin.tar.gz"
mkdir -p "${autoscaler_dir}/src/binaries/jdk"
mv "${java_dir}/${binary_name}" "${autoscaler_dir}/src/binaries/jdk/${binary_name}"
#source "${script_dir}/download_java.sh" "${JAVA_VERSION}"
#printf "\n"

# Step 2 --> upload blob to blobstore
pushd "${autoscaler_dir}" > /dev/null
  echo "- Adding and uploading blobs to release blobstore "
  bosh add-blob "src/binaries/jdk/${binary_name}" "${binary_name}"
  create_bosh_config
  bosh upload-blobs
popd > /dev/null

printf "\n"

# Step 3 --> update bosh java package references
echo "- updating bosh java packages..."
# shellcheck disable=SC2038
# only change java references if major versions are different
if [ "${current_major_version}" != "${desired_major_version}" ]; then
  find "${autoscaler_dir}" -type f ! -name "*.yml" ! -name "update_java_package.sh" ! -path '*/\.*' -exec grep -l "openjdk-${current_major_version}" {} \;| xargs sed -i "s/openjdk-${current_major_version}/openjdk-${desired_major_version}/g"
  mv "${autoscaler_dir}/packages/openjdk-${current_major_version}" "${autoscaler_dir}/packages/openjdk-${desired_major_version}"
  exit 0
fi

echo " - creating spec file"
cat > "${autoscaler_dir}/packages/openjdk-$desired_major_version/spec" <<EOF
---
name: openjdk-${desired_major_version}
dependencies: []
files:
- ${binary_name}  # from https://github.com/SAP/SapMachine/releases/download/sapmachine-${JAVA_VERSION}/${binary_name}
EOF

echo " - creating packaging script "
cat > "${autoscaler_dir}/packages/openjdk-$desired_major_version/packaging" <<EOF
set -ex
# compile and runtime bosh env vars
export JAVA_HOME=/var/vcap/packages/openjdk-$desired_major_version
export PATH=\$JAVA_HOME/bin:\$PATH

cd \${BOSH_INSTALL_TARGET}

tar zxvf \${BOSH_COMPILE_TARGET}/*.tar.gz --strip 1
EOF

#required for PR creation
echo -n "${JAVA_VERSION}" > "${autoscaler_dir}/version"
echo -n "${JAVA_VERSION}" > "${autoscaler_dir}/vendored-commit"
echo -n "${JAVA_VERSION}" > "${autoscaler_dir}/packages/openjdk-${desired_major_version}/version"
