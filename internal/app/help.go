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
			Usage:       "pacto status [--root .] [--mode compat|strict] [--format table|json] [--fail-on policy]",
			Description: "Scans plans from pacto states, extracts task/progress signals, verifies claims (paths/symbols/endpoints/test refs), and emits a consolidated report.",
			Examples: []string{
				"pacto status",
				"pacto status --mode strict --format table",
				"pacto status --format json --fail-on partial",
			},
		},
		{
			Name:        "new",
			Summary:     "Create a new plan scaffold and update root index.",
			Usage:       "pacto new <current|to-implement|done|outdated> <slug> [--title ...] [--owner ...] [--root .] [--allow-minimal-root]",
			Description: "Generates plan folder with README + PLAN file from template and updates root README counters, links, and last update date.",
			Examples: []string{
				"pacto new to-implement polling-contactos-v2",
				"pacto new current api-contract-refresh --title \"API Contract Refresh\" --owner \"Backend Team\"",
				"pacto new to-implement sandbox --root ./samples/mock-pacto-repo --allow-minimal-root",
			},
		},
		{
			Name:        "init",
			Summary:     "Initialize local Pacto workspace in .pacto/plans.",
			Usage:       "pacto init [--root .] [--with-agents] [--force]",
			Description: "Bootstraps a project-local planning workspace and optional AGENTS.md managed guidance block.",
			Examples: []string{
				"pacto init",
				"pacto init --root . --with-agents",
				"pacto init --force",
			},
		},
		{
			Name:        "exec",
			Summary:     "Execute plan slices and register deltas (planned).",
			Usage:       "pacto exec <path-to-plan-md>",
			Description: "Reserved command for guided phase/task execution with evidence and delta registration.",
			Examples: []string{
				"pacto exec ./current/agentes-y-herramientas/README.md",
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
