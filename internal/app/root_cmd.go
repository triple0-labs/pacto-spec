package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"pacto/internal/i18n"
)

var exitCode int

// Execute is the primary CLI entrypoint.
func Execute() int {
	return ExecuteArgs(os.Args[1:])
}

// ExecuteArgs runs pacto with explicit args for deterministic tests.
func ExecuteArgs(args []string) int {
	exitCode = 0
	setGlobalLangOverride("")

	normalized := normalizeLegacyLongFlags(args)
	rootCmd := newRootCommand(normalized)
	rootCmd.SetArgs(normalized)

	err := rootCmd.Execute()
	setGlobalLangOverride("")
	if err != nil {
		if exitCode != 0 {
			return exitCode
		}
		return 1
	}
	return exitCode
}

func newRootCommand(rawArgs []string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "pacto",
		Short: "Pacto: A Spec-Driven Development Framework",
		Long:  "Pacto is a specialized CLI tool for managing spec-driven development workspaces.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			lang, _ := cmd.Flags().GetString("lang")
			if lang != "" {
				if _, ok := i18n.ParseLanguage(lang); !ok {
					exitCode = 2
					cmd.SilenceUsage = true
					cmd.SilenceErrors = true
					fmt.Fprintf(os.Stderr, "invalid --lang value %q (allowed: en|es)\n", lang)
					return fmt.Errorf("invalid --lang")
				}
				setGlobalLangOverride(lang)
			}

			noColor, _ := cmd.Flags().GetBool("no-color")
			if noColor {
				os.Setenv("NO_COLOR", "1")
			}

			fullArgs := append([]string{"pacto"}, rawArgs...)
			_, allow := stripAllowGuardrailArg(fullArgs)
			if code, blocked := runGuardrailsIfNeeded(cmd.Name(), fullArgs, allow, false); blocked {
				exitCode = code
				cmd.SilenceUsage = true
				cmd.SilenceErrors = true
				return fmt.Errorf("guardrail blocked")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	rootCmd.PersistentFlags().String("lang", "", "Language override (en|es)")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable color output")
	rootCmd.PersistentFlags().StringSlice("allow-guardrail", []string{}, "Comma-separated list of guardrails to bypass")
	rootCmd.PersistentFlags().String("root", "", "Project root path (auto-discovered if omitted)")

	rootCmd.AddCommand(newExploreCommand())
	rootCmd.AddCommand(newMoveCommand())
	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newStatusCommand())
	rootCmd.AddCommand(newDoctorCommand())
	rootCmd.AddCommand(newInstallCommand())
	rootCmd.AddCommand(newUpdateCommand())
	rootCmd.AddCommand(newNewCommand())
	rootCmd.AddCommand(newExecCommand())
	rootCmd.AddCommand(newNormalizeCommand())
	rootCmd.AddCommand(newPluginCommand())
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print Pacto version.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(VersionLine())
		},
	})

	return rootCmd
}

func newExploreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explore [slug]",
		Short: "Create, update, show, or list ideas in the exploration workspace.",
		Run: func(cmd *cobra.Command, args []string) {
			forwarded := forwardChangedFlags(cmd, args, []string{"root", "title", "note", "list", "show"})
			exitCode = RunExplore(forwarded)
		},
	}
	cmd.Flags().String("title", "", "Optional idea title")
	cmd.Flags().String("note", "", "Append exploration note and refresh update timestamp")
	cmd.Flags().Bool("list", false, "List saved ideas")
	cmd.Flags().String("show", "", "Show a saved idea by slug")
	return cmd
}

func newMoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move <from-state> <slug> <to-state>",
		Short: "Move a plan to a different state directory.",
		Run: func(cmd *cobra.Command, args []string) {
			forwarded := forwardChangedFlags(cmd, args, []string{"root", "reason", "force"})
			exitCode = RunMove(forwarded)
		},
	}
	cmd.Flags().String("reason", "", "Optional reason to record in plan README")
	cmd.Flags().Bool("force", false, "Overwrite destination if it exists")
	return cmd
}

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Pacto workspace.",
		Run: func(cmd *cobra.Command, args []string) {
			forwarded := forwardChangedFlags(cmd, args, []string{"root", "with-agents", "force", "lang", "no-interactive", "tools", "yes", "no-install", "dry-run", "editor", "language"})
			exitCode = RunInit(forwarded)
		},
	}
	cmd.Flags().Bool("with-agents", false, "Create or update optional AGENTS.md hand-off block (canonical guidance remains in PACTO.md)")
	cmd.Flags().Bool("force", false, "Overwrite init-managed files when they already exist")
	cmd.Flags().Bool("no-interactive", false, "Disable Bubble Tea onboarding and use fallback profile resolution")
	cmd.Flags().String("tools", "", "Tools to configure during init: all, none, or comma-separated IDs (codex,cursor,claude,opencode)")
	cmd.Flags().Bool("yes", false, "Auto-approve install preview in interactive mode")
	cmd.Flags().Bool("no-install", false, "Skip skill/command installation during init")
	cmd.Flags().Bool("dry-run", false, "Show intended actions without writing files")
	cmd.Flags().String("editor", "", "Deprecated")
	_ = cmd.Flags().MarkHidden("editor")
	cmd.Flags().String("language", "", "Deprecated")
	_ = cmd.Flags().MarkHidden("language")
	return cmd
}

func newStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current workspace status.",
		Run: func(cmd *cobra.Command, args []string) {
			forwarded := forwardChangedFlags(cmd, args, []string{"root", "plans-root", "repo-root", "mode", "lang", "format", "config", "fail-on", "state", "include-archive", "max-next-actions", "max-blockers", "verbose"})
			exitCode = RunStatus(forwarded)
		},
	}
	cmd.Flags().String("plans-root", "", "Deprecated: path to plans root (use --root)")
	cmd.Flags().String("repo-root", "", "Path to repository root for evidence verification")
	cmd.Flags().String("mode", "compat", "Parsing mode: compat|strict")
	cmd.Flags().String("format", "table", "Output format: table|json")
	cmd.Flags().String("config", "", "Optional path to .pacto-engine.yaml")
	cmd.Flags().String("fail-on", "none", "Fail policy: none|unverified|partial|blocked")
	cmd.Flags().String("state", "all", "State filter: current|to-implement|done|outdated|all")
	cmd.Flags().Bool("include-archive", false, "Include archive plans")
	cmd.Flags().Int("max-next-actions", 3, "Max next actions per plan")
	cmd.Flags().Int("max-blockers", 3, "Max blockers per plan")
	cmd.Flags().Bool("verbose", false, "Print config and debug warnings")
	return cmd
}

func newInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install managed skills and command prompts.",
		Run: func(cmd *cobra.Command, args []string) {
			forwarded := forwardChangedFlags(cmd, args, []string{"root", "tools", "force"})
			exitCode = RunInstall(forwarded)
		},
	}
	cmd.Flags().String("tools", "", "Tools to configure: all, none, or comma-separated list (codex,cursor,claude,opencode)")
	cmd.Flags().Bool("force", false, "Overwrite unmanaged existing files")
	return cmd
}

func newDoctorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Audit managed integration artifacts for drift.",
		Run: func(cmd *cobra.Command, args []string) {
			forwarded := forwardChangedFlags(cmd, args, []string{"root", "tools", "format", "fail-on", "verbose"})
			exitCode = RunDoctor(forwarded)
		},
	}
	cmd.Flags().String("tools", "", "Tools to analyze: all, none, or comma-separated list (codex,cursor,claude,opencode)")
	cmd.Flags().String("format", "table", "Output format: table|json")
	cmd.Flags().String("fail-on", "none", "Failure policy: none|drift|legacy|any")
	cmd.Flags().Bool("verbose", false, "Print config and debug warnings")
	return cmd
}

func newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update pacto binary or refresh managed artifacts.",
		Run: func(cmd *cobra.Command, args []string) {
			forwarded := forwardChangedFlags(cmd, args, []string{"root", "artifacts", "tools", "force", "check", "yes", "version", "repo"})
			exitCode = RunUpdate(forwarded)
		},
	}
	cmd.Flags().Bool("artifacts", false, "Use legacy mode: refresh installed tool artifacts instead of updating pacto binary")
	cmd.Flags().String("tools", "", "Legacy artifacts mode only: tools to configure")
	cmd.Flags().Bool("force", false, "Legacy artifacts mode only: overwrite unmanaged files")
	cmd.Flags().Bool("check", false, "Check latest release and report status without installing")
	cmd.Flags().Bool("yes", false, "Auto-confirm binary replacement")
	cmd.Flags().String("version", "", "Install a specific release version (for example: v0.1.15)")
	cmd.Flags().String("repo", defaultUpdateRepo, "GitHub repo for releases (owner/name)")
	return cmd
}

func newNewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new <current|to-implement|done|outdated> <slug>",
		Short: "Create a new plan.",
		Run: func(cmd *cobra.Command, args []string) {
			forwarded := forwardChangedFlags(cmd, args, []string{"root", "title", "owner", "allow-minimal-root", "lang"})
			exitCode = RunNew(forwarded)
		},
	}
	cmd.Flags().String("title", "", "Optional plan title")
	cmd.Flags().String("owner", "Platform Team", "Owner for generated plan")
	cmd.Flags().Bool("allow-minimal-root", false, "Allow creating plans in lightweight/non-canonical roots")
	return cmd
}

func newExecCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec <state> <slug>",
		Short: "Execute a plan step.",
		Run: func(cmd *cobra.Command, args []string) {
			forwarded := forwardChangedFlags(cmd, args, []string{"root", "step", "note", "blocker", "evidence", "dry-run"})
			exitCode = RunExec(forwarded)
		},
	}
	cmd.Flags().String("step", "", "Target task id (e.g. 1.2)")
	cmd.Flags().String("note", "", "Append execution note")
	cmd.Flags().String("blocker", "", "Append blocker")
	cmd.Flags().String("evidence", "", "Append evidence reference")
	cmd.Flags().Bool("dry-run", false, "Show intended changes without writing files")
	return cmd
}

func newPluginCommand() *cobra.Command {
	pluginCmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage workspace plugins.",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		Run: func(cmd *cobra.Command, args []string) {
			forwarded := forwardChangedFlags(cmd, args, []string{"root", "format"})
			exitCode = runPluginList(forwarded)
		},
	}
	listCmd.Flags().String("format", "table", "Output format: table|json")

	listAvailableCmd := &cobra.Command{
		Use:   "list-available",
		Short: "List built-in installable plugins",
		Run: func(cmd *cobra.Command, args []string) {
			forwarded := forwardChangedFlags(cmd, args, []string{"format"})
			exitCode = runPluginListAvailable(forwarded)
		},
	}
	listAvailableCmd.Flags().String("format", "table", "Output format: table|json")

	installCmd := &cobra.Command{
		Use:   "install <id>",
		Short: "Install a built-in plugin",
		Run: func(cmd *cobra.Command, args []string) {
			forwarded := forwardChangedFlags(cmd, args, []string{"root", "force", "no-enable"})
			exitCode = runPluginInstall(forwarded)
		},
	}
	installCmd.Flags().Bool("force", false, "Overwrite existing plugin files")
	installCmd.Flags().Bool("no-enable", false, "Install files but do not auto-enable plugin")

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate discovered plugin manifests",
		Run: func(cmd *cobra.Command, args []string) {
			forwarded := forwardChangedFlags(cmd, args, []string{"root", "plugin"})
			exitCode = runPluginValidate(forwarded)
		},
	}
	validateCmd.Flags().String("plugin", "", "Validate a specific plugin ID")

	enableCmd := &cobra.Command{
		Use:   "enable <id>",
		Short: "Enable plugin",
		Run: func(cmd *cobra.Command, args []string) {
			forwarded := forwardChangedFlags(cmd, args, []string{"root"})
			exitCode = runPluginEnable(forwarded)
		},
	}

	disableCmd := &cobra.Command{
		Use:   "disable <id>",
		Short: "Disable plugin",
		Run: func(cmd *cobra.Command, args []string) {
			forwarded := forwardChangedFlags(cmd, args, []string{"root"})
			exitCode = runPluginDisable(forwarded)
		},
	}

	pluginCmd.AddCommand(listCmd, listAvailableCmd, installCmd, validateCmd, enableCmd, disableCmd)
	return pluginCmd
}

func newNormalizeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "normalize",
		Short: "Normalize plan structure and report compliance drift.",
		Run: func(cmd *cobra.Command, args []string) {
			forwarded := forwardChangedFlags(cmd, args, []string{"root", "state", "include-archive", "write", "format"})
			exitCode = RunNormalize(forwarded)
		},
	}
	cmd.Flags().String("state", "all", "State filter: current|to-implement|done|outdated|all")
	cmd.Flags().Bool("include-archive", false, "Include archive plans")
	cmd.Flags().Bool("write", false, "Apply normalized content to files (default is dry-run)")
	cmd.Flags().String("format", "table", "Output format: table|json")
	return cmd
}

func forwardChangedFlags(cmd *cobra.Command, pos []string, names []string) []string {
	out := make([]string, 0, len(pos)+len(names)*2)
	for _, name := range names {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			flag = cmd.InheritedFlags().Lookup(name)
		}
		if flag == nil || !flag.Changed {
			continue
		}

		switch flag.Value.Type() {
		case "bool":
			if flag.Value.String() == "true" {
				out = append(out, "--"+name)
			} else {
				out = append(out, fmt.Sprintf("--%s=false", name))
			}
		default:
			out = append(out, "--"+name, flag.Value.String())
		}
	}
	out = append(out, pos...)
	return out
}

func normalizeLegacyLongFlags(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		if strings.HasPrefix(a, "--") || !strings.HasPrefix(a, "-") || len(a) <= 2 {
			out = append(out, a)
			continue
		}

		// Preserve short flags and numeric values.
		first := a[1]
		if (first < 'a' || first > 'z') && (first < 'A' || first > 'Z') {
			out = append(out, a)
			continue
		}

		out = append(out, "--"+a[1:])
	}
	return out
}
