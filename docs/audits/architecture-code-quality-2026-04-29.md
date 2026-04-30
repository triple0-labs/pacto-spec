# Architecture and Code Quality Audit (2026-04-29)

## Scope

Audit target: current `main` of the `pacto` repository at commit
`dd9713f` (post-Wave 3 use-case extraction).

Reference baseline: previous audit
[architecture-code-quality-2026-02-28.md](architecture-code-quality-2026-02-28.md).

Focus: validate the recent DDD restructuring (`internal/domain`,
`internal/app/*`, `internal/command/*`) and assess overall code quality.

## Methodology

1. `go test ./...` — full test suite.
2. `go test ./... -cover` — package-level coverage.
3. `go vet ./...` — standard static analysis.
4. `golangci-lint run ./...` (v2, config: [.golangci.yml](../../.golangci.yml)) —
   bug-focused linter set (errcheck, govet, ineffassign, staticcheck,
   unused, bodyclose, copyloopvar, unconvert) plus gofmt/goimports.
5. Import-graph inspection across `internal/{domain,app,command,cli}` to
   verify layering.
6. Per-file size and per-function length sampling on hotspots.
7. Review of recent `git log` (Waves 1–3) to map the DDD intent against
   the resulting structure.

## Repository Baseline

1. Tooling: Go 1.23, Cobra, Bubble Tea/Lipgloss for TUI.
2. Single binary entrypoint: [cmd/pacto/main.go](../../cmd/pacto/main.go).
   The legacy `cmd/pacto-engine/` alias was dropped (commit `f7bdcfd`).
3. Layered package map (post-DDD):
   - **Domain** (pure, std-lib only):
     [internal/domain/plan](../../internal/domain/plan/plan.go),
     [internal/domain/claim](../../internal/domain/claim/claim.go),
     [internal/domain/report](../../internal/domain/report/report.go).
   - **Application / use cases** (no I/O on stdout, no Cobra):
     `internal/app/{doctor,execplan,explore,initws,installtools,move,
     newplan,normalizeplans,status}`.
   - **Supporting domain services**:
     `internal/{discovery,parser,claims,verify,analyze,planfmt,
     integrations,plugins,onboarding,workspace,context,config}`.
   - **Presentation / CLI wiring**:
     `internal/cli`, `internal/command/*`, `internal/ui`,
     `internal/tui/*`, `internal/render`, `internal/i18n`,
     `internal/cmdutil`.

## Test, Lint, and Coverage Snapshot

1. `go test ./...` — **all packages pass**, no flakes observed.
2. `go vet ./...` — clean.
3. `golangci-lint run ./...` — clean (zero findings under the configured
   linter set).
4. Coverage by package (selected; full list below):

   | Package | Coverage |
   | --- | --- |
   | internal/parser | 87.6% |
   | internal/app/status | 84.1% |
   | internal/app/newplan | 84.1% |
   | internal/app/explore | 85.1% |
   | internal/context | 84.0% |
   | internal/command/execmd | 82.8% |
   | internal/claims | 80.8% |
   | internal/plugins/builtin | 78.0% |
   | internal/app/initws | 78.4% |
   | internal/app/execplan | 77.8% |
   | internal/app/normalizeplans | 77.8% |
   | internal/app/move | 77.6% |
   | internal/app/installtools | 77.1% |
   | internal/integrations | 76.2% |
   | internal/analyze | 74.6% |
   | internal/command/status | 73.9% |
   | internal/command/newcmd | 70.0% |
   | internal/plugins | 69.7% |
   | internal/app/doctor | 69.0% |
   | internal/command/doctor | 68.8% |
   | internal/command/install | 68.4% |
   | internal/command/normalize | 65.2% |
   | internal/command/explore | 64.4% |
   | internal/verify | 63.3% |
   | internal/render | 62.7% |
   | internal/planfmt | 60.5% |
   | internal/discovery | 57.1% |
   | internal/command/plugincmd | 57.4% |
   | internal/cli | 52.1% |
   | internal/command/initcmd | 51.3% |
   | internal/config | 66.3% |
   | internal/workspace | 27.4% |
   | internal/onboarding | 16.2% |
   | internal/cmdutil | 13.1% |
   | internal/{ui,tui/init,tui/status,exitcode,i18n,markdown,yamlutil,assets,testutil,domain/*} | 0% (no tests) |

## DDD Restructuring — Findings

### 1. Domain layer is genuinely pure

- [internal/domain/plan/plan.go](../../internal/domain/plan/plan.go),
  [internal/domain/claim/claim.go](../../internal/domain/claim/claim.go),
  [internal/domain/report/report.go](../../internal/domain/report/report.go)
  declare value types with **no imports beyond stdlib** (and
  `report` → `claim`). Confirmed via import inspection.
- All consumers (`analyze`, `parser`, `verify`, `claims`, `discovery`,
  `render`, `tui/status`, `app/status`, `command/status`, `exitcode`)
  use `internal/domain/*`. No surviving references to legacy
  `internal/model` or `internal/report` packages remain.

### 2. Application layer respects boundaries

- `internal/app/*` packages contain **zero `fmt.Print*`/`fmt.Fprint*`
  calls** to stdout/stderr. Output is returned via typed `Result`
  structs to be rendered by `internal/command/*`. Verified by grep
  across the tree.
- **No cross-app imports**: no `internal/app/*` package imports another
  `internal/app/*` package. Each use case is independently testable.
- App packages depend only downward (domain + supporting services).
  Sample import graph:

  | App package | Downward deps |
  | --- | --- |
  | `app/status` | `analyze`, `claims`, `discovery`, `parser`, `verify`, `context`, `domain/{claim,report}` |
  | `app/execplan` | `i18n`, `workspace` |
  | `app/doctor` | `integrations`, `workspace` |
  | `app/installtools` | `integrations`, `workspace` |
  | `app/initws` | `assets`, `i18n`, `integrations`, `onboarding`, `context` |
  | `app/normalizeplans` | `discovery`, `planfmt`, `workspace` |
  | `app/newplan` | `cmdutil`, `i18n`, `workspace` |
  | `app/explore` | `i18n`, `markdown` |
  | `app/move` | `i18n` |

### 3. Command layer wiring is clean (with one caveat)

- [internal/cli/root_cmd.go](../../internal/cli/root_cmd.go) is the only
  composition root that fans out to all `internal/command/*` packages.
- Cross-command imports exist where commands legitimately chain
  bootstrap behavior (e.g. `command/execmd` → `command/initcmd`,
  `command/newcmd`; `command/movecmd` → `command/initcmd`,
  `command/newcmd`; `command/doctor` → `command/install`;
  `command/plugincmd` → `command/initcmd`). These are intentional UX
  shortcuts (run `init`/`new` lazily before the user's actual command)
  but they create a small dependency web inside `internal/command`.
- The single reverse import — `command/plugincmd` → `internal/cli` —
  is **test-only**
  ([test_helpers_test.go](../../internal/command/plugincmd/test_helpers_test.go))
  and therefore not a runtime layering violation.

## Code Quality — Findings

### 1. Lint and vet are clean baselines

CI now enforces `golangci-lint v2` (commit `d101ae5`). The configured
set is intentionally bug-focused (no stylistic noise). The current
working tree passes with zero findings.

### 2. Hotspots by file size

| LOC | File | Note |
| --- | --- | --- |
| 615 | [internal/parser/parser.go](../../internal/parser/parser.go) | One 127-line `parseStructuredDeltas`; otherwise small helpers. |
| 571 | [internal/cli/root_cmd.go](../../internal/cli/root_cmd.go) | 17 thin Cobra constructors; size is by-design fan-out. |
| 499 | [internal/app/execplan/execplan.go](../../internal/app/execplan/execplan.go) | One 98-line `applyExecTaskUpdate`, 78-line `Apply`, 52-line `upsertPlanLastModified`. Splittable. |
| 421 | [internal/tui/init/wizard.go](../../internal/tui/init/wizard.go) | Bubble Tea state machine; 0% covered (TUI, untested). |
| 421 | [internal/integrations/templates.go](../../internal/integrations/templates.go) | Mostly template literals. |
| 359 | [internal/command/install/install.go](../../internal/command/install/install.go) | CLI rendering + dispatch. |

The longest functions worth attention:

- `parser.parseStructuredDeltas` (127 lines): a state machine over plan
  delta blocks. Candidate for extraction into smaller field-handlers.
- `execplan.applyExecTaskUpdate` (98 lines) and
  `execplan.upsertPlanLastModified` (52 lines): regex-heavy text
  surgery. Both are well-tested but would benefit from named
  sub-helpers to make the control flow grep-friendly.

### 3. Coverage gaps worth tracking

- `internal/cmdutil` (13.1%) and `internal/workspace` (27.4%) are
  shared utility packages used everywhere; raising their floor would
  buy broad regression protection.
- `internal/onboarding` (16.2%) — same concern flagged in the
  2026-02-28 audit; persistence helpers (`persist.go`, ~180 LOC) and
  the `valueOrTODO` template path remain largely uncovered.
- Domain packages (`internal/domain/{plan,claim,report}`) have no
  tests. They are pure data carriers, but adding minimal JSON
  round-trip tests on `report.StatusReport` and `claim.Result` would
  freeze the wire format that the `pacto status --format=json` contract
  depends on.
- TUI packages (`internal/tui/{init,status}`) and `internal/ui` have no
  tests. This is acceptable given the Bubble Tea harness cost, but the
  411-line `tui/init/wizard.go` is the largest untested file in the
  tree.

### 4. Minor smells (non-blocking)

- **i18n in the application layer.** `app/execplan`, `app/explore`,
  `app/move`, `app/newplan`, and `app/initws` import
  `internal/i18n`. Strictly, language selection is a presentation
  concern; today it leaks into the use cases because plan-document
  language is autodetected and influences how text is appended/
  upserted. This is defensible (the rule "match the document's
  language" belongs in the use case) but worth documenting as an
  explicit policy in `docs/architecture.md`.
- **`internal/exitcode`** depends on `internal/domain/report`. Exit
  codes are presentation-policy. The current shape is fine because
  `exitcode` only translates a domain `StatusReport` into a number,
  but it should not grow CLI-specific concerns.
- **Cross-command chaining.** `command/{execmd,movecmd,newcmd,
  plugincmd,initcmd}` cross-import to lazily run `init`/`new` before
  the user's command. Centralizing that "ensure-bootstrapped"
  behavior in a single `internal/cli` helper would flatten the
  dependency graph and make it easier to add new commands without
  knowing the bootstrap chain.
- **Embedded asset panics.**
  [internal/assets/assets.go](../../internal/assets/assets.go) panics
  on missing template lookups. This is the documented contract and is
  excluded from the linter, but it is the only `panic` in the
  non-asset code path — keep it confined there.

## Status of Prior (2026-02-28) Findings

| Finding | Status |
| --- | --- |
| Config clobber risk during init | Still fixed; merge-preserving `WriteConfig` in place. |
| Coverage gaps in `claims`/`discovery`/`report`/`onboarding` | `claims` at 80.8%, `discovery` at 57.1%, `render`/`report` at 62.7%, `onboarding` still at 16.2% (regressed slightly from 17.1% as new code was added). |
| Single-binary entrypoint | Done — `cmd/pacto-engine` removed (`f7bdcfd`). |
| Use-case/CLI separation | Substantially advanced by Waves 1–3; all CLI commands now wrap a thin `internal/app/*` use case. |

## Recommendations

Priority is roughly highest → lowest.

1. **Raise `cmdutil` and `workspace` coverage to ≥60%.** They underpin
   every command; small, focused tests will pay back broadly.
2. **Add JSON-shape tests for `domain/report.StatusReport` and
   `domain/claim.Result`.** Freezes the public contract of
   `pacto status --format=json`.
3. **Document the "i18n in app layer" policy** in
   [docs/architecture.md](../architecture.md) so the leak is
   intentional rather than accidental.
4. **Extract sub-helpers from `parser.parseStructuredDeltas` and
   `execplan.applyExecTaskUpdate`.** Lowers cognitive load on the two
   hottest code paths in the plan pipeline.
5. **Centralize the "ensure init/new before command" bootstrap** in
   `internal/cli` to remove cross-`internal/command` imports.
6. **Add a smoke test for `internal/tui/init/wizard.go`** that drives a
   handful of `Update` transitions; even 20–30% coverage on the
   wizard would catch the most common regressions.
7. **Add an `ARCHITECTURE.md`-level diagram** of the
   `domain → app → command → cli` flow so the DDD intent is visible
   without grepping imports. (The audits document it; the architecture
   doc should too.)

## Verdict

The Wave 1–3 DDD work landed cleanly:

- Domain is pure, application is I/O-free, presentation is the only
  layer touching Cobra/stdout.
- All tests, `go vet`, and `golangci-lint` are green.
- No layering violations in non-test code.

The remaining work is incremental: coverage on shared utilities,
splitting two long functions, and documenting the few intentional
exceptions (i18n in app layer, panic in `assets`, command chaining).
The codebase is in good shape to absorb future feature work without
further structural refactors.
