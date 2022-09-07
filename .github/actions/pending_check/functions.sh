#!/bin/bash

function check_create {

echo '::group::Creating new check'
curl -vf -POST \
   --retry 5 \
   -H "Accept: application/vnd.github+json" \
   -H "Authorization: token ${{ env.GITHUB_TOKEN }}" \
   "https://api.github.com/repos/${{ env.GITHUB_REPOSITORY }}/check-runs" \
-d '
    {
    "name":        "${{env.check_name}}",
    "head_sha":    "${{env.PR_SHA}}",
    "status":      "in_progress",
    "external_id": "${{env.GH_RUN_ID}}",
    "output": {
                "title":   "${{ env.WORKFLOW_NAME }} running",
                "summary": "pending check for commit ${{env.PR_SHA}}",
                "text":    "Awaiting check result..."
    }
    }
' \
-o new_check.json

id=$(jq -r '.id' new_check.json)

if [ -z "${id}" ]; then
  echo "ERROR: Failed to create the required check job"
  echo "Result of curl:"
  cat new_check.json
  exit 1
fi
echo "::endgroup::"

echo "Id is: ${id}"

}