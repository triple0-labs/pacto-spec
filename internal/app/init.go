package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pacto/internal/assets"
	"pacto/internal/i18n"
	"pacto/internal/integrations"
	"pacto/internal/onboarding"
	initui "pacto/internal/tui/init"
	"pacto/internal/ui"
)

const (
	agentsManagedStart = "<!-- pacto:init:start -->"
	agentsManagedEnd   = "<!-- pacto:init:end -->"
)

type initOptions struct {
	root          string
	withAgents    bool
	force         bool
	lang          string
	noInteractive bool
	tools         string
	yes           bool
	noInstall     bool
	dryRun        bool
}

func RunInit(args []string) int {
	opts, code, ok := parseInitOptions(args)
	if !ok {
		return code
	}
	if strings.TrimSpace(opts.lang) != "" {
		setGlobalLangOverride(opts.lang)
		defer setGlobalLangOverride("")
	}

	projectRoot, err := filepath.Abs(opts.root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}
	plansRoot := filepath.Join(projectRoot, ".pacto", "plans")

	base := onboarding.DetectProfile(projectRoot)
	answered := onboarding.Profile{}
	interactive := !opts.noInteractive && isTerminal(os.Stdin) && isTerminal(os.Stdout)
	if interactive {
		ans, confirmed, err := initui.Run(base)
		if err != nil {
			fmt.Fprintf(os.Stderr, "run init wizard: %v\n", err)
			return 3
		}
		if !confirmed {
			fmt.Fprintln(os.Stderr, "init cancelled")
			return 2
		}
		answered = ans
	}
	profile, err := onboarding.ResolveProfile(base, answered, onboarding.Overrides{ToolsCSV: opts.tools})
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve init profile: %v\n", err)
		return 2
	}
	if strings.TrimSpace(opts.lang) != "" {
	}
	resolvedUILang := i18n.NormalizeLanguage(profile.UILanguage)
	if strings.TrimSpace(opts.lang) != "" {
		resolvedUILang = i18n.NormalizeLanguage(opts.lang)
	} else if strings.TrimSpace(profile.UILanguage) == "" {
		resolvedUILang = effectiveLanguage(projectRoot)
	}
	profile.UILanguage = string(resolvedUILang)
	if !interactive {
		applyInitFallbacks(&profile)
	}
	lang := i18n.NormalizeLanguage(profile.UILanguage)
	validation := onboarding.ValidateProfile(profile)
	if len(validation.Errors) > 0 {
		fmt.Fprintln(os.Stderr, tr(lang, "init profile is incomplete:", "el perfil de init está incompleto:"))
		for _, msg := range validation.Errors {
			fmt.Fprintf(os.Stderr, "  - %s\n", msg)
		}
		if opts.noInteractive {
			fmt.Fprintln(os.Stderr, tr(lang, "rerun without --no-interactive to complete onboarding prompts", "ejecuta nuevamente sin --no-interactive para completar el onboarding"))
		}
		return 2
	}
	for _, warn := range validation.Warnings {
		fmt.Fprintf(os.Stderr, "%s: %s\n", tr(lang, "warning", "advertencia"), warn)
	}

	if opts.dryRun {
		fmt.Println(ui.ActionHeader(tr(lang, "Init Dry Run", "Simulación de Init"), displayPath(projectRoot)))
		technologies := append([]string{}, profile.Languages...)
		technologies = append(technologies, profile.CustomLanguages...)
		fmt.Printf("technologies=%s tools=%s\n", strings.Join(technologies, ","), strings.Join(profile.Tools, ","))
		fmt.Println(pathLine("created", plansRoot))
		fmt.Println(pathLine("updated", filepath.Join(projectRoot, ".pacto", "config.yaml")))
		fmt.Println(pathLine("updated", filepath.Join(projectRoot, "prd.md")))
		if opts.withAgents {
			fmt.Println(pathLine("updated", filepath.Join(projectRoot, "AGENTS.md")))
		}
		return 0
	}

	var created, updated, skipped []string
	if err := bootstrapWorkspace(plansRoot, profile.UILanguage, opts.force, &created, &updated, &skipped); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 3
	}

	cfgPath := filepath.Join(projectRoot, ".pacto", "config.yaml")
	cfgExisted := pathExists(cfgPath)
	cfgWritten, err := onboarding.WriteConfig(projectRoot, profile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "write .pacto/config.yaml: %v\n", err)
		return 3
	}
	if cfgExisted {
		updated = append(updated, cfgWritten)
	} else {
		created = append(created, cfgWritten)
	}

	prdPath := filepath.Join(projectRoot, "prd.md")
	prdExisted := pathExists(prdPath)
	writtenPRD, prdChanged, err := onboarding.WritePRD(projectRoot, profile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "write prd.md: %v\n", err)
		return 3
	}
	if prdChanged {
		if prdExisted {
			updated = append(updated, writtenPRD)
		} else {
			created = append(created, writtenPRD)
		}
	} else {
		skipped = append(skipped, writtenPRD)
	}

	if opts.withAgents {
		agentsPath := filepath.Join(projectRoot, "AGENTS.md")
		act, aerr := writeAgentsManagedBlock(agentsPath, assets.MustTemplateLang(profile.UILanguage, "AGENTS.md"))
		if aerr != nil {
			fmt.Fprintf(os.Stderr, "update AGENTS.md: %v\n", aerr)
			return 3
		}
		switch act {
		case "created":
			created = append(created, agentsPath)
		case "updated":
			updated = append(updated, agentsPath)
		case "skipped":
			skipped = append(skipped, agentsPath)
		}
	}

	if !opts.noInstall && len(profile.Tools) > 0 {
		approve := true
		if interactive && !opts.yes {
			fmt.Printf("%s %s ? [Y/n]: ", tr(lang, "Install tool artifacts for:", "¿Instalar artefactos para herramientas:"), strings.Join(profile.Tools, ", "))
			approve = promptYesNo(true)
		}
		if approve {
			installCreated, installUpdated, installSkipped, installFailed := applyInstallPlan(projectRoot, profile.Tools, opts.force)
			created = append(created, installCreated...)
			updated = append(updated, installUpdated...)
			skipped = append(skipped, installSkipped...)
			if len(installFailed) > 0 {
				for _, e := range installFailed {
					fmt.Fprintf(os.Stderr, "install error: %s\n", e)
				}
				return 3
			}
		}
	}

	sort.Strings(created)
	sort.Strings(updated)
	sort.Strings(skipped)

	printInitSummary(lang, plansRoot, profile, created, updated, skipped)
	return 0
}

func parseInitOptions(args []string) (initOptions, int, bool) {
	if flagName := removedInitFlag(args); flagName != "" {
		fmt.Fprintf(os.Stderr, "flag --%s is no longer supported; use interactive onboarding for technologies and install targets\n", flagName)
		return initOptions{}, 2, false
	}

	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  pacto init [--root .] [--with-agents] [--force] [--tools <all|none|csv>] [--no-interactive] [--yes] [--no-install] [--dry-run]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		fs.PrintDefaults()
	}

	opts := initOptions{}
	fs.StringVar(&opts.root, "root", ".", "Project root path")
	fs.BoolVar(&opts.withAgents, "with-agents", false, "Create or update optional AGENTS.md hand-off block (canonical guidance remains in PACTO.md)")
	fs.BoolVar(&opts.force, "force", false, "Overwrite init-managed files when they already exist")
	fs.StringVar(&opts.lang, "lang", "", "Output language override: en|es")
	fs.BoolVar(&opts.noInteractive, "no-interactive", false, "Disable Bubble Tea onboarding and use fallback profile resolution")
	fs.StringVar(&opts.tools, "tools", "", "Tools to configure during init: all, none, or comma-separated IDs (codex,cursor,claude,opencode)")
	fs.BoolVar(&opts.yes, "yes", false, "Auto-approve install preview in interactive mode")
	fs.BoolVar(&opts.noInstall, "no-install", false, "Skip skill/command installation during init")
	fs.BoolVar(&opts.dryRun, "dry-run", false, "Show intended actions without writing files")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.Usage()
			return initOptions{}, 0, false
		}
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		return initOptions{}, 2, false
	}
	if strings.TrimSpace(opts.lang) != "" {
		if _, ok := i18n.ParseLanguage(opts.lang); !ok {
			fmt.Fprintf(os.Stderr, "invalid --lang value %q (allowed: en|es)\n", opts.lang)
			return initOptions{}, 2, false
		}
	}
	if len(fs.Args()) > 0 {
		fs.Usage()
		return initOptions{}, 2, false
	}
	return opts, 0, true
}

func applyInitFallbacks(profile *onboarding.Profile) {
	if profile == nil {
		return
	}
	if len(profile.Languages) == 0 && len(profile.CustomLanguages) == 0 {
		profile.Languages = []string{"unknown"}
		if strings.TrimSpace(profile.Sources.Languages) == "" || profile.Sources.Languages == "auto" {
			profile.Sources.Languages = "fallback"
		}
	}
	if strings.TrimSpace(profile.Intents.Problem) == "" {
		profile.Intents.Problem = "TODO: define the core problem"
	}
	if strings.TrimSpace(profile.UILanguage) == "" {
		profile.UILanguage = string(i18n.English)
	}
}

func removedInitFlag(args []string) string {
	for _, arg := range args {
		if arg == "--editor" || strings.HasPrefix(arg, "--editor=") {
			return "editor"
		}
		if arg == "--language" || strings.HasPrefix(arg, "--language=") {
			return "language"
		}
	}
	return ""
}

func bootstrapWorkspace(plansRoot, lang string, force bool, created, updated, skipped *[]string) error {
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		p := filepath.Join(plansRoot, st)
		if info, err := os.Stat(p); err == nil {
			if !info.IsDir() {
				return fmt.Errorf("state path exists but is not a directory: %s", p)
			}
			*skipped = append(*skipped, p)
			continue
		}
		if err := os.MkdirAll(p, 0o775); err != nil {
			return fmt.Errorf("create state dir %q: %w", p, err)
		}
		*created = append(*created, p)
	}

	workspaceFiles := map[string]string{
		filepath.Join(plansRoot, "README.md"):               assets.MustTemplateLang(lang, "README.md"),
		filepath.Join(plansRoot, "PACTO.md"):                assets.MustTemplateLang(lang, "PACTO.md"),
		filepath.Join(plansRoot, "PLANTILLA_PACTO_PLAN.md"): assets.MustTemplateLang(lang, "PLANTILLA_PACTO_PLAN.md"),
		filepath.Join(plansRoot, "SLASH_COMMANDS.md"):       assets.MustTemplateLang(lang, "SLASH_COMMANDS.md"),
	}

	for path, content := range workspaceFiles {
		wc, wu, ws, werr := writeManagedFile(path, content, force)
		if werr != nil {
			return fmt.Errorf("write file %q: %w", path, werr)
		}
		if wc {
			*created = append(*created, path)
		}
		if wu {
			*updated = append(*updated, path)
		}
		if ws {
			*skipped = append(*skipped, path)
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

func applyInstallPlan(projectRoot string, tools []string, force bool) (created []string, updated []string, skipped []string, failed []string) {
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

func promptYesNo(defaultYes bool) bool {
	var raw string
	if _, err := fmt.Scanln(&raw); err != nil {
		return defaultYes
	}
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return defaultYes
	}
	switch s {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return defaultYes
	}
}

func printInitSummary(lang i18n.Language, plansRoot string, profile onboarding.Profile, created, updated, skipped []string) {
	fmt.Println(ui.ActionHeader(tr(lang, "Workspace Ready", "Workspace listo"), displayPath(plansRoot)))
	fmt.Println(tr(lang, "Setup complete. Here's what changed:", "Configuración completada. Esto cambió:"))
	fmt.Printf("  %s: +%d  ~%d  =%d\n", tr(lang, "Files/Folders", "Archivos/Directorios"), len(created), len(updated), len(skipped))
	fmt.Printf("  %s: %s\n", tr(lang, "Language", "Idioma"), readableLanguage(lang))
	fmt.Printf("  %s: %s\n", tr(lang, "Technologies", "Tecnologías"), joinOrNone(append(append([]string{}, profile.Languages...), profile.CustomLanguages...), lang))
	fmt.Printf("  %s: %s\n", tr(lang, "Tools", "Herramientas"), joinOrNone(profile.Tools, lang))
	fmt.Println("")

	printPathGroup(lang, tr(lang, "Created", "Creados"), "created", created)
	printPathGroup(lang, tr(lang, "Updated", "Actualizados"), "updated", updated)
	printPathGroup(lang, tr(lang, "Unchanged", "Sin cambios"), "skipped", skipped)

	fmt.Println("")
	fmt.Printf("%s: %s\n", tr(lang, "Next", "Siguiente paso"), tr(lang, "run `pacto status` to inspect your plans", "ejecuta `pacto status` para revisar tus planes"))
}

func printPathGroup(lang i18n.Language, label, action string, paths []string) {
	if len(paths) == 0 {
		return
	}
	fmt.Printf("%s (%d)\n", label, len(paths))
	for _, p := range paths {
		fmt.Println(pathLine(action, p))
	}
}

func joinOrNone(items []string, lang i18n.Language) string {
	if len(items) == 0 {
		return tr(lang, "none", "ninguna")
	}
	return strings.Join(items, ",")
}

func readableLanguage(lang i18n.Language) string {
	switch lang {
	case i18n.Spanish:
		return "es (Español)"
	default:
		return "en (English)"
	}
}
