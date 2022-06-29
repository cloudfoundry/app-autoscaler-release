# Using ACT

__Deploy PR + Run Acceptance tests__

    act -W ./.github/workflows/acceptance_tests.yaml\
      -j deploy_autoscaler  \
      --eventpath .github/test/event.json \
      --secret-file ~/SAPDevelop/_secrets/acceptance_test.secrets \
      -s GITHUB_TOKEN=$(cat ~/SAPDevelop/_secrets/github_token-marcin)
