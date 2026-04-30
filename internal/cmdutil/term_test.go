package cmdutil

import (
	"bytes"
	"testing"
)

func TestIsTerminalReturnsFalseForNonFile(t *testing.T) {
	if IsTerminal(&bytes.Buffer{}) {
		t.Error("buffer is not a terminal")
	}
}
