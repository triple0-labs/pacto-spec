package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	PluginGuardrailStart = "<!-- pacto:plugin-guardrails:start -->"
	PluginGuardrailEnd   = "<!-- pacto:plugin-guardrails:end -->"
)

func CollectAgentContributions(active []Plugin, toolID, workflowID string) []AgentContribution {
	out := make([]AgentContribution, 0)
	for _, plugin := range active {
		for _, g := range plugin.Manifest.Spec.AgentGuardrails {
			if !matchesTool(toolID, g.Tools) {
				continue
			}
			if !matchesWorkflow(workflowID, g.Workflows) {
				continue
			}
			mdPath := filepath.Clean(filepath.Join(plugin.Dir, g.MarkdownFile))
			b, err := os.ReadFile(mdPath)
			if err != nil {
				continue
			}
			text := strings.TrimSpace(string(b))
			if text == "" {
				continue
			}
			out = append(out, AgentContribution{PluginID: plugin.Manifest.Metadata.ID, GuardrailID: g.ID, Markdown: text})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].PluginID == out[j].PluginID {
			return out[i].GuardrailID < out[j].GuardrailID
		}
		return out[i].PluginID < out[j].PluginID
	})
	return out
}

func matchesTool(tool string, tools []string) bool {
	if len(tools) == 0 {
		return true
	}
	for _, t := range tools {
		if strings.EqualFold(strings.TrimSpace(t), tool) {
			return true
		}
	}
	return false
}

func matchesWorkflow(workflow string, workflows []string) bool {
	if len(workflows) == 0 {
		return true
	}
	for _, w := range workflows {
		if strings.EqualFold(strings.TrimSpace(w), workflow) {
			return true
		}
	}
	return false
}

func RenderPluginGuardrailSection(items []AgentContribution) string {
	if len(items) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n## Plugin Guardrails\n\n")
	b.WriteString(PluginGuardrailStart)
	b.WriteString("\n")
	for _, item := range items {
		b.WriteString(fmt.Sprintf("### %s\n\n", item.FullID()))
		b.WriteString(strings.TrimSpace(item.Markdown))
		b.WriteString("\n\n")
	}
	b.WriteString(PluginGuardrailEnd)
	b.WriteString("\n")
	return b.String()
}
