# Contributing

We treat this repo as "Open Source" within Redis: anyone who clears the bar below is welcome to contribute.

## Local setup

```bash
git clone git@github.com:redis-performance/redis-zbench-go.git
cd redis-zbench-go
# Download all Go module dependencies
GO111MODULE=on go get -t -v ./...
# Build the binary
make build
```

Requires **Go 1.16 or later**. No other system dependencies are needed beyond a running Redis instance for integration tests.

## Branch naming

```
<type>/<short-description>
```

Types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`

Example: `feat/add-pipeline-mode`

## Coding standards

- Keep changes focused; one logical change per PR.
- Follow the conventions already present in the codebase (formatting, naming, error handling).
- No dead code, no commented-out blocks.

## Submitting changes

1. Fork or create a branch from `main`.
2. Make your changes with clear, atomic commits.
3. Open a pull request against `main` with a descriptive title and summary.
4. Address review comments promptly; force-push to the same branch to update.

## Testing

- All new behaviour must be covered by tests.
- Existing tests must pass: run the test suite locally before opening a PR.
- Coverage should not decrease.

Run the full test suite (requires a Redis instance on `127.0.0.1:6379`):

```bash
make test
```

This runs `gofmt` for formatting checks and `go test -race -covermode=atomic ./...` under the hood.

To also produce a coverage report:

```bash
make coverage
```

## Review process

- At least one maintainer approval is required before merge.
- CI must be green.
- Maintainers may request changes or close PRs that do not meet the bar — this is normal and not personal.