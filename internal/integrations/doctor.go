package integrations

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pacto/internal/plugins"
)

func codexHomeDir() string {
	home := strings.TrimSpace(os.Getenv("CODEX_HOME"))
	if home == "" {
		u, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		home = filepath.Join(u, ".codex")
	}
	return home
}

func AnalyzeDrift(projectRoot string, tools []string) ([]DriftRecord, error) {
	root := strings.TrimSpace(projectRoot)
	if root == "" {
		return nil, fmt.Errorf("project root is required")
	}
	root = filepath.Clean(root)
	activePlugins, _ := plugins.LoadActive(root)

	records := make([]DriftRecord, 0)
	seenSkillPath := map[string]bool{}
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
			skillPath, err := adapter.SkillFilePath(root, wf.WorkflowID)
			if err == nil {
				if seenSkillPath[skillPath] {
					continue
				}
				seenSkillPath[skillPath] = true
				skillBody, skillMeta := skillBodyAndMetadata(activePlugins, toolID, wf)
				records = append(records, inspectManagedArtifact(toolID, "skill", wf.WorkflowID, skillPath, skillBody, skillMeta))
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
	if !ok {
		return nil
	}
	records := make([]DriftRecord, 0)

	if toolID == "codex" {
		legacyCodexSkills, _ := filepath.Glob(filepath.Join(projectRoot, ".codex", "skills", "pacto-*", "SKILL.md"))
		for _, p := range legacyCodexSkills {
			records = append(records, DriftRecord{
				Tool:           toolID,
				Kind:           "skill",
				Path:           p,
				Status:         DriftLegacyPattern,
				Reason:         "legacy .codex/skills path detected",
				RecommendedFix: "remove legacy .codex/skills pacto-* entries and use .agents/skills for Codex",
			})
		}
		legacyCodexPrompts, _ := filepath.Glob(filepath.Join(codexHomeDir(), "prompts", "pacto-*.md"))
		for _, p := range legacyCodexPrompts {
			records = append(records, DriftRecord{
				Tool:           toolID,
				Kind:           "command",
				Path:           p,
				Status:         DriftLegacyPattern,
				Reason:         "legacy Codex prompt file (commands are no longer generated)",
				RecommendedFix: "remove if unused; workflows are under .agents/skills as Agent Skills",
			})
		}
	}
	if toolID == "cursor" {
		legacyCursorSkills, _ := filepath.Glob(filepath.Join(projectRoot, ".cursor", "skills", "pacto-*", "SKILL.md"))
		for _, p := range legacyCursorSkills {
			records = append(records, DriftRecord{
				Tool:           toolID,
				Kind:           "skill",
				Path:           p,
				Status:         DriftLegacyPattern,
				Reason:         "legacy .cursor/skills path detected",
				RecommendedFix: "remove legacy .cursor/skills pacto-* entries; project skills live under .agents/skills",
			})
		}
	}
	if !dirExists(base) {
		return dedupeLegacyRecords(records)
	}

	legacyCommands, _ := filepath.Glob(filepath.Join(base, "commands", "pa-*.md"))
	for _, p := range legacyCommands {
		records = append(records, DriftRecord{
			Tool:           toolID,
			Kind:           "command",
			Path:           p,
			Status:         DriftLegacyPattern,
			Reason:         "legacy pa-* command file detected",
			RecommendedFix: "remove if unused; Pacto no longer generates command prompts",
		})
	}
	deprecatedPactoCmds, _ := filepath.Glob(filepath.Join(base, "commands", "pacto-*.md"))
	for _, p := range deprecatedPactoCmds {
		records = append(records, DriftRecord{
			Tool:           toolID,
			Kind:           "command",
			Path:           p,
			Status:         DriftLegacyPattern,
			Reason:         "legacy pacto command file (commands are no longer generated)",
			RecommendedFix: "remove if unused; use skills under .agents/skills (Cursor/Codex) or tool-specific skills directories",
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
			RecommendedFix: "remove if no longer needed; prefer workflows in skills under .agents/skills or tool-specific skill dirs",
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
	skillFiles, _ := filepath.Glob(filepath.Join(base, "skills", "pacto-*", "SKILL.md"))
	return skillFiles
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
		return filepath.Join(projectRoot, ".agents"), true
	default:
		return "", false
	}
}
