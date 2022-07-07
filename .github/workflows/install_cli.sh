#!/bin/bash
set -euo pipefail

export BOSH_VERSION=7.0.1
export BBL_VERSION=8.4.89
export CREDHUB_VERSION=2.9.3
export YQ_VERSION=4.25.2
export CF_VERSION=7.5.0

echo "# Installing all deployment cli requirements"
bin_folder="${HOME}/bin"
mkdir -p "${bin_folder}"
echo "${bin_folder}" >> "${GITHUB_PATH}"


if [ ! -e "${bin_folder}/bosh" ]; then
 echo " - installing bosh cli"
 curl -sLf "https://github.com/cloudfoundry/bosh-cli/releases/download/v${BOSH_VERSION}/bosh-cli-${BOSH_VERSION}-linux-amd64" -o "${bin_folder}/bosh"
 chmod +x "${bin_folder}/bosh"
 "${bin_folder}/bosh" --version
fi


if [ ! -e "${bin_folder}/bbl" ]; then
 echo " - installing bbl cli"
 curl -sLf "https://github.com/cloudfoundry/bosh-bootloader/releases/download/v${BBL_VERSION}/bbl-v${BBL_VERSION}_linux_x86-64" -o "${bin_folder}/bbl"
 chmod +x "${bin_folder}/bbl"
 "${bin_folder}/bbl" version
fi


if [ ! -e "${bin_folder}/credhub" ]; then
 echo " - installing credhub cli"
 curl -sLf "https://github.com/cloudfoundry/credhub-cli/releases/download/${CREDHUB_VERSION}/credhub-linux-${CREDHUB_VERSION}.tgz" | tar -zx -C "${bin_folder}"
 "${bin_folder}/credhub" --version
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