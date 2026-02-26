<h1 align="center">Pacto</h1>

<p align="center">Spec-first planning and verification for AI-assisted engineering.</p>

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

## Why Pacto?

AI assistants move fast, but plans drift unless requirements, progress, and evidence stay structured.
Pacto gives teams a small, durable workflow:

- Create plan slices with a repeatable folder structure.
- Track state across `current`, `to-implement`, `done`, and `outdated`.
- Verify implementation claims from plan documents against repository evidence.
- Generate table or JSON reports for humans and automation.

## See It In Action

```text
You: pacto new to-implement improve-auth-flow --root ./plans
CLI: Created plan: to-implement/improve-auth-flow
     - ./plans/to-implement/improve-auth-flow/README.md
     - ./plans/to-implement/improve-auth-flow/PLAN_IMPROVE_AUTH_FLOW_2026-02-26.md
     Updated index: ./plans/README.md

You: pacto status --root ./plans --format table
CLI: State summary + per-plan progress/blockers + verification outcomes

You: pacto status --root ./plans --format json --fail-on partial
CLI: JSON report suitable for CI gates
```

## Install

Requires Go.

```bash
go install ./cmd/pacto
```

Or build binaries directly:

```bash
go build -o pacto ./cmd/pacto
go build -o pacto-engine ./cmd/pacto-engine
```

## Quick Start

```bash
pacto help
pacto version

# Auto-detects ./plans if present
pacto status

# Explicit root
pacto status --root ./plans --format table

# Create a plan scaffold
pacto new to-implement my-plan-slug --root ./plans
```

## Commands

- `pacto status`
  - Discovers plans and computes status/progress.
  - Extracts blockers/next actions.
  - Verifies claims (`paths`, `symbols`, `endpoints`, `test_refs`).
  - Supports `--mode`, `--format`, `--fail-on`, `--state`, `--include-archive`.
- `pacto new`
  - Creates plan folder scaffold (`README.md` + `PLAN_*.md`).
  - Updates root index metadata.
  - Supports canonical and minimal roots (`--allow-minimal-root`).
- `pacto exec`
  - Planned command, not implemented yet.

## Workspace Layout

```text
.
├── cmd/                 # CLI entrypoints: pacto, pacto-engine
├── internal/            # parser, verify, analyze, report, discovery, config
├── plans/               # canonical planning workspace
│   ├── current/
│   ├── to-implement/
│   ├── done/
│   ├── outdated/
│   ├── PACTO.md
│   ├── PLANTILLA_PACTO_PLAN.md
│   └── README.md
└── samples/mock-pacto-repo/
```

## Tiny Smoke Test

```bash
make tiny-smoke
```

This target downloads the latest GitHub release artifact, verifies checksums, and runs a minimal end-to-end flow (`status -> new -> status`) in `/tmp/pacto-tiny-smoke/mock`.

## Release Flow

```bash
git tag v0.1.0
git push origin v0.1.0
```

CI runs tests on pushes/PRs, and tags matching `v*` publish release artifacts via GoReleaser.

## Notes

- CLI output is English-only (`--lang` is deprecated and ignored).
- Plan content can still be authored in any language.
- JSON output is the stable interface for automation.
