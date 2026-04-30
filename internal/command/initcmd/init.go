package initcmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pacto/internal/app/initws"
	"pacto/internal/i18n"
	"pacto/internal/onboarding"
	initui "pacto/internal/tui/init"
	"pacto/internal/ui"
)

// Re-exported here for the package-internal test suite which references the
// AGENTS.md managed-block markers directly.
const (
	agentsManagedStart = "<!-- pacto:init:start -->"
	agentsManagedEnd   = "<!-- pacto:init:end -->"
)

type Options struct {
	Root          string
	WithAgents    bool
	Force         bool
	Lang          string
	NoInteractive bool
	Tools         string
	Yes           bool
	NoInstall     bool
	DryRun        bool
}

func Run(opts Options) int {
	if strings.TrimSpace(opts.Root) == "" {
		opts.Root = "."
	}
	if strings.TrimSpace(opts.Lang) != "" {
		if _, ok := i18n.ParseLanguage(opts.Lang); !ok {
			fmt.Fprintf(os.Stderr, "invalid --lang value %q (allowed: en|es)\n", opts.Lang)
			return 2
		}
	}
	if strings.TrimSpace(opts.Lang) != "" {
		setGlobalLangOverride(opts.Lang)
		defer setGlobalLangOverride("")
	}

	projectRoot, err := filepath.Abs(opts.Root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}
	plansRoot := filepath.Join(projectRoot, ".pacto", "plans")

	base := onboarding.DetectProfile(projectRoot)
	answered := onboarding.Profile{}
	interactive := !opts.NoInteractive && isTerminal(os.Stdin) && isTerminal(os.Stdout)
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
	profile, err := onboarding.ResolveProfile(base, answered, onboarding.Overrides{ToolsCSV: opts.Tools})
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve init profile: %v\n", err)
		return 2
	}
	resolvedUILang := i18n.NormalizeLanguage(profile.UILanguage)
	if strings.TrimSpace(opts.Lang) != "" {
		resolvedUILang = i18n.NormalizeLanguage(opts.Lang)
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
		if opts.NoInteractive {
			fmt.Fprintln(os.Stderr, tr(lang, "rerun without --no-interactive to complete onboarding prompts", "ejecuta nuevamente sin --no-interactive para completar el onboarding"))
		}
		return 2
	}
	for _, warn := range validation.Warnings {
		fmt.Fprintf(os.Stderr, "%s: %s\n", tr(lang, "warning", "advertencia"), warn)
	}

	if opts.DryRun {
		fmt.Println(ui.ActionHeader(tr(lang, "Init Dry Run", "Simulación de Init"), displayPath(projectRoot)))
		technologies := append([]string{}, profile.Languages...)
		technologies = append(technologies, profile.CustomLanguages...)
		fmt.Printf("technologies=%s tools=%s\n", strings.Join(technologies, ","), strings.Join(profile.Tools, ","))
		fmt.Println(pathLine("created", plansRoot))
		fmt.Println(pathLine("updated", filepath.Join(projectRoot, ".pacto", "config.yaml")))
		fmt.Println(pathLine("updated", filepath.Join(projectRoot, "prd.md")))
		if opts.WithAgents {
			fmt.Println(pathLine("updated", filepath.Join(projectRoot, "AGENTS.md")))
		}
		return 0
	}

	// Decide whether to defer the install step behind a confirmation prompt.
	skipInstallNow := opts.NoInstall
	deferredInstall := false
	if !opts.NoInstall && interactive && !opts.Yes && len(profile.Tools) > 0 {
		fmt.Printf("%s %s ? [Y/n]: ", tr(lang, "Install tool artifacts for:", "¿Instalar artefactos para herramientas:"), strings.Join(profile.Tools, ", "))
		if promptYesNo(true) {
			deferredInstall = true
		}
		// In either case we don't want the use case to install yet — we
		// either run it ourselves below (deferredInstall) or skip entirely.
		skipInstallNow = true
	}

	res, err := initws.Apply(initws.Input{
		ProjectRoot: projectRoot,
		Profile:     profile,
		WithAgents:  opts.WithAgents,
		Force:       opts.Force,
		NoInstall:   skipInstallNow,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		if errors.Is(err, initws.ErrInvalid) {
			return 2
		}
		return 3
	}
	if len(res.InstallFailed) > 0 {
		for _, e := range res.InstallFailed {
			fmt.Fprintf(os.Stderr, "install error: %s\n", e)
		}
		return 3
	}

	if deferredInstall {
		c, u, s, f := initws.ApplyInstall(projectRoot, profile.Tools, opts.Force)
		res.Created = append(res.Created, c...)
		res.Updated = append(res.Updated, u...)
		res.Skipped = append(res.Skipped, s...)
		if len(f) > 0 {
			for _, e := range f {
				fmt.Fprintf(os.Stderr, "install error: %s\n", e)
			}
			return 3
		}
	}

	sort.Strings(res.Created)
	sort.Strings(res.Updated)
	sort.Strings(res.Skipped)

	printInitSummary(lang, res.PlansRoot, profile, res.Created, res.Updated, res.Skipped)
	return 0
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

	if len(profile.Tools) > 0 {
		fmt.Println(ui.ActionHeader(tr(lang, "IDE Setup", "Configuración de IDE"), ""))
		for _, tool := range profile.Tools {
			switch tool {
			case "cursor":
				fmt.Println("  - Cursor: Use Composer (Ctrl+I / Cmd+I) or mention @pacto to start working with your plans.")
			case "claude":
				fmt.Println("  - Claude: Attach your project root to Claude Desktop to grant it access to the Pacto skills.")
			case "codex":
				fmt.Println("  - Codex: Use the Codex sidebar or terminal integration to run actions.")
			case "opencode":
				fmt.Println("  - OpenCode: Use the OpenCode assistant to interact with the project context.")
			case "other":
				fmt.Println("  - Custom IDE: Configure your assistant to read the .pacto/plans directory.")
			}
		}
		fmt.Println("")
	}

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
