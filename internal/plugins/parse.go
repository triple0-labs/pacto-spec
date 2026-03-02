package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func parseManifest(path string) (Manifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}

	var out Manifest
	if err := yaml.Unmarshal(b, &out); err != nil {
		return Manifest{}, err
	}
	normalizeManifest(&out)
	return out, nil
}

func normalizeManifest(m *Manifest) {
	if m == nil {
		return
	}
	m.APIVersion = strings.TrimSpace(m.APIVersion)
	m.Kind = strings.TrimSpace(m.Kind)
	m.Metadata.ID = strings.TrimSpace(m.Metadata.ID)
	m.Metadata.Version = strings.TrimSpace(m.Metadata.Version)
	for i := range m.Spec.CLIGuardrails {
		g := &m.Spec.CLIGuardrails[i]
		g.ID = strings.TrimSpace(g.ID)
		g.Run.Script = strings.TrimSpace(g.Run.Script)
		g.OnFail.Message = strings.TrimSpace(g.OnFail.Message)
		if g.Run.TimeoutMS <= 0 {
			g.Run.TimeoutMS = 5000
		}
	}
	for i := range m.Spec.AgentGuardrails {
		g := &m.Spec.AgentGuardrails[i]
		g.ID = strings.TrimSpace(g.ID)
		g.MarkdownFile = strings.TrimSpace(g.MarkdownFile)
	}
}

func validateManifest(m Manifest, pluginDir string) error {
	if m.APIVersion != ManifestAPIVersion {
		return fmt.Errorf("apiVersion must be %q", ManifestAPIVersion)
	}
	if m.Kind != ManifestKind {
		return fmt.Errorf("kind must be %q", ManifestKind)
	}
	if strings.TrimSpace(m.Metadata.ID) == "" {
		return fmt.Errorf("metadata.id is required")
	}
	for _, g := range m.Spec.CLIGuardrails {
		if strings.TrimSpace(g.ID) == "" {
			return fmt.Errorf("spec.cliGuardrails[].id is required")
		}
		if strings.TrimSpace(g.Run.Script) == "" {
			return fmt.Errorf("spec.cliGuardrails[%s].run.script is required", g.ID)
		}
		scriptPath := filepath.Clean(filepath.Join(pluginDir, g.Run.Script))
		if !strings.HasPrefix(scriptPath, filepath.Clean(pluginDir)+string(os.PathSeparator)) && scriptPath != filepath.Clean(pluginDir) {
			return fmt.Errorf("script path escapes plugin directory: %s", g.Run.Script)
		}
		if _, err := os.Stat(scriptPath); err != nil {
			return fmt.Errorf("script not found: %s", g.Run.Script)
		}
	}
	for _, g := range m.Spec.AgentGuardrails {
		if strings.TrimSpace(g.ID) == "" {
			return fmt.Errorf("spec.agentGuardrails[].id is required")
		}
		if strings.TrimSpace(g.MarkdownFile) == "" {
			return fmt.Errorf("spec.agentGuardrails[%s].markdownFile is required", g.ID)
		}
		mdPath := filepath.Clean(filepath.Join(pluginDir, g.MarkdownFile))
		if !strings.HasPrefix(mdPath, filepath.Clean(pluginDir)+string(os.PathSeparator)) && mdPath != filepath.Clean(pluginDir) {
			return fmt.Errorf("markdown path escapes plugin directory: %s", g.MarkdownFile)
		}
		if _, err := os.Stat(mdPath); err != nil {
			return fmt.Errorf("markdown file not found: %s", g.MarkdownFile)
		}
	}
	return nil
}
