# Design: Spec Baseline and Requirements Grammar

## Metadata

- Owner: pacto-core
- Created: 2026-04-29
- Last Modified: 2026-04-29
- State: to-implement
- Slug: spec-baseline-and-requirements

## Context

Pacto already has the building blocks this plan needs:

- [internal/parser/parser.go](../../../../internal/parser/parser.go) parses
  plan markdown (phases, tasks, structured Delta History entries).
- [internal/planfmt](../../../../internal/planfmt/) validates and normalizes
  plan documents — strict mode for active plans.
- [internal/context/context.go](../../../../internal/context/context.go)
  already manages `.pacto/context/domains/<domain>/` and demonstrates the
  pattern of "init creates folder, move done writes into it".
- [internal/app/move/move.go](../../../../internal/app/move/move.go) is the
  single hook for archive-time work.
- [internal/assets/templates/](../../../../internal/assets/templates/) and
  `templates_es/` host the markdown templates emitted by `pacto new`.

This plan extends the same patterns to a new dimension: capability-scoped
baseline files plus a structured Requirement grammar inside plan specs.

## Goals

- Add a baseline tree under `.pacto/specs/<capability>/spec.md` that survives
  plan completion.
- Define a parseable Requirement / Scenario grammar that downstream commands
  can consume without regex archaeology.
- Make plans express their changes as deltas against capabilities and merge
  those deltas on `move done`.
- Stay backwards compatible: every plan that does not adopt the new sections
  continues to work.

## Non-Goals

- Implementing the analyze / clarify commands (separate plans).
- Building a UI for browsing the baseline (out of scope).
- Allowing multi-file capabilities (one capability == one baseline file v1).

## Decisions

### D1: New package `internal/specsbaseline`

Houses pure functions: parsing, validation helpers, and merge logic. Keeping
it separate from `internal/parser` (plan parsing) avoids bloating that
package with capability concerns and lets it import `parser` types if
needed without cycles.

Alternatives rejected:
- Extending `internal/parser` directly — would conflate plan parsing with
  baseline merging.
- Putting it under `internal/context` — context is about domain knowledge
  notes, not authoritative requirements.

### D2: Grammar shape

Inside a plan `spec.md`:

```markdown
## Capabilities

- New Capabilities: [auth, profile]
- Modified Capabilities: [billing]

## Capability: auth

### ADDED Requirements

#### Requirement: User can sign in with OAuth
The system SHALL allow users to authenticate via Google OAuth.

##### Scenario: Successful sign in
- WHEN the user completes the OAuth flow
- THEN the system creates a session and redirects to the dashboard
```

Rules:
- Capability sections live at `## Capability: <slug>`.
- Delta op headers are `### ADDED Requirements`,
  `### MODIFIED Requirements`, `### REMOVED Requirements`,
  `### RENAMED Requirements` (mirrors OpenSpec to keep cognitive load low).
- Each Requirement is `#### Requirement: <name>`; each Scenario is
  `##### Scenario: <name>` with `WHEN`/`THEN` bullet lines.

For non-delta plan specs (greenfield Requirements not tied to baseline merge),
authors may use a simpler `## Requirements` block with the same Requirement /
Scenario shape one heading-level higher; the parser detects which form a
section uses by its parent header.

Alternatives rejected:
- OpenSpec's exact layout (`### Requirement` / `#### Scenario`) at the top
  level — collides with Pacto's existing `## Acceptance Criteria` and
  `## Phase N` headers and would break planfmt rules.
- YAML frontmatter for requirements — loses readability and breaks the
  markdown-first principle.

### D3: Stable IDs

Requirements and Scenarios get IDs assigned by parse order within a file:
`R-001`, `R-002`, `S-001` (scoped per Requirement). IDs are not stored in
the markdown — they are derived. Authors who want stable cross-references
across renames may add an HTML comment marker
(`<!-- id: R-014 -->`) which the parser honors when present.

Alternatives rejected:
- Forcing authors to write IDs by hand — friction and merge conflicts.
- Hash-based IDs — opaque and change on any wording edit.

### D4: Baseline merge semantics

`pacto move done <slug>` triggers merge **only** when the plan declares one or
more `## Capability:` sections. For each capability:

- ADDED → append the Requirement block to baseline; if a Requirement of the
  same name already exists, return an error and abort the move.
- MODIFIED → locate the matching Requirement in baseline (whitespace-
  insensitive name match) and replace its block; error if missing.
- REMOVED → delete the Requirement block; insert
  `<!-- removed by <slug> on <YYYY-MM-DD> -->` at the deletion site.
- RENAMED → rewrite the `#### Requirement:` header line only.

Atomicity: each baseline file is written via temp file + `os.Rename`. If any
capability fails, no baseline file is written and the move is aborted with a
clear error message; the user fixes the spec and re-runs `move done`.

### D5: Status coverage

`pacto status` already builds a per-plan report. Extend the report to:

- Parse Requirements from each active plan's `spec.md` (Capability blocks +
  `## Requirements` block).
- Walk `tasks.md` looking for `R-NNN` references in task descriptions; count
  matches per Requirement.
- Reuse existing claim/evidence machinery for the Requirement's evidence
  count (a Requirement may include `Evidence: paths=...` lines parsed the
  same way claims are today).
- Emit a new `requirements` block in JSON output and a new column in the
  table view.

### D6: Idempotent init

`pacto init` calls a new helper `specsbaseline.InitBaseline(plansRoot)` that
creates `.pacto/specs/` plus `README.md` only when missing. Existing files
are preserved. Mirrors `context.InitContext`.

### D7: i18n

Spanish templates get parallel sections:
- `## Capabilities` → `## Capacidades`
- `## Requirements` → `## Requisitos`
- `### Requirement: …` → `### Requisito: …`
- `#### Scenario: …` → `#### Escenario: …`

The parser recognizes both English and Spanish header forms (already the
pattern in `internal/parser`).

## Risks / Trade-offs

- **Risk**: Authors adopt the Capability/Requirement grammar inconsistently,
  producing partial baselines.
  **Mitigation**: planfmt validation in strict mode + clear error messages
  in `move done`.

- **Risk**: Merge conflicts between two plans modifying the same Requirement.
  **Mitigation**: detect on `move done` ("Requirement X already MODIFIED by
  plan Y not yet archived") via a simple cross-plan scan; defer full
  resolution to a follow-up plan.

- **Risk**: Baseline files diverge from code reality (the very problem Pacto
  exists to solve).
  **Mitigation**: out of scope for this plan, but the baseline is the
  natural target for a future `pacto verify --baseline` that runs the
  existing rg-based verification against Requirement-level claims.

- **Risk**: Legacy plans break.
  **Mitigation**: every new code path checks for the presence of the new
  sections and degrades to no-op when absent. AC-009 + dedicated tests.

## Migration Plan

This is additive. No migration of existing plans is required. Adoption is
driven by templates so new plans get the grammar by default; existing plans
can opt in by adding the sections by hand.

## Open Questions

- Should the baseline file allow a `## Notes` free-form section alongside
  Requirements? (Lean: yes, for human context.)
- Do we need a `pacto baseline show <capability>` read-only viewer in this
  plan, or punt to a follow-up? (Lean: punt — `cat` works.)
- How should two plans both declare `New Capabilities: [foo]` be handled? At
  `move done` time the second one would collide. Likely: detect and error
  with a hint to coordinate.
