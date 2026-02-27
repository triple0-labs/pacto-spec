package integrations

func GenerateForTool(projectRoot, toolID string, force bool) []ArtifactResult {
	results := make([]ArtifactResult, 0)
	adapter, ok := GetAdapter(toolID)
	if !ok {
		return []ArtifactResult{{Tool: toolID, Err: errUnsupportedTool(toolID)}}
	}

	for _, wf := range Workflows() {
		skillPath, err := adapter.SkillFilePath(projectRoot, wf.WorkflowID)
		if err != nil {
			results = append(results, ArtifactResult{Tool: toolID, Kind: "skill", WorkflowID: wf.WorkflowID, Err: err})
		} else {
			wr, werr := WriteManaged(skillPath, RenderSkill(toolID, wf), force)
			results = append(results, ArtifactResult{Tool: toolID, Kind: "skill", WorkflowID: wf.WorkflowID, Path: skillPath, Outcome: wr.Outcome, Reason: wr.Reason, Err: werr})
		}

		commandPath, err := adapter.CommandFilePath(projectRoot, wf.CommandID)
		if err != nil {
			results = append(results, ArtifactResult{Tool: toolID, Kind: "command", WorkflowID: wf.WorkflowID, Err: err})
			continue
		}
		wr, werr := WriteManaged(commandPath, RenderCommand(toolID, wf), force)
		results = append(results, ArtifactResult{Tool: toolID, Kind: "command", WorkflowID: wf.WorkflowID, Path: commandPath, Outcome: wr.Outcome, Reason: wr.Reason, Err: werr})
	}

	return results
}

func errUnsupportedTool(toolID string) error {
	return &unsupportedToolError{toolID: toolID}
}

type unsupportedToolError struct{ toolID string }

func (e *unsupportedToolError) Error() string { return "unsupported tool: " + e.toolID }
