# Tasks: Catalog-based i18n with locale in context

## Execution Metadata

- Status: Draft
- Owner: Platform Team
- Created: 2026-04-30
- Last Modified: 2026-04-30
- State: to-implement
- Slug: i18n-message-catalog
- Language: en

## Implementation Plan by Phases

## Phase 1: Catalog runtime + locale-in-context

- [ ] 1.1 Create `internal/i18n/catalog.go` with `M`, `WithLocale`, `LocaleFrom`,
      private `localeKey`, and a YAML loader using `//go:embed messages.*.yaml`.
- [ ] 1.2 Implement template rendering with `text/template` (named args via
      `map[string]any`; positional args via `{{.0}}` etc.).
- [ ] 1.3 Add `internal/i18n/messages.en.yaml` and `internal/i18n/messages.es.yaml`
      starter files with `common.warning`, `common.error`, and a self-test ID
      `meta.smoke`.
- [ ] 1.4 Unit tests for `M` covering: lookup hit, missing-ID fallback, missing
      locale-key fallback to English, named-arg substitution, positional-arg
      substitution, unknown-locale defaults to English.
- [ ] 1.5 Mark `i18n.T` as `// Deprecated:` in a doc comment but keep it
      functional (delegates to `M` only when both messages happen to be present
      in the catalog; otherwise returns the inline pair).

## Phase 2: Parity + lint guardrails

- [ ] 2.1 `internal/i18n/parity_test.go` — load every `messages.*.yaml`,
      compute the union of keys, fail listing missing keys per locale.
- [ ] 2.2 `internal/assets/parity_test.go` — list `templates/` and
      `templates_es/` directories, fail on file-set diff.
- [ ] 2.3 `internal/i18n/lint/lint_test.go` — Go AST walker over
      `internal/command/...`, `internal/app/...`, `internal/cli/...`,
      `internal/render/...`. Flag `fmt.{F,}Print{ln,f}?` calls whose first
      string-literal argument is ≥ 8 chars and contains a space, unless the
      line carries `//pacto:i18n-ignore` or the file is allow-listed.
- [ ] 2.4 Seed the lint allow-list with the JSON encoder output paths and any
      genuinely machine-readable diagnostic prints discovered while running it.
- [ ] 2.5 Run the lint test in advisory mode (logs only, doesn't fail) until
      Phase 4 lands; flip to hard-fail in 4.4.

## Phase 3: Wire locale through context at the CLI entry point

- [ ] 3.1 In `internal/cli/root_cmd.go`'s `PersistentPreRunE`, resolve the
      effective locale via the existing `cmdutil.EffectiveLanguage` and call
      `cmd.SetContext(i18n.WithLocale(cmd.Context(), lang))`.
- [ ] 3.2 Add a regression test for `cli/root_cmd.go` confirming
      `cmd.Context()` carries the right locale after `--lang es`.
- [ ] 3.3 Update `cmdutil.EffectiveLanguage` callers in command-tree code to
      prefer `i18n.LocaleFrom(ctx)` where a `ctx` is already in scope; keep the
      file-walk fallback for entry points that don't have one yet.

## Phase 4: Migrate package-by-package to `i18n.M`

Each sub-task lands as its own commit. For each package: extract every
`i18n.T(lang, en, es)` pair into the catalog, replace the call with
`i18n.M(ctx, "<id>", ...)`, drop `lang` from the function signature where it
becomes unused, run `go test ./...`, and run the lint test for the package.

- [ ] 4.1 `internal/cli/help.go` (13 sites — `cli.help.*`).
- [ ] 4.2 `internal/cli/root_cmd.go` and `internal/cli/guardrails.go`
      (untranslated error paths — `cli.guardrails.*`, `cli.flag.invalid_lang`).
- [ ] 4.3 `internal/render/report.go` (12 sites — `render.report.*`).
- [ ] 4.4 `internal/command/install/install.go` and `helpers.go`
      (18 sites + the `"Running "+opts.Command` bug — `install.*`). Fix the
      Spanish branch to use a templated `"install.running"`.
- [ ] 4.5 `internal/command/initcmd/init.go` and `helpers.go`
      (17 sites — `init.*`).
- [ ] 4.6 `internal/command/explore/explore.go` and `helpers.go` plus
      `internal/app/explore/explore.go` (~21 sites — `explore.*`).
- [ ] 4.7 `internal/command/execmd/exec.go` and `helpers.go` plus
      `internal/app/execplan/execplan.go` (~7 sites — `exec.*`).
- [ ] 4.8 `internal/command/doctor/doctor.go` (`doctor.*`).
- [ ] 4.9 `internal/command/newcmd/new.go` and `internal/app/newplan/newplan.go`
      (~10 sites — `new.*`).
- [ ] 4.10 `internal/command/movecmd/move.go` (untranslated, `move.*`).
- [ ] 4.11 `internal/command/normalize/normalize.go` (untranslated,
      `normalize.*`).
- [ ] 4.12 `internal/tui/init/wizard.go` and `internal/tui/status/model.go`
      (TUI placeholders and labels — `tui.init.*`, `tui.status.*`).
- [ ] 4.13 Flip the lint test from advisory to hard-failing.
- [ ] 4.14 Delete the three duplicate wrappers: `cmdutil.Tr`, `cli.tr`,
      `execplan.trLang`. Update remaining call sites to `i18n.M`.

## Phase 5: Plan-declared language

- [ ] 5.1 Teach `internal/planfmt` to recognise `Language: <code>` in the
      Metadata block; expose it on the parsed plan struct.
- [ ] 5.2 Update `internal/app/execplan/execplan.go` to use the declared
      language from the plan struct; delete `detectPlanDocLanguage` and the
      bilingual phrase list.
- [ ] 5.3 Update `internal/render/report.go` and the status TUI to call
      `i18n.WithLocale(ctx, plan.Language)` per-plan when rendering.
- [ ] 5.4 Update `pacto new` to write `Language: <effective-locale>` into the
      generated metadata blocks.
- [ ] 5.5 Update `internal/assets/templates/` and `internal/assets/templates_es/`
      so generated PACTO/AGENTS/README documents include `Language: en`/`es`
      in their Metadata sections (parity test still passes).
- [ ] 5.6 Refactor `internal/planfmt/planfmt.go::canonicalAliases` to be built
      from a per-locale registry (one map of canonical→synonyms per locale)
      rather than a single bilingual flat map.

## Phase 6: Remove the deprecated shim

- [ ] 6.1 Verify zero remaining call sites of `i18n.T` via
      `rg -n 'i18n\.T\(' internal/`.
- [ ] 6.2 Delete `i18n.T` from `internal/i18n/i18n.go`.
- [ ] 6.3 Update `CHANGELOG.md` under `## Unreleased`.
- [ ] 6.4 Tag a release per the `pacto-dev-publish` skill.

## Evidence

- 2026-04-30 audit `internal/i18n/i18n.go` — confirmed single 40-line helper
- 2026-04-30 audit `rg -c 'i18n\.T\(' internal/` — 95 call sites across 18 files
- 2026-04-30 audit `rg 'fmt\.F?Print' ... | wc -l` — 67 untranslated user-facing prints
- 2026-04-30 audit `internal/app/execplan/execplan.go:495` — phrase-counting heuristic
- 2026-04-30 audit `internal/command/install/install.go:75` — `"Running "+opts.Command` mix bug

## Blockers

- <YYYY-MM-DD HH:MM> <blocker>

## Next Steps

1. Begin Phase 1: implement `internal/i18n/catalog.go` and seed the catalog
   files with `common.*` plus a smoke-test ID.
2. Land Phase 2 guardrails in advisory mode so the migration in Phase 4 has
   feedback as it proceeds.
