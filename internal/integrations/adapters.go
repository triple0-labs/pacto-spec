package integrations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type localAdapter struct {
	toolID      string
	skillsDir   string
	commandsDir string
}

func (a localAdapter) ToolID() string { return a.toolID }

func (a localAdapter) SkillFilePath(projectRoot, workflowID string) (string, error) {
	if workflowID == "" {
		return "", fmt.Errorf("workflow ID is required")
	}
	return filepath.Join(projectRoot, a.skillsDir, "skills", "pacto-"+workflowID, "SKILL.md"), nil
}

func (a localAdapter) CommandFilePath(projectRoot, commandID string) (string, error) {
	if commandID == "" {
		return "", fmt.Errorf("command ID is required")
	}
	return filepath.Join(projectRoot, a.commandsDir, commandID+".md"), nil
}

type codexAdapter struct{}

func (a codexAdapter) ToolID() string { return "codex" }

func (a codexAdapter) SkillFilePath(projectRoot, workflowID string) (string, error) {
	if workflowID == "" {
		return "", fmt.Errorf("workflow ID is required")
	}
	return filepath.Join(projectRoot, ".codex", "skills", "pacto-"+workflowID, "SKILL.md"), nil
}

func (a codexAdapter) CommandFilePath(_ string, commandID string) (string, error) {
	if commandID == "" {
		return "", fmt.Errorf("command ID is required")
	}
	home := strings.TrimSpace(os.Getenv("CODEX_HOME"))
	if home == "" {
		u, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		home = filepath.Join(u, ".codex")
	}
	return filepath.Join(home, "prompts", commandID+".md"), nil
}

func adapters() map[string]Adapter {
	return map[string]Adapter{
		"codex":    codexAdapter{},
		"cursor":   localAdapter{toolID: "cursor", skillsDir: ".cursor", commandsDir: ".cursor/commands"},
		"claude":   localAdapter{toolID: "claude", skillsDir: ".claude", commandsDir: ".claude/commands"},
		"opencode": localAdapter{toolID: "opencode", skillsDir: ".opencode", commandsDir: ".opencode/commands"},
	}
}

func GetAdapter(toolID string) (Adapter, bool) {
	a, ok := adapters()[toolID]
	return a, ok
}
