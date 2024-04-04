#!/usr/bin/env bash
set -euo pipefail

compile() {
  local suites operating_systems architectures
  suites=("api" "app" "broker" "post_upgrade" "pre_upgrade" "run_performance" "setup_performance")
  operating_systems=("linux" "darwin")
  architectures=("amd64" "arm64") # [amd64: intel 64 bit chips]  [arm64: apple silicon chips]

  for suite in "${suites[@]}"; do
      for os in "${operating_systems[@]}"; do
        for arch in "${architectures[@]}"; do
          GOOS="${os}" GOARCH="${arch}" ginkgo build "src/acceptance/${suite}"

          # adjust binary name to include os and architecture information
          mv "src/acceptance/${suite}/${suite}.test" "src/acceptance/${suite}/${suite}_${os}_${arch}.test"
        done
      done
  done
}

main() {
  compile
}

main
