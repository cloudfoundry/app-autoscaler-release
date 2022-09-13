#!/bin/bash

# Checking run under ACT with override option
eval "${ACT_RUN}"
set -euo pipefail

checkruns_url="https://api.github.com/repos/${GITHUB_REPOSITORY}/check-runs"
checkruns_commit_url="https://api.github.com/repos/${GITHUB_REPOSITORY}/commits/${PR_SHA}/check-runs"
curlopts=(-f --retry 5 -H "Accept: application/vnd.github+json" -H "Authorization: token ${GITHUB_TOKEN}" )


function main {
    case ${PENDING_CHECK} in
        "create")
          check_create;;
        "verify")
          check_verify;;
        *)
        echo "ERROR: invalid action ${PENDING_CHECK} for pending check. Allowed: create or verify"
        exit 1
    esac
}

#--------------------------------------------------------------------------------------------------
function check_create {
echo '::group::Creating new check'
curl -v "${curlopts[@]}" -POST "${checkruns_url}" -o new_check.json \
-d @- << END;
    {
    "name":        "${CHECK_NAME}",
    "head_sha":    "${PR_SHA}",
    "status":      "in_progress",
    "external_id": "${GH_RUN_ID}",
    "output": {
                "title":   "${WORKFLOW_NAME} check running",
                "summary": "pending check for commit ${PR_SHA}",
                "text":    "Awaiting result..."
            }
    }
END
id=$(jq -r '.id' new_check.json)

if [ -z "${id}" ]; then
  echo "ERROR: Failed to create check run"
  echo "Result of curl:"
  cat new_check.json
  exit 1
fi
echo "::endgroup::"
echo "Id is: ${id}"
}


#--------------------------------------------------------------------------------------------------
function send_conclusion() {

echo "Verifying: ${checkruns_url}/${id}"
curl -s "${curlopts[@]}" -X PATCH "${checkruns_url}/${id}" \
-d @- << END;
    { "name": "${CHECK_NAME}", "conclusion": "$1" }
END
}


#--------------------------------------------------------------------------------------------------
function check_verify {

echo "::group::Getting checkruns for commit ${PR_SHA}"
curl -s "${curlopts[@]}" "${checkruns_commit_url}" -o checkruns.json

echo "Looking for the last result"
jq '[.check_runs[] | select(.name=="'"${CHECK_NAME}"'")]' checkruns.json > results.json
jq '.|last' results.json > latest_result.json

id=$( jq '.id' latest_result.json )
number_of_checks=$(jq '. | length' results.json)

echo "== Latest ${CHECK_NAME} check result =="
echo
  cat latest_result.json
echo "::endgroup::"

echo "::group::Check Info"
echo "Latest check id:${id}"
echo "Number of checks for commit ${PR_SHA} ${number_of_checks}"
echo "::endgroup::"

if [ "${number_of_checks}" -eq 0 ]; then
  echo "ERROR: no checks were found this commit!"
  exit 1
fi

echo "::group::Retrieving status of jobs (checks_filter: ${CHECK_FILTER})"
jq '.check_runs[] | select(.conclusion == "failure") | select(.name? | match("'"${CHECK_FILTER}"'")) | " - \(.name): \(.html_url)"' checkruns.json \
  > bad_jobs.txt
echo "::endgroup::"
if [ ! -s bad_jobs.txt ]; then
  echo "OK: all jobs passed!"

  echo "::group::Sending success conclusion to the workflow check"
    send_conclusion "success"
  echo "::endgroup::"

else
  echo "=========================="
  echo "List of failed checks:"
    cat bad_jobs.txt
  echo "=========================="

  echo "::group::Sending failure conclusion to the workflow check"
    send_conclusion "failure"
  echo "::endgroup::"
  exit 1
fi
}
#----------------------------------------------------------------------------------------------------------------------


# ++ start ++
main