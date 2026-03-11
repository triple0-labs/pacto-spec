package main

import (
	"testing"

	"pacto/internal/app"
)

func TestRunMatchesSharedRouterExitCodes(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want int
	}{
		{name: "root-help", args: []string{}, want: 0},
		{name: "version", args: []string{"version"}, want: 0},
		{name: "help", args: []string{"help"}, want: 0},
		{name: "unknown", args: []string{"unknown"}, want: 1},
		{name: "exec-missing-args", args: []string{"exec"}, want: 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := app.ExecuteArgs(tc.args)
			if got != tc.want {
				t.Fatalf("ExecuteArgs(%v) = %d, want %d", tc.args, got, tc.want)
			}
		})
	}
}
