package cmdutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pacto/internal/i18n"
)

func TestVersionLineDefault(t *testing.T) {
	got := VersionLine()
	if !strings.HasPrefix(got, "pacto version ") {
		t.Errorf("unexpected prefix: %q", got)
	}
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("expected trailing newline: %q", got)
	}
}

func TestVersionLineHonorsInjectedVersion(t *testing.T) {
	prev := Version
	t.Cleanup(func() { Version = prev })
	Version = "1.2.3"
	if got := VersionLine(); got != "pacto version 1.2.3\n" {
		t.Errorf("got %q", got)
	}
}

func TestSetGlobalLangOverrideAndEffectiveLanguage(t *testing.T) {
	t.Cleanup(func() { SetGlobalLangOverride("") })

	SetGlobalLangOverride("es")
	if got := EffectiveLanguage(""); got != i18n.Spanish {
		t.Errorf("expected Spanish, got %v", got)
	}

	SetGlobalLangOverride("en")
	if got := EffectiveLanguage(""); got != i18n.English {
		t.Errorf("expected English, got %v", got)
	}

	// Unknown overrides fall through to the default detection path.
	SetGlobalLangOverride("klingon")
	if got := EffectiveLanguage(""); got != i18n.English {
		t.Errorf("expected English fallback, got %v", got)
	}
}

func TestEffectiveLanguageReadsConfig(t *testing.T) {
	t.Cleanup(func() { SetGlobalLangOverride("") })
	SetGlobalLangOverride("")

	root := t.TempDir()
	cfgDir := filepath.Join(root, ".pacto")
	if err := os.MkdirAll(cfgDir, 0o775); err != nil {
		t.Fatal(err)
	}
	cfg := "ui:\n  language: es\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(cfg), 0o664); err != nil {
		t.Fatal(err)
	}

	if got := EffectiveLanguage(root); got != i18n.Spanish {
		t.Errorf("expected Spanish from config, got %v", got)
	}
}

func TestEffectiveLanguageDefaultsToEnglish(t *testing.T) {
	t.Cleanup(func() { SetGlobalLangOverride("") })
	SetGlobalLangOverride("")
	if got := EffectiveLanguage(t.TempDir()); got != i18n.English {
		t.Errorf("expected English default, got %v", got)
	}
}

func TestTrSelectsByLanguage(t *testing.T) {
	if got := Tr(i18n.English, "hello", "hola"); got != "hello" {
		t.Errorf("got %q", got)
	}
	if got := Tr(i18n.Spanish, "hello", "hola"); got != "hola" {
		t.Errorf("got %q", got)
	}
}
