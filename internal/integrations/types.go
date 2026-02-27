package integrations

import "fmt"

type Adapter interface {
	ToolID() string
	SkillFilePath(projectRoot, workflowID string) (string, error)
	CommandFilePath(projectRoot, commandID string) (string, error)
}

type WorkflowSpec struct {
	WorkflowID string
	CommandID  string
	Title      string
	Summary    string
	Command    string
}

type WriteOutcome string

const (
	OutcomeCreated WriteOutcome = "created"
	OutcomeUpdated WriteOutcome = "updated"
	OutcomeSkipped WriteOutcome = "skipped"
)

type WriteResult struct {
	Outcome WriteOutcome
	Reason  string
}

type ArtifactResult struct {
	Tool       string
	Kind       string
	WorkflowID string
	Path       string
	Outcome    WriteOutcome
	Reason     string
	Err        error
}

func SupportedTools() []string {
	return []string{"codex", "cursor", "claude", "opencode"}
}

func ParseToolsArg(raw string) ([]string, error) {
	v := normalize(raw)
	if v == "" {
		return nil, fmt.Errorf("--tools expects a value: all, none, or comma-separated tool IDs")
	}
	if v == "all" {
		return SupportedTools(), nil
	}
	if v == "none" {
		return []string{}, nil
	}

	valid := map[string]bool{}
	for _, t := range SupportedTools() {
		valid[t] = true
	}

	out := make([]string, 0)
	seen := map[string]bool{}
	for _, tok := range splitCSV(v) {
		if tok == "all" || tok == "none" {
			return nil, fmt.Errorf("cannot combine %q with specific tool IDs", tok)
		}
		if !valid[tok] {
			return nil, fmt.Errorf("invalid tool %q (allowed: %s)", tok, joinSupported())
		}
		if !seen[tok] {
			out = append(out, tok)
			seen[tok] = true
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("--tools expects at least one tool ID")
	}
	return out, nil
}
