package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pacto/internal/plugins"
)

func shouldRunGuardrails(cmd string, args []string) bool {
	if wantsHelp(args) {
		return false
	}
	switch cmd {
	case "init", "new", "move", "install", "update":
		if hasBoolFlag(args, "--dry-run") {
			return false
		}
		return true
	case "exec":
		return !hasBoolFlag(args, "--dry-run")
	case "explore":
		if hasBoolFlag(args, "--list") || hasStringFlag(args, "--show") {
			return false
		}
		return true
	case "status":
		return true
	default:
		return false
	}
}

func hasBoolFlag(args []string, key string) bool {
	for _, a := range args {
		if a == key {
			return true
		}
		if strings.HasPrefix(a, key+"=") {
			v := strings.TrimSpace(strings.TrimPrefix(a, key+"="))
			return v == "" || strings.EqualFold(v, "true") || v == "1" || strings.EqualFold(v, "yes")
		}
	}
	return false
}

func hasStringFlag(args []string, key string) bool {
	for i, a := range args {
		if a == key && i+1 < len(args) {
			return true
		}
		if strings.HasPrefix(a, key+"=") {
			return true
		}
	}
	return false
}

func findProjectRootForPlugins(cwd string) (string, bool) {
	cur := cleanAbs(cwd)
	for {
		pactoDir := filepath.Join(cur, ".pacto")
		if info, err := os.Stat(pactoDir); err == nil && info.IsDir() {
			return cur, true
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	return "", false
}

func runGuardrailsIfNeeded(cmd string, args []string, allow map[string]bool, verbose bool) (int, bool) {
	if !shouldRunGuardrails(cmd, args) {
		return 0, false
	}
	cwd, err := filepath.Abs(".")
	if err != nil {
		return 0, false
	}
	projectRoot, ok := findProjectRootForPlugins(cwd)
	if !ok {
		return 0, false
	}
	active, errs := plugins.LoadActive(projectRoot)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "plugin error: %v\n", e)
		}
		return 3, true
	}
	if len(active) == 0 {
		return 0, false
	}
	violations := plugins.EvaluateGuardrails(active, plugins.HookRequest{
		Command:     cmd,
		Args:        args,
		ProjectRoot: projectRoot,
		Allow:       allow,
		Verbose:     verbose,
	})
	blocked := false
	for _, v := range violations {
		if v.Allowed {
			continue
		}
		blocked = true
		status := "failed"
		if v.TimedOut {
			status = "timed out"
		}
		fmt.Fprintf(os.Stderr, "guardrail blocked: %s (%s)\n", v.FullID(), status)
		if strings.TrimSpace(v.Message) != "" {
			fmt.Fprintf(os.Stderr, "  %s\n", strings.TrimSpace(v.Message))
		}
		if verbose {
			if strings.TrimSpace(v.Stdout) != "" {
				fmt.Fprintf(os.Stderr, "  stdout:\n%s\n", v.Stdout)
			}
			if strings.TrimSpace(v.Stderr) != "" {
				fmt.Fprintf(os.Stderr, "  stderr:\n%s\n", v.Stderr)
			}
		}
	}
	if blocked {
		fmt.Fprintln(os.Stderr, "Use --allow-guardrail <id> to bypass specific guardrails for this run.")
		return 3, true
	}
	return 0, false
}

func stripAllowGuardrailArg(args []string) ([]string, map[string]bool) {
	allow := map[string]bool{}
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "--allow-guardrail=") {
			for _, id := range strings.Split(strings.TrimPrefix(a, "--allow-guardrail="), ",") {
				id = strings.ToLower(strings.TrimSpace(id))
				if id != "" {
					allow[id] = true
				}
			}
			continue
		}
		if a == "--allow-guardrail" {
			if i+1 < len(args) {
				for _, id := range strings.Split(args[i+1], ",") {
					id = strings.ToLower(strings.TrimSpace(id))
					if id != "" {
						allow[id] = true
					}
				}
				i++
				continue
			}
			continue
		}
		out = append(out, a)
	}
	return out, allow
}
