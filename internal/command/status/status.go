package statuscmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	appstatus "pacto/internal/app/status"
	"pacto/internal/config"
	"pacto/internal/domain/report"
	"pacto/internal/exitcode"
	"pacto/internal/i18n"
	"pacto/internal/render"
	statusui "pacto/internal/tui/status"
)

type Options struct {
	Root           string
	PlansRoot      string
	RepoRoot       string
	Mode           string
	Lang           string
	Format         string
	ConfigPath     string
	FailOn         string
	State          string
	IncludeArchive bool
	Verify         bool
	MaxNextActions int
	MaxBlockers    int
	Verbose        bool
	Provided       map[string]bool
}

var runStatusUI = statusui.Run

func Run(opts Options) int {
	provided := map[string]bool{}
	for k, v := range opts.Provided {
		provided[k] = v
	}

	if strings.TrimSpace(opts.Lang) != "" {
		if _, ok := i18n.ParseLanguage(opts.Lang); !ok {
			fmt.Fprintf(os.Stderr, "invalid --lang value %q (allowed: en|es)\n", opts.Lang)
			return 2
		}
		setGlobalLangOverride(opts.Lang)
		defer setGlobalLangOverride("")
	}

	cfg, cfgWarnings, runtimeWarnings, code, ok := buildStatusConfig(opts, provided)
	if !ok {
		return code
	}
	for _, w := range append([]string{}, append(cfgWarnings, runtimeWarnings...)...) {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}

	rep, code, ok := buildStatusReport(cfg, append(cfgWarnings, runtimeWarnings...))
	if !ok {
		return code
	}

	lang := effectiveLanguage(cfg.RepoRoot)
	if isTerminal(os.Stdout) && !provided["format"] {
		if err := runStatusUI(rep, lang); err != nil {
			fmt.Fprintf(os.Stderr, "run status tui: %v\n", err)
			return 3
		}
		return 0
	}

	out, err := render.RenderWithLanguage(rep, cfg.Format, lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render report: %v\n", err)
		return 3
	}
	fmt.Println(out)

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "config: mode=%s format=%s fail-on=%s state=%s include-archive=%t verify=%t root=%s plans-root=%s repo-root=%s\n", cfg.Mode, cfg.Format, cfg.FailOn, cfg.State, cfg.IncludeArchive, cfg.Verify, cfg.Root, cfg.PlansRoot, cfg.RepoRoot)
	}
	return exitcode.Evaluate(cfg.FailOn, rep)
}

func buildStatusConfig(values Options, provided map[string]bool) (config.Config, []string, []string, int, bool) {
	cwdAbs, err := filepath.Abs(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve cwd: %v\n", err)
		return config.Config{}, nil, nil, 2, false
	}
	cfg, cfgWarnings, err := config.Load(values.ConfigPath, cwdAbs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		return config.Config{}, nil, nil, 2, false
	}

	applyOverrides(&cfg, provided, values.Root, values.PlansRoot, values.RepoRoot, values.Mode, values.Format, values.FailOn, values.State, values.IncludeArchive, values.Verify, values.MaxNextActions, values.MaxBlockers)
	cfg = normalizeConfig(cfg)

	runtimeWarnings := make([]string, 0, 2)
	if provided["plans-root"] {
		runtimeWarnings = append(runtimeWarnings, "flag --plans-root is deprecated for status; use --root")
	}
	if provided["plans-root"] && provided["root"] {
		runtimeWarnings = append(runtimeWarnings, "flag --plans-root takes precedence over --root")
	}

	plansRoot, repoRoot, err := resolveStatusRoots(cfg, provided, cwdAbs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve roots: %v\n", err)
		return config.Config{}, nil, nil, 2, false
	}
	cfg.PlansRoot = plansRoot
	cfg.RepoRoot = repoRoot

	if err := validateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "invalid config: %v\n", err)
		return config.Config{}, nil, nil, 2, false
	}
	return cfg, cfgWarnings, runtimeWarnings, 0, true
}

func resolveStatusRoots(cfg config.Config, provided map[string]bool, cwdAbs string) (string, string, error) {
	projectRoot := ""
	if provided["root"] && !provided["plans-root"] {
		projectRoot = cfg.Root
	} else if strings.TrimSpace(cfg.Root) != "" && !provided["plans-root"] {
		projectRoot = cfg.Root
	}
	if projectRoot == "" {
		projectRoot = cwdAbs
	}
	projectRoot = cleanAbs(projectRoot)

	plansRoot := strings.TrimSpace(cfg.PlansRoot)
	detectedProjectRoot := ""
	if plansRoot == "" {
		if provided["root"] || strings.TrimSpace(cfg.Root) != "" {
			if resolved, ok := resolvePlanRoot(projectRoot); ok {
				plansRoot = resolved
				detectedProjectRoot = projectRoot
			} else {
				return "", "", fmt.Errorf("could not resolve plans root from %s (expected .pacto/plans)", projectRoot)
			}
		} else if resolved, foundProjectRoot, ok := resolvePlanRootFrom(projectRoot); ok {
			plansRoot = resolved
			detectedProjectRoot = foundProjectRoot
		} else {
			return "", "", fmt.Errorf("could not resolve plans root from %s or parents (expected .pacto/plans)", projectRoot)
		}
	}
	plansRoot = cleanAbs(plansRoot)
	if !hasStateDirs(plansRoot) {
		return "", "", fmt.Errorf("plans root missing state folders: %s", plansRoot)
	}

	repoRoot := strings.TrimSpace(cfg.RepoRoot)
	if repoRoot == "" {
		if detectedProjectRoot != "" {
			repoRoot = detectedProjectRoot
		} else {
			repoRoot = projectRoot
		}
	}
	repoRoot = cleanAbs(repoRoot)
	info, err := os.Stat(repoRoot)
	if err != nil || !info.IsDir() {
		return "", "", fmt.Errorf("repo root does not exist or is not a directory: %s", repoRoot)
	}

	return plansRoot, repoRoot, nil
}

func buildStatusReport(cfg config.Config, cfgWarnings []string) (report.StatusReport, int, bool) {
	rep, err := appstatus.BuildReport(appstatus.Input{
		PlansRoot:      cfg.PlansRoot,
		RepoRoot:       cfg.RepoRoot,
		Mode:           cfg.Mode,
		State:          cfg.State,
		IncludeArchive: cfg.IncludeArchive,
		Verify:         cfg.Verify,
		ClaimsPaths:    cfg.ClaimsPaths,
		MaxNextActions: cfg.MaxNextActions,
		MaxBlockers:    cfg.MaxBlockers,
		Warnings:       cfgWarnings,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return report.StatusReport{}, 3, false
	}
	return rep, 0, true
}

func applyOverrides(cfg *config.Config, provided map[string]bool, root, plansRoot, repoRoot, mode, format, failOn, state string, includeArchive bool, verify bool, maxNext, maxBlockers int) {
	if provided["root"] {
		if abs, err := filepath.Abs(root); err == nil {
			cfg.Root = abs
		}
	}
	if provided["plans-root"] {
		if abs, err := filepath.Abs(plansRoot); err == nil {
			cfg.PlansRoot = abs
		}
	}
	if provided["repo-root"] {
		if abs, err := filepath.Abs(repoRoot); err == nil {
			cfg.RepoRoot = abs
		}
	}
	if provided["mode"] {
		cfg.Mode = mode
	}
	if provided["format"] {
		cfg.Format = format
	}
	if provided["fail-on"] {
		cfg.FailOn = failOn
	}
	if provided["state"] {
		cfg.State = state
	}
	if provided["include-archive"] {
		cfg.IncludeArchive = includeArchive
	}
	if provided["verify"] {
		cfg.Verify = verify
	}
	if provided["max-next-actions"] {
		cfg.MaxNextActions = maxNext
	}
	if provided["max-blockers"] {
		cfg.MaxBlockers = maxBlockers
	}
}

func normalizeConfig(cfg config.Config) config.Config {
	cfg.Mode = strings.ToLower(strings.TrimSpace(cfg.Mode))
	cfg.Format = strings.ToLower(strings.TrimSpace(cfg.Format))
	cfg.FailOn = strings.ToLower(strings.TrimSpace(cfg.FailOn))
	cfg.State = strings.ToLower(strings.TrimSpace(cfg.State))
	cfg.Root = strings.TrimSpace(cfg.Root)
	cfg.PlansRoot = strings.TrimSpace(cfg.PlansRoot)
	cfg.RepoRoot = strings.TrimSpace(cfg.RepoRoot)
	return cfg
}

func validateConfig(cfg config.Config) error {
	if cfg.Mode != "compat" && cfg.Mode != "strict" {
		return fmt.Errorf("mode must be compat|strict")
	}
	if cfg.Format != "table" && cfg.Format != "json" {
		return fmt.Errorf("format must be table|json")
	}
	switch cfg.FailOn {
	case "none", "unverified", "partial", "blocked":
	default:
		return fmt.Errorf("fail-on must be none|unverified|partial|blocked")
	}
	validStates := []string{"all", "current", "to-implement", "done", "outdated"}
	ok := false
	for _, s := range validStates {
		if strings.EqualFold(cfg.State, s) {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("state must be all|current|to-implement|done|outdated")
	}
	if cfg.MaxNextActions < 1 || cfg.MaxBlockers < 1 {
		return fmt.Errorf("max limits must be >=1")
	}
	return nil
}

func cleanAbs(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(path)
}
