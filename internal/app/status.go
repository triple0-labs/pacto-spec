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
	"pacto/internal/verify"
)

func RunStatus(args []string) int {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  pacto status [options]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		fs.PrintDefaults()
	}
	root := fs.String("root", ".", "Path to pacto root")
	mode := fs.String("mode", "compat", "Parsing mode: compat|strict")
	lang := fs.String("lang", "", "Deprecated: ignored, CLI output is English-only")
	format := fs.String("format", "table", "Output format: table|json")
	cfgPath := fs.String("config", "", "Optional path to .pacto-engine.yaml")
	failOn := fs.String("fail-on", "none", "Fail policy: none|unverified|partial|blocked")
	state := fs.String("state", "all", "State filter: current|to-implement|done|outdated|all")
	includeArchive := fs.Bool("include-archive", false, "Include archive plans")
	maxNext := fs.Int("max-next-actions", 3, "Max next actions per plan")
	maxBlockers := fs.Int("max-blockers", 3, "Max blockers per plan")
	verbose := fs.Bool("verbose", false, "Print config and debug warnings")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.Usage()
			return 0
		}
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		return 2
	}
	if strings.TrimSpace(*lang) != "" || hasLangArg(args) {
		fmt.Fprintln(os.Stderr, "warning: --lang is deprecated and ignored; CLI output is English-only")
	}

	provided := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { provided[f.Name] = true })

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}
	defaultPlanRoot, foundPlanRoot := resolvePlanRoot(absRoot)

	cfg, cfgWarnings, err := config.Load(*cfgPath, absRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		return 2
	}

	applyOverrides(&cfg, provided, *root, *mode, *format, *failOn, *state, *includeArchive, *maxNext, *maxBlockers)
	if !provided["root"] && cfg.Root == absRoot && foundPlanRoot {
		cfg.Root = defaultPlanRoot
	}
	if resolved, ok := resolvePlanRoot(cfg.Root); ok {
		cfg.Root = resolved
	}
	if err := validateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "invalid config: %v\n", err)
		return 2
	}

	plans, err := discovery.FindPlans(cfg.Root, discovery.Options{StateFilter: cfg.State, IncludeArchive: cfg.IncludeArchive})
	if err != nil {
		fmt.Fprintf(os.Stderr, "discover plans: %v\n", err)
		return 3
	}

	parsed := make([]parser.ParsedPlan, 0, len(plans))
	claimsByPlan := map[string][]model.ClaimResult{}
	warningsByPlan := map[string][]string{}
	verifier := verify.New(cfg.Root)
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

	rep := analyze.Build(analyze.Input{Root: cfg.Root, Mode: cfg.Mode, Plans: parsed, Claims: claimsByPlan, Warnings: warningsByPlan}, analyze.Options{MaxNextActions: cfg.MaxNextActions, MaxBlockers: cfg.MaxBlockers})
	out, err := report.Render(rep, cfg.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render report: %v\n", err)
		return 3
	}
	fmt.Println(out)

	if *verbose {
		fmt.Fprintf(os.Stderr, "config: mode=%s format=%s fail-on=%s state=%s include-archive=%t\n", cfg.Mode, cfg.Format, cfg.FailOn, cfg.State, cfg.IncludeArchive)
	}
	return exitcode.Evaluate(cfg.FailOn, rep)
}

func hasLangArg(args []string) bool {
	for _, a := range args {
		if a == "--lang" || a == "-lang" || strings.HasPrefix(a, "--lang=") {
			return true
		}
	}
	return false
}

func applyOverrides(cfg *config.Config, provided map[string]bool, root, mode, format, failOn, state string, includeArchive bool, maxNext, maxBlockers int) {
	if provided["root"] {
		if abs, err := filepath.Abs(root); err == nil {
			cfg.Root = abs
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
