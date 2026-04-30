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

## Domain Context

As plans complete, Pacto accumulates a system source of truth under `.pacto/context/domains/`.

Each domain gets its own folder:

```text
.pacto/context/domains/
└── auth/
    ├── context.md    ← bounded context: purpose, boundary, key terms, rules, collaborators
    └── decisions.md  ← architectural decisions (ADR-style): why things are the way they are
```

Pacto creates the folder and stubs when a plan first declares that domain in `## Domains Affected` and moves to `done`. After that, the files are human/agent territory — Pacto never overwrites them.

`pacto status` detects when two active plans declare the same domain and surfaces a warning.

## Capability Baseline and Requirements

Alongside the free-form domain notes, Pacto maintains a structured **capability baseline** under `.pacto/specs/`:

```text
.pacto/specs/
├── README.md
└── auth/
    └── spec.md       ← `### Requirement:` blocks with nested `#### Scenario:`
```

Each baseline `spec.md` answers "what does this capability do today?" with addressable units:

- `### Requirement: <name>` — the smallest verifiable behaviour (auto-assigned `R-NNN` IDs)
- `#### Scenario: <name>` — concrete `WHEN`/`THEN` example for that requirement (`S-NNN`)

Plans express changes against the baseline using delta blocks inside their own `spec.md`:

```markdown
## Capabilities

- New Capabilities: [auth]
- Modified Capabilities: [billing]

## Capability: auth

### ADDED Requirements

#### Requirement: User can sign in with OAuth
The system SHALL allow users to authenticate via Google OAuth.

##### Scenario: Successful sign in
- WHEN the user completes the OAuth flow
- THEN the system creates a session
```

Supported delta ops: `ADDED`, `MODIFIED`, `REMOVED`, `RENAMED` (use `- to: <new name>` inside the body).

On `pacto move done`, Pacto pre-validates every delta against the current baseline and only renames the plan folder if the merge would succeed. Then it folds the deltas in atomically (temp file + rename per capability), creating `.pacto/specs/<slug>/spec.md` if the capability is new.

`pacto status` reports per-Requirement coverage: how many tasks reference `R-NNN` in `tasks.md` and how many evidence lines mention it. Requirements with zero task references are flagged `uncovered`.

Spanish keywords `Capacidad`, `Requisito`, `Escenario`, `Capacidades`, `Requisitos` are accepted alongside their English forms.

## Workspace vs Product Docs

- `docs/`: canonical product/user documentation.
- `.pacto/plans/*`: workspace artifacts and templates generated/used by CLI.

This separation keeps user docs stable while workspace templates remain operational.
