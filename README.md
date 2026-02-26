# Pacto CLI Repository

This repository contains:
- the `pacto` CLI engine (Go),
- canonical planning/spec documents,
- active and historical plans,
- sample repositories for testing.

## Features

- `pacto status`:
  - discovers plans across lifecycle folders and computes per-plan status
  - parses progress, tasks, blockers, and next actions from plan docs
  - verifies implementation claims and emits evidence in table or JSON output
- Plan lifecycle structure:
  - standardized states: `current/`, `to-implement/`, `done/`, `outdated/` (optionally `archive/`)
  - each plan uses a folder with `README.md` and plan docs (`PLAN*.md` or fallback non-README `*.md`)
- `pacto new`:
  - scaffolds a new plan folder from templates
  - generates `README.md` and initial `PLAN_*.md`
  - updates workspace index metadata
- Verification engine:
  - path claim verification bounded to configured root
  - symbol/endpoint/test-ref verification via repository search
  - plan docs excluded as implementation evidence (`plan_doc_only`)
  - outside-root path claims reported explicitly (`outside_root`)
- Configuration (`.pacto-engine.yaml`):
  - root/mode/format/fail behavior/state filters/archive inclusion
  - claim-type toggles (`paths`, `symbols`, `endpoints`, `test_refs`)
  - report limits (next actions, blockers)
- Reporting model:
  - per-plan status, progress, blockers, verification, confidence, and claim references
  - repository-level summaries by state and verification outcome

## Repository Layout

- `cmd/`:
  - CLI entrypoints (`pacto`, `pacto-engine` compatibility wrapper)
- `internal/`:
  - core packages (`parser`, `verify`, `analyze`, `report`, etc.)
- `plans/`:
  - full planning workspace:
    - `current/`, `to-implement/`, `done/`, `outdated/`, `archive/`
  - governance/spec docs:
    - `plans/PACTO.md`
    - `plans/PLANTILLA_PACTO_PLAN.md`
    - `plans/SLASH_COMMANDS.md`
  - index of plans:
    - `plans/README.md`
- `samples/`:
  - test fixtures, including `samples/mock-pacto-repo/`

## Quick Start

```bash
# Show help
pacto help

# Show version
pacto version

# Status for default workspace (auto-detects ./plans)
pacto status

# Status for explicit workspace
pacto status --root ./plans

# Create a new plan
pacto new to-implement my-plan-slug

# Create inside sample fixture
pacto new to-implement demo-plan --root ./samples/mock-pacto-repo
```

## Shipping Releases

- CI workflow: `.github/workflows/ci.yml` runs `go test ./...` on PRs and pushes to `main`.
- Release workflow: `.github/workflows/release.yml` publishes GitHub releases on tags matching `v*`.
- GoReleaser config: `.goreleaser.yaml` builds `pacto` and `pacto-engine` for:
  - `linux/amd64`, `linux/arm64`
  - `darwin/amd64`, `darwin/arm64`
- Version injection:
  - binaries embed version with `-X pacto/internal/app.Version={{.Version}}`
  - validate with `pacto version`

Release flow:

```bash
git tag v0.1.0
git push origin v0.1.0
```

## Notes

- CLI output is English-only.
- Plan content can still be in Spanish or other languages.
- JSON output remains canonical for automation.
