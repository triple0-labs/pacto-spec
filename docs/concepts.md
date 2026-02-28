# Concepts

Pacto is an SDD tool: plans act as executable specs, and status is validated against repository evidence.

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
- `PLAN_<TOPIC>_<YYYY-MM-DD>.md`

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

## Workspace vs Product Docs

- `docs/`: canonical product/user documentation.
- `.pacto/plans/*`: workspace artifacts and templates generated/used by CLI.

This separation keeps user docs stable while workspace templates remain operational.
