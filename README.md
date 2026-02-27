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
- Emit table or JSON outputs for humans and CI automation.

## Core Workflow

```text
pacto init  ->  pacto new  ->  pacto status
```

- `pacto init`: bootstrap `.pacto/plans` workspace.
- `pacto new`: create a plan slice from template and update the index.
- `pacto status`: parse plans, verify claims, and report readiness.

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

# create a plan
pacto new to-implement improve-auth-flow

# verify status and evidence
pacto status --format table

# CI-friendly output
pacto status --format json --fail-on partial
```

## Docs

- [Getting Started](./docs/getting-started.md)
- [Concepts](./docs/concepts.md)
- [Commands](./docs/commands.md)
- [Integrations](./docs/integrations.md)
- [Contributing](./docs/contributing.md)
- [Releasing](./RELEASING.md)

## Notes

- CLI output is English-only (`--lang` is deprecated and ignored).
- `pacto exec` is planned and not implemented yet.
- `plans/` and `.pacto/plans/` files are workspace artifacts/templates; canonical product docs are in `docs/`.
