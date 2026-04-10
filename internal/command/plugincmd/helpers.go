package plugincmd

import "pacto/internal/cmdutil"

func normalizeArgs(args []string, withValue map[string]bool) ([]string, error) {
	return cmdutil.NormalizeArgs(args, withValue)
}

func pathLine(kind, path string) string {
	return cmdutil.PathLine(kind, path)
}
