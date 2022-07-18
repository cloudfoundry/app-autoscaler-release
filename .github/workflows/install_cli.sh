#!/bin/bash
set -euxo pipefail

export BOSH_VERSION=7.0.1
export BBL_VERSION=8.4.89
export CREDHUB_VERSION=2.9.3
export YQ_VERSION=4.25.2
export CF_VERSION=7.5.0

echo "# Installing all deployment cli requirements"
bin_folder="${HOME}/bin"
mkdir -p "${bin_folder}/unchecked"
echo "${bin_folder}" >> "${GITHUB_PATH}"


if [ ! -e "${bin_folder}/bosh" ]; then
 echo " - installing bosh cli"
 sha256=$( curl -sf -H "Accept: application/vnd.github+json"  "https://api.github.com/repos/cloudfoundry/bosh-cli/releases/tags/v${BOSH_VERSION}" | jq -r '.body' | grep -E '.*-linux-amd64' | awk '{ print $1}')
 curl -sLf "https://github.com/cloudfoundry/bosh-cli/releases/download/v${BOSH_VERSION}/bosh-cli-${BOSH_VERSION}-linux-amd64" -o "${bin_folder}/unchecked/bosh"
 echo "${sha256} ${bin_folder}/unchecked/bosh" | sha256sum -c
 chmod +x "${bin_folder}/unchecked/bosh"
 "${bin_folder}/unchecked/bosh" --version
 mv "${bin_folder}/unchecked/bosh" "${bin_folder}/bosh"
fi


if [ ! -e "${bin_folder}/bbl" ]; then
 echo " - installing bbl cli"
 sha256=$( curl -sf -H "Accept: application/vnd.github+json"  "https://api.github.com/repos/cloudfoundry/bosh-bootloader/releases/tags/v8.4.89" | jq -r '.body' | grep "Linux sha256" | sed 's/.*`\(.*\)`\*/\1/')
 curl -sLf "https://github.com/cloudfoundry/bosh-bootloader/releases/download/v${BBL_VERSION}/bbl-v${BBL_VERSION}_linux_x86-64" -o "${bin_folder}/unchecked/bbl"
 echo "${sha256} ${bin_folder}/unchecked/bbl" | sha256sum -c
 chmod +x "${bin_folder}/unchecked/bbl"
 "${bin_folder}/unchecked/bbl" version
 mv "${bin_folder}/unchecked/bbl" "${bin_folder}/bbl"
fi


if [ ! -e "${bin_folder}/credhub" ]; then
 echo " - installing credhub cli"
 sha256=$( curl -sf -H "Accept: application/vnd.github+json" "https://api.github.com/repos/cloudfoundry/credhub-cli/releases/tags/${CREDHUB_VERSION}" | jq -r '.body' | grep "Linux sha256" | sed 's/.*`\(.*\)`\*/\1/' | xargs)
 curl -sLf "https://github.com/cloudfoundry/credhub-cli/releases/download/${CREDHUB_VERSION}/credhub-linux-${CREDHUB_VERSION}.tgz" | tar -zx -C "${bin_folder}/unchecked"
 echo "${sha256} ${bin_folder}/unchecked/credhub" | sha256sum -c
 chmod +x "${bin_folder}/unchecked/credhub"
 "${bin_folder}/unchecked/credhub" --version
 mv "${bin_folder}/unchecked/credhub" "${bin_folder}/credhub"
fi


if [ ! -e "${bin_folder}/yq" ]; then
 echo " - installing yq"
 curl -sLf "https://github.com/mikefarah/yq/releases/download/v${YQ_VERSION}/yq_linux_amd64" -o "${bin_folder}/yq"
 chmod a+x "${bin_folder}/yq"
 "${bin_folder}/yq" --version
fi


if [ ! -e "${bin_folder}/cf" ]; then
 echo " - installing cf cli"
 curl -sLf "https://packages.cloudfoundry.org/stable?release=linux64-binary&version=${CF_VERSION}&source=github-rel" | tar -zx -C "${bin_folder}"
 "${bin_folder}/cf" version
fi

echo " - installing cf-uaac"
which uaac || gem install cf-uaac