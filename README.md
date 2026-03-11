<h1 align="center">Pacto</h1>

<p align="center">Spec-driven development (SDD) planning and verification for AI-assisted engineering.</p>

<p align="center">
  <a href="https://github.com/triple0-labs/pacto-spec/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/triple0-labs/pacto-spec/actions/workflows/ci.yml/badge.svg" /></a>
  <a href="https://github.com/triple0-labs/pacto-spec/releases"><img alt="Releases" src="https://img.shields.io/github/v/release/triple0-labs/pacto-spec?style=flat-square" /></a>
  <a href="./LICENSE"><img alt="License" src="https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square" /></a>
</p>

Our philosophy:

```text
-> specs before code
-> evidence over assumptions
-> lightweight over ceremony
-> practical for brownfield and greenfield
```

## Why Pacto

Pacto keeps AI-assisted work anchored in executable specs:

- Define plan slices before implementation.
- Track progress through explicit states (`to-implement`, `current`, `done`, `outdated`).
- Verify plan claims against repository evidence.
- Render interactive status in TTY and emit table/JSON in non-TTY for automation.

## Core Workflow

```text
pacto init  ->  pacto status  ->  pacto new  ->  pacto exec  ->  pacto move  ->  pacto status
```

- `pacto init`: bootstrap `.pacto/plans` workspace.
- `pacto status`: inspect current plan/evidence state before acting.
- `pacto new`: create a plan slice from template and update the index.
- `pacto exec`: update execution progress/evidence in plan docs.
- `pacto move`: perform explicit state transitions (`to-implement -> current -> done`).

Primary source of truth is `<plans-root>/PACTO.md` and plan artifacts.
`AGENTS.md` (when generated via `pacto init --with-agents`) is only a hand-off layer for compatible assistants.

Optional ideation flow:

```text
pacto explore
```

## Install

### Option 1: curl (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/triple0-labs/pacto-spec/main/install.sh | bash
```

### Option 2: Go

```bash
go install ./cmd/pacto
```

## Quick Start

```bash
pacto help
pacto version

# initialize project workspace
pacto init

# non-interactive setup with explicit profile
pacto init --no-interactive --tools codex,cursor --yes

# create a plan
pacto new to-implement improve-auth-flow

# verify status and evidence
pacto status

# CI-friendly output
pacto status --format json --fail-on partial
pacto doctor --format json --fail-on any

# list and validate local plugins
pacto plugin list-available
pacto plugin install git-guardrails
pacto plugin list
pacto plugin validate
```

## Drift Check

Use `pacto doctor` to detect stale generated artifacts and legacy integration patterns:

```bash
pacto doctor --format table
pacto doctor --format json --fail-on any
```

Recommended remediation when drift is reported:

```bash
pacto update --artifacts
```

## CI Snippet

```yaml
name: pacto-artifact-drift
on: [pull_request]
jobs:
  drift:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24"
      - run: go build -o ./bin/pacto ./cmd/pacto
      - run: ./bin/pacto doctor --format json --fail-on any
```

## Docs

- [Getting Started](./docs/getting-started.md)
- [Concepts](./docs/concepts.md)
- [Commands](./docs/commands.md)
- [Integrations](./docs/integrations.md)
- [Plugins](./docs/plugins.md)
- [Contributing](./docs/contributing.md)
- [Releasing](./RELEASING.md)

## Notes

- CLI supports `--lang en|es`; `pacto init` persists workspace language in `.pacto/config.yaml`.
- `pacto exec` updates execution artifacts in plan docs (no source-code edits).
- `.pacto/plans/` files are workspace artifacts/templates; canonical product docs are in `docs/`.
