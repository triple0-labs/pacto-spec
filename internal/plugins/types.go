package plugins

import "time"

const (
	ManifestAPIVersion = "pacto/v1alpha1"
	ManifestKind       = "Plugin"
)

type Plugin struct {
	Dir      string
	Manifest Manifest
}

type Manifest struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

type Metadata struct {
	ID       string `yaml:"id"`
	Version  string `yaml:"version"`
	Priority int    `yaml:"priority"`
}

type Spec struct {
	CLIGuardrails   []CLIGuardrail   `yaml:"cliGuardrails"`
	AgentGuardrails []AgentGuardrail `yaml:"agentGuardrails"`
}

type CLIGuardrail struct {
	ID       string   `yaml:"id"`
	Commands []string `yaml:"commands"`
	Run      RunSpec  `yaml:"run"`
	OnFail   OnFail   `yaml:"onFail"`
}

type RunSpec struct {
	Script    string `yaml:"script"`
	TimeoutMS int    `yaml:"timeoutMs"`
}

type OnFail struct {
	Message string `yaml:"message"`
}

type AgentGuardrail struct {
	ID           string   `yaml:"id"`
	Tools        []string `yaml:"tools"`
	Workflows    []string `yaml:"workflows"`
	MarkdownFile string   `yaml:"markdownFile"`
}

type ActiveConfig struct {
	Enabled []string
}

type HookRequest struct {
	Command     string
	Args        []string
	ProjectRoot string
	Allow       map[string]bool
	Verbose     bool
}

type GuardrailViolation struct {
	PluginID    string
	GuardrailID string
	Message     string
	Script      string
	ExitCode    int
	TimedOut    bool
	Stdout      string
	Stderr      string
	Duration    time.Duration
	Allowed     bool
}

func (v GuardrailViolation) FullID() string {
	if v.PluginID == "" {
		return v.GuardrailID
	}
	return v.PluginID + "/" + v.GuardrailID
}

type AgentContribution struct {
	PluginID    string
	GuardrailID string
	Markdown    string
}

func (a AgentContribution) FullID() string {
	if a.PluginID == "" {
		return a.GuardrailID
	}
	return a.PluginID + "/" + a.GuardrailID
}
