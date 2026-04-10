# Design: System Source of Truth

## Metadata

- Owner: pacto-core
- Created: 2026-04-08
- Last Modified: 2026-04-09
- State: to-implement
- Slug: system-source-of-truth

## Technical Context

- Language/Version: Go 1.24
- Dependencies: standard library only (no new deps)
- Constraints: markdown-first, no AST or formal parser; follow existing section-reading pattern from `move.go` (find heading, collect list items)

## Architecture Decisions

1. Decision: Use a directory of per-domain markdown docs under `.pacto/context/domains/`, not a single merged context file | Rationale: domains are the natural unit of ownership and conflict. Per-domain docs reduce merge noise, keep context scoped, and make it easier to enrich one area without touching unrelated system knowledge.

2. Decision: Two-tier extraction (mechanical + agent-driven) | Rationale: Domains are declarative and binary — pacto reads them mechanically from a structured section. Decisions and constraints require judgment — pacto surfaces the opportunity and the human/agent enriches. Avoids fragile regex on narrative text.

3. Decision: Domain reading uses the same pattern as Delta History | Rationale: pacto already reads structured markdown sections (e.g. `## Delta History` in plan README). Domains follow the same `find heading → collect list items` approach. No new parsing paradigm.

4. Decision: Overlap detection is a warning, not a block | Rationale: plans may intentionally overlap. The source of truth makes conflict visible, not impossible.

5. Decision: `context/` lives inside `.pacto/` alongside `plans/`, with `README.md` as overview and `domains/` as the source-of-truth subtree | Rationale: git-versioned through `.pacto/`, consistent with pacto's layout, discoverable by agents reading `.pacto/`, and flexible enough to evolve beyond domains later.

6. Decision: Testing in three layers — Go unit, Go integration, bash feature matrix | Rationale: Unit tests cover pure logic (domain parsing, merge, overlap detection). Integration tests cover cross-command state (init → new → fill → move → verify). Feature matrix covers CLI contract (flags, output, exit codes). The feature matrix alone is insufficient because bash can't easily assert file contents or state accumulation across commands.

## File Layout

```
.pacto/
├── context/
│   ├── README.md            ← Overview and conventions
│   └── domains/
│       ├── auth.md          ← Domain-specific context
│       └── session.md
├── plans/
│   ├── current/
│   │   └── add-auth/
│   │       ├── spec.md       ← declares ## Domains Affected
│   │       └── ...
│   └── ...
```

### `.pacto/context/README.md` Template

```markdown
# System Context

This directory holds the system source of truth for pacto.

## Structure

- `domains/` contains one markdown file per domain, for example `auth.md` or `billing.md`
- pacto creates domain docs mechanically from completed plans
- humans or agents enrich those docs with decisions, constraints, and notable references

## Conventions

- Use lowercase dash-separated domain filenames
- Keep one domain per file
- Preserve manual notes; pacto should only create missing scaffolds or append missing references safely
```

### Domain Doc Template

Each domain file starts as a small scaffold:

```markdown
# Domain: auth

## Summary

<!-- What this domain owns. -->

## Related Plans

- <state>/<slug>

## Decisions

<!-- Add key architectural choices for this domain. -->

## Constraints

<!-- Add rules future plans must respect. -->
```

### Spec Template Addition

The `## Domains Affected` section is added after `## Acceptance Criteria` in the spec template:

```markdown
## Domains Affected

- <domain>
```

## New Internal Package: `internal/context`

- `context.go` — types and constants for the context workspace
- `read.go` — list existing domain docs and read domain filenames/metadata
- `merge.go` — create or update per-domain docs under `context/domains/`
- `extract.go` — extract domains from a plan's spec.md
- `overlap.go` — detect overlapping domains across active plans
- `slug.go` — normalize domain names into stable filenames

### Key Functions

```go
// ExtractDomains reads ## Domains Affected from a spec.md and returns the list.
func ExtractDomains(specPath string) []string

// NormalizeDomainSlug turns freeform domain labels into stable filenames.
func NormalizeDomainSlug(raw string) string

// ReadContextDomains lists domain docs already present under context/domains/.
func ReadContextDomains(contextDir string) []string

// EnsureDomainDocs creates or updates per-domain docs under context/domains/.
func EnsureDomainDocs(contextDir string, domains []string, planRef string) error

// DetectOverlaps takes a map of plan -> domains and returns overlapping groups.
func DetectOverlaps(planDomains map[string][]string) []Overlap

// InitContext creates context/README.md and context/domains/.
func InitContext(plansRoot string) error
```

## Integration Points

### `pacto init`

After creating the plans directory structure, call `context.InitContext(plansRoot)` to create `.pacto/context/README.md` and `.pacto/context/domains/`. On re-init (idempotent path), do not overwrite existing context docs.

### `pacto new`

Add `## Domains Affected` section to `defaultSpecTemplate`. Same for Spanish template (`## Dominios Afectados`). Single placeholder line `- <domain>` (or `- <dominio>`).

### `pacto status`

After collecting active plans, read each plan's spec.md domains via `context.ExtractDomains`. Run `context.DetectOverlaps` and append warnings to the status output (both table and JSON formats). In JSON, add an `"overlaps"` field. In table, print warning lines after the plan table.

### `pacto move done`

After the existing move logic (rename dir, rewrite README), extract domains from the moved plan's spec.md and create/update the corresponding files under `.pacto/context/domains/`. Also add a `Related Plans` reference for the completed plan. Print Tier 2 enrichment prompt to stdout naming the affected domain docs.

## Testing Strategy

### Layer 1: Go Unit Tests (`internal/context/`)

New file `internal/context/context_test.go` covering:

| Test Case | What It Asserts |
|-----------|----------------|
| `TestExtractDomainsFromSpec` | Reads spec with `## Domains Affected` listing `auth, session` → returns `["auth", "session"]` |
| `TestExtractDomainsEmptySection` | Spec has heading but no items → returns empty slice |
| `TestExtractDomainsMissingSection` | Spec has no `## Domains Affected` → returns empty slice, no error |
| `TestExtractDomainsMalformed` | Spec has non-list content under heading → returns only valid `- item` lines |
| `TestExtractDomainsWhitespace` | Items with leading/trailing spaces → trimmed correctly |
| `TestNormalizeDomainSlug` | `Auth Session` → `auth-session` |
| `TestReadContextDomains` | Reads `context/domains/*.md` and returns correct domain list |
| `TestReadContextDomainsEmpty` | Empty `context/domains/` → returns empty |
| `TestReadContextDomainsMissingDir` | `context/domains/` doesn't exist → returns empty, no error |
| `TestEnsureDomainDocsNew` | Creates `auth.md` and `billing.md` with default scaffolds |
| `TestEnsureDomainDocsIdempotent` | Ensuring `auth` twice does not duplicate sections or references |
| `TestEnsureDomainDocsPreservesManualContent` | Existing Decisions/Constraints survive mechanical updates |
| `TestEnsureDomainDocsAppendsRelatedPlan` | Completed plan reference is added under `## Related Plans` |
| `TestDetectOverlapsNone` | No shared domains → empty result |
| `TestDetectOverlapsTwoPlans` | Two plans share `auth` → one overlap reported |
| `TestDetectOverlapsMultipleDomains` | Two plans share `auth` and `billing` → both reported |
| `TestDetectOverlapsThreePlans` | Three plans share one domain → all three listed |
| `TestDetectOverlapsSinglePlan` | Only one plan → no overlap |
| `TestInitContext` | Creates context/README.md and `context/domains/` |
| `TestInitContextIdempotent` | Calling twice does not overwrite existing content |
| `TestInitContextPreservesOnReInit` | Existing README and domain docs are preserved |

### Layer 2: Go Integration Tests (`internal/command/*`)

Extended/added tests in existing test files:

| Test Case | Commands | What It Asserts |
|-----------|----------|----------------|
| `TestInitCreatesContextWorkspace` | `RunInit` | `.pacto/context/README.md` and `.pacto/context/domains/` exist |
| `TestInitIdempotentPreservesContext` | `RunInit` twice | Custom context content survives re-init |
| `TestNewSpecIncludesDomainsAffected` | `RunNew` | spec.md contains `## Domains Affected` section |
| `TestNewSpecIncludesDominiosAfectados` | `RunNew` with `--lang es` | spec.md contains `## Dominios Afectados` section |
| `TestMoveDoneCreatesDomainDocs` | `RunNew`, write spec with domains, `RunMove done` | `auth.md` and `session.md` are created |
| `TestMoveDoneUpdatesExistingDomainDoc` | Move two plans with shared domain "auth" | `auth.md` stays single-file and gains both related plans |
| `TestMoveDonePrintsEnrichmentPrompt` | `RunMove done` | stdout contains Tier 2 prompt text and affected domain doc names |
| `TestStatusShowsOverlapWarning` | Create two plans with shared domain, `RunStatus` | output contains overlap warning |
| `TestStatusNoOverlapNoWarning` | Create two plans with different domains, `RunStatus` | no overlap warning |
| `TestStatusOverlapsInJSON` | Two plans with shared domain, `--format json` | JSON contains `"overlaps"` field |
| `TestMoveDoneNoDomainsGraceful` | Move plan without domains section | succeeds, context unchanged |
| `TestFullLifecycleIntegration` | init → new → fill domains → status (no overlap) → new overlapping plan → status (overlap) → move done → check domain docs | end-to-end state accumulation |

### Layer 3: Feature Matrix Updates (`bin/pacto-feature-matrix.sh`)

Add new sections to the existing script:

| Assertion | What It Checks |
|-----------|---------------|
| `context workspace exists after init` | `.pacto/context/README.md` and `domains/` are created |
| `spec.md has Domains Affected section` | New plan spec includes the section |
| `pacto status shows overlap` | Two plans with shared domain trigger warning |
| `pacto move done creates domain docs` | `.pacto/context/domains/<domain>.md` appears after move |
| `pacto move done prints enrichment prompt` | stdout mentions reviewing design.md and affected domain docs |
| `idempotent domain docs on second move` | Moving second plan with same domain does not create duplicate files |

## Out of Scope

- Automatic extraction of decisions or constraints from design.md (Tier 2 stays agent-driven)
- Domain versioning or history tracking
- Formal merge engine with deltas
- Diagrams or visualization
- ADR numbering scheme
