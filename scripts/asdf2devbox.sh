#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

while IFS=' ' read -r program version
do
  if [[ "$program" == "bosh" ]]; then
    program="bosh-cli"
  fi
  if [[ "$program" == "cf" ]]; then
    program="cloudfoundry-cli"
    version="latest" # nixos doesn't seem to package java versions too quickly, so we'll just use the latest for now
  fi
  if [[ "$program" == "concourse" ]]; then
    program="fly"
  fi
  if [[ "$program" == "gcloud" ]]; then
    program="google-cloud-sdk"
  fi
  if [[ "$program" == "golang" ]]; then
    program="go"
  fi
  if [[ "$program" == "java" ]]; then
    program="temurin-bin-17"
    #version="${version#temurin-}"
    version="latest" # nixos doesn't seem to package java versions too quickly, so we'll just use the latest for now
  fi
  if [[ "$program" == "make" ]]; then
    program="gnumake"
  fi
  if [[ "$program" == "shellcheck" ]]; then
    version="latest" # nixos doesn't seem to package shellcheck versions too quickly, so we'll just use the latest for now
  fi
  if [[ "$program" == "yq" ]]; then
    program="yq-go"
  fi
  devbox add "$program"@"$version"
done < "$SCRIPT_DIR/../.tool-versions"

