package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pacto/internal/model"
)

var defaultStates = []string{"current", "to-implement", "done", "outdated"}

type Options struct {
	StateFilter    string
	IncludeArchive bool
}

func FindPlans(root string, opts Options) ([]model.PlanRef, error) {
	states := selectedStates(opts)
	res := make([]model.PlanRef, 0)
	for _, st := range states {
		plans, err := collectState(root, st)
		if err != nil {
			return nil, err
		}
		res = append(res, plans...)
	}
	if opts.IncludeArchive {
		plans, err := collectState(root, "archive")
		if err == nil {
			res = append(res, plans...)
		}
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].State == res[j].State {
			return res[i].Slug < res[j].Slug
		}
		return res[i].State < res[j].State
	})
	return res, nil
}

func selectedStates(opts Options) []string {
	state := strings.ToLower(strings.TrimSpace(opts.StateFilter))
	if state == "" || state == "all" {
		return append([]string{}, defaultStates...)
	}
	for _, st := range defaultStates {
		if st == state {
			return []string{state}
		}
	}
	return append([]string{}, defaultStates...)
}

func collectState(root, st string) ([]model.PlanRef, error) {
	stateDir := filepath.Join(root, st)
	ents, err := os.ReadDir(stateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read state dir %s: %w", stateDir, err)
	}

	res := make([]model.PlanRef, 0)
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		d := filepath.Join(stateDir, e.Name())
		readme := filepath.Join(d, "README.md")
		if _, err := os.Stat(readme); err != nil {
			continue
		}
		planDocs, _ := filepath.Glob(filepath.Join(d, "PLAN*.md"))
		if len(planDocs) == 0 {
			planDocs, _ = filepath.Glob(filepath.Join(d, "*.md"))
			filtered := make([]string, 0, len(planDocs))
			for _, f := range planDocs {
				if strings.EqualFold(filepath.Base(f), "README.md") {
					continue
				}
				filtered = append(filtered, f)
			}
			planDocs = filtered
		}
		sort.Strings(planDocs)
		res = append(res, model.PlanRef{State: st, Slug: e.Name(), Dir: d, Readme: readme, PlanDocs: planDocs})
	}
	return res, nil
}
