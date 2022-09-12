.github/docs/act.md# Using ACT

__Deploy PR + Run Acceptance tests__

See `.github/test/acceptance_test.example.secrets` for `${ACCEPTANCE_TEST_SECRET_FILE}` example.

```console
    # Make sure you have your credentials in a shell-variable. In this example we use: "GITHUB_TOKEN".
    act --workflows ./.github/workflows/acceptance_tests.yaml \
        --job deploy_autoscaler \
        --eventpath .github/test/event.json \
        --secret-file "${ACCEPTANCE_TEST_SECRET_FILE}" \
        --secret GITHUB_TOKEN="${GITHUB_TOKEN}"
```