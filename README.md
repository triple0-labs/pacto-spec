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
You: pacto init
CLI: Initialized Pacto workspace: ./.pacto/plans

You: pacto new to-implement improve-auth-flow
CLI: Created plan: to-implement/improve-auth-flow
     - ./.pacto/plans/to-implement/improve-auth-flow/README.md
     - ./.pacto/plans/to-implement/improve-auth-flow/PLAN_IMPROVE_AUTH_FLOW_2026-02-26.md
     Updated index: ./.pacto/plans/README.md

You: pacto status --format table
CLI: State summary + per-plan progress/blockers + verification outcomes

You: pacto status --format json --fail-on partial
CLI: JSON report suitable for CI gates
```

## Install

Pacto supports curl install, native Go install, and npm/npx wrapper install.

### Option 1: curl (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/triple0-labs/pacto-spec/main/install.sh | bash
```

Optional overrides:

```bash
# install a specific version
curl -fsSL https://raw.githubusercontent.com/triple0-labs/pacto-spec/main/install.sh | PACTO_VERSION=0.1.1 bash

# install into a custom directory
curl -fsSL https://raw.githubusercontent.com/triple0-labs/pacto-spec/main/install.sh | INSTALL_DIR=$HOME/.local/bin bash
```

### Option 2: Go Native

```bash
# from this repository root
go install ./cmd/pacto
```

Or build binaries directly:

```bash
go build -o pacto ./cmd/pacto
go build -o pacto-engine ./cmd/pacto-engine
```

### Option 3: npm / npx Wrapper

Requires Node.js 18+.

```bash
# one-off execution via npx
npx -y @triple0-labs/pacto-spec --help

# global shim
npm install -g @triple0-labs/pacto-spec
pacto --help
```

On first run, the wrapper downloads the matching `pacto` GitHub release binary for your OS/arch, verifies checksums, caches it locally, and then executes it.

## Quick Start

```bash
pacto help
pacto version

# Bootstrap local workspace (default: ./.pacto/plans)
pacto init

# Auto-detects plans root from current dir and parents
# Resolution order per directory: direct state dirs, ./.pacto/plans, ./plans
pacto status

# Explicit split roots
pacto status --plans-root ./.pacto/plans --repo-root . --format table

# Create a plan scaffold (also auto-detects plans root from current dir and parents)
pacto new to-implement my-plan-slug
```

## Commands

- `pacto status`
  - Discovers plans and computes status/progress.
  - Extracts blockers/next actions.
  - Auto-detects plans root from current directory or parent directories.
  - Verifies claims (`paths`, `symbols`, `endpoints`, `test_refs`) against `repo-root`.
  - Supports `--plans-root`, `--repo-root`, `--mode`, `--format`, `--fail-on`, `--state`, `--include-archive`.
- `pacto new`
  - Creates plan folder scaffold (`README.md` + `PLAN_*.md`).
  - Auto-detects plans root from current directory or parent directories when `--root` is omitted.
  - Updates root index metadata.
  - Supports explicit root override (`--root`) and minimal roots (`--allow-minimal-root`).
- `pacto init`
  - Bootstraps a project-local workspace at `./.pacto/plans`.
  - Creates canonical docs/templates and state folders.
  - Optional `--with-agents` adds a managed Pacto block in `AGENTS.md`.
  - Supports `--force` to overwrite init-managed files.
- `pacto exec`
  - Planned command, not implemented yet.

## Workspace Layout

```text
.
├── cmd/                 # CLI entrypoints: pacto, pacto-engine
├── internal/            # parser, verify, analyze, report, discovery, config
├── .pacto/
│   └── plans/           # default workspace created by `pacto init`
│       ├── current/
│       ├── to-implement/
│       ├── done/
│       ├── outdated/
│       ├── PACTO.md
│       ├── PLANTILLA_PACTO_PLAN.md
│       ├── SLASH_COMMANDS.md
│       └── README.md
├── plans/               # optional legacy workspace (still supported)
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

CI runs tests on pushes/PRs. Tags matching `v*` publish Go release artifacts via GoReleaser, and then publish the npm wrapper package automatically via npm Trusted Publishing (OIDC).

Detailed checklist: [RELEASING.md](./RELEASING.md)

## Notes

- CLI output is English-only (`--lang` is deprecated and ignored).
- For `status`, `--root` is deprecated; use `--plans-root` and `--repo-root`.
- `status` and `new` can be run from nested directories; root is auto-discovered.
- Plan content can still be authored in any language.
- JSON output is the stable interface for automation.
