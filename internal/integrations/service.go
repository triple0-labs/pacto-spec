package integrations

import (
	"fmt"
	"strings"

	"pacto/internal/plugins"
)

func agentsSkillTool(toolID string) bool {
	return toolID == "cursor" || toolID == "codex"
}

func skillPluginSections(active []plugins.Plugin, toolID, workflowID string) []string {
	if agentsSkillTool(toolID) {
		return mergePluginSectionStrings(
			collectPluginSections(active, "cursor", workflowID),
			collectPluginSections(active, "codex", workflowID),
		)
	}
	return collectPluginSections(active, toolID, workflowID)
}

func mergePluginSectionStrings(a, b []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(a)+len(b))
	for _, s := range append(append([]string{}, a...), b...) {
		t := strings.TrimSpace(s)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, s)
	}
	return out
}

func skillArtifactName(toolID, workflowID string) string {
	if agentsSkillTool(toolID) {
		return fmt.Sprintf("agents/skill/pacto-%s", workflowID)
	}
	return fmt.Sprintf("%s/skill/pacto-%s", toolID, workflowID)
}

func skillBodyAndMetadata(active []plugins.Plugin, toolID string, wf WorkflowSpec) (body string, meta ManagedMetadata) {
	sections := skillPluginSections(active, toolID, wf.WorkflowID)
	target := SkillToolTargetForIntegration(toolID)
	body = RenderSkill(target, wf, sections...)
	meta = ManagedMetadata{
		Artifact:    skillArtifactName(toolID, wf.WorkflowID),
		Workflow:    wf.WorkflowID,
		Contract:    ContractVersion,
		TemplateSHA: TemplateChecksum(body),
		GeneratedBy: "pacto",
	}
	return body, meta
}

func GenerateForTool(projectRoot, toolID string, force bool) []ArtifactResult {
	results := make([]ArtifactResult, 0)
	adapter, ok := GetAdapter(toolID)
	if !ok {
		return []ArtifactResult{{Tool: toolID, Err: errUnsupportedTool(toolID)}}
	}
	activePlugins, _ := plugins.LoadActive(projectRoot)

	for _, wf := range Workflows() {
		skillPath, err := adapter.SkillFilePath(projectRoot, wf.WorkflowID)
		if err != nil {
			results = append(results, ArtifactResult{Tool: toolID, Kind: "skill", WorkflowID: wf.WorkflowID, Err: err})
			continue
		}
		body, meta := skillBodyAndMetadata(activePlugins, toolID, wf)
		prefix := skillFrontmatter(toolID, wf)
		wr, werr := WriteManagedWithPrefix(skillPath, body, meta, force, prefix)
		results = append(results, ArtifactResult{Tool: toolID, Kind: "skill", WorkflowID: wf.WorkflowID, Path: skillPath, Outcome: wr.Outcome, Reason: wr.Reason, Err: werr})
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

func skillFrontmatter(toolID string, wf WorkflowSpec) string {
	switch toolID {
	case "codex", "cursor", "claude":
		return fmt.Sprintf("---\nname: pacto-%s\ndescription: Agent contract for the Pacto %s workflow.\n---", wf.WorkflowID, wf.WorkflowID)
	case "opencode":
		return fmt.Sprintf("---\nname: pacto-%s\ndescription: Agent contract for the Pacto %s workflow.\ncompatibility: opencode\n---", wf.WorkflowID, wf.WorkflowID)
	default:
		return ""
	}
}
