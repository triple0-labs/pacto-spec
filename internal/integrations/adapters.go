package integrations

import (
	"fmt"
	"path/filepath"
)

type localAdapter struct {
	toolID    string
	skillsDir string
}

func (a localAdapter) ToolID() string { return a.toolID }

func (a localAdapter) SkillFilePath(projectRoot, workflowID string) (string, error) {
	if workflowID == "" {
		return "", fmt.Errorf("workflow ID is required")
	}
	return filepath.Join(projectRoot, a.skillsDir, "skills", "pacto-"+workflowID, "SKILL.md"), nil
}

type codexAdapter struct{}

func (a codexAdapter) ToolID() string { return "codex" }

func (a codexAdapter) SkillFilePath(projectRoot, workflowID string) (string, error) {
	if workflowID == "" {
		return "", fmt.Errorf("workflow ID is required")
	}
	return filepath.Join(projectRoot, ".agents", "skills", "pacto-"+workflowID, "SKILL.md"), nil
}

func adapters() map[string]Adapter {
	return map[string]Adapter{
		"codex":    codexAdapter{},
		"cursor":   localAdapter{toolID: "cursor", skillsDir: ".agents"},
		"claude":   localAdapter{toolID: "claude", skillsDir: ".claude"},
		"opencode": localAdapter{toolID: "opencode", skillsDir: ".opencode"},
	}
}

func GetAdapter(toolID string) (Adapter, bool) {
	a, ok := adapters()[toolID]
	return a, ok
}
