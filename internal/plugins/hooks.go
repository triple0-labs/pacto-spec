package plugins

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func EvaluateGuardrails(active []Plugin, req HookRequest) []GuardrailViolation {
	violations := make([]GuardrailViolation, 0)
	for _, p := range active {
		pid := p.Manifest.Metadata.ID
		for _, g := range p.Manifest.Spec.CLIGuardrails {
			if !matchesCommand(req.Command, g.Commands) {
				continue
			}
			fullID := pid + "/" + g.ID
			allowed := req.Allow[fullID] || req.Allow[g.ID]
			v := runHook(p.Dir, pid, g, req)
			if v == nil {
				continue
			}
			v.Allowed = allowed
			violations = append(violations, *v)
		}
	}
	sort.Slice(violations, func(i, j int) bool {
		if violations[i].PluginID == violations[j].PluginID {
			return violations[i].GuardrailID < violations[j].GuardrailID
		}
		return violations[i].PluginID < violations[j].PluginID
	})
	return violations
}

func runHook(pluginDir, pluginID string, g CLIGuardrail, req HookRequest) *GuardrailViolation {
	scriptPath := filepath.Clean(filepath.Join(pluginDir, g.Run.Script))
	timeout := g.Run.TimeoutMS
	if timeout < 500 {
		timeout = 500
	}
	if timeout > 60000 {
		timeout = 60000
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/bin/sh", scriptPath)
	cmd.Dir = req.ProjectRoot
	cmd.Env = append(cmd.Environ(),
		"PACTO_PLUGIN_ID="+pluginID,
		"PACTO_GUARDRAIL_ID="+g.ID,
		"PACTO_COMMAND="+req.Command,
		"PACTO_PROJECT_ROOT="+req.ProjectRoot,
		"PACTO_ARGS="+strings.Join(req.Args, " "),
	)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	start := time.Now()
	err := cmd.Run()
	dur := time.Since(start)
	if err == nil {
		return nil
	}
	v := &GuardrailViolation{
		PluginID:    pluginID,
		GuardrailID: g.ID,
		Message:     strings.TrimSpace(g.OnFail.Message),
		Script:      g.Run.Script,
		Stdout:      trimOutput(outb.String()),
		Stderr:      trimOutput(errb.String()),
		Duration:    dur,
	}
	if v.Message == "" {
		v.Message = fmt.Sprintf("guardrail %s failed", v.FullID())
	}
	if ctx.Err() == context.DeadlineExceeded {
		v.TimedOut = true
		v.ExitCode = 124
		return v
	}
	if ee, ok := err.(*exec.ExitError); ok {
		v.ExitCode = ee.ExitCode()
	} else {
		v.ExitCode = 1
	}
	return v
}

func matchesCommand(command string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, c := range allowed {
		if strings.EqualFold(strings.TrimSpace(c), command) {
			return true
		}
	}
	return false
}

func trimOutput(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 4000 {
		return s[:4000]
	}
	return s
}
