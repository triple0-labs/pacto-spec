package initws

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pacto/internal/i18n"
	"pacto/internal/onboarding"
)

func minimalProfile() onboarding.Profile {
	return onboarding.Profile{
		UILanguage:      string(i18n.English),
		Languages:       []string{"go"},
		CustomLanguages: nil,
		Tools:           nil,
		Intents:         onboarding.Intents{Problem: "test problem"},
	}
}

func TestApply_HappyPath(t *testing.T) {
	root := t.TempDir()
	res, err := Apply(Input{ProjectRoot: root, Profile: minimalProfile()})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !strings.HasSuffix(res.PlansRoot, filepath.Join(".pacto", "plans")) {
		t.Fatalf("unexpected PlansRoot: %s", res.PlansRoot)
	}
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		p := filepath.Join(res.PlansRoot, st)
		if info, err := os.Stat(p); err != nil || !info.IsDir() {
			t.Fatalf("missing state dir %s: %v", p, err)
		}
	}
	for _, f := range []string{"README.md", "PACTO.md"} {
		if _, err := os.Stat(filepath.Join(res.PlansRoot, f)); err != nil {
			t.Fatalf("missing %s: %v", f, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".pacto", "config.yaml")); err != nil {
		t.Fatalf("missing config: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "prd.md")); err != nil {
		t.Fatalf("missing prd: %v", err)
	}
	if len(res.Created) == 0 {
		t.Fatal("expected Created entries")
	}
	if len(res.InstallFailed) != 0 {
		t.Fatalf("unexpected install failures: %v", res.InstallFailed)
	}
}

func TestApply_RerunMarksFilesSkippedOrUpdated(t *testing.T) {
	root := t.TempDir()
	if _, err := Apply(Input{ProjectRoot: root, Profile: minimalProfile()}); err != nil {
		t.Fatalf("first Apply: %v", err)
	}
	res, err := Apply(Input{ProjectRoot: root, Profile: minimalProfile()})
	if err != nil {
		t.Fatalf("second Apply: %v", err)
	}
	if len(res.Skipped) == 0 {
		t.Fatal("expected at least one Skipped entry on rerun")
	}
}

func TestApply_RejectsStateAsFile(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".pacto", "plans"), 0o775); err != nil {
		t.Fatal(err)
	}
	clash := filepath.Join(root, ".pacto", "plans", "current")
	if err := os.WriteFile(clash, []byte("not a dir"), 0o664); err != nil {
		t.Fatal(err)
	}
	_, err := Apply(Input{ProjectRoot: root, Profile: minimalProfile()})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestApply_WithAgentsCreatesManagedBlock(t *testing.T) {
	root := t.TempDir()
	res, err := Apply(Input{ProjectRoot: root, Profile: minimalProfile(), WithAgents: true})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	agents := filepath.Join(root, "AGENTS.md")
	b, err := os.ReadFile(agents)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if !strings.Contains(string(b), agentsManagedStart) || !strings.Contains(string(b), agentsManagedEnd) {
		t.Fatalf("missing managed block markers in AGENTS.md")
	}
	found := false
	for _, p := range res.Created {
		if p == agents {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected AGENTS.md in Created, got %v", res.Created)
	}
}

func TestApply_WithAgents_RewritesExistingBlock(t *testing.T) {
	root := t.TempDir()
	agents := filepath.Join(root, "AGENTS.md")
	original := "# header\n\n" + agentsManagedStart + "\nold\n" + agentsManagedEnd + "\n"
	if err := os.WriteFile(agents, []byte(original), 0o664); err != nil {
		t.Fatal(err)
	}
	res, err := Apply(Input{ProjectRoot: root, Profile: minimalProfile(), WithAgents: true})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	updated := false
	for _, p := range res.Updated {
		if p == agents {
			updated = true
			break
		}
	}
	if !updated {
		t.Fatalf("expected AGENTS.md in Updated, got %v", res.Updated)
	}
	b, _ := os.ReadFile(agents)
	if !strings.Contains(string(b), "# header") {
		t.Fatal("lost user content outside managed block")
	}
}

func TestApply_InstallToolsGeneratesArtifacts(t *testing.T) {
	root := t.TempDir()
	p := minimalProfile()
	p.Tools = []string{"cursor"}
	res, err := Apply(Input{ProjectRoot: root, Profile: p})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".agents", "skills", "pacto-status", "SKILL.md")); err != nil {
		t.Fatalf("expected cursor artifacts written: %v", err)
	}
	if len(res.InstallFailed) != 0 {
		t.Fatalf("unexpected install failures: %v", res.InstallFailed)
	}
}

func TestApply_NoInstallSkipsArtifacts(t *testing.T) {
	root := t.TempDir()
	p := minimalProfile()
	p.Tools = []string{"cursor"}
	if _, err := Apply(Input{ProjectRoot: root, Profile: p, NoInstall: true}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".agents", "skills", "pacto-status", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("expected no artifacts when NoInstall, stat err=%v", err)
	}
}

func TestApplyInstall_DirectCall(t *testing.T) {
	root := t.TempDir()
	c, _, _, f := ApplyInstall(root, []string{"cursor"}, false)
	if len(f) != 0 {
		t.Fatalf("unexpected failures: %v", f)
	}
	if len(c) == 0 {
		t.Fatal("expected created paths from ApplyInstall")
	}
}
