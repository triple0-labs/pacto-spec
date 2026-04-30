# Design: Catalog-based i18n with locale in context

## Metadata

- Owner: Platform Team
- Created: 2026-04-30
- Last Modified: 2026-04-30
- State: to-implement
- Slug: i18n-message-catalog
- Language: en

## Technical Context

- Language/Version: Go 1.23 (module `pacto`)
- Dependencies: `gopkg.in/yaml.v3` (already in `go.mod`); no third-party i18n
  library — keep the runtime dependency surface zero.
- Constraints:
  - Catalogs MUST be embedded via `//go:embed` so a single binary still ships
    every locale.
  - Existing `i18n.T(lang, en, es)` MUST keep working during the migration
    (deprecated shim → eventual removal in a follow-up tag).
  - No new public Go interfaces should leak `*Catalog` types — the API surface
    is just `i18n.M`, `i18n.WithLocale`, `i18n.LocaleFrom`.
  - Lint and parity tests MUST run as ordinary `go test` (no extra build tools)
    so they piggy-back on the existing CI lint job.

## Architecture Decisions

1. **Decision:** Catalogs are embedded YAML files at
   `internal/i18n/messages.<code>.yaml`, loaded once at package init.
   **Rationale:** Editable by translators in a plain text editor, diff-friendly
   in PRs, supports comments to give translators context, easy to validate.

2. **Decision:** Message IDs use **dotted lowercase namespaces** mirroring the
   package path (`install.no_tools_selected`, `move.error.bad_slug`,
   `cli.help.usage`, `render.report.col.state`). Reserve `common.*` for
   cross-cutting words (warning, error, ok, dim labels).
   **Rationale:** Searchable, collision-free, and translators can scope work to
   one prefix.

3. **Decision:** `i18n.M(ctx context.Context, id string, args ...any) string`
   is the sole runtime entry point. Format strings use `text/template` syntax
   (`{{.tools}}`) rather than `printf` verbs because named args survive
   reordering between locales.
   **Rationale:** Spanish word order often differs ("Install tools for X" vs
   "Instalar herramientas para X y…"). `printf` positional args break.

4. **Decision:** Active locale lives on `context.Context` via an unexported
   key. `i18n.WithLocale(ctx, lang)` wraps the context, `i18n.LocaleFrom(ctx)`
   reads it (default English on cache miss).
   **Rationale:** Eliminates threading `lang i18n.Language` through 25+
   call chains; matches stdlib idioms (`net/http`, `database/sql`).

5. **Decision:** The CLI persistent pre-run resolves the effective locale once
   (CLI override → workspace `ui.language` → English) and stores it on
   `cmd.Context()` via `cmd.SetContext(i18n.WithLocale(ctx, lang))`. Every
   command's `RunE` receives that context and propagates it.
   **Rationale:** Single source of truth for "what locale am I rendering in",
   resolved at exactly one point.

6. **Decision:** Plans gain an optional `Language: <code>` metadata field.
   Per-plan rendering/exec wraps `ctx` with that locale before calling
   formatters. The `detectPlanDocLanguage` heuristic in
   `internal/app/execplan/execplan.go` is deleted in the same commit that
   teaches `planfmt` to read `Language:`.
   **Rationale:** Authors know the language of their own document; sniffing is
   a workaround for missing data.

7. **Decision:** Heading aliases in `planfmt` are derived from a small
   per-locale registry (English canonical key → ES synonyms) instead of a
   single bilingual flat map. The registry lives at
   `internal/planfmt/headings.go` and is loaded at init.
   **Rationale:** Adding a new locale's heading synonyms becomes one block of
   data, not edits scattered across a 70-line map.

8. **Decision:** A parity test (`internal/i18n/parity_test.go`) loads every
   `messages.*.yaml`, compares the key sets, and fails listing any missing
   IDs per locale. A second parity test (`internal/assets/parity_test.go`)
   compares the file set under `templates/` vs `templates_es/`.
   **Rationale:** Cheap, deterministic, runs in the regular test job.

9. **Decision:** A static-analysis test (`internal/i18n/lint/lint_test.go`)
   walks AST of `internal/command/...`, `internal/app/...`, `internal/cli/...`,
   `internal/render/...` and flags any `fmt.{F,}Print{ln,f}?` call whose
   first non-Writer argument is a string literal of length ≥ 8 containing
   a space, unless the line carries a `//pacto:i18n-ignore` directive or
   the file appears in an explicit allow-list (e.g. JSON encoders).
   **Rationale:** Catches the regression class the audit identified (~67 raw
   English prints) without requiring a custom linter binary.

10. **Decision:** The migration is performed package-by-package, with each
    package landing its own commit on `main`. The deprecated `i18n.T` shim
    stays alive throughout the migration and is removed in the final commit
    once all callers are gone.
    **Rationale:** Keeps each PR reviewable; the build stays green throughout.

## Catalog file format

```yaml
# internal/i18n/messages.en.yaml
common:
  warning: warning
  error: error
  ok: ok

install:
  no_tools_selected: "No tools selected; nothing to do."
  running: "Running {{.command}}"
  skipped_unmanaged: "skipped unmanaged file (use --force)"

move:
  error:
    bad_slug: "move requires <from-state> <slug> <to-state>"
    parse_capabilities: "move error: parse capability deltas: {{.err}}"

cli:
  help:
    usage: "Usage:"
    commands: "Commands:"
```

```yaml
# internal/i18n/messages.es.yaml — identical key set
common:
  warning: advertencia
  error: error
  ok: ok

install:
  no_tools_selected: "No se seleccionaron herramientas; nada por hacer."
  running: "Ejecutando {{.command}}"
  skipped_unmanaged: "archivo no gestionado omitido (usa --force)"

move:
  error:
    bad_slug: "move requiere <from-state> <slug> <to-state>"
    parse_capabilities: "error de move: analizar deltas de capacidades: {{.err}}"

cli:
  help:
    usage: "Uso:"
    commands: "Comandos:"
```

## Public API sketch

```go
// internal/i18n/catalog.go
package i18n

import (
    "context"
    _ "embed"
    "text/template"
)

type localeKey struct{}

func WithLocale(ctx context.Context, lang Language) context.Context {
    return context.WithValue(ctx, localeKey{}, lang)
}

func LocaleFrom(ctx context.Context) Language {
    if v, ok := ctx.Value(localeKey{}).(Language); ok && v != "" {
        return v
    }
    return English
}

// M renders a message by ID using the locale stored in ctx.
// args is converted to a template data map; positional args fill {{.0}}, {{.1}},
// while a single map[string]any fills named placeholders.
func M(ctx context.Context, id string, args ...any) string { ... }

// Deprecated: use M with a stable ID. Kept temporarily for migration.
func T(lang Language, en, es string) string { ... }
```

## Migration plan (high level)

1. Land the catalog runtime + parity tests + lint test (no behaviour change).
2. Seed `messages.en.yaml` and `messages.es.yaml` with every existing
   `i18n.T` pair, key by ID; replace each call site one package at a time.
3. Wire `WithLocale` into the root command and remove explicit `lang` arguments
   from the migrated packages.
4. Add `Language:` recognition to `planfmt` and remove `detectPlanDocLanguage`.
5. Backfill catalog entries for the ~67 untranslated `fmt.Print*` lines and
   migrate them.
6. Enable the lint test in CI as a hard failure (it can run in advisory mode
   while the migration is in flight).
7. Delete the three duplicate `tr` wrappers and remove the `i18n.T` shim.

## Out of scope

- Pluralization rules (CLDR-style). Add later as `i18n.MN(ctx, id, n, args...)`
  if a real need shows up.
- Locale-aware date/number formatting. Pacto only renders RFC3339 timestamps,
  which are locale-neutral.
- A web-style translation memory tool. The YAML files are the authoritative
  source; if/when external translators want a UI, point them at a YAML editor.
- Right-to-left layout — irrelevant for a CLI.
