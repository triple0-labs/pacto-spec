package app

import (
	"fmt"
	"strings"
)

type CommandHelp struct {
	Name        string
	Summary     string
	Usage       string
	Description string
	Examples    []string
}

func RootHelp() string {
	var b strings.Builder
	b.WriteString("Pacto CLI\n\n")
	b.WriteString("Usage:\n")
	b.WriteString("  pacto <command> [options]\n\n")
	b.WriteString("Commands:\n")
	for _, c := range commandCatalog() {
		b.WriteString("  " + padRight(c.Name, 8) + c.Summary + "\n")
	}
	b.WriteString("\n")
	b.WriteString("Use \"pacto help <command>\" for command details.\n")
	return b.String()
}

func HelpFor(name string) string {
	for _, c := range commandCatalog() {
		if c.Name == name {
			var b strings.Builder
			b.WriteString("Command: " + c.Name + "\n\n")
			b.WriteString("Summary:\n")
			b.WriteString("  " + c.Summary + "\n\n")
			b.WriteString("Usage:\n")
			b.WriteString("  " + c.Usage + "\n\n")
			if strings.TrimSpace(c.Description) != "" {
				b.WriteString("Description:\n")
				b.WriteString("  " + c.Description + "\n\n")
			}
			if len(c.Examples) > 0 {
				b.WriteString("Examples:\n")
				for _, ex := range c.Examples {
					b.WriteString("  " + ex + "\n")
				}
			}
			return b.String()
		}
	}
	return ""
}

func commandCatalog() []CommandHelp {
	return []CommandHelp{
		{
			Name:        "status",
			Summary:     "Verify plan status, blockers, and evidence claims.",
			Usage:       "pacto status [--root <path>] [--repo-root <path>] [--mode compat|strict] [--format table|json] [--fail-on policy]",
			Description: "Scans plans from plans root, extracts task/progress signals, verifies claims (paths/symbols/endpoints/test refs) against repo root, and emits a consolidated report. If roots are omitted, auto-discovers from current directory and parents.",
			Examples: []string{
				"pacto status",
				"pacto status # from nested directory",
				"pacto status --root . --repo-root .",
				"pacto status --mode strict --format table",
				"pacto status --format json --fail-on partial",
			},
		},
		{
			Name:        "new",
			Summary:     "Create a new plan scaffold and update root index.",
			Usage:       "pacto new <current|to-implement|done|outdated> <slug> [--title ...] [--owner ...] [--root <path>] [--allow-minimal-root]",
			Description: "Generates plan folder with README + PLAN file from template and updates root README counters, links, and last update date. If --root is omitted, auto-discovers from current directory and parents.",
			Examples: []string{
				"pacto new to-implement polling-contactos-v2",
				"pacto new to-implement polling-contactos-v2 # from nested directory",
				"pacto new current api-contract-refresh --title \"API Contract Refresh\" --owner \"Backend Team\"",
				"pacto new to-implement sandbox --root ./samples/mock-pacto-repo --allow-minimal-root",
			},
		},
		{
			Name:        "explore",
			Summary:     "Capture and revisit ideas without implementation.",
			Usage:       "pacto explore <slug> [--title ...] [--note ...] [--root <path>] | --list | --show <slug>",
			Description: "Creates and manages idea workspaces in .pacto/ideas with Created At and Updated At timestamps.",
			Examples: []string{
				"pacto explore auth-refresh --title \"Auth refresh ideas\"",
				"pacto explore auth-refresh --note \"Compare token vs session approach\"",
				"pacto explore --list",
				"pacto explore --show auth-refresh",
			},
		},
		{
			Name:        "init",
			Summary:     "Initialize local Pacto workspace in .pacto/plans.",
			Usage:       "pacto init [--root .] [--with-agents] [--force]",
			Description: "Bootstraps a project-local planning workspace. `--with-agents` adds an optional AGENTS.md hand-off block; canonical guidance remains in PACTO.md.",
			Examples: []string{
				"pacto init",
				"pacto init --root . --with-agents",
				"pacto init --force",
			},
		},
		{
			Name:        "install",
			Summary:     "Install Pacto skills and command prompts for AI tools.",
			Usage:       "pacto install [--tools <all|none|csv>] [--force]",
			Description: "Generates managed Pacto skills and command files for supported tools (codex,cursor,claude,opencode). If --tools is omitted, tools are auto-detected from project directories.",
			Examples: []string{
				"pacto install",
				"pacto install --tools codex,cursor",
				"pacto install --tools all",
			},
		},
		{
			Name:        "update",
			Summary:     "Refresh previously installed Pacto tool artifacts.",
			Usage:       "pacto update [--tools <all|none|csv>] [--force]",
			Description: "Refreshes managed Pacto blocks in generated skills and command files for supported tools.",
			Examples: []string{
				"pacto update",
				"pacto update --tools claude,opencode",
				"pacto update --force",
			},
		},
		{
			Name:        "exec",
			Summary:     "Execute plan tasks and append execution evidence.",
			Usage:       "pacto exec <current|to-implement|done|outdated> <slug> [--root <path>] [--step <task-id>] [--note <text>] [--blocker <text>] [--evidence <claim>] [--dry-run]",
			Description: "Runs guided execution updates on plan markdown artifacts only (no source-code edits). Execution is allowed only for plans in `current` state.",
			Examples: []string{
				"pacto exec current improve-auth-flow",
				"pacto exec current improve-auth-flow --step T3 --note \"Validated staging behavior\" --evidence src/auth/flow.go",
				"pacto exec current improve-auth-flow --dry-run",
			},
		},
		{
			Name:        "move",
			Summary:     "Move a plan slice between states.",
			Usage:       "pacto move <from-state> <slug> <to-state> [--root <path>] [--reason <text>] [--force]",
			Description: "Performs explicit state transitions (to-implement/current/done/outdated), updates plan README status, and refreshes plans index links/counts.",
			Examples: []string{
				"pacto move to-implement improve-auth-flow current",
				"pacto move current improve-auth-flow done --reason \"Tasks complete and evidence verified\"",
			},
		},
		{
			Name:        "help",
			Summary:     "Show root or command-specific help.",
			Usage:       "pacto help [command]",
			Description: "Displays command catalog and detailed command usage.",
			Examples: []string{
				"pacto help",
				"pacto help status",
			},
		},
		{
			Name:        "version",
			Summary:     "Show CLI version.",
			Usage:       "pacto version",
			Description: "Prints the current pacto CLI version.",
			Examples: []string{
				"pacto version",
				"pacto --version",
			},
		},
	}
}

func CommandPlannedMessage(cmd string) string {
	return fmt.Sprintf("command %q is planned but not implemented yet\nsee: pacto help %s\n", cmd, cmd)
}

func UnknownCommandMessage(cmd string) string {
	return fmt.Sprintf("unknown command: %s\n\n", cmd)
}

func UnknownHelpTopicMessage(topic string) string {
	return fmt.Sprintf("unknown help topic: %s\n\n", topic)
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s + " "
	}
	return s + strings.Repeat(" ", n-len(s))
}
