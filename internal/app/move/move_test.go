package move

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pacto/internal/i18n"
)

func writePlan(t *testing.T, root, state, slug, body string) string {
	t.Helper()
	dir := filepath.Join(root, state, slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "README.md")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestMovePlan_HappyPath_RewritesStatus(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "to-implement", "alpha", "# alpha\n\n**Status:** Pending (To Implement)  \n\nbody\n")
	if err := os.MkdirAll(filepath.Join(root, "current"), 0o755); err != nil {
		t.Fatal(err)
	}

	out, err := MovePlan(root, "to-implement", "alpha", "current", false, "", i18n.English)
	if err != nil {
		t.Fatalf("MovePlan: %v", err)
	}
	if filepath.Base(out) != "README.md" {
		t.Fatalf("expected README path, got %s", out)
	}
	if _, err := os.Stat(filepath.Join(root, "to-implement", "alpha")); !os.IsNotExist(err) {
		t.Fatalf("source dir should be gone")
	}
	b, _ := os.ReadFile(out)
	if !strings.Contains(string(b), "In Progress (Current)") {
		t.Fatalf("status not rewritten:\n%s", b)
	}
}

func TestMovePlan_AppendsMoveHistoryWhenReasonProvided(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "current", "beta", "# beta\n\n**Status:** In Progress (Current)  \n")
	if err := os.MkdirAll(filepath.Join(root, "done"), 0o755); err != nil {
		t.Fatal(err)
	}

	out, err := MovePlan(root, "current", "beta", "done", false, "shipped v1", i18n.English)
	if err != nil {
		t.Fatalf("MovePlan: %v", err)
	}
	b, _ := os.ReadFile(out)
	body := string(b)
	if !strings.Contains(body, "## Move History") {
		t.Fatalf("expected Move History section:\n%s", body)
	}
	if !strings.Contains(body, "shipped v1") {
		t.Fatalf("expected reason in history:\n%s", body)
	}
}

func TestMovePlan_RejectsInvalidStates(t *testing.T) {
	_, err := MovePlan(t.TempDir(), "bogus", "x", "current", false, "", i18n.English)
	if err == nil || !strings.Contains(err.Error(), "invalid state transition") {
		t.Fatalf("expected invalid state error, got %v", err)
	}
}

func TestMovePlan_RejectsSameSourceAndDestination(t *testing.T) {
	_, err := MovePlan(t.TempDir(), "current", "x", "current", false, "", i18n.English)
	if err == nil || !strings.Contains(err.Error(), "same") {
		t.Fatalf("expected same-state error, got %v", err)
	}
}

func TestMovePlan_RejectsInvalidSlug(t *testing.T) {
	_, err := MovePlan(t.TempDir(), "current", "Bad Slug!", "done", false, "", i18n.English)
	if err == nil || !strings.Contains(err.Error(), "invalid slug") {
		t.Fatalf("expected invalid slug error, got %v", err)
	}
}

func TestMovePlan_MissingSourcePlanFails(t *testing.T) {
	_, err := MovePlan(t.TempDir(), "current", "ghost", "done", false, "", i18n.English)
	if err == nil || !strings.Contains(err.Error(), "source plan not found") {
		t.Fatalf("expected missing source error, got %v", err)
	}
}

func TestMovePlan_DestinationExistsRequiresForce(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "current", "gamma", "# gamma\n\n**Status:** In Progress (Current)  \n")
	writePlan(t, root, "done", "gamma", "# old\n")

	_, err := MovePlan(root, "current", "gamma", "done", false, "", i18n.English)
	if err == nil || !strings.Contains(err.Error(), "destination already exists") {
		t.Fatalf("expected destination-exists error, got %v", err)
	}

	out, err := MovePlan(root, "current", "gamma", "done", true, "", i18n.English)
	if err != nil {
		t.Fatalf("MovePlan with force: %v", err)
	}
	b, _ := os.ReadFile(out)
	if !strings.Contains(string(b), "# gamma") {
		t.Fatalf("expected new content to win on --force:\n%s", b)
	}
}

func TestMovePlan_SpanishStatusRewrite(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "to-implement", "delta", "# delta\n\n**Estado:** Pendiente (To Implement)  \n")
	if err := os.MkdirAll(filepath.Join(root, "current"), 0o755); err != nil {
		t.Fatal(err)
	}

	out, err := MovePlan(root, "to-implement", "delta", "current", false, "", i18n.English)
	if err != nil {
		t.Fatalf("MovePlan: %v", err)
	}
	b, _ := os.ReadFile(out)
	body := string(b)
	// rewritePlanReadmeStatus currently forces English labels; ensure the
	// existing **Estado:** line is still updated in place rather than duplicated.
	if strings.Count(body, "**Status:**")+strings.Count(body, "**Estado:**") != 1 {
		t.Fatalf("expected exactly one status line after move:\n%s", body)
	}
}
