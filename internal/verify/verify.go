package verify

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"pacto/internal/model"
)

type Verifier struct {
	Root          string
	ExcludedFiles map[string]struct{}
}

func New(root string) Verifier {
	return Verifier{Root: root, ExcludedFiles: collectPlanDocs(root)}
}

func (v Verifier) VerifyClaim(plan model.PlanRef, c model.ClaimResult) model.ClaimResult {
	switch c.ClaimType {
	case model.ClaimPath:
		return v.verifyPath(c)
	case model.ClaimSymbol:
		return v.verifySearch(c)
	case model.ClaimEndpoint:
		query := c.SourceText
		if idx := strings.Index(query, " "); idx > 0 {
			query = strings.TrimSpace(query[idx+1:])
		}
		c.SourceText = strings.TrimSpace(c.SourceText)
		return v.verifySearchToken(c, query)
	case model.ClaimTestRef:
		return v.verifyTestRef(c)
	default:
		c.Result = "partial"
		return c
	}
}

func (v Verifier) verifyPath(c model.ClaimResult) model.ClaimResult {
	raw := strings.TrimSpace(c.SourceText)
	cand := []string{}
	if filepath.IsAbs(raw) {
		cand = append(cand, raw)
	} else {
		raw2 := strings.TrimPrefix(raw, "./")
		cand = append(cand, filepath.Join(v.Root, raw), filepath.Join(v.Root, raw2))
	}

	planOnly := []string{}
	outsideRoot := []string{}
	rootAbs := cleanAbs(v.Root)
	for _, p := range cand {
		abs := cleanAbs(p)
		if !isWithinRoot(rootAbs, abs) {
			outsideRoot = append(outsideRoot, abs)
			continue
		}
		if st, err := os.Stat(abs); err == nil && !st.IsDir() {
			if v.isExcluded(abs) {
				planOnly = append(planOnly, abs)
				continue
			}
			c.Result = "verified"
			c.References = []string{abs}
			return c
		}
	}
	if len(planOnly) > 0 {
		c.Result = "unverified"
		c.Evidence = "plan_doc_only"
		c.References = truncateRefs(planOnly, 3)
		return c
	}
	if len(outsideRoot) > 0 {
		c.Result = "unverified"
		c.Evidence = "outside_root"
		c.References = truncateRefs(outsideRoot, 3)
		return c
	}
	c.Result = "unverified"
	return c
}

func (v Verifier) verifySearch(c model.ClaimResult) model.ClaimResult {
	return v.verifySearchToken(c, c.SourceText)
}

func (v Verifier) verifySearchToken(c model.ClaimResult, token string) model.ClaimResult {
	ok, refs, planOnly := v.search(token)
	if ok {
		c.Result = "verified"
		c.Evidence = "repo_search"
		c.References = refs
		return c
	}
	if planOnly {
		c.Result = "unverified"
		c.Evidence = "plan_doc_only"
		c.References = refs
		return c
	}
	c.Result = "unverified"
	c.Evidence = "repo_search"
	return c
}

func (v Verifier) verifyTestRef(c model.ClaimResult) model.ClaimResult {
	raw := strings.TrimSpace(c.SourceText)
	if strings.Contains(raw, "/") || strings.HasSuffix(raw, ".py") || strings.HasSuffix(raw, ".go") || strings.HasSuffix(raw, ".ts") || strings.HasSuffix(raw, ".tsx") {
		return v.verifyPath(c)
	}
	return v.verifySearch(c)
}

func (v Verifier) search(token string) (bool, []string, bool) {
	token = strings.TrimSpace(token)
	if token == "" {
		return false, nil, false
	}
	refs, planRefs, err := v.searchRG(token)
	if err == nil {
		if len(refs) > 0 {
			return true, refs, false
		}
		if len(planRefs) > 0 {
			return false, planRefs, true
		}
		return false, nil, false
	}
	return v.searchWalk(token)
}

func (v Verifier) searchRG(token string) ([]string, []string, error) {
	cmd := exec.Command("rg", "-n", "--fixed-strings", token, v.Root, "-g", "!archive/**")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	refs := make([]string, 0, 3)
	planRefs := make([]string, 0, 3)
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		path := parseRGPath(ln)
		if path == "" {
			continue
		}
		abs := cleanAbs(path)
		if !filepath.IsAbs(abs) {
			abs = cleanAbs(filepath.Join(v.Root, path))
		}
		ref := truncateRef(ln)
		if v.isExcluded(abs) {
			if len(planRefs) < 3 {
				planRefs = append(planRefs, ref)
			}
			continue
		}
		refs = append(refs, ref)
		if len(refs) >= 3 {
			break
		}
	}
	return refs, planRefs, nil
}

func (v Verifier) searchWalk(token string) (bool, []string, bool) {
	refs := make([]string, 0, 3)
	planRefs := make([]string, 0, 3)
	_ = filepath.WalkDir(v.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if strings.Contains(path, string(filepath.Separator)+"archive") {
				return filepath.SkipDir
			}
			return nil
		}
		abs := cleanAbs(path)
		if len(refs) >= 3 {
			return filepath.SkipDir
		}
		f, e := os.Open(abs)
		if e != nil {
			return nil
		}
		defer f.Close()
		s := bufio.NewScanner(f)
		lineNo := 0
		for s.Scan() {
			lineNo++
			if strings.Contains(s.Text(), token) {
				ref := fmt.Sprintf("%s:%d", abs, lineNo)
				if v.isExcluded(abs) {
					if len(planRefs) < 3 {
						planRefs = append(planRefs, ref)
					}
				} else {
					refs = append(refs, ref)
				}
				break
			}
		}
		return nil
	})
	if len(refs) > 0 {
		return true, refs, false
	}
	if len(planRefs) > 0 {
		return false, planRefs, true
	}
	return false, nil, false
}

func collectPlanDocs(root string) map[string]struct{} {
	excluded := map[string]struct{}{}
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		dir := filepath.Join(root, st)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			planDir := filepath.Join(dir, e.Name())
			readme := cleanAbs(filepath.Join(planDir, "README.md"))
			if _, err := os.Stat(readme); err == nil {
				excluded[readme] = struct{}{}
			}
			docs, _ := filepath.Glob(filepath.Join(planDir, "PLAN*.md"))
			if len(docs) == 0 {
				docs, _ = filepath.Glob(filepath.Join(planDir, "*.md"))
				filtered := make([]string, 0, len(docs))
				for _, d := range docs {
					if strings.EqualFold(filepath.Base(d), "README.md") {
						continue
					}
					filtered = append(filtered, d)
				}
				docs = filtered
			}
			sort.Strings(docs)
			for _, d := range docs {
				excluded[cleanAbs(d)] = struct{}{}
			}
		}
	}
	return excluded
}

func isWithinRoot(rootAbs, candidateAbs string) bool {
	rel, err := filepath.Rel(rootAbs, candidateAbs)
	if err != nil {
		return false
	}
	if rel == "." || rel == "" {
		return true
	}
	rel = filepath.Clean(rel)
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return !filepath.IsAbs(rel)
}

func (v Verifier) isExcluded(path string) bool {
	_, ok := v.ExcludedFiles[cleanAbs(path)]
	return ok
}

func cleanAbs(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(path)
}

func parseRGPath(line string) string {
	parts := strings.SplitN(line, ":", 3)
	if len(parts) < 2 {
		return ""
	}
	return parts[0]
}

func truncateRef(s string) string {
	if len(s) <= 220 {
		return s
	}
	return s[:217] + "..."
}

func truncateRefs(items []string, n int) []string {
	if len(items) <= n {
		return items
	}
	return items[:n]
}
