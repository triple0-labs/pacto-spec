package main

import (
	"fmt"
	"os"
	"strings"

	"pacto/internal/app"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	args, langUsed := stripDeprecatedLangArg(args)
	if langUsed {
		fmt.Fprintln(os.Stderr, "warning: --lang is deprecated and ignored; CLI output is English-only")
	}
	if len(args) == 0 {
		fmt.Print(app.RootHelp())
		return 0
	}

	cmd := strings.ToLower(strings.TrimSpace(args[0]))
	rest := args[1:]

	switch cmd {
	case "-h", "--help", "help":
		return runHelp(rest)
	case "-v", "--version", "version":
		fmt.Print(app.VersionLine())
		return 0
	case "status":
		if wantsHelp(rest) {
			fmt.Print(app.HelpFor("status"))
			return 0
		}
		return app.RunStatus(rest)
	case "new":
		if wantsHelp(rest) {
			fmt.Print(app.HelpFor("new"))
			return 0
		}
		return app.RunNew(rest)
	case "init":
		if wantsHelp(rest) {
			fmt.Print(app.HelpFor("init"))
			return 0
		}
		return app.RunInit(rest)
	case "exec":
		if wantsHelp(rest) {
			fmt.Print(app.HelpFor("exec"))
			return 0
		}
		fmt.Fprint(os.Stderr, app.CommandPlannedMessage(cmd))
		return 2
	default:
		fmt.Fprint(os.Stderr, app.UnknownCommandMessage(cmd))
		fmt.Print(app.RootHelp())
		return 2
	}
}

func runHelp(args []string) int {
	if len(args) == 0 {
		fmt.Print(app.RootHelp())
		return 0
	}
	name := strings.ToLower(strings.TrimSpace(args[0]))
	h := app.HelpFor(name)
	if h == "" {
		fmt.Fprint(os.Stderr, app.UnknownHelpTopicMessage(name))
		fmt.Print(app.RootHelp())
		return 2
	}
	fmt.Print(h)
	return 0
}

func wantsHelp(args []string) bool {
	for _, a := range args {
		if a == "-h" || a == "--help" || a == "help" {
			return true
		}
	}
	return false
}

func stripDeprecatedLangArg(args []string) ([]string, bool) {
	used := false
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "--lang=") {
			used = true
			continue
		}
		if a == "--lang" || a == "-lang" {
			used = true
			if i+1 < len(args) {
				i++
			}
			continue
		}
		out = append(out, a)
	}
	return out, used
}
