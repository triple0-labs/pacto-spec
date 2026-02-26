package app

import (
	"os"
	"path/filepath"
)

func resolvePlanRoot(path string) (string, bool) {
	if hasStateDirs(path) {
		return path, true
	}
	cand := filepath.Join(path, "plans")
	if hasStateDirs(cand) {
		return cand, true
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
