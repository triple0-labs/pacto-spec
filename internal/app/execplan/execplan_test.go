package execplan

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pacto/internal/i18n"
)

const samplePlan = "# Sample\n\n" +
	"**Date:** 2026-01-01\n\n" +
	"## Phase 1: Setup\n\n" +
	"- [ ] 1.1 First task\n" +
	"- [ ] 1.2 Second task\n\n" +
	"## Phase 2: Build\n\n" +
	"- [ ] 2.1 Third task\n"

func setupPlan(t *testing.T, content string) (string, string, string) {
	t.Helper()
	root := t.TempDir()
	plansRoot := filepath.Join(root, ".pacto", "plans")
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		if err := os.MkdirAll(filepath.Join(plansRoot, st), 0o775); err != nil {
			t.Fatal(err)
		}
	}
	dir := filepath.Join(plansRoot, "current", "alpha")
	if err := os.MkdirAll(dir, 0o775); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# alpha\n"), 0o664); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(dir, "tasks.md")
	if err := os.WriteFile(planPath, []byte(content), 0o664); err != nil {
		t.Fatal(err)
	}
	return root, dir, planPath
}

func fixedNow() time.Time {
	return time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC)
}

func TestApply_CompletesNextTask(t *testing.T) {
	root, _, planPath := setupPlan(t, samplePlan)
	res, err := Apply(Input{
		Root: root, State: "current", Slug: "alpha",
		Lang: i18n.English, Now: fixedNow(),
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if res.PlanPath != planPath {
		t.Fatalf("unexpected PlanPath: %s", res.PlanPath)
	}
	if len(res.Actions) != 1 || res.Actions[0] != "completed 1.1" {
		t.Fatalf("unexpected actions: %v", res.Actions)
	}
	b, _ := os.ReadFile(planPath)
	if !strings.Contains(string(b), "- [x] 1.1") {
		t.Fatalf("expected first task completed, got:\n%s", b)
	}
	if !strings.Contains(string(b), "**Last Modified:** 2026-04-29 12:00") {
		t.Fatalf("expected Last Modified inserted, got:\n%s", b)
	}
}

func TestApply_TargetsSpecificStep(t *testing.T) {
	root, _, planPath := setupPlan(t, samplePlan)
	res, err := Apply(Input{
		Root: root, State: "current", Slug: "alpha", Step: "2.1",
		Lang: i18n.English, Now: fixedNow(),
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(res.Actions) != 1 || res.Actions[0] != "completed 2.1" {
		t.Fatalf("unexpected actions: %v", res.Actions)
	}
	b, _ := os.ReadFile(planPath)
	if !strings.Contains(string(b), "- [x] 2.1") {
		t.Fatalf("expected 2.1 completed")
	}
	if strings.Contains(string(b), "- [x] 1.1") {
		t.Fatalf("did not expect 1.1 completed")
	}
}

func TestApply_AppendsNoteBlockerEvidence(t *testing.T) {
	root, _, planPath := setupPlan(t, samplePlan)
	res, err := Apply(Input{
		Root: root, State: "current", Slug: "alpha", Step: "1.1",
		Note: "took an hour", Blocker: "needs review", Evidence: "make build",
		Lang: i18n.English, Now: fixedNow(),
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(res.Actions) != 4 {
		t.Fatalf("expected 4 actions, got %v", res.Actions)
	}
	b, _ := os.ReadFile(planPath)
	s := string(b)
	if !strings.Contains(s, "## Execution Notes") || !strings.Contains(s, "took an hour") {
		t.Fatalf("missing execution note section: %s", s)
	}
	if !strings.Contains(s, "## Blockers") || !strings.Contains(s, "needs review") {
		t.Fatalf("missing blockers section")
	}
	if !strings.Contains(s, "## Evidence") || !strings.Contains(s, "`make build`") {
		t.Fatalf("missing evidence section")
	}
}

func TestApply_DryRunDoesNotWrite(t *testing.T) {
	root, _, planPath := setupPlan(t, samplePlan)
	originalBytes, _ := os.ReadFile(planPath)
	res, err := Apply(Input{
		Root: root, State: "current", Slug: "alpha", DryRun: true,
		Lang: i18n.English, Now: fixedNow(),
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !res.DryRun {
		t.Fatal("expected DryRun=true on Result")
	}
	after, _ := os.ReadFile(planPath)
	if string(after) != string(originalBytes) {
		t.Fatal("file should not be modified on dry-run")
	}
}

func TestApply_NoChangeWhenAllDone(t *testing.T) {
	allDone := "# Sample\n\n## Phase 1\n\n- [x] 1.1 done\n"
	root, _, _ := setupPlan(t, allDone)
	res, err := Apply(Input{
		Root: root, State: "current", Slug: "alpha",
		Lang: i18n.English, Now: fixedNow(),
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !res.NoChange {
		t.Fatalf("expected NoChange, got %+v", res)
	}
}

func TestApply_RejectsNonCurrentState(t *testing.T) {
	root, _, _ := setupPlan(t, samplePlan)
	_, err := Apply(Input{Root: root, State: "to-implement", Slug: "alpha", Lang: i18n.English})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestApply_RejectsBadState(t *testing.T) {
	_, err := Apply(Input{Root: t.TempDir(), State: "weird", Slug: "alpha", Lang: i18n.English})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestApply_RejectsBadSlug(t *testing.T) {
	_, err := Apply(Input{Root: t.TempDir(), State: "current", Slug: "BadSlug", Lang: i18n.English})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestApply_RejectsLegacyStepFormat(t *testing.T) {
	root, _, _ := setupPlan(t, samplePlan)
	_, err := Apply(Input{
		Root: root, State: "current", Slug: "alpha", Step: "T1",
		Lang: i18n.English,
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestApply_RejectsInvalidStepFormat(t *testing.T) {
	root, _, _ := setupPlan(t, samplePlan)
	_, err := Apply(Input{
		Root: root, State: "current", Slug: "alpha", Step: "1",
		Lang: i18n.English,
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestApply_RejectsPlanWithoutPhases(t *testing.T) {
	noPhases := "# Sample\n\n## Notes\n\n- nothing\n"
	root, _, _ := setupPlan(t, noPhases)
	_, err := Apply(Input{Root: root, State: "current", Slug: "alpha", Lang: i18n.English})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestApply_RejectsMissingPlan(t *testing.T) {
	_, err := Apply(Input{Root: t.TempDir(), State: "current", Slug: "missing", Lang: i18n.English})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestApply_SpanishPlanUsesSpanishHeadings(t *testing.T) {
	es := "# Plan\n\n**Fecha:** 2026-01-01\n\n## Fase 1\n\n- [ ] 1.1 hacer cosa\n\n## Evidencia\n\n## Bloqueadores\n"
	root, _, planPath := setupPlan(t, es)
	_, err := Apply(Input{
		Root: root, State: "current", Slug: "alpha", Step: "1.1",
		Note: "hecho", Blocker: "esperar revisión", Evidence: "make build",
		Lang: i18n.English, Now: fixedNow(),
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	b, _ := os.ReadFile(planPath)
	s := string(b)
	if !strings.Contains(s, "## Notas de Ejecución") {
		t.Fatalf("expected Spanish notes heading: %s", s)
	}
	if !strings.Contains(s, "**Última Modificación:**") {
		t.Fatalf("expected Spanish Last Modified label: %s", s)
	}
}
