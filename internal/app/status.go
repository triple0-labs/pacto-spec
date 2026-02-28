package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pacto/internal/analyze"
	"pacto/internal/claims"
	"pacto/internal/config"
	"pacto/internal/discovery"
	"pacto/internal/exitcode"
	"pacto/internal/model"
	"pacto/internal/parser"
	"pacto/internal/report"
	statusui "pacto/internal/tui/status"
	"pacto/internal/verify"
)

type statusFlagValues struct {
	root           string
	plansRoot      string
	repoRoot       string
	mode           string
	lang           string
	format         string
	configPath     string
	failOn         string
	state          string
	includeArchive bool
	maxNext        int
	maxBlockers    int
	verbose        bool
}

func RunStatus(args []string) int {
	values, provided, code, ok := parseStatusFlags(args)
	if !ok {
		return code
	}

	cfg, cfgWarnings, runtimeWarnings, code, ok := buildStatusConfig(values, provided)
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

	if isTerminal(os.Stdout) {
		if provided["format"] {
			fmt.Fprintln(os.Stderr, "flag --format is only supported in non-TTY mode for pacto status")
			fmt.Fprintln(os.Stderr, "hint: run without --format for interactive status, or pipe output for table/json")
			return 2
		}
		if err := statusui.Run(rep); err != nil {
			fmt.Fprintf(os.Stderr, "run status tui: %v\n", err)
			return 3
		}
		return 0
	}

	out, err := report.Render(rep, cfg.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render report: %v\n", err)
		return 3
	}
	fmt.Println(out)

	if values.verbose {
		fmt.Fprintf(os.Stderr, "config: mode=%s format=%s fail-on=%s state=%s include-archive=%t root=%s plans-root=%s repo-root=%s\n", cfg.Mode, cfg.Format, cfg.FailOn, cfg.State, cfg.IncludeArchive, cfg.Root, cfg.PlansRoot, cfg.RepoRoot)
	}
	return exitcode.Evaluate(cfg.FailOn, rep)
}

func parseStatusFlags(args []string) (statusFlagValues, map[string]bool, int, bool) {
	values := statusFlagValues{}
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  pacto status [options]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		fs.PrintDefaults()
	}

	fs.StringVar(&values.root, "root", "", "Project root (auto-resolves .pacto/plans)")
	fs.StringVar(&values.plansRoot, "plans-root", "", "Deprecated: path to plans root (use --root)")
	fs.StringVar(&values.repoRoot, "repo-root", "", "Path to repository root for evidence verification")
	fs.StringVar(&values.mode, "mode", "compat", "Parsing mode: compat|strict")
	fs.StringVar(&values.lang, "lang", "", "Deprecated: ignored, CLI output is English-only")
	fs.StringVar(&values.format, "format", "table", "Output format: table|json")
	fs.StringVar(&values.configPath, "config", "", "Optional path to .pacto-engine.yaml")
	fs.StringVar(&values.failOn, "fail-on", "none", "Fail policy: none|unverified|partial|blocked")
	fs.StringVar(&values.state, "state", "all", "State filter: current|to-implement|done|outdated|all")
	fs.BoolVar(&values.includeArchive, "include-archive", false, "Include archive plans")
	fs.IntVar(&values.maxNext, "max-next-actions", 3, "Max next actions per plan")
	fs.IntVar(&values.maxBlockers, "max-blockers", 3, "Max blockers per plan")
	fs.BoolVar(&values.verbose, "verbose", false, "Print config and debug warnings")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.Usage()
			return statusFlagValues{}, nil, 0, false
		}
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		return statusFlagValues{}, nil, 2, false
	}
	if strings.TrimSpace(values.lang) != "" || hasLangArg(args) {
		fmt.Fprintln(os.Stderr, "warning: --lang is deprecated and ignored; CLI output is English-only")
	}

	provided := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { provided[f.Name] = true })
	return values, provided, 0, true
}

func buildStatusConfig(values statusFlagValues, provided map[string]bool) (config.Config, []string, []string, int, bool) {
	cwdAbs, err := filepath.Abs(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve cwd: %v\n", err)
		return config.Config{}, nil, nil, 2, false
	}
	cfg, cfgWarnings, err := config.Load(values.configPath, cwdAbs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		return config.Config{}, nil, nil, 2, false
	}

	applyOverrides(&cfg, provided, values.root, values.plansRoot, values.repoRoot, values.mode, values.format, values.failOn, values.state, values.includeArchive, values.maxNext, values.maxBlockers)
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

func buildStatusReport(cfg config.Config, cfgWarnings []string) (model.StatusReport, int, bool) {
	plans, err := discovery.FindPlans(cfg.PlansRoot, discovery.Options{StateFilter: cfg.State, IncludeArchive: cfg.IncludeArchive})
	if err != nil {
		fmt.Fprintf(os.Stderr, "discover plans: %v\n", err)
		return model.StatusReport{}, 3, false
	}

	parsed := make([]parser.ParsedPlan, 0, len(plans))
	claimsByPlan := map[string][]model.ClaimResult{}
	warningsByPlan := map[string][]string{}
	verifier := verify.New(cfg.RepoRoot, cfg.PlansRoot)
	claimOpts := claims.Options{Paths: cfg.ClaimsPaths, Symbols: cfg.ClaimsSymbols, Endpoints: cfg.ClaimsEndpoints, TestRefs: cfg.ClaimsTestRefs}

	for _, plan := range plans {
		pp, pErr := parser.ParsePlan(plan, cfg.Mode)
		if pErr != nil {
			pp.ParseError = pErr.Error()
		}
		parsed = append(parsed, pp)
		key := plan.State + "/" + plan.Slug
		rawClaims := claims.Extract(pp, claimOpts)
		verifiedClaims := make([]model.ClaimResult, 0, len(rawClaims))
		for _, c := range rawClaims {
			verifiedClaims = append(verifiedClaims, verifier.VerifyClaim(plan, c))
		}
		claimsByPlan[key] = verifiedClaims
		if len(cfgWarnings) > 0 {
			warningsByPlan[key] = append(warningsByPlan[key], cfgWarnings...)
		}
	}

	rep := analyze.Build(analyze.Input{
		Root:      cfg.PlansRoot,
		PlansRoot: cfg.PlansRoot,
		RepoRoot:  cfg.RepoRoot,
		Mode:      cfg.Mode,
		Plans:     parsed,
		Claims:    claimsByPlan,
		Warnings:  warningsByPlan,
	}, analyze.Options{MaxNextActions: cfg.MaxNextActions, MaxBlockers: cfg.MaxBlockers})
	return rep, 0, true
}

func hasLangArg(args []string) bool {
	for _, a := range args {
		if a == "--lang" || a == "-lang" || strings.HasPrefix(a, "--lang=") {
			return true
		}
	}
	return false
}

func applyOverrides(cfg *config.Config, provided map[string]bool, root, plansRoot, repoRoot, mode, format, failOn, state string, includeArchive bool, maxNext, maxBlockers int) {
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
