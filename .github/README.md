# Github Actions


## Testing workflow

example:

    act -W ./.github/workflows/acceptance_tests.yaml -j acceptance_test --secret-file .github/test/acceptance_test.example.secrets -s GITHUB_TOKEN=YOUR_GITHUB_TOKEN
