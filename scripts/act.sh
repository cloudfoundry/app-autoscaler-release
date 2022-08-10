
#!/bin/bash

set -euo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
workflow_dir="${script_dir}/../.github/workflows"

pushd "$workflow_dir" > /dev/null
workflows=($(ls |grep -E "yml|yaml" | cut -d'.' -f1))
popd

select workflow in "${workflows[@]}"
do
  workflow_file="$workflow_dir/$workflow.yaml"
  jobs=($( yq '.jobs | keys | .[]' "$workflow_file"))
  select job in "${jobs[@]}"
  do
    echo "Running $workflow - job: $job"
    act -W "$workflow_file" -j "$job"
    break
  done
done

