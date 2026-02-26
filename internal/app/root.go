package app

import (
	"os"
	"path/filepath"
)

func resolvePlanRoot(path string) (string, bool) {
	if hasStateDirs(path) {
		return path, true
	}
	for _, rel := range []string{filepath.Join(".pacto", "plans"), "plans"} {
		cand := filepath.Join(path, rel)
		if hasStateDirs(cand) {
			return cand, true
		}
	}
	return path, false
}

func hasStateDirs(path string) bool {
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		info, err := os.Stat(filepath.Join(path, st))
		if err != nil || !info.IsDir() {
			return false
		}
	}
	return true
}
