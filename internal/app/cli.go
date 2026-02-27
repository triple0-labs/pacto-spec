package app

import (
	"fmt"
	"os"
	"strings"
)

func Run(args []string) int {
	args, langUsed := stripDeprecatedLangArg(args)
	if langUsed {
		fmt.Fprintln(os.Stderr, "warning: --lang is deprecated and ignored; CLI output is English-only")
	}
	if len(args) == 0 {
		fmt.Print(RootHelp())
		return 0
	}

	cmd := strings.ToLower(strings.TrimSpace(args[0]))
	rest := args[1:]

	switch cmd {
	case "-h", "--help", "help":
		return runHelp(rest)
	case "-v", "--version", "version":
		fmt.Print(VersionLine())
		return 0
	case "status":
		if wantsHelp(rest) {
			fmt.Print(HelpFor("status"))
			return 0
		}
		return RunStatus(rest)
	case "new":
		if wantsHelp(rest) {
			fmt.Print(HelpFor("new"))
			return 0
		}
		return RunNew(rest)
	case "explore":
		if wantsHelp(rest) {
			fmt.Print(HelpFor("explore"))
			return 0
		}
		return RunExplore(rest)
	case "init":
		if wantsHelp(rest) {
			fmt.Print(HelpFor("init"))
			return 0
		}
		return RunInit(rest)
	case "install":
		if wantsHelp(rest) {
			fmt.Print(HelpFor("install"))
			return 0
		}
		return RunInstall(rest)
	case "update":
		if wantsHelp(rest) {
			fmt.Print(HelpFor("update"))
			return 0
		}
		return RunUpdate(rest)
	case "exec":
		if wantsHelp(rest) {
			fmt.Print(HelpFor("exec"))
			return 0
		}
		fmt.Fprint(os.Stderr, CommandPlannedMessage(cmd))
		return 2
	default:
		fmt.Fprint(os.Stderr, UnknownCommandMessage(cmd))
		fmt.Print(RootHelp())
		return 2
	}
}

func runHelp(args []string) int {
	if len(args) == 0 {
		fmt.Print(RootHelp())
		return 0
	}
	name := strings.ToLower(strings.TrimSpace(args[0]))
	h := HelpFor(name)
	if h == "" {
		fmt.Fprint(os.Stderr, UnknownHelpTopicMessage(name))
		fmt.Print(RootHelp())
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
