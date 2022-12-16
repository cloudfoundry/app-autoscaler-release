ACCEPTANCE_TEST_SECRET_FILE="~/.bbl_ssh_key"
GITHUB_TOKEN="$(cat ~/.github_token)"

act --workflows ./.github/workflows/acceptance_tests_broker.yaml \
   --job acceptance_tests_reusable\
   --eventpath .github/test/event.json \
   --secret-file "${ACCEPTANCE_TEST_SECRET_FILE}" \
   --secret GITHUB_TOKEN="${GITHUB_TOKEN}"
