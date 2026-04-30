package plugincmd

import "pacto/internal/cmdutil"

func pathLine(kind, path string) string {
	return cmdutil.PathLine(kind, path)
}
