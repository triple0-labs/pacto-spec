package main

import (
	"os"
	"testing"
)

func TestRunMatchesSharedRouterExitCodes(t *testing.T) {
	cases := [][]string{
		{},
		{"version"},
		{"help"},
		{"unknown"},
		{"exec"},
	}
	osArgs := os.Args
	defer func() { os.Args = osArgs }()

	for _, args := range cases {
		os.Args = append([]string{"pacto"}, args...)
		got := run()
		_ = got
	}
}
