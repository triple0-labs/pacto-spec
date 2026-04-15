# Spec: System Source of Truth

## Metadata

- Owner: pacto-core
- Created: 2026-04-08
- Last Modified: 2026-04-09
- State: to-implement
- Slug: system-source-of-truth

## Domains Affected

- status
- move
- new
- init

## Problem Statement

Pacto plans are silos. Two plans in `current/` or `to-implement/` can silently contradict each other because there is nothing they are both measured against. When a plan completes (`done/`), its decisions and rationale disappear into a graveyard. The next person (or AI agent) starting a new plan has no shared understanding of what the system already is.

We need a lightweight, markdown-first shared understanding that makes plan conflicts visible and accumulates organizational knowledge as plans complete — without becoming a heavyweight spec engine.

## User Scenarios

### Scenario: domain overlap detection

- **GIVEN** two active plans that both declare domain "auth"
- **WHEN** the user runs `pacto status`
- **THEN** a warning is displayed showing the overlapping domain and which plans share it

### Scenario: knowledge accumulation on move done

- **GIVEN** a completed plan with `## Domains Affected` listing "auth" and "session"
- **WHEN** the user runs `pacto move done <slug>`
- **THEN** pacto mechanically creates or updates `.pacto/context/domains/auth.md` and `.pacto/context/domains/session.md`
- **AND** prints a prompt suggesting the agent/dev review `design.md` and enrich the relevant domain docs with decisions and constraints worth preserving

### Scenario: greenfield startup

- **GIVEN** a freshly initialized workspace with no completed plans
- **WHEN** `pacto init` runs
- **THEN** `.pacto/context/README.md` is created as an overview plus `.pacto/context/domains/` as the source-of-truth directory
- **AND** the first plan's spec template includes a `## Domains Affected` section

### Scenario: existing plans without domains

- **GIVEN** plans created before this feature (no `## Domains Affected` section)
- **WHEN** `pacto status` reads those plans
- **THEN** no overlap warning is shown and no error occurs — missing domains degrade to empty

### Scenario: agent enrichment prompt

- **GIVEN** a plan moved to done with a non-empty `design.md`
- **WHEN** the move completes
- **THEN** pacto prints a Tier 2 prompt: "This plan may contain decisions or constraints worth preserving. Review design.md and update the affected domain docs under `.pacto/context/domains/`"

## Acceptance Criteria

- AC-001: `pacto init` creates `.pacto/context/README.md` as a context overview and `.pacto/context/domains/` as the per-domain source-of-truth directory.
- AC-002: `pacto new` generates spec.md with a `## Domains Affected` section containing a placeholder `- <domain>` entry.
- AC-003: `pacto status` reads `## Domains Affected` from all active plans (current + to-implement) and warns when two or more plans declare the same domain.
- AC-004: `pacto move done <slug>` extracts domains from spec.md, creates or updates `.pacto/context/domains/<domain>.md` files for each declared domain, and prints a Tier 2 enrichment prompt.
- AC-005: Plans without `## Domains Affected` are handled gracefully — no crash, no false overlap warnings.
- AC-006: Overlap warnings include the plan names and domain for easy identification.
- AC-007: Existing Go unit tests continue passing (no regression in move, status, new, init).
- AC-008: `go test ./...` continues passing after template changes (no separate bash feature matrix).
- AC-009: All new logic is covered by Go unit tests (domain extraction, domain-doc initialization/update, overlap detection, domain slug normalization).
- AC-010: Go integration tests cover the full command sequences: init → new → fill spec → move done → verify `.pacto/context/domains/<domain>.md`.
