package cmdutil

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func IsTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func PromptYesNo(defaultYes bool) bool {
	var raw string
	if _, err := fmt.Scanln(&raw); err != nil {
		return defaultYes
	}
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return defaultYes
	}
	switch s {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return defaultYes
	}
}
