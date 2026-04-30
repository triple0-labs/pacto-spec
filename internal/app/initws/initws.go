// Package initws is the use-case layer for the non-interactive half of
// `pacto init`: given an already-resolved onboarding profile and target
// project root, bootstrap the workspace skeleton (`.pacto/plans/<state>/`,
// README.md, PACTO.md), the context directory, the config file, the PRD
// scaffold, the optional managed AGENTS.md block, and (optionally) tool
// artifacts. It returns categorized lists of created/updated/skipped paths
// plus any per-tool install failures so the CLI can render the summary.
//
// The CLI layer is responsible for: profile detection, the interactive
// wizard, validation/warnings, dry-run output, the final summary, and any
// confirm-before-install prompt.
package initws

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pacto/internal/assets"
	plancontext "pacto/internal/context"
	"pacto/internal/integrations"
	"pacto/internal/onboarding"
	"pacto/internal/specsbaseline"
)

const (
	agentsManagedStart = "<!-- pacto:init:start -->"
	agentsManagedEnd   = "<!-- pacto:init:end -->"
)

// ErrInvalid wraps invalid filesystem state encountered while bootstrapping
// the workspace (e.g. a state path exists as a file, not a directory). The
// CLI maps this to exit code 2.
var ErrInvalid = errors.New("invalid state")

// Input is the request for an init apply.
type Input struct {
	// ProjectRoot is the resolved absolute project root.
	ProjectRoot string
	// Profile is the fully-resolved onboarding profile (post-validation).
	Profile onboarding.Profile
	// WithAgents writes/refreshes the managed block in AGENTS.md.
	WithAgents bool
	// Force overwrites existing managed files when applicable.
	Force bool
	// NoInstall skips per-tool artifact generation even when the profile
	// declares tools.
	NoInstall bool
}

// Result is the categorized outcome of an init apply.
type Result struct {
	PlansRoot string
	Created   []string
	Updated   []string
	Skipped   []string
	// InstallFailed contains preformatted per-tool failure lines, suitable
	// for printing one-per-line on stderr.
	InstallFailed []string
}

// Apply runs the bootstrap pipeline. It does not print anything.
func Apply(in Input) (Result, error) {
	res := Result{PlansRoot: filepath.Join(in.ProjectRoot, ".pacto", "plans")}

	if err := bootstrapWorkspace(res.PlansRoot, in.Profile.UILanguage, in.Force, &res); err != nil {
		return res, err
	}

	contextDir := plancontext.ContextDirFromPlansRoot(res.PlansRoot)
	contextReadme := filepath.Join(contextDir, "README.md")
	contextDomains := filepath.Join(contextDir, "domains")
	readmeExists := pathExists(contextReadme)
	domainsExists := pathExists(contextDomains)
	if err := plancontext.InitContext(res.PlansRoot); err != nil {
		return res, fmt.Errorf("init context: %w", err)
	}
	if pathExists(contextReadme) {
		if readmeExists {
			res.Skipped = append(res.Skipped, contextReadme)
		} else {
			res.Created = append(res.Created, contextReadme)
		}
	}
	if pathExists(contextDomains) {
		if domainsExists {
			res.Skipped = append(res.Skipped, contextDomains)
		} else {
			res.Created = append(res.Created, contextDomains)
		}
	}

	specsDir := specsbaseline.SpecsDirFromPlansRoot(res.PlansRoot)
	specsReadme := filepath.Join(specsDir, "README.md")
	specsReadmeExisted := pathExists(specsReadme)
	specsDirExisted := pathExists(specsDir)
	if err := specsbaseline.InitBaseline(res.PlansRoot); err != nil {
		return res, fmt.Errorf("init specs baseline: %w", err)
	}
	if pathExists(specsDir) {
		if specsDirExisted {
			// already known; only record README transitions below
		} else {
			res.Created = append(res.Created, specsDir)
		}
	}
	if pathExists(specsReadme) {
		if specsReadmeExisted {
			res.Skipped = append(res.Skipped, specsReadme)
		} else {
			res.Created = append(res.Created, specsReadme)
		}
	}

	cfgPath := filepath.Join(in.ProjectRoot, ".pacto", "config.yaml")
	cfgExisted := pathExists(cfgPath)
	cfgWritten, err := onboarding.WriteConfig(in.ProjectRoot, in.Profile)
	if err != nil {
		return res, fmt.Errorf("write .pacto/config.yaml: %w", err)
	}
	if cfgExisted {
		res.Updated = append(res.Updated, cfgWritten)
	} else {
		res.Created = append(res.Created, cfgWritten)
	}

	prdPath := filepath.Join(in.ProjectRoot, "prd.md")
	prdExisted := pathExists(prdPath)
	writtenPRD, prdChanged, err := onboarding.WritePRD(in.ProjectRoot, in.Profile)
	if err != nil {
		return res, fmt.Errorf("write prd.md: %w", err)
	}
	if prdChanged {
		if prdExisted {
			res.Updated = append(res.Updated, writtenPRD)
		} else {
			res.Created = append(res.Created, writtenPRD)
		}
	} else {
		res.Skipped = append(res.Skipped, writtenPRD)
	}

	if in.WithAgents {
		agentsPath := filepath.Join(in.ProjectRoot, "AGENTS.md")
		act, aerr := writeAgentsManagedBlock(agentsPath, assets.MustTemplateLang(in.Profile.UILanguage, "AGENTS.md"))
		if aerr != nil {
			return res, fmt.Errorf("update AGENTS.md: %w", aerr)
		}
		switch act {
		case "created":
			res.Created = append(res.Created, agentsPath)
		case "updated":
			res.Updated = append(res.Updated, agentsPath)
		case "skipped":
			res.Skipped = append(res.Skipped, agentsPath)
		}
	}

	if !in.NoInstall && len(in.Profile.Tools) > 0 {
		c, u, s, f := applyInstallPlan(in.ProjectRoot, in.Profile.Tools, in.Force)
		res.Created = append(res.Created, c...)
		res.Updated = append(res.Updated, u...)
		res.Skipped = append(res.Skipped, s...)
		res.InstallFailed = append(res.InstallFailed, f...)
	}

	return res, nil
}

// ApplyInstall runs only the per-tool artifact generation step. It exists so
// the CLI can defer the install branch behind an interactive confirmation
// prompt that lives outside the use case.
func ApplyInstall(projectRoot string, tools []string, force bool) (created, updated, skipped, failed []string) {
	return applyInstallPlan(projectRoot, tools, force)
}

func bootstrapWorkspace(plansRoot, lang string, force bool, res *Result) error {
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		p := filepath.Join(plansRoot, st)
		if info, err := os.Stat(p); err == nil {
			if !info.IsDir() {
				return fmt.Errorf("%w: state path exists but is not a directory: %s", ErrInvalid, p)
			}
			res.Skipped = append(res.Skipped, p)
			continue
		}
		if err := os.MkdirAll(p, 0o775); err != nil {
			return fmt.Errorf("create state dir %q: %w", p, err)
		}
		res.Created = append(res.Created, p)
	}

	workspaceFiles := map[string]string{
		filepath.Join(plansRoot, "README.md"): assets.MustTemplateLang(lang, "README.md"),
		filepath.Join(plansRoot, "PACTO.md"):  assets.MustTemplateLang(lang, "PACTO.md"),
	}

	for path, content := range workspaceFiles {
		wc, wu, ws, werr := writeManagedFile(path, content, force)
		if werr != nil {
			return fmt.Errorf("write file %q: %w", path, werr)
		}
		if wc {
			res.Created = append(res.Created, path)
		}
		if wu {
			res.Updated = append(res.Updated, path)
		}
		if ws {
			res.Skipped = append(res.Skipped, path)
		}
	}
	return nil
}

func writeManagedFile(path, content string, force bool) (created, updated, skipped bool, err error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o775); err != nil {
		return false, false, false, err
	}
	_, statErr := os.Stat(path)
	exists := statErr == nil
	if exists && !force {
		return false, false, true, nil
	}
	if err := os.WriteFile(path, []byte(content), 0o664); err != nil {
		return false, false, false, err
	}
	if exists {
		return false, true, false, nil
	}
	return true, false, false, nil
}

func applyInstallPlan(projectRoot string, tools []string, force bool) (created, updated, skipped, failed []string) {
	for _, toolID := range tools {
		results := integrations.GenerateForTool(projectRoot, toolID, force)
		for _, r := range results {
			if r.Err != nil {
				failed = append(failed, fmt.Sprintf("tool=%s kind=%s workflow=%s err=%v", r.Tool, r.Kind, r.WorkflowID, r.Err))
				continue
			}
			switch r.Outcome {
			case integrations.OutcomeCreated:
				created = append(created, r.Path)
			case integrations.OutcomeUpdated:
				updated = append(updated, r.Path)
			case integrations.OutcomeSkipped:
				skipped = append(skipped, r.Path)
			}
		}
	}
	return created, updated, skipped, failed
}

func writeAgentsManagedBlock(path, template string) (string, error) {
	block := agentsManagedStart + "\n" + strings.TrimSpace(template) + "\n" + agentsManagedEnd + "\n"

	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(block), 0o664); err != nil {
				return "", err
			}
			return "created", nil
		}
		return "", err
	}

	s := string(b)
	start := strings.Index(s, agentsManagedStart)
	end := strings.Index(s, agentsManagedEnd)
	if start >= 0 && end >= 0 && end > start {
		end += len(agentsManagedEnd)
		next := s[:start] + block + s[end:]
		if next == s {
			return "skipped", nil
		}
		if err := os.WriteFile(path, []byte(next), 0o664); err != nil {
			return "", err
		}
		return "updated", nil
	}

	trimmed := strings.TrimRight(s, "\n")
	next := trimmed + "\n\n" + block
	if next == s {
		return "skipped", nil
	}
	if err := os.WriteFile(path, []byte(next), 0o664); err != nil {
		return "", err
	}
	return "updated", nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
