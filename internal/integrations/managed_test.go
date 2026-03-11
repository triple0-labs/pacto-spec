package integrations

import (
	"strings"
	"testing"
)

func TestManagedMetadataRoundTrip(t *testing.T) {
	meta := ManagedMetadata{
		Artifact:    "cursor/command/pacto-status.md",
		Workflow:    "status",
		Contract:    ContractVersion,
		TemplateSHA: TemplateChecksum("hello"),
		GeneratedBy: "pacto",
		GeneratedAt: "2026-03-06T00:00:00Z",
	}
	line := BuildManagedMetaLine(meta)
	got, err := ParseManagedMetaLine(line)
	if err != nil {
		t.Fatalf("ParseManagedMetaLine error: %v", err)
	}
	if got.Artifact != meta.Artifact || got.Workflow != meta.Workflow || got.Contract != meta.Contract || got.TemplateSHA != meta.TemplateSHA {
		t.Fatalf("parsed metadata mismatch: got=%+v want=%+v", got, meta)
	}
}

func TestFindManagedMeta(t *testing.T) {
	meta := ManagedMetadata{
		Artifact:    "cursor/skill/pacto-status",
		Workflow:    "status",
		Contract:    ContractVersion,
		TemplateSHA: TemplateChecksum("abc"),
		GeneratedBy: "pacto",
		GeneratedAt: "2026-03-06T00:00:00Z",
	}
	content := WrapManaged("abc", meta)
	got, ok, err := FindManagedMeta(content)
	if err != nil {
		t.Fatalf("FindManagedMeta error: %v", err)
	}
	if !ok {
		t.Fatal("expected managed metadata to be found")
	}
	if got.Artifact != meta.Artifact || got.Workflow != meta.Workflow {
		t.Fatalf("unexpected metadata: got=%+v want=%+v", got, meta)
	}
}

func TestMetadataDiff(t *testing.T) {
	a := ManagedMetadata{Artifact: "a", Workflow: "status", Contract: ContractVersion, TemplateSHA: "x"}
	b := ManagedMetadata{Artifact: "b", Workflow: "status", Contract: ContractVersion, TemplateSHA: "y"}
	diff := MetadataDiff(a, b)
	joined := strings.Join(diff, ",")
	if !strings.Contains(joined, "artifact") || !strings.Contains(joined, "template_sha256") {
		t.Fatalf("expected artifact/template diff, got %v", diff)
	}
}

