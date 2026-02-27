package integrations

import "fmt"

func Workflows() []WorkflowSpec {
	return []WorkflowSpec{
		{WorkflowID: "status", CommandID: "pacto-status", Title: "Pacto Status", Summary: "Verify plan status, blockers, and evidence claims.", Command: "pacto status --format table"},
		{WorkflowID: "new", CommandID: "pacto-new", Title: "Pacto New", Summary: "Create a new plan scaffold and update plan index metadata.", Command: "pacto new to-implement my-plan-slug"},
		{WorkflowID: "exec", CommandID: "pacto-exec", Summary: "Execute plan slices and capture deltas/evidence.", Title: "Pacto Exec", Command: "pacto exec <path-to-plan-md>"},
		{WorkflowID: "init", CommandID: "pacto-init", Summary: "Initialize a project-local Pacto workspace.", Title: "Pacto Init", Command: "pacto init"},
	}
}

func RenderSkill(toolID string, wf WorkflowSpec) string {
	return fmt.Sprintf(`# %s Skill

Use this skill to run the %s workflow in projects using Pacto.

## Objective

%s

## Default Command

- %s

## Rules

1. Prefer JSON output for automation (`+"`--format json`"+`).
2. Keep plan evidence current and concrete.
3. Use absolute dates in updates and reports.
`, wf.Title, wf.WorkflowID, wf.Summary, wf.Command)
}

func RenderCommand(toolID string, wf WorkflowSpec) string {
	return fmt.Sprintf(`# %s

Run the Pacto %s workflow.

- Goal: %s
- Recommended command: `+"`%s`"+`
`, wf.CommandID, wf.WorkflowID, wf.Summary, wf.Command)
}
