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
	for _, name := range []string{"auth", "billing"} {
		if err := os.MkdirAll(filepath.Join(domainsDir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// non-directory file should be excluded
	if err := os.WriteFile(filepath.Join(domainsDir, "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := ReadContextDomains(contextDir)
	if len(got) != 2 || got[0] != "auth" || got[1] != "billing" {
		t.Fatalf("unexpected context domains: %v", got)
	}
}

func TestEnsureDomainFoldersCreatesStubs(t *testing.T) {
	root := t.TempDir()
	contextDir := filepath.Join(root, ".pacto", "context")
	if err := EnsureDomainFolders(contextDir, []string{"auth"}); err != nil {
		t.Fatal(err)
	}

	folderPath := filepath.Join(contextDir, "domains", "auth")
	for _, name := range []string{"context.md", "decisions.md"} {
		if _, err := os.Stat(filepath.Join(folderPath, name)); err != nil {
			t.Fatalf("expected file %s: %v", name, err)
		}
	}

	b, err := os.ReadFile(filepath.Join(folderPath, "context.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "# auth") {
		t.Fatalf("expected slug heading in context.md, got %q", string(b))
	}
}

func TestEnsureDomainFoldersIdempotent(t *testing.T) {
	root := t.TempDir()
	contextDir := filepath.Join(root, ".pacto", "context")
	if err := EnsureDomainFolders(contextDir, []string{"auth"}); err != nil {
		t.Fatal(err)
	}
	if err := EnsureDomainFolders(contextDir, []string{"auth"}); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(filepath.Join(contextDir, "domains", "auth"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected exactly 2 files in domain folder, got %d", len(entries))
	}
}

func TestEnsureDomainFoldersPreservesManualContent(t *testing.T) {
	root := t.TempDir()
	contextDir := filepath.Join(root, ".pacto", "context")
	folderPath := filepath.Join(contextDir, "domains", "auth")
	if err := os.MkdirAll(folderPath, 0o755); err != nil {
		t.Fatal(err)
	}
	manualContent := "# auth\n\nManual bounded context description.\n"
	if err := os.WriteFile(filepath.Join(folderPath, "context.md"), []byte(manualContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// calling again must not touch the existing folder
	if err := EnsureDomainFolders(contextDir, []string{"auth"}); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(filepath.Join(folderPath, "context.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != manualContent {
		t.Fatalf("expected context.md untouched, got %q", string(b))
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
