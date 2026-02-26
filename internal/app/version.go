package app

import "fmt"

// Version is injected at build time via -ldflags.
var Version = "dev"

func VersionLine() string {
	return fmt.Sprintf("pacto version %s\n", Version)
}
