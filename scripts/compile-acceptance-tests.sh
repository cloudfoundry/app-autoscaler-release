#!/usr/bin/env bash
set -euo pipefail

readonly SUITES=("api" "app" "broker" "post_upgrade" "pre_upgrade" "run_performance" "setup_performance")
readonly OPERATING_SYSTEMS=("linux" "darwin")
readonly ARCHITECTURES=("amd64" "arm64") # [amd64: intel 64 bit chips]  [arm64: apple silicon chips]

compile_suites() {
  for suite in "${SUITES[@]}"; do
      for os in "${OPERATING_SYSTEMS[@]}"; do
        for arch in "${ARCHITECTURES[@]}"; do
          GOOS="${os}" GOARCH="${arch}" ginkgo build "src/acceptance/${suite}"

          # adjust binary name to include os and architecture information
          mv "src/acceptance/${suite}/${suite}.test" "src/acceptance/${suite}/${suite}_${os}_${arch}.test"
        done
      done
  done
}

compile_ginkgo() {
  pushd ./src/acceptance > /dev/null
    for os in "${OPERATING_SYSTEMS[@]}"; do
      for arch in "${ARCHITECTURES[@]}"; do
        binary_name="ginkgo_v2_${os}_${arch}"
        GOOS="${os}" GOARCH="${arch}" go build -o "ginkgo_v2_${os}_${arch}" github.com/onsi/ginkgo/v2/ginkgo
        chmod +x "${binary_name}"
      done
    done
  popd > /dev/null
}

main() {
  compile_suites
  compile_ginkgo
}

main
