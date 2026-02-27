package integrations

import (
	"fmt"
	"strings"
)

func Workflows() []WorkflowSpec {
	return []WorkflowSpec{
		{
			WorkflowID: "status",
			CommandID:  "pacto-status",
			Title:      "Pacto Status",
			Summary:    "Verify plan status, blockers, and evidence claims.",
			Command:    "pacto status --format table",
			WhenToUse:  "Use when you need a consolidated status report for plans and claim verification against repo evidence.",
			RequiredInputs: []string{
				"None when auto-discovery can resolve plans root from current directory or parents.",
			},
			OptionalInputs: []string{
				"`--plans-root <path>` to pin plan discovery root.",
				"`--repo-root <path>` to pin evidence verification root.",
				"`--format table|json`, `--fail-on`, `--state`, `--include-archive`.",
				"`--mode compat|strict`, `--config`, `--max-next-actions`, `--max-blockers`, `--verbose`.",
			},
			OutputContract: []string{
				"Produces `table` or `json` report with state summary, blockers, next actions, and verification outcomes.",
				"Verification classifications are `verified`, `partial`, or `unverified`.",
				"Exit code follows `--fail-on` policy for CI automation.",
			},
			ValidationChecklist: []string{
				"Confirm resolved roots are correct for the user's intent.",
				"Confirm report includes expected plans/states.",
				"If CI use case, ensure `--format json` and explicit `--fail-on` are set.",
			},
			FailureModes: []string{
				"Root resolution failure when no valid plans root is discoverable.",
				"Invalid config/flags or unsupported flag values.",
				"Partial verification due to missing or stale repository evidence.",
			},
			FallbackAction: "Ask for explicit `--plans-root` and `--repo-root` when auto-discovery fails.",
			Implemented:    true,
		},
		{
			WorkflowID: "new",
			CommandID:  "pacto-new",
			Title:      "Pacto New",
			Summary:    "Create a new plan scaffold and update plan index metadata.",
			Command:    "pacto new to-implement my-plan-slug",
			WhenToUse:  "Use when a new plan slice must be created in one of the canonical states.",
			RequiredInputs: []string{
				"`<state>` in `current|to-implement|done|outdated`.",
				"`<slug>` matching `[a-z0-9][a-z0-9-]*`.",
			},
			OptionalInputs: []string{
				"`--title`, `--owner` for richer metadata.",
				"`--root <path>` for explicit plan root.",
				"`--allow-minimal-root` to bootstrap missing root files.",
			},
			OutputContract: []string{
				"Creates `<state>/<slug>/README.md` and `PLAN_<TOPIC>_<YYYY-MM-DD>.md`.",
				"Updates root `README.md` counts, section links, and last update date.",
				"Prints created paths and updated index path.",
			},
			ValidationChecklist: []string{
				"Verify state and slug validity before execution.",
				"Confirm plan directory does not already exist.",
				"Confirm index update succeeded in root `README.md`.",
			},
			FailureModes: []string{
				"Invalid state or invalid slug format.",
				"Invalid root (missing canonical files/folders) when minimal root is not allowed.",
				"Plan already exists for the same state/slug.",
			},
			FallbackAction: "If root validation fails, retry with explicit `--root` or `--allow-minimal-root` when appropriate.",
			Implemented:    true,
		},
		{
			WorkflowID: "explore",
			CommandID:  "pacto-explore",
			Title:      "Pacto Explore",
			Summary:    "Capture and revisit ideas without implementing them.",
			Command:    "pacto explore <slug> [--title <title>] [--note <note>]",
			WhenToUse:  "Use for ideation and notes when work is not ready for a formal plan slice.",
			RequiredInputs: []string{
				"`<slug>` for create/update flows, or one of `--list` / `--show <slug>`.",
			},
			OptionalInputs: []string{
				"`--title` for the initial idea heading.",
				"`--note` to append timestamped exploration notes.",
				"`--root <path>` to target a specific project root.",
			},
			OutputContract: []string{
				"Stores ideas in `.pacto/ideas/<slug>/README.md`.",
				"Tracks `Created At` and `Updated At` timestamps.",
				"Returns list/show output for discovery and review.",
			},
			ValidationChecklist: []string{
				"Confirm idea slug resolves to intended workspace.",
				"Confirm notes append with timestamp and preserve prior history.",
				"Use `--show` to verify resulting content when needed.",
			},
			FailureModes: []string{
				"Missing slug for create/show usage.",
				"Invalid flag combinations such as conflicting mode flags.",
				"Permission/path issues creating `.pacto/ideas` files.",
			},
			FallbackAction: "If idea lookup fails, run `pacto explore --list` to discover available slugs.",
			Implemented:    true,
		},
		{
			WorkflowID: "init",
			CommandID:  "pacto-init",
			Title:      "Pacto Init",
			Summary:    "Initialize a project-local Pacto workspace.",
			Command:    "pacto init",
			WhenToUse:  "Use once per project to create canonical `.pacto/plans` workspace scaffolding.",
			RequiredInputs: []string{
				"None for current directory initialization.",
			},
			OptionalInputs: []string{
				"`--root <path>` to initialize a specific project root.",
				"`--with-agents` to manage `AGENTS.md` guidance block.",
				"`--force` to overwrite init-managed files.",
			},
			OutputContract: []string{
				"Creates canonical state directories and template files under `.pacto/plans`.",
				"Reports created/updated/skipped items.",
				"Optionally creates or updates managed Pacto block in `AGENTS.md`.",
			},
			ValidationChecklist: []string{
				"Confirm `.pacto/plans/{current,to-implement,done,outdated}` exist.",
				"Confirm core docs (`README.md`, `PACTO.md`, template, slash commands) exist.",
				"When `--with-agents`, confirm managed block markers are present in `AGENTS.md`.",
			},
			FailureModes: []string{
				"State path already exists as non-directory.",
				"Filesystem permission errors while creating workspace.",
				"Force-less init skips existing managed files by design.",
			},
			FallbackAction: "If files are skipped unexpectedly, rerun with `--force` only for init-managed artifacts.",
			Implemented:    true,
		},
		{
			WorkflowID: "install",
			CommandID:  "pacto-install",
			Title:      "Pacto Install",
			Summary:    "Install managed Pacto skills and command prompts for supported tools.",
			Command:    "pacto install [--tools <all|none|csv>] [--force]",
			WhenToUse:  "Use to bootstrap Pacto-generated skills/prompts for compatible AI tools.",
			RequiredInputs: []string{
				"Either detectable tool directories or explicit `--tools` selection.",
			},
			OptionalInputs: []string{
				"`--tools <all|none|csv>` for explicit selection.",
				"`--force` to overwrite unmanaged existing files.",
			},
			OutputContract: []string{
				"Generates managed skill and command files per workflow and selected tool.",
				"Returns per-file outcome summary: created, updated, skipped, failed.",
			},
			ValidationChecklist: []string{
				"Confirm selected/detected tools match user intent.",
				"Check warnings for unmanaged file skips.",
				"Confirm generated artifacts are wrapped with managed markers.",
			},
			FailureModes: []string{
				"No tools detected when `--tools` is omitted.",
				"Invalid `--tools` argument values.",
				"Filesystem write failures for target tool paths.",
			},
			FallbackAction: "If detection fails, rerun with explicit `--tools` list.",
			Implemented:    true,
		},
		{
			WorkflowID: "update",
			CommandID:  "pacto-update",
			Title:      "Pacto Update",
			Summary:    "Refresh previously installed managed Pacto artifacts.",
			Command:    "pacto update [--tools <all|none|csv>] [--force]",
			WhenToUse:  "Use after upgrading Pacto to refresh managed blocks in generated artifacts.",
			RequiredInputs: []string{
				"Previously installed tool artifacts, or explicit `--tools` target.",
			},
			OptionalInputs: []string{
				"`--tools <all|none|csv>` for explicit tool selection.",
				"`--force` to overwrite unmanaged files when needed.",
			},
			OutputContract: []string{
				"Updates managed blocks in skill and command artifacts in place.",
				"Reports created/updated/skipped/failed counts.",
				"Preserves unmanaged files unless `--force` is set.",
			},
			ValidationChecklist: []string{
				"Confirm managed marker replacement happened for existing files.",
				"Review skipped unmanaged warnings and decide if force is appropriate.",
				"Spot-check one skill and one command artifact for expected template updates.",
			},
			FailureModes: []string{
				"Unsupported or invalid tool selection.",
				"Unmanaged files skipped without `--force`.",
				"Filesystem write errors.",
			},
			FallbackAction: "Use `--force` only when intentional overwrite of unmanaged files is acceptable.",
			Implemented:    true,
		},
		{
			WorkflowID: "exec",
			CommandID:  "pacto-exec",
			Title:      "Pacto Exec",
			Summary:    "Execute plan slices and capture deltas/evidence.",
			Command:    "pacto exec <path-to-plan-md>",
			WhenToUse:  "Reserved workflow for guided plan execution when command implementation becomes available.",
			RequiredInputs: []string{
				"`<path-to-plan-md>` of the plan to execute (future behavior).",
			},
			OptionalInputs: []string{
				"None currently, command is planned.",
			},
			OutputContract: []string{
				"Current expected output is a planned/not-implemented message.",
				"Do not represent this workflow as executing plan slices today.",
			},
			ValidationChecklist: []string{
				"Explicitly communicate that `pacto exec` is planned and not implemented.",
				"Offer a fallback workflow (`status`, `new`, or `explore`) aligned with user intent.",
			},
			FailureModes: []string{
				"Users may expect execution; command currently cannot execute slices.",
			},
			FallbackAction: "Use `pacto status` for verification, `pacto new` for scaffolding, and `pacto explore` for ideation until exec ships.",
			Implemented:    false,
		},
	}
}

func RenderSkill(toolID string, wf WorkflowSpec) string {
	return fmt.Sprintf(`# %s Skill

Use this skill as an agent contract for the %s workflow in Pacto projects.

## Objective

%s

## When To Use

%s

## Input Contract

### Required Inputs
%s

### Optional Inputs
%s

## Execution Contract

- Tool target: %s
- Recommended command: %s

## Output Contract
%s

## Validation Checklist
%s

## Failure Modes and Handling
%s

## Implementation Status

- Status: **%s**
- Fallback: %s
`, wf.Title, wf.WorkflowID, wf.Summary, wf.WhenToUse, asBullets(wf.RequiredInputs), asBullets(wf.OptionalInputs), toolID, wf.Command, asBullets(wf.OutputContract), asBullets(wf.ValidationChecklist), asBullets(wf.FailureModes), implementationStatusLabel(wf.Implemented), wf.FallbackAction)
}

func RenderCommand(toolID string, wf WorkflowSpec) string {
	return fmt.Sprintf(`# %s

Agent contract for %s.

## Objective

%s

## Input Contract

### Required Inputs
%s

### Optional Inputs
%s

## Execution Contract

- Tool target: %s
- Recommended command: %s

## Output Contract
%s

## Validation Checklist
%s

## Failure Modes and Handling
%s

## Implementation Status

- Status: **%s**
- Fallback: %s
`, wf.CommandID, wf.WorkflowID, wf.Summary, asBullets(wf.RequiredInputs), asBullets(wf.OptionalInputs), toolID, wf.Command, asBullets(wf.OutputContract), asBullets(wf.ValidationChecklist), asBullets(wf.FailureModes), implementationStatusLabel(wf.Implemented), wf.FallbackAction)
}

func asBullets(items []string) string {
	if len(items) == 0 {
		return "- None."
	}
	var b strings.Builder
	for _, item := range items {
		t := strings.TrimSpace(item)
		if t == "" {
			continue
		}
		b.WriteString("- ")
		b.WriteString(t)
		b.WriteString("\n")
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "- None."
	}
	return out
}

func implementationStatusLabel(implemented bool) string {
	if implemented {
		return "Implemented"
	}
	return "Planned (Not Implemented)"
}
