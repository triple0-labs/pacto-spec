package main

import (
	"testing"

	"pacto/internal/app"
)

func TestRunMatchesSharedRouterExitCodes(t *testing.T) {
	cases := [][]string{
		{},
		{"version"},
		{"help"},
		{"unknown"},
		{"exec"},
	}
	for _, args := range cases {
		got := run(args)
		want := app.Run(args)
		if got != want {
			t.Fatalf("run(%v)=%d, want %d", args, got, want)
		}
	}
}
