#!/bin/bash

set -euo pipefail

TARGET=autoscaler

PIPELINE_NAME=app-autoscaler-release

fly -t $TARGET set-pipeline --config=pipeline.yml --pipeline=$PIPELINE_NAME
