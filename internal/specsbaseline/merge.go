package specsbaseline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MergeFile is a planned write to a baseline capability file.
type MergeFile struct {
	Path     string // absolute path to .pacto/specs/<slug>/spec.md
	Slug     string
	Content  string
	Existed  bool
	Mutation []string // human-readable summary of changes (e.g. "added 2 requirements")
}

// PlanMerge computes the resulting baseline file content for each capability
// without writing anything. It returns the planned writes along with any
// error (collisions, missing modify targets, etc.). PlanSlug is recorded in
// REMOVED audit comments.
func PlanMerge(specsDir, planSlug string, caps []Capability) ([]MergeFile, error) {
	out := make([]MergeFile, 0, len(caps))
	for _, c := range caps {
		path := CapabilityPath(specsDir, c.Slug)
		existing, existed, err := readBaseline(path)
		if err != nil {
			return nil, err
		}
		merged, summary, err := applyDeltas(existing, c, planSlug)
		if err != nil {
			return nil, fmt.Errorf("capability %q: %w", c.Slug, err)
		}
		out = append(out, MergeFile{
			Path:     path,
			Slug:     c.Slug,
			Content:  merged,
			Existed:  existed,
			Mutation: summary,
		})
	}
	return out, nil
}

// CommitMerge writes the planned files atomically per file (temp + rename).
// It does not roll back already-written files if a later write fails — callers
// should validate via PlanMerge first to make failures here unlikely.
func CommitMerge(files []MergeFile) error {
	for _, f := range files {
		if err := os.MkdirAll(filepath.Dir(f.Path), 0o775); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(f.Path), err)
		}
		tmp, err := os.CreateTemp(filepath.Dir(f.Path), ".spec-*.tmp")
		if err != nil {
			return fmt.Errorf("temp %s: %w", f.Path, err)
		}
		if _, err := tmp.WriteString(f.Content); err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
			return fmt.Errorf("write temp %s: %w", f.Path, err)
		}
		if err := tmp.Close(); err != nil {
			_ = os.Remove(tmp.Name())
			return err
		}
		if err := os.Rename(tmp.Name(), f.Path); err != nil {
			_ = os.Remove(tmp.Name())
			return fmt.Errorf("rename %s: %w", f.Path, err)
		}
	}
	return nil
}

func readBaseline(path string) (string, bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	return string(b), true, nil
}

// applyDeltas applies the deltas of one capability to its baseline text and
// returns the new content plus a human-readable summary of mutations.
func applyDeltas(baseline string, cap Capability, planSlug string) (string, []string, error) {
	if baseline == "" {
		baseline = newBaselineHeader(cap.Slug)
	}
	summary := []string{}
	for _, d := range cap.Deltas {
		switch d.Op {
		case DeltaAdded:
			for _, r := range d.Requirements {
				if findRequirementBlock(baseline, r.Name) != nil {
					return "", nil, fmt.Errorf("ADDED Requirement %q already exists in baseline", r.Name)
				}
				baseline = appendRequirement(baseline, r)
			}
			summary = append(summary, fmt.Sprintf("added %d requirement(s)", len(d.Requirements)))
		case DeltaModified:
			for _, r := range d.Requirements {
				loc := findRequirementBlock(baseline, r.Name)
				if loc == nil {
					return "", nil, fmt.Errorf("MODIFIED Requirement %q not found in baseline", r.Name)
				}
				baseline = replaceRange(baseline, loc, renderRequirement(r))
			}
			summary = append(summary, fmt.Sprintf("modified %d requirement(s)", len(d.Requirements)))
		case DeltaRemoved:
			for _, r := range d.Requirements {
				loc := findRequirementBlock(baseline, r.Name)
				if loc == nil {
					return "", nil, fmt.Errorf("REMOVED Requirement %q not found in baseline", r.Name)
				}
				audit := fmt.Sprintf("<!-- removed by %s on %s -->\n", planSlug, time.Now().Format("2006-01-02"))
				baseline = replaceRange(baseline, loc, audit)
			}
			summary = append(summary, fmt.Sprintf("removed %d requirement(s)", len(d.Requirements)))
		case DeltaRenamed:
			for _, r := range d.Requirements {
				loc := findRequirementBlock(baseline, r.Name)
				if loc == nil {
					return "", nil, fmt.Errorf("RENAMED Requirement %q not found in baseline", r.Name)
				}
				baseline = renameRequirementHeader(baseline, loc, r.NewName)
			}
			summary = append(summary, fmt.Sprintf("renamed %d requirement(s)", len(d.Requirements)))
		default:
			return "", nil, fmt.Errorf("unsupported delta op %q", d.Op)
		}
	}
	return baseline, summary, nil
}

func newBaselineHeader(slug string) string {
	return fmt.Sprintf("# Capability: %s\n\n## Requirements\n\n", slug)
}

// reqLoc points to a Requirement block in baseline text (line indices).
type reqLoc struct {
	Start int // first line of "### Requirement: ..."
	End   int // exclusive — first line not part of the block
}

// findRequirementBlock locates a `### Requirement: <name>` block in baseline
// text (case-insensitive name match, trimmed). Returns nil if not found.
//
// The block runs from the heading line until the next heading at the same
// or shallower level (`#`, `##`, `###`).
func findRequirementBlock(text, name string) *reqLoc {
	lines := strings.Split(text, "\n")
	target := strings.ToLower(strings.TrimSpace(name))
	for i, line := range lines {
		lvl, head, ok := headingLevel(line)
		if !ok || lvl != 3 {
			continue
		}
		nm, isReq := reqHeading(head)
		if !isReq {
			continue
		}
		if strings.ToLower(strings.TrimSpace(nm)) != target {
			continue
		}
		end := len(lines)
		for j := i + 1; j < len(lines); j++ {
			lvl2, _, ok2 := headingLevel(lines[j])
			if ok2 && lvl2 <= 3 {
				end = j
				break
			}
		}
		return &reqLoc{Start: i, End: end}
	}
	return nil
}

// replaceRange replaces lines [loc.Start, loc.End) in text with replacement.
// The replacement may be empty or multiline.
func replaceRange(text string, loc *reqLoc, replacement string) string {
	lines := strings.Split(text, "\n")
	repLines := strings.Split(strings.TrimRight(replacement, "\n"), "\n")
	out := make([]string, 0, len(lines)-(loc.End-loc.Start)+len(repLines))
	out = append(out, lines[:loc.Start]...)
	if replacement != "" {
		out = append(out, repLines...)
	}
	out = append(out, lines[loc.End:]...)
	return strings.Join(out, "\n")
}

// renameRequirementHeader rewrites the `### Requirement:` heading at loc.Start
// to use the new name; the rest of the block is preserved.
func renameRequirementHeader(text string, loc *reqLoc, newName string) string {
	lines := strings.Split(text, "\n")
	lines[loc.Start] = "### Requirement: " + strings.TrimSpace(newName)
	return strings.Join(lines, "\n")
}

// appendRequirement appends a Requirement block to the baseline text. If the
// baseline already contains a `## Requirements` section, the block is added
// at its end; otherwise the block is appended to the end of the document.
func appendRequirement(text string, r Requirement) string {
	block := renderRequirement(r)
	// Try to insert at end of the Requirements section.
	lines := strings.Split(text, "\n")
	sectionStart := -1
	sectionLevel := 0
	for i, line := range lines {
		lvl, head, ok := headingLevel(line)
		if !ok {
			continue
		}
		if isRequirementsSection(head) {
			sectionStart = i
			sectionLevel = lvl
			break
		}
	}
	if sectionStart == -1 {
		if !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		return text + "\n" + block
	}
	// find end of section
	end := len(lines)
	for j := sectionStart + 1; j < len(lines); j++ {
		lvl2, _, ok2 := headingLevel(lines[j])
		if ok2 && lvl2 <= sectionLevel {
			end = j
			break
		}
	}
	// trim trailing blank lines inside the section to insert cleanly
	insert := end
	for insert > sectionStart+1 && strings.TrimSpace(lines[insert-1]) == "" {
		insert--
	}
	out := make([]string, 0, len(lines)+8)
	out = append(out, lines[:insert]...)
	if insert > sectionStart+1 {
		out = append(out, "")
	}
	out = append(out, strings.Split(strings.TrimRight(block, "\n"), "\n")...)
	out = append(out, lines[insert:]...)
	return strings.Join(out, "\n")
}

// renderRequirement renders a Requirement block as baseline markdown
// (`### Requirement: <name>` + body + nested `#### Scenario:` blocks).
func renderRequirement(r Requirement) string {
	var b strings.Builder
	b.WriteString("### Requirement: ")
	b.WriteString(strings.TrimSpace(r.Name))
	b.WriteString("\n")
	for _, l := range trimEdges(r.Body) {
		b.WriteString(l)
		b.WriteString("\n")
	}
	for _, s := range r.Scenarios {
		b.WriteString("\n#### Scenario: ")
		b.WriteString(strings.TrimSpace(s.Name))
		b.WriteString("\n")
		for _, l := range trimEdges(s.Lines) {
			b.WriteString(l)
			b.WriteString("\n")
		}
	}
	return b.String()
}

// trimEdges removes leading and trailing empty lines from a slice.
func trimEdges(lines []string) []string {
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	end := len(lines)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	return lines[start:end]
}
