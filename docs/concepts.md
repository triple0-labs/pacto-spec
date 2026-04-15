# Concepts

Pacto is an **SDD (spec-driven development)** tool: plans act as executable specs, and `pacto status` is validated against repository evidence — not assumed from narrative text alone.

For the default workflow in order, see [Getting started](./getting-started.md). For package layering, see [Architecture](./architecture.md).

## Planning Model

Each plan belongs to one state:

- `to-implement`: not started yet
- `current`: in progress
- `done`: completed
- `outdated`: superseded or stale

Each plan slice lives at:

```text
<plans-root>/<state>/<slug>/
```

Minimum files per slice:

- `README.md`
- `spec.md`
- `design.md`
- `tasks.md`

## Evidence-First Verification

`pacto status` does not rely only on narrative plan text.

It parses plan documents, extracts claims, and verifies them against `repo-root` evidence.

Claim categories:

- `paths`
- `symbols`
- `endpoints`
- `test_refs`

Verification outcomes:

- `verified`
- `partial`
- `unverified`

## Delta Schema (Canonical + i18n)

Structured delta parsing uses an English canonical schema:

- `Delta History`
- `Delta D-YYYY-MM-DD-XX`
- fields like `Date`, `Status`, `Changes`, `Next Delta`

For compatibility and localization, parser aliases accept Spanish headings and
field labels (for example `Historial de deltas`, `Fecha`, `Siguiente delta`).

Canonical code paths, enums, and config remain English-first.

## Workspace vs Product Docs

- `docs/`: canonical product/user documentation.
- `.pacto/plans/*`: workspace artifacts and templates generated/used by CLI.

This separation keeps user docs stable while workspace templates remain operational.
