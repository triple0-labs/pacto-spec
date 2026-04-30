# Spec: Catalog-based i18n with locale in context

## Metadata

- Owner: Platform Team
- Created: 2026-04-30
- Last Modified: 2026-04-30
- State: to-implement
- Slug: i18n-message-catalog
- Language: en

## Problem Statement

Pacto's current i18n consists of a single helper `i18n.T(lang, en, es string)`
invoked inline at ~70 call sites, three duplicate `tr`/`Tr`/`trLang` wrappers,
two parallel asset template trees (`templates/`, `templates_es/`) maintained by
hand, and a hand-written bilingual heading-alias table inside `internal/planfmt`.
Consequences observed in the audit:

1. ~67 user-facing `fmt.Print*` calls bypass `T()` entirely and are
   English-only (most of `internal/command/movecmd`, `cli/guardrails.go`,
   `cli/root_cmd.go`, `command/normalize`).
2. Single commands mix localized and hard-coded strings (e.g.
   `internal/command/install/install.go:75` concatenates the English literal
   `"Running "+opts.Command` even on the Spanish branch).
3. Translators cannot review strings without reading Go source: messages live as
   pairs of literal arguments scattered across 25+ files.
4. Plan documents have to be sniffed at runtime by `detectPlanDocLanguage` in
   `internal/app/execplan` because plans never declare their authored language.
5. Adding a third locale requires editing every `i18n.T` call site, every alias
   table, and every template directory — there is no single source of truth.

The scope of this plan is to replace the inline-pair model with an embedded
**message catalog** keyed by stable IDs, propagate the active locale through
`context.Context` instead of as a function argument, declare the locale on plan
metadata, and add CI guardrails preventing regressions (missing IDs in any
locale, untranslated `fmt.Print*` in user-facing packages, drift between
`templates/` and `templates_es/`).

## User Scenarios

### Scenario: Adding a new user-facing message

- **GIVEN** a developer needs to print a new error or status line from a command
- **WHEN** they add the call site
- **THEN** they reference a message ID (`i18n.M(ctx, "install.no_tools_selected")`)
  and add the corresponding entry to `messages.en.yaml` and `messages.es.yaml`,
  and CI fails the build if either entry is missing.

### Scenario: Translator updates Spanish copy

- **GIVEN** a translator with no Go knowledge
- **WHEN** they open `internal/i18n/messages.es.yaml`
- **THEN** they see every Spanish string in the product in one file, grouped by
  domain prefix, and can edit them without touching Go source.

### Scenario: Adding a new locale

- **GIVEN** a contributor wants to add `pt-BR`
- **WHEN** they create `internal/i18n/messages.pt-BR.yaml` with the same set of
  IDs as `messages.en.yaml`
- **THEN** the catalog parity test passes, the new locale becomes selectable via
  `--lang pt-BR` or `ui.language: pt-BR`, and every previously translated UI
  surface renders in Portuguese without further code changes.

### Scenario: Plan declares its language

- **GIVEN** an author writes a plan in Spanish
- **WHEN** they set `Language: es` in the plan's metadata block (or use
  `pacto new --lang es`)
- **THEN** parsing, normalization, validation, and execution all read the
  declared language directly without phrase-counting heuristics.

### Scenario: Locale flows through one command end-to-end

- **GIVEN** the user runs `pacto status --lang es`
- **WHEN** any callee deep in the call chain needs to render a label
- **THEN** it pulls the locale from `context.Context` rather than receiving it
  as an explicit parameter from every intermediate function.

### Scenario: Untranslated print is rejected by CI

- **GIVEN** a contributor adds `fmt.Fprintln(os.Stderr, "raw English message")`
  to a file under `internal/command/`, `internal/app/`, or `internal/cli/`
- **WHEN** CI runs the i18n lint
- **THEN** the build fails with a pointer to the offending line.

### Scenario: Templates parity is enforced

- **GIVEN** a contributor edits `internal/assets/templates/PACTO.md`
- **WHEN** they forget to update `internal/assets/templates_es/PACTO.md`
- **THEN** `go test ./internal/assets/...` fails with a parity error listing
  the missing or out-of-date files.

## Acceptance Criteria

- AC-001: All current `i18n.T(lang, en, es)` call sites are migrated to
  `i18n.M(ctx, id, args...)` and the inline-pair signature is removed (or kept
  only as a deprecated thin shim with a vet directive).
- AC-002: Every message ID present in any locale catalog is present in *all*
  locale catalogs; the set difference is asserted empty in CI.
- AC-003: `messages.en.yaml` and `messages.es.yaml` cover every string previously
  passed to `i18n.T`, plus the ~67 currently English-only `fmt.Print*` user-
  facing messages identified in the audit.
- AC-004: Active locale is carried in `context.Context`; no command-package
  function in the post-migration world receives `lang i18n.Language` as an
  explicit parameter when it could pull it from `ctx`.
- AC-005: Plan documents may declare `Language: <code>` in their metadata; when
  present the value is authoritative and `detectPlanDocLanguage` is removed.
- AC-006: `internal/assets` exposes a parity test that fails when the file set
  in `templates/` differs from `templates_es/`.
- AC-007: A `go test`-based i18n lint flags any `fmt.F?Print*` literal-string
  call in `internal/command/`, `internal/app/`, `internal/cli/`, or
  `internal/render/` whose first format-string argument is a non-trivial English
  literal not produced by `i18n.M`.
- AC-008: The three duplicate translation wrappers (`cmdutil.Tr`, `cli.tr`,
  `execplan.trLang`) are removed; only one `i18n` entry point remains.
- AC-009: The Spanish branch of every install/init/exec/status command no
  longer concatenates raw English fragments.
- AC-010: Adding a brand-new locale requires only adding one
  `messages.<code>.yaml` (and optionally a `templates_<code>/` directory) — no
  code changes to switch on the locale.

## Capabilities

- New Capabilities: [i18n-catalog]
- Modified Capabilities: [plan-grammar, cli-output]

## Requirements

### Requirement: Message catalog is the single source of truth for UI strings

The system SHALL load all user-facing UI strings from embedded per-locale
message catalogs keyed by stable string IDs, and SHALL fall back to the English
catalog when a locale-specific entry is missing while logging a build-time
warning so CI surfaces the gap.

#### Scenario: Lookup by ID

- WHEN code calls `i18n.M(ctx, "install.no_tools_selected")` with a locale of `es`
- THEN the system returns the value from `messages.es.yaml` for that ID.

#### Scenario: Missing translation falls back to English

- WHEN the active locale is `es` but the ID has no entry in `messages.es.yaml`
- THEN the system returns the English value and CI's parity test fails the build.

#### Scenario: Unknown ID is a programmer error

- WHEN code calls `i18n.M(ctx, "nonexistent.id")`
- THEN the system returns the literal ID string at runtime and the i18n lint
  test fails the build naming the offending call site.

### Requirement: Active locale propagates via context.Context

The system SHALL store the active locale in `context.Context` at the CLI entry
point and SHALL provide `i18n.WithLocale(ctx, lang)` and `i18n.LocaleFrom(ctx)`
helpers so no command-tree function needs to receive `lang` as an explicit
parameter.

#### Scenario: CLI entry sets the locale once

- WHEN a command's `RunE` is invoked
- THEN the persistent pre-run resolves the effective locale once and stores it
  on the request context before any business logic runs.

#### Scenario: Deep callee reads the locale

- WHEN a function 5 levels deep needs to format a label
- THEN it calls `i18n.M(ctx, "...")` directly without `lang` being threaded
  through intermediate signatures.

### Requirement: Plans declare their authored language

The system SHALL recognise an optional `Language: <code>` field on plan
metadata (in `spec.md` and `tasks.md`) and SHALL use that value as the source
of truth for parsing, normalization, validation, rendering, and execution
touching that plan.

#### Scenario: Plan declares Spanish

- WHEN a plan's metadata block contains `Language: es`
- THEN `pacto status`, `pacto exec`, and `pacto move` treat all of its content
  as Spanish without invoking phrase-counting heuristics.

#### Scenario: Plan omits language

- WHEN a plan does not declare a `Language:` field
- THEN the system treats it as the workspace default declared in
  `.pacto/config.yaml::ui.language`.

### Requirement: Catalog parity is enforced by CI

The system SHALL provide an automated test that fails when a message ID is
present in one locale catalog but missing from another, and SHALL provide a
similar test that fails when the file set in `internal/assets/templates/`
differs from `internal/assets/templates_es/`.

#### Scenario: Missing key in es catalog

- WHEN a developer adds `mykey:` to `messages.en.yaml` and forgets the es file
- THEN `go test ./internal/i18n/...` fails listing `mykey` as missing in `es`.

#### Scenario: Templates drift

- WHEN a developer adds `templates/NEWFILE.md` without adding
  `templates_es/NEWFILE.md`
- THEN `go test ./internal/assets/...` fails listing `NEWFILE.md` as missing in
  `templates_es/`.

### Requirement: Untranslated user-facing prints are rejected by lint

The system SHALL provide a static-analysis test that scans
`internal/command/`, `internal/app/`, `internal/cli/`, and `internal/render/`
for `fmt.F?Print*` calls whose first format-string argument is a non-trivial
English string literal that did not originate from `i18n.M`, and SHALL fail
when any are found, except for explicitly allow-listed exceptions.

#### Scenario: Hard-coded error string

- WHEN a contributor adds `fmt.Fprintln(os.Stderr, "move error: bad slug")`
  under `internal/command/movecmd/`
- THEN `go test ./internal/i18n/lint/...` fails with the file:line and a
  suggested replacement `i18n.M(ctx, "move.error.bad_slug")`.

#### Scenario: Allow-list escape hatch

- WHEN a print is intentionally untranslated (JSON output, machine-readable
  diagnostics)
- THEN the contributor can mark the line with `//pacto:i18n-ignore` and the
  lint passes.

## Domains Affected

- internal/i18n
- internal/cmdutil
- internal/cli
- internal/command (all subpackages)
- internal/app (newplan, execplan, explore, doctor, status)
- internal/render
- internal/tui (init wizard, status TUI)
- internal/onboarding
- internal/assets
- internal/planfmt

## Capability: i18n-catalog

### ADDED Requirements

- Message catalog is the single source of truth for UI strings
- Active locale propagates via context.Context
- Catalog parity is enforced by CI
- Untranslated user-facing prints are rejected by lint

## Capability: plan-grammar

### MODIFIED Requirements

- Plans declare their authored language

## Capability: cli-output

### MODIFIED Requirements

- All command output goes through the catalog (no inline `i18n.T(lang, en, es)`
  pairs remain in command-tree code).
