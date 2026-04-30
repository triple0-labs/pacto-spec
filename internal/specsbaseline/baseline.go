package specsbaseline

import (
	"fmt"
	"os"
	"path/filepath"
)

const defaultBaselineReadme = `# Specs Baseline

This directory holds the persistent capability baseline for pacto.

## Structure

- one folder per capability slug — for example ` + "`auth/`" + ` or ` + "`billing/`" + `
- each capability folder contains ` + "`spec.md`" + ` with ` + "`### Requirement:`" + ` and nested ` + "`#### Scenario:`" + ` blocks
- baseline files answer "what does the system already do?" — they survive plan completion
- pacto creates and merges baseline files mechanically on ` + "`pacto move done`" + ` from plan delta blocks

## Conventions

- Use lowercase dash-separated capability folder names
- Plans express changes via ` + "`## Capability: <slug>`" + ` plus ` + "`### ADDED|MODIFIED|REMOVED|RENAMED Requirements`" + ` blocks in their ` + "`spec.md`" + `
- humans or agents may edit baseline files directly between plan moves
`

// SpecsDirFromPlansRoot returns the absolute path to the specs baseline tree
// (sibling of plans under .pacto/).
func SpecsDirFromPlansRoot(plansRoot string) string {
	return filepath.Join(filepath.Dir(filepath.Clean(plansRoot)), "specs")
}

// CapabilityPath returns the baseline spec.md path for a capability slug.
func CapabilityPath(specsDir, slug string) string {
	return filepath.Join(specsDir, slug, "spec.md")
}

// InitBaseline ensures the baseline tree exists with a README.md describing
// it. Idempotent — preserves any existing README and capability folders.
func InitBaseline(plansRoot string) error {
	dir := SpecsDirFromPlansRoot(plansRoot)
	if err := os.MkdirAll(dir, 0o775); err != nil {
		return fmt.Errorf("create specs baseline dir: %w", err)
	}
	readme := filepath.Join(dir, "README.md")
	if _, err := os.Stat(readme); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat baseline README: %w", err)
	}
	if err := os.WriteFile(readme, []byte(defaultBaselineReadme), 0o664); err != nil {
		return fmt.Errorf("write baseline README: %w", err)
	}
	return nil
}
