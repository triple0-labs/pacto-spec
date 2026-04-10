package context

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractDomainsFromSpec(t *testing.T) {
	root := t.TempDir()
	specPath := filepath.Join(root, "spec.md")
	spec := "# Spec\n\n## Acceptance Criteria\n\n- AC-001: x\n\n## Domains Affected\n\n- auth\n- session\n\n## Notes\n\ntext\n"
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}

	got := ExtractDomains(specPath)
	if len(got) != 2 || got[0] != "auth" || got[1] != "session" {
		t.Fatalf("unexpected domains: %v", got)
	}
}

func TestExtractDomainsIgnoresPlaceholderAndMalformedLines(t *testing.T) {
	root := t.TempDir()
	specPath := filepath.Join(root, "spec.md")
	spec := "# Spec\n\n## Domains Affected\n\n- <domain>\nnot-a-list-item\n- Auth Session\n"
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}

	got := ExtractDomains(specPath)
	if len(got) != 1 || got[0] != "auth-session" {
		t.Fatalf("unexpected domains: %v", got)
	}
}

func TestNormalizeDomainSlug(t *testing.T) {
	if got := NormalizeDomainSlug(" Auth Session / Login "); got != "auth-session-login" {
		t.Fatalf("unexpected slug: %q", got)
	}
}

func TestReadContextDomains(t *testing.T) {
	root := t.TempDir()
	contextDir := filepath.Join(root, ".pacto", "context")
	domainsDir := filepath.Join(contextDir, "domains")
	if err := os.MkdirAll(domainsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"auth.md", "billing.md", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(domainsDir, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got := ReadContextDomains(contextDir)
	if len(got) != 2 || got[0] != "auth" || got[1] != "billing" {
		t.Fatalf("unexpected context domains: %v", got)
	}
}

func TestEnsureDomainDocsCreatesAndUpdatesWithoutDuplication(t *testing.T) {
	root := t.TempDir()
	contextDir := filepath.Join(root, ".pacto", "context")
	if err := EnsureDomainDocs(contextDir, []string{"auth"}, "done/plan-a"); err != nil {
		t.Fatal(err)
	}
	if err := EnsureDomainDocs(contextDir, []string{"auth"}, "done/plan-b"); err != nil {
		t.Fatal(err)
	}
	if err := EnsureDomainDocs(contextDir, []string{"auth"}, "done/plan-b"); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(filepath.Join(contextDir, "domains", "auth.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if strings.Count(text, "- done/plan-a") != 1 || strings.Count(text, "- done/plan-b") != 1 {
		t.Fatalf("expected unique related plans, got %q", text)
	}
	if !strings.Contains(text, "## Decisions") || !strings.Contains(text, "## Constraints") {
		t.Fatalf("expected scaffold sections, got %q", text)
	}
}

func TestEnsureDomainDocsPreservesManualContent(t *testing.T) {
	root := t.TempDir()
	contextDir := filepath.Join(root, ".pacto", "context")
	domainPath := filepath.Join(contextDir, "domains", "auth.md")
	if err := os.MkdirAll(filepath.Dir(domainPath), 0o755); err != nil {
		t.Fatal(err)
	}
	initial := "# Domain: auth\n\n## Summary\n\nManual summary.\n\n## Related Plans\n\n- done/existing\n\n## Decisions\n\nKeep this decision.\n\n## Constraints\n\nKeep this constraint.\n"
	if err := os.WriteFile(domainPath, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := EnsureDomainDocs(contextDir, []string{"auth"}, "done/new-plan"); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(domainPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if !strings.Contains(text, "Keep this decision.") || !strings.Contains(text, "Keep this constraint.") {
		t.Fatalf("expected manual content preserved, got %q", text)
	}
	if !strings.Contains(text, "- done/new-plan") {
		t.Fatalf("expected new related plan added, got %q", text)
	}
}

func TestDetectOverlaps(t *testing.T) {
	got := DetectOverlaps(map[string][]string{
		"current/plan-a":      {"auth", "billing"},
		"to-implement/plan-b": {"auth"},
		"current/plan-c":      {"session"},
	})
	if len(got) != 1 {
		t.Fatalf("expected one overlap, got %v", got)
	}
	if got[0].Domain != "auth" {
		t.Fatalf("unexpected overlap domain: %v", got)
	}
	if len(got[0].Plans) != 2 {
		t.Fatalf("unexpected overlap plans: %v", got)
	}
}

func TestInitContextIsIdempotent(t *testing.T) {
	root := t.TempDir()
	plansRoot := filepath.Join(root, ".pacto", "plans")
	if err := os.MkdirAll(plansRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := InitContext(plansRoot); err != nil {
		t.Fatal(err)
	}
	contextDir := ContextDirFromPlansRoot(plansRoot)
	readmePath := filepath.Join(contextDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("custom context"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := InitContext(plansRoot); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(b)) != "custom context" {
		t.Fatalf("expected context README preserved, got %q", string(b))
	}
	if _, err := os.Stat(filepath.Join(contextDir, "domains")); err != nil {
		t.Fatalf("expected domains dir: %v", err)
	}
}
