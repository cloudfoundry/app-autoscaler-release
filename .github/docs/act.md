.github/docs/act.md# Using ACT

__Deploy PR + Run Acceptance tests__

see `.github/test/acceptance_test.example.secrets` for `$ACCEPTANCE_TEST_SECRET_FILE` example

```
    act -W ./.github/workflows/acceptance_tests.yaml\
        -j deploy_autoscaler  \
        --eventpath .github/test/event.json \
        --secret-file "$ACCEPTANCE_TEST_SECRET_FILE"  \
        -s GITHUB_TOKEN="$(cat $PERSONAL_GITHUB_TOKEN_FILE)"
```
