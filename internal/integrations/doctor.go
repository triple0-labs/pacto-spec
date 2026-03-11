package integrations

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pacto/internal/plugins"
)

func AnalyzeDrift(projectRoot string, tools []string) ([]DriftRecord, error) {
	root := strings.TrimSpace(projectRoot)
	if root == "" {
		return nil, fmt.Errorf("project root is required")
	}
	root = filepath.Clean(root)
	activePlugins, _ := plugins.LoadActive(root)

	records := make([]DriftRecord, 0)
	for _, toolID := range dedupe(tools) {
		adapter, ok := GetAdapter(toolID)
		if !ok {
			records = append(records, DriftRecord{
				Tool:           toolID,
				Status:         DriftMetaMismatch,
				Reason:         "unsupported tool",
				RecommendedFix: "",
			})
			continue
		}

		for _, wf := range Workflows() {
			sections := collectPluginSections(activePlugins, toolID, wf.WorkflowID)
			skillPath, err := adapter.SkillFilePath(root, wf.WorkflowID)
			if err == nil {
				skillBody := RenderSkill(toolID, wf, sections...)
				skillMeta := ManagedMetadata{
					Artifact:    fmt.Sprintf("%s/skill/pacto-%s", toolID, wf.WorkflowID),
					Workflow:    wf.WorkflowID,
					Contract:    ContractVersion,
					TemplateSHA: TemplateChecksum(skillBody),
					GeneratedBy: "pacto",
				}
				records = append(records, inspectManagedArtifact(toolID, "skill", wf.WorkflowID, skillPath, skillBody, skillMeta))
			}

			commandPath, err := adapter.CommandFilePath(root, wf.CommandID)
			if err == nil {
				commandBody := RenderCommand(toolID, wf, sections...)
				commandMeta := ManagedMetadata{
					Artifact:    fmt.Sprintf("%s/command/%s", toolID, filepath.Base(commandPath)),
					Workflow:    wf.WorkflowID,
					Contract:    ContractVersion,
					TemplateSHA: TemplateChecksum(commandBody),
					GeneratedBy: "pacto",
				}
				records = append(records, inspectManagedArtifact(toolID, "command", wf.WorkflowID, commandPath, commandBody, commandMeta))
			}
		}

		records = append(records, detectLegacyPatterns(root, toolID)...)
	}

	sort.Slice(records, func(i, j int) bool {
		if records[i].Tool != records[j].Tool {
			return records[i].Tool < records[j].Tool
		}
		if records[i].Path != records[j].Path {
			return records[i].Path < records[j].Path
		}
		if records[i].WorkflowID != records[j].WorkflowID {
			return records[i].WorkflowID < records[j].WorkflowID
		}
		return records[i].Status < records[j].Status
	})
	return records, nil
}

func inspectManagedArtifact(toolID, kind, workflowID, path, expectedBody string, expectedMeta ManagedMetadata) DriftRecord {
	rec := DriftRecord{
		Tool:           toolID,
		Kind:           kind,
		WorkflowID:     workflowID,
		Path:           path,
		Status:         DriftOK,
		Reason:         "managed artifact is current",
		RecommendedFix: fmt.Sprintf("pacto update --artifacts --tools %s", toolID),
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			rec.Status = DriftMissing
			rec.Reason = "managed artifact is missing"
			return rec
		}
		rec.Status = DriftMetaMismatch
		rec.Reason = "read failed: " + err.Error()
		return rec
	}
	content := string(b)
	start := strings.Index(content, ManagedStart)
	end := strings.Index(content, ManagedEnd)
	if start < 0 || end <= start {
		rec.Status = DriftUnmanaged
		rec.Reason = "file exists without managed markers"
		rec.RecommendedFix = fmt.Sprintf("pacto update --artifacts --tools %s --force", toolID)
		return rec
	}
	meta, hasMeta, err := FindManagedMeta(content)
	if err != nil {
		rec.Status = DriftMetaMismatch
		rec.Reason = "invalid managed metadata: " + err.Error()
		return rec
	}
	if !hasMeta {
		rec.Status = DriftLegacyManaged
		rec.Reason = "managed markers found without metadata"
		return rec
	}
	if diffs := MetadataDiff(meta, expectedMeta); len(diffs) > 0 {
		rec.Status = DriftMetaMismatch
		rec.Reason = "metadata mismatch: " + strings.Join(diffs, ", ")
		return rec
	}
	body := extractManagedBody(content)
	if strings.TrimSpace(body) != strings.TrimSpace(expectedBody) {
		rec.Status = DriftStale
		rec.Reason = "managed body differs from current template"
		return rec
	}
	return rec
}

func extractManagedBody(content string) string {
	start := strings.Index(content, ManagedStart)
	end := strings.Index(content, ManagedEnd)
	if start < 0 || end <= start {
		return ""
	}
	segment := content[start+len(ManagedStart) : end]
	lines := strings.Split(segment, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			filtered = append(filtered, line)
			continue
		}
		if strings.HasPrefix(trimmed, ManagedMeta) {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.TrimSpace(strings.Join(filtered, "\n"))
}

func detectLegacyPatterns(projectRoot, toolID string) []DriftRecord {
	base, ok := toolBaseDir(projectRoot, toolID)
	if !ok || !dirExists(base) {
		return nil
	}
	records := make([]DriftRecord, 0)

	legacyCommands, _ := filepath.Glob(filepath.Join(base, "commands", "pa-*.md"))
	for _, p := range legacyCommands {
		records = append(records, DriftRecord{
			Tool:           toolID,
			Kind:           "command",
			Path:           p,
			Status:         DriftLegacyPattern,
			Reason:         "legacy pa-* command file detected",
			RecommendedFix: "remove or replace with pacto-* command artifacts",
		})
	}
	pactoPlanPath := filepath.Join(base, "commands", "pacto-plan.md")
	if _, err := os.Stat(pactoPlanPath); err == nil {
		records = append(records, DriftRecord{
			Tool:           toolID,
			Kind:           "command",
			Path:           pactoPlanPath,
			Status:         DriftLegacyPattern,
			Reason:         "legacy pacto-plan compatibility wrapper detected",
			RecommendedFix: "remove if no longer needed; prefer pacto-* command files",
		})
	}

	patternChecks := []struct {
		label string
		find  string
	}{
		{label: "deprecated --plans-root guidance", find: "--plans-root"},
		{label: "legacy pacto/ root guidance", find: "`pacto/`"},
		{label: "obsolete exec-not-implemented guidance", find: "planned but not implemented"},
	}
	for _, p := range legacyArtifactCandidates(base) {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		s := string(b)
		for _, check := range patternChecks {
			if strings.Contains(s, check.find) {
				records = append(records, DriftRecord{
					Tool:           toolID,
					Path:           p,
					Status:         DriftLegacyPattern,
					Reason:         check.label,
					RecommendedFix: fmt.Sprintf("pacto update --artifacts --tools %s", toolID),
				})
			}
		}
	}
	return dedupeLegacyRecords(records)
}

func legacyArtifactCandidates(base string) []string {
	paths := make([]string, 0)
	cmdFiles, _ := filepath.Glob(filepath.Join(base, "commands", "pacto-*.md"))
	paths = append(paths, cmdFiles...)
	skillFiles, _ := filepath.Glob(filepath.Join(base, "skills", "pacto-*", "SKILL.md"))
	paths = append(paths, skillFiles...)
	return paths
}

func dedupeLegacyRecords(records []DriftRecord) []DriftRecord {
	out := make([]DriftRecord, 0, len(records))
	seen := map[string]bool{}
	for _, r := range records {
		key := strings.Join([]string{r.Tool, r.Path, string(r.Status), r.Reason}, "|")
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, r)
	}
	return out
}

func toolBaseDir(projectRoot, toolID string) (string, bool) {
	switch toolID {
	case "cursor":
		return filepath.Join(projectRoot, ".cursor"), true
	case "claude":
		return filepath.Join(projectRoot, ".claude"), true
	case "opencode":
		return filepath.Join(projectRoot, ".opencode"), true
	case "codex":
		return filepath.Join(projectRoot, ".codex"), true
	default:
		return "", false
	}
}
