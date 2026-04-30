package cmdutil

import (
	"strings"
	"testing"
)

func TestNormalizeArgsReordersPositionalsAfterFlags(t *testing.T) {
	got, err := NormalizeArgs(
		[]string{"pos1", "--flag", "pos2"},
		map[string]bool{},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"--flag", "pos1", "pos2"}
	if !equalSlices(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestNormalizeArgsConsumesFlagValue(t *testing.T) {
	got, err := NormalizeArgs(
		[]string{"--root", "/x", "alpha"},
		map[string]bool{"--root": true},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"--root", "/x", "alpha"}
	if !equalSlices(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestNormalizeArgsErrorOnDanglingFlag(t *testing.T) {
	_, err := NormalizeArgs(
		[]string{"--root"},
		map[string]bool{"--root": true},
	)
	if err == nil {
		t.Fatal("expected error for missing value")
	}
	if !strings.Contains(err.Error(), "expects a value") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNormalizeArgsHandlesEmpty(t *testing.T) {
	got, err := NormalizeArgs(nil, map[string]bool{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestNormalizeArgsKeepsShortFlagOrder(t *testing.T) {
	got, err := NormalizeArgs(
		[]string{"a", "-x", "b", "-y", "c"},
		map[string]bool{},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"-x", "-y", "a", "b", "c"}
	if !equalSlices(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
