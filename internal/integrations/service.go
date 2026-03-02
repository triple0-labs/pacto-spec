package integrations

import (
	"fmt"
	"strings"

	"pacto/internal/plugins"
)

func GenerateForTool(projectRoot, toolID string, force bool) []ArtifactResult {
	results := make([]ArtifactResult, 0)
	adapter, ok := GetAdapter(toolID)
	if !ok {
		return []ArtifactResult{{Tool: toolID, Err: errUnsupportedTool(toolID)}}
	}
	activePlugins, _ := plugins.LoadActive(projectRoot)

	for _, wf := range Workflows() {
		pluginSections := collectPluginSections(activePlugins, toolID, wf.WorkflowID)
		skillPath, err := adapter.SkillFilePath(projectRoot, wf.WorkflowID)
		if err != nil {
			results = append(results, ArtifactResult{Tool: toolID, Kind: "skill", WorkflowID: wf.WorkflowID, Err: err})
		} else {
			wr, werr := WriteManaged(skillPath, RenderSkill(toolID, wf, pluginSections...), force)
			results = append(results, ArtifactResult{Tool: toolID, Kind: "skill", WorkflowID: wf.WorkflowID, Path: skillPath, Outcome: wr.Outcome, Reason: wr.Reason, Err: werr})
		}

		commandPath, err := adapter.CommandFilePath(projectRoot, wf.CommandID)
		if err != nil {
			results = append(results, ArtifactResult{Tool: toolID, Kind: "command", WorkflowID: wf.WorkflowID, Err: err})
			continue
		}
		wr, werr := WriteManaged(commandPath, RenderCommand(toolID, wf, pluginSections...), force)
		results = append(results, ArtifactResult{Tool: toolID, Kind: "command", WorkflowID: wf.WorkflowID, Path: commandPath, Outcome: wr.Outcome, Reason: wr.Reason, Err: werr})
	}

	return results
}

func errUnsupportedTool(toolID string) error {
	return &unsupportedToolError{toolID: toolID}
}

type unsupportedToolError struct{ toolID string }

func (e *unsupportedToolError) Error() string { return "unsupported tool: " + e.toolID }

func collectPluginSections(active []plugins.Plugin, toolID, workflowID string) []string {
	contribs := plugins.CollectAgentContributions(active, toolID, workflowID)
	out := make([]string, 0, len(contribs))
	for _, c := range contribs {
		out = append(out, fmt.Sprintf("### %s\n\n%s", c.FullID(), strings.TrimSpace(c.Markdown)))
	}
	return out
}
