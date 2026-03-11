---
name: pacto-local-dev
description: Local development workflow for building, installing, and validating the pacto binary from this repository.
---

# Pacto Local Dev Skill

Use this skill when developing Pacto itself and validating local changes with Codex.

## Objective

Build, install, and test the local `pacto` binary safely during development.

## When To Use

- You changed Pacto code and want to test behavior in a real project.
- You need to verify the active `pacto` binary is the one built from this repo.
- You want a repeatable local dev loop before release.

## Input Contract

### Required Inputs
- Repo root path for this project.

### Optional Inputs
- Target install path (default: `~/.local/bin/pacto`).
- Version suffix/tag for local builds.
- Smoke project path to run manual checks.

## Execution Contract

- Tool target: codex
- Recommended commands:
  - `go test ./...`
  - `go build -ldflags "-X pacto/internal/app.Version=dev-local.<YYYYMMDDHHMM>.<sha>" -o ~/.local/bin/pacto ./cmd/pacto`
  - `which pacto && pacto --version`

## Local Dev Loop

1. Run test suite:
   - `go test ./...`
2. Build local binary with explicit local version stamp.
3. Install to `~/.local/bin/pacto`.
4. Verify active binary path and version.
5. Run a quick smoke flow in a sandbox repo:
   - `pacto init`
   - `pacto status`
   - `pacto new to-implement <slug>`
   - `pacto status`

## Output Contract

- Local binary is installed and executable.
- `pacto --version` includes local version stamp.
- Basic workflow commands run successfully in a test project.

## Validation Checklist

- `which pacto` points to expected install path.
- `pacto --version` matches local build tag.
- `go test ./...` passes.
- `pacto init/status/new` smoke checks complete.

## Failure Modes and Handling

- Permission denied writing install path:
  - rebuild/install with appropriate permissions.
- Wrong binary on PATH:
  - inspect `which pacto` and PATH order.
- Version not updated:
  - rebuild with `-ldflags` and re-run verification.
- Regression in behavior:
  - capture failing command and reproduce in isolated temp repo.
