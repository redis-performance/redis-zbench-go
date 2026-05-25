# Agent guidelines

Instructions for AI coding agents (Claude Code, Copilot, Cursor, etc.) working in this repo.

## Project overview

`redis-zbench-go` is a Go benchmark tool for Redis **Sorted Sets**. It drives `ZADD` (load mode) and `ZRANGEBYLEX` / `ZRANGEBYSCORE` (query mode) workloads against a standalone or OSS-cluster Redis deployment, with configurable concurrency, pipelining, multi-exec transactions, keyspace size, element size, and RPS limits. The tool emits throughput and HDR-histogram latency summaries. It is published as a single self-contained binary for Linux and macOS (amd64 and arm64).

## Local setup

```bash
git clone git@github.com:redis-performance/redis-zbench-go.git
cd redis-zbench-go
# Download all Go module dependencies
GO111MODULE=on go get -t -v ./...
# Build the binary
make build
# The compiled binary lands in the current directory as ./redis-zbench-go
```

Requires **Go 1.16 or later**. A running Redis instance is needed for integration tests (default: `127.0.0.1:6379`).

## Branch naming

Same as human contributors: `<type>/<short-description>` (e.g. `fix/off-by-one-in-pipeline`).

## Coding standards

- Match the style already in the file you are editing.
- Prefer clear, minimal changes over large refactors unless explicitly asked.
- Do not add comments that describe *what* the code does — only add comments when the *why* is non-obvious.
- Do not introduce new dependencies without checking with the maintainer.

## Running tests

```bash
make test
```

This formats all Go source files with `gofmt`, then runs `go test -race -covermode=atomic ./...`. A Redis instance must be reachable on `127.0.0.1:6379`.

To also generate a coverage report:

```bash
make coverage
```

Always run tests before declaring a task complete.

## How to submit changes

1. Create a branch: `git checkout -b <type>/<description>`.
2. Commit with a clear message focused on *why*, not *what*.
3. Open a pull request against `main`.
4. Do **not** push directly to `main`.

## What to avoid

- Do not reformat files unrelated to your change.
- Do not remove error handling or tests.
- Do not commit secrets, credentials, or large binary files.
- Do not amend published commits.