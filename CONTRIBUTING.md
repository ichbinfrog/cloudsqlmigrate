# Contributing guidelines

Feel free to open PRs to improve the code base, tests or add further preflight and postflight checks.

Currently E2E tests are run manually based on the infrastructure bootstrapped in `e2e/terraform`. It's highly recommended that you run both unit tests and integration tests on your PRs


**Unit test (also on the CI)**

```console
go test -v ./...
```

**Integration tests (WIP)**

```console
# Initial setup
cd e2e/terraform
terraform apply

# Running tests
go run main.go \
    --src-project=$SRC_PROJECT \
    --src-instance=$SRC_INSTANCE \
    --dst-project=$DST_PROJECT \
    --dst-instance=$DST_INSTANCE
```
