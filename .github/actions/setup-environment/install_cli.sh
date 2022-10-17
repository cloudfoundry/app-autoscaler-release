#!/bin/bash
set -euo pipefail

export BOSH_VERSION=7.0.1
export BBL_VERSION=8.4.89
export CREDHUB_VERSION=2.9.3
export YQ_VERSION=4.26.1
export CF_PACKAGE_VERSION=1.38.0
export CF_VERSION=8.3.0

echo "# Installing all deployment cli requirements"
bin_folder="${HOME}/bin"
mkdir -p "${bin_folder}/unchecked"
echo "${bin_folder}" >> "${GITHUB_PATH}"


if [ ! -e "${bin_folder}/bosh" ]; then
 echo " - installing bosh cli"
 sha256=$( curl -sf --compressed -H "Accept: application/vnd.github+json"  "https://api.github.com/repos/cloudfoundry/bosh-cli/releases/tags/v${BOSH_VERSION}" | jq -r '.body' | grep -E '.*-linux-amd64' | awk '{ print $1}')
 curl -sLf --compressed "https://github.com/cloudfoundry/bosh-cli/releases/download/v${BOSH_VERSION}/bosh-cli-${BOSH_VERSION}-linux-amd64" -o "${bin_folder}/unchecked/bosh"
 echo "${sha256} ${bin_folder}/unchecked/bosh" | sha256sum -c
 chmod +x "${bin_folder}/unchecked/bosh"
 "${bin_folder}/unchecked/bosh" --version
 mv "${bin_folder}/unchecked/bosh" "${bin_folder}/bosh"
fi


if [ ! -e "${bin_folder}/bbl" ]; then
 echo " - installing bbl cli"
 # shellcheck disable=SC2016
 sha256=$( curl -sf --compressed  -H "Accept: application/vnd.github+json"  "https://api.github.com/repos/cloudfoundry/bosh-bootloader/releases/tags/v${BBL_VERSION}" | jq -r '.body' | grep "Linux sha256" | sed 's/.*`\(.*\)`\*/\1/')
 curl -sLf --compressed "https://github.com/cloudfoundry/bosh-bootloader/releases/download/v${BBL_VERSION}/bbl-v${BBL_VERSION}_linux_x86-64" -o "${bin_folder}/unchecked/bbl"
 echo "${sha256} ${bin_folder}/unchecked/bbl" | sha256sum -c
 chmod +x "${bin_folder}/unchecked/bbl"
 "${bin_folder}/unchecked/bbl" version
 mv "${bin_folder}/unchecked/bbl" "${bin_folder}/bbl"
fi


if [ ! -e "${bin_folder}/credhub" ]; then
 echo " - installing credhub cli"
 # shellcheck disable=SC2016
 sha256=$( curl -sf --compressed -H "Accept: application/vnd.github+json" "https://api.github.com/repos/cloudfoundry/credhub-cli/releases/tags/${CREDHUB_VERSION}" | jq -r '.body' | grep "Linux sha256" | sed 's/.*`\(.*\)`\*/\1/' | tr -d '\r')
 curl -sLf --compressed "https://github.com/cloudfoundry/credhub-cli/releases/download/${CREDHUB_VERSION}/credhub-linux-${CREDHUB_VERSION}.tgz" -o "${bin_folder}/unchecked/credhub.tar.gz"
 echo "${sha256} ${bin_folder}/unchecked/credhub.tar.gz" | sha256sum -c
 tar -zxf "${bin_folder}/unchecked/credhub.tar.gz" -C "${bin_folder}/unchecked"
 chmod +x "${bin_folder}/unchecked/credhub"
 "${bin_folder}/unchecked/credhub" --version
 mv "${bin_folder}/unchecked/credhub" "${bin_folder}/credhub"
fi


if [ ! -e "${bin_folder}/yq" ]; then
 echo " - installing yq"
 sha256=$(curl -sLf --compressed "https://github.com/mikefarah/yq/releases/download/v${YQ_VERSION}/checksums" | grep -E '^yq_linux_amd64.tar.gz' | awk '{ print $19 }')
 curl -sLf --compressed "https://github.com/mikefarah/yq/releases/download/v${YQ_VERSION}/yq_linux_amd64.tar.gz" -o "${bin_folder}/unchecked/yq.tar.gz"
 echo "${sha256} ${bin_folder}/unchecked/yq.tar.gz" | sha256sum -c
 tar -zxf "${bin_folder}/unchecked/yq.tar.gz" -C "${bin_folder}/unchecked"
 chmod a+x "${bin_folder}/unchecked/yq_linux_amd64"
 "${bin_folder}/unchecked/yq_linux_amd64" --version
 mv "${bin_folder}/unchecked/yq_linux_amd64" "${bin_folder}/yq"
fi


if [ ! -e "${bin_folder}/cf" ]; then
 echo " - installing cf cli"
 sha1=$(curl -sLf --compressed "https://raw.githubusercontent.com/bosh-packages/cf-cli-release/v${CF_PACKAGE_VERSION}/config/blobs.yml"  | "${bin_folder}/yq" ".\"cf8-cli_${CF_VERSION}_linux_x86-64.tgz\".sha" )
 curl -sLf --compressed "https://packages.cloudfoundry.org/stable?release=linux64-binary&version=${CF_VERSION}&source=github-rel" -o "${bin_folder}/unchecked/cf.tar.gz"
 echo "${sha1} ${bin_folder}/unchecked/cf.tar.gz" | sha1sum -c
 tar -zxf "${bin_folder}/unchecked/cf.tar.gz" -C "${bin_folder}/unchecked"
 chmod a+x "${bin_folder}/unchecked/cf8"
 "${bin_folder}/unchecked/cf8" --version
 mv "${bin_folder}/unchecked/cf8" "${bin_folder}/cf"
 "${bin_folder}/cf" install-plugin -f -r CF-Community app-autoscaler-plugin && "${bin_folder}/cf" plugins
fi

rm -rf "${bin_folder}/unchecked" || true

echo " - installing cf-uaac"
which uaac || gem install cf-uaac