package verify

import (
	"os"
	"path/filepath"
	"testing"

	"pacto/internal/model"
)

func TestVerifyPathVerifiedInRoot(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "src", "main.go"), "package main\n")

	v := New(root)
	c := model.ClaimResult{ClaimType: model.ClaimPath, SourceText: "src/main.go"}
	got := v.VerifyClaim(model.PlanRef{}, c)

	if got.Result != "verified" {
		t.Fatalf("expected verified result, got %q", got.Result)
	}
	if len(got.References) != 1 {
		t.Fatalf("expected 1 reference, got %d", len(got.References))
	}
}

func TestVerifyPathRejectsTraversalOutsideRoot(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "repo")
	mustMkdirAll(t, root)
	writeFile(t, filepath.Join(parent, "outside.txt"), "outside\n")

	v := New(root)
	c := model.ClaimResult{ClaimType: model.ClaimPath, SourceText: "../outside.txt"}
	got := v.VerifyClaim(model.PlanRef{}, c)

	if got.Result != "unverified" {
		t.Fatalf("expected unverified result, got %q", got.Result)
	}
	if got.Evidence != "outside_root" {
		t.Fatalf("expected outside_root evidence, got %q", got.Evidence)
	}
	if len(got.References) == 0 {
		t.Fatalf("expected at least one outside-root reference")
	}
}

func TestVerifyPathRejectsAbsoluteOutsideRoot(t *testing.T) {
	root := t.TempDir()
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "outside.txt")
	writeFile(t, outsideFile, "outside\n")

	v := New(root)
	c := model.ClaimResult{ClaimType: model.ClaimPath, SourceText: outsideFile}
	got := v.VerifyClaim(model.PlanRef{}, c)

	if got.Result != "unverified" {
		t.Fatalf("expected unverified result, got %q", got.Result)
	}
	if got.Evidence != "outside_root" {
		t.Fatalf("expected outside_root evidence, got %q", got.Evidence)
	}
}

func TestVerifyPathPlanDocOnlyFromFallbackDocs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "current", "plan-a", "README.md"), "# plan a\n")
	writeFile(t, filepath.Join(root, "current", "plan-a", "DESIGN.md"), "details\n")

	v := New(root)
	c := model.ClaimResult{ClaimType: model.ClaimPath, SourceText: "current/plan-a/DESIGN.md"}
	got := v.VerifyClaim(model.PlanRef{}, c)

	if got.Result != "unverified" {
		t.Fatalf("expected unverified result, got %q", got.Result)
	}
	if got.Evidence != "plan_doc_only" {
		t.Fatalf("expected plan_doc_only evidence, got %q", got.Evidence)
	}
}

func TestCollectPlanDocsFallbackIncludesAllMarkdownExceptReadme(t *testing.T) {
	root := t.TempDir()
	readme := filepath.Join(root, "current", "plan-a", "README.md")
	design := filepath.Join(root, "current", "plan-a", "DESIGN.md")
	notes := filepath.Join(root, "current", "plan-a", "NOTES.md")
	writeFile(t, readme, "# plan a\n")
	writeFile(t, design, "design\n")
	writeFile(t, notes, "notes\n")

	excluded := collectPlanDocs(root)

	assertExcluded(t, excluded, readme, true)
	assertExcluded(t, excluded, design, true)
	assertExcluded(t, excluded, notes, true)
}

func TestCollectPlanDocsPrefersPLANFiles(t *testing.T) {
	root := t.TempDir()
	readme := filepath.Join(root, "current", "plan-b", "README.md")
	plan := filepath.Join(root, "current", "plan-b", "PLAN_TOPIC.md")
	design := filepath.Join(root, "current", "plan-b", "DESIGN.md")
	writeFile(t, readme, "# plan b\n")
	writeFile(t, plan, "plan\n")
	writeFile(t, design, "design\n")

	excluded := collectPlanDocs(root)

	assertExcluded(t, excluded, readme, true)
	assertExcluded(t, excluded, plan, true)
	assertExcluded(t, excluded, design, false)
}

func TestVerifySearchTokenPlanDocOnlyInFallbackDocs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "current", "plan-a", "README.md"), "# plan a\n")
	writeFile(t, filepath.Join(root, "current", "plan-a", "DESIGN.md"), "UNIQUE_PLAN_TOKEN\n")

	v := New(root)
	c := model.ClaimResult{ClaimType: model.ClaimSymbol, SourceText: "UNIQUE_PLAN_TOKEN"}
	got := v.VerifyClaim(model.PlanRef{}, c)

	if got.Result != "unverified" {
		t.Fatalf("expected unverified result, got %q", got.Result)
	}
	if got.Evidence != "plan_doc_only" {
		t.Fatalf("expected plan_doc_only evidence, got %q", got.Evidence)
	}
}

func TestVerifySearchTokenVerifiedInRepoFile(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "src", "service.go"), "const UniqueRepoToken = 1\n")

	v := New(root)
	c := model.ClaimResult{ClaimType: model.ClaimSymbol, SourceText: "UniqueRepoToken"}
	got := v.VerifyClaim(model.PlanRef{}, c)

	if got.Result != "verified" {
		t.Fatalf("expected verified result, got %q", got.Result)
	}
	if got.Evidence != "repo_search" {
		t.Fatalf("expected repo_search evidence, got %q", got.Evidence)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	mustMkdirAll(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func assertExcluded(t *testing.T, excluded map[string]struct{}, path string, want bool) {
	t.Helper()
	_, ok := excluded[cleanAbs(path)]
	if ok != want {
		t.Fatalf("path %s excluded=%t, want %t", path, ok, want)
	}
}
