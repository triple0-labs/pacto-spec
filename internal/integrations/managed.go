package integrations

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	ManagedStart = "<!-- pacto:managed:start -->"
	ManagedEnd   = "<!-- pacto:managed:end -->"
)

func WrapManaged(body string) string {
	return ManagedStart + "\n" + strings.TrimSpace(body) + "\n" + ManagedEnd + "\n"
}

func WriteManaged(path, body string, force bool) (WriteResult, error) {
	wrapped := WrapManaged(body)
	if err := os.MkdirAll(filepath.Dir(path), 0o775); err != nil {
		return WriteResult{}, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(wrapped), 0o664); err != nil {
				return WriteResult{}, err
			}
			return WriteResult{Outcome: OutcomeCreated}, nil
		}
		return WriteResult{}, err
	}

	s := string(b)
	start := strings.Index(s, ManagedStart)
	end := strings.Index(s, ManagedEnd)
	if start >= 0 && end > start {
		end += len(ManagedEnd)
		next := s[:start] + strings.TrimRight(wrapped, "\n") + s[end:]
		if next == s {
			return WriteResult{Outcome: OutcomeSkipped, Reason: "unchanged"}, nil
		}
		if err := os.WriteFile(path, []byte(next), 0o664); err != nil {
			return WriteResult{}, err
		}
		return WriteResult{Outcome: OutcomeUpdated}, nil
	}

	if !force {
		return WriteResult{Outcome: OutcomeSkipped, Reason: "unmanaged_exists"}, nil
	}
	if err := os.WriteFile(path, []byte(wrapped), 0o664); err != nil {
		return WriteResult{}, err
	}
	return WriteResult{Outcome: OutcomeUpdated, Reason: "force_overwrite"}, nil
}
