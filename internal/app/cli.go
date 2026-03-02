package app

import (
	"fmt"
	"os"
	"strings"

	"pacto/internal/i18n"
)

func Run(args []string) int {
	args, langArg, langErr := stripGlobalLangArg(args)
	if langErr != nil {
		fmt.Fprintln(os.Stderr, langErr.Error())
		return 2
	}
	setGlobalLangOverride(langArg)
	defer setGlobalLangOverride("")
	args, allowGuardrails := stripAllowGuardrailArg(args)
	args, noColorUsed := stripNoColorArg(args)
	if noColorUsed {
		_ = os.Setenv("NO_COLOR", "1")
	}
	if len(args) == 0 {
		fmt.Print(RootHelpLang(effectiveLanguage("")))
		return 0
	}

	cmd := strings.ToLower(strings.TrimSpace(args[0]))
	rest := args[1:]
	lang := effectiveLanguage("")
	if code, handled := runGuardrailsIfNeeded(cmd, rest, allowGuardrails, hasVerboseArg(rest)); handled {
		return code
	}

	switch cmd {
	case "-h", "--help", "help":
		return runHelp(rest)
	case "-v", "--version", "version":
		fmt.Print(VersionLine())
		return 0
	case "status":
		if wantsHelp(rest) {
			fmt.Print(HelpForLang("status", lang))
			return 0
		}
		return RunStatus(rest)
	case "new":
		if wantsHelp(rest) {
			fmt.Print(HelpForLang("new", lang))
			return 0
		}
		return RunNew(rest)
	case "explore":
		if wantsHelp(rest) {
			fmt.Print(HelpForLang("explore", lang))
			return 0
		}
		return RunExplore(rest)
	case "init":
		if wantsHelp(rest) {
			fmt.Print(HelpForLang("init", lang))
			return 0
		}
		return RunInit(rest)
	case "install":
		if wantsHelp(rest) {
			fmt.Print(HelpForLang("install", lang))
			return 0
		}
		return RunInstall(rest)
	case "update":
		if wantsHelp(rest) {
			fmt.Print(HelpForLang("update", lang))
			return 0
		}
		return RunUpdate(rest)
	case "exec":
		if wantsHelp(rest) {
			fmt.Print(HelpForLang("exec", lang))
			return 0
		}
		return RunExec(rest)
	case "move":
		if wantsHelp(rest) {
			fmt.Print(HelpForLang("move", lang))
			return 0
		}
		return RunMove(rest)
	case "plugin":
		if wantsHelp(rest) {
			fmt.Print(HelpForLang("plugin", lang))
			return 0
		}
		return RunPlugin(rest)
	default:
		fmt.Fprint(os.Stderr, UnknownCommandMessage(cmd))
		fmt.Print(RootHelpLang(lang))
		return 2
	}
}

func runHelp(args []string) int {
	lang := effectiveLanguage("")
	if len(args) == 0 {
		fmt.Print(RootHelpLang(lang))
		return 0
	}
	name := strings.ToLower(strings.TrimSpace(args[0]))
	h := HelpForLang(name, lang)
	if h == "" {
		fmt.Fprint(os.Stderr, UnknownHelpTopicMessage(name))
		fmt.Print(RootHelpLang(lang))
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

func stripGlobalLangArg(args []string) ([]string, string, error) {
	lang := ""
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "--lang=") {
			lang = strings.TrimSpace(strings.TrimPrefix(a, "--lang="))
			if _, ok := i18n.ParseLanguage(lang); !ok {
				return nil, "", fmt.Errorf("invalid --lang value %q (allowed: en|es)", lang)
			}
			continue
		}
		if a == "--lang" || a == "-lang" {
			if i+1 >= len(args) {
				return nil, "", fmt.Errorf("flag --lang expects a value (en|es)")
			}
			lang = strings.TrimSpace(args[i+1])
			if _, ok := i18n.ParseLanguage(lang); !ok {
				return nil, "", fmt.Errorf("invalid --lang value %q (allowed: en|es)", lang)
			}
			i++
			continue
		}
		out = append(out, a)
	}
	return out, lang, nil
}

func stripNoColorArg(args []string) ([]string, bool) {
	used := false
	out := make([]string, 0, len(args))
	for _, a := range args {
		if a == "--no-color" {
			used = true
			continue
		}
		out = append(out, a)
	}
	return out, used
}

func hasVerboseArg(args []string) bool {
	for _, a := range args {
		if a == "--verbose" || strings.HasPrefix(a, "--verbose=") {
			return true
		}
	}
	return false
}
