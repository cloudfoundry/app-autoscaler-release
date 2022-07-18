#!/bin/bash

set -euo pipefail
mkdir combined-ops/operations \
 && cp -r ops-files/operations/* combined-ops/operations \
 && cp -r custom-ops/ci/operations/* combined-ops/operations
