package testutil

import (
	"io"
	"os"
	"testing"
)

func CaptureOutput(t *testing.T, fn func()) (stdout string, stderr string) {
	t.Helper()
	oldOut := os.Stdout
	oldErr := os.Stderr

	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = outW
	os.Stderr = errW
	defer func() {
		os.Stdout = oldOut
		os.Stderr = oldErr
	}()

	fn()

	_ = outW.Close()
	_ = errW.Close()

	outB, err := io.ReadAll(outR)
	if err != nil {
		t.Fatal(err)
	}
	errB, err := io.ReadAll(errR)
	if err != nil {
		t.Fatal(err)
	}
	return string(outB), string(errB)
}
