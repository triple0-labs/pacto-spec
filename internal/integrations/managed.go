package integrations

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	ManagedStart = "<!-- pacto:managed:start -->"
	ManagedEnd   = "<!-- pacto:managed:end -->"
	ManagedMeta  = "<!-- pacto:managed:meta"

	ContractVersion = "v1"
)

type ManagedMetadata struct {
	Artifact     string
	Workflow     string
	Contract     string
	TemplateSHA  string
	GeneratedBy  string
	GeneratedAt  string
}

func (m ManagedMetadata) normalize() ManagedMetadata {
	out := m
	out.Artifact = strings.TrimSpace(out.Artifact)
	out.Workflow = strings.TrimSpace(out.Workflow)
	out.Contract = strings.TrimSpace(out.Contract)
	out.TemplateSHA = strings.TrimSpace(out.TemplateSHA)
	out.GeneratedBy = strings.TrimSpace(out.GeneratedBy)
	out.GeneratedAt = strings.TrimSpace(out.GeneratedAt)
	if out.Contract == "" {
		out.Contract = ContractVersion
	}
	if out.GeneratedBy == "" {
		out.GeneratedBy = "pacto"
	}
	if out.GeneratedAt == "" {
		out.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	}
	return out
}

func TemplateChecksum(body string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(body)))
	return hex.EncodeToString(sum[:])
}

func BuildManagedMetaLine(meta ManagedMetadata) string {
	m := meta.normalize()
	parts := []string{
		"artifact=" + m.Artifact,
		"workflow=" + m.Workflow,
		"contract=" + m.Contract,
		"template_sha256=" + m.TemplateSHA,
		"generated_by=" + m.GeneratedBy,
		"generated_at=" + m.GeneratedAt,
	}
	return ManagedMeta + " " + strings.Join(parts, " ") + " -->"
}

func ParseManagedMetaLine(line string) (ManagedMetadata, error) {
	raw := strings.TrimSpace(line)
	if !strings.HasPrefix(raw, ManagedMeta) || !strings.HasSuffix(raw, "-->") {
		return ManagedMetadata{}, fmt.Errorf("managed metadata marker not found")
	}
	payload := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(raw, ManagedMeta), "-->"))
	if payload == "" {
		return ManagedMetadata{}, fmt.Errorf("empty managed metadata payload")
	}
	meta := ManagedMetadata{}
	for _, tok := range strings.Fields(payload) {
		parts := strings.SplitN(tok, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		switch k {
		case "artifact":
			meta.Artifact = v
		case "workflow":
			meta.Workflow = v
		case "contract":
			meta.Contract = v
		case "template_sha256":
			meta.TemplateSHA = v
		case "generated_by":
			meta.GeneratedBy = v
		case "generated_at":
			meta.GeneratedAt = v
		}
	}
	meta = meta.normalize()
	if meta.Artifact == "" || meta.Workflow == "" || meta.TemplateSHA == "" {
		return ManagedMetadata{}, fmt.Errorf("managed metadata missing required fields")
	}
	return meta, nil
}

func WrapManaged(body string, meta ManagedMetadata) string {
	line := BuildManagedMetaLine(meta)
	return ManagedStart + "\n" + line + "\n" + strings.TrimSpace(body) + "\n" + ManagedEnd + "\n"
}

func WriteManaged(path, body string, meta ManagedMetadata, force bool) (WriteResult, error) {
	wrapped := WrapManaged(body, meta)
	if err := os.MkdirAll(filepath.Dir(path), 0o775); err != nil {
		return WriteResult{}, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(wrapped), 0o664); err != nil {
				return WriteResult{}, err
			}
			return WriteResult{Outcome: OutcomeCreated}, nil
		}
		return WriteResult{}, err
	}

	s := string(b)
	start := strings.Index(s, ManagedStart)
	end := strings.Index(s, ManagedEnd)
	if start >= 0 && end > start {
		end += len(ManagedEnd)
		next := s[:start] + strings.TrimRight(wrapped, "\n") + s[end:]
		if next == s {
			return WriteResult{Outcome: OutcomeSkipped, Reason: "unchanged"}, nil
		}
		if err := os.WriteFile(path, []byte(next), 0o664); err != nil {
			return WriteResult{}, err
		}
		return WriteResult{Outcome: OutcomeUpdated}, nil
	}

	if !force {
		return WriteResult{Outcome: OutcomeSkipped, Reason: "unmanaged_exists"}, nil
	}
	if err := os.WriteFile(path, []byte(wrapped), 0o664); err != nil {
		return WriteResult{}, err
	}
	return WriteResult{Outcome: OutcomeUpdated, Reason: "force_overwrite"}, nil
}

func FindManagedMeta(content string) (ManagedMetadata, bool, error) {
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, "pacto:managed:meta") {
			meta, err := ParseManagedMetaLine(line)
			if err != nil {
				return ManagedMetadata{}, false, err
			}
			return meta, true, nil
		}
	}
	return ManagedMetadata{}, false, nil
}

func MetadataDiff(actual, expected ManagedMetadata) []string {
	issues := make([]string, 0, 3)
	if strings.TrimSpace(actual.Artifact) != strings.TrimSpace(expected.Artifact) {
		issues = append(issues, "artifact")
	}
	if strings.TrimSpace(actual.Workflow) != strings.TrimSpace(expected.Workflow) {
		issues = append(issues, "workflow")
	}
	if strings.TrimSpace(actual.Contract) != strings.TrimSpace(expected.Contract) {
		issues = append(issues, "contract")
	}
	if strings.TrimSpace(actual.TemplateSHA) != strings.TrimSpace(expected.TemplateSHA) {
		issues = append(issues, "template_sha256")
	}
	sort.Strings(issues)
	return issues
}
