# Using ACT


kk

    act -W ./.github/workflows/acceptance_tests.yaml -j acceptance_test --secret-file .github/test/acceptance_test.example.secrets -s GITHUB_TOKEN=YOUR_GITHUB_TOKEN
k

Deploy PR + Run Acceptance tests

    act -W ./.github/workflows/acceptance_tests.yaml\
      -j deploy_autoscaler  \
      --eventpath .github/test/event.json \
      --secret-file ~/SAPDevelop/_secrets/acceptance_test.secrets \
      -s GITHUB_TOKEN=$(cat ~/SAPDevelop/_secrets/github_token-marcin)
