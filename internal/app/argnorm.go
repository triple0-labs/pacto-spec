package app

import (
	"fmt"
	"strings"
)

func normalizeArgs(args []string, withValue map[string]bool) ([]string, error) {
	flags := make([]string, 0, len(args))
	pos := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "--") || strings.HasPrefix(a, "-") {
			if withValue[a] {
				if i+1 >= len(args) {
					return nil, fmt.Errorf("flag %s expects a value", a)
				}
				flags = append(flags, a, args[i+1])
				i++
				continue
			}
			flags = append(flags, a)
			continue
		}
		pos = append(pos, a)
	}
	return append(flags, pos...), nil
}
