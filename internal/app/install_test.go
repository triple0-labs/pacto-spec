package app

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInstallExplicitToolCreatesArtifacts(t *testing.T) {
	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	stdout, _ := captureOutput(t, func() {
		if code := RunInstall([]string{"--tools", "cursor"}); code != 0 {
			t.Fatalf("RunInstall returned %d", code)
		}
	})
	if !strings.Contains(stdout, "Created:") {
		t.Fatalf("expected summary in stdout, got %q", stdout)
	}

	assertExists(t, filepath.Join(root, ".cursor", "skills", "pacto-status", "SKILL.md"))
	assertExists(t, filepath.Join(root, ".cursor", "commands", "pacto-status.md"))
	assertExists(t, filepath.Join(root, ".cursor", "skills", "pacto-new", "SKILL.md"))
	assertExists(t, filepath.Join(root, ".cursor", "commands", "pacto-new.md"))
}

func TestRunInstallAutoDetectsOpenCode(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".opencode"), 0o755); err != nil {
		t.Fatal(err)
	}
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	if code := RunInstall([]string{}); code != 0 {
		t.Fatalf("RunInstall returned %d", code)
	}
	assertExists(t, filepath.Join(root, ".opencode", "skills", "pacto-status", "SKILL.md"))
	assertExists(t, filepath.Join(root, ".opencode", "commands", "pacto-status.md"))
}

func TestRunUpdateSkipsUnmanagedWithoutForce(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".cursor", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ".cursor", "commands", "pacto-status.md")
	if err := os.WriteFile(path, []byte("custom\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	_, stderr := captureOutput(t, func() {
		if code := RunUpdate([]string{"--tools", "cursor"}); code != 0 {
			t.Fatalf("RunUpdate returned %d", code)
		}
	})
	if !strings.Contains(stderr, "skipped unmanaged file") {
		t.Fatalf("expected unmanaged warning, got %q", stderr)
	}
}

func TestRunUpdateCheckReportsAvailableVersion(t *testing.T) {
	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/test/repo/releases/latest" {
			_, _ = w.Write([]byte(`{"tag_name":"v9.9.9"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	oldAPI := selfUpdateAPIBase
	oldHTTP := selfUpdateHTTPClient
	oldVersion := Version
	selfUpdateAPIBase = ts.URL
	selfUpdateHTTPClient = ts.Client()
	Version = "0.1.0"
	defer func() {
		selfUpdateAPIBase = oldAPI
		selfUpdateHTTPClient = oldHTTP
		Version = oldVersion
	}()

	stdout, _ := captureOutput(t, func() {
		if code := RunUpdate([]string{"--check", "--repo", "test/repo"}); code != 0 {
			t.Fatalf("RunUpdate returned %d", code)
		}
	})
	if !strings.Contains(stdout, "Update available:") {
		t.Fatalf("expected update available output, got %q", stdout)
	}
}

func TestRunUpdateDefaultInstallsBinary(t *testing.T) {
	root := t.TempDir()
	exePath := filepath.Join(root, "pacto")
	if err := os.WriteFile(exePath, []byte("old-binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	newBinary := []byte("new-binary")
	archive := buildTarGz(t, map[string][]byte{"pacto": newBinary})
	artifact := "pacto_1.2.3_linux_amd64.tar.gz"
	sum := sha256.Sum256(archive)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), artifact)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/test/repo/releases/latest":
			_, _ = w.Write([]byte(`{"tag_name":"v1.2.3"}`))
		case "/test/repo/releases/download/v1.2.3/" + artifact:
			_, _ = w.Write(archive)
		case "/test/repo/releases/download/v1.2.3/checksums.txt":
			_, _ = w.Write([]byte(checksums))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	oldAPI := selfUpdateAPIBase
	oldDL := selfUpdateDownloadBase
	oldHTTP := selfUpdateHTTPClient
	oldExe := selfUpdateExecutablePath
	oldOS := selfUpdateGOOS
	oldArch := selfUpdateGOARCH
	oldVersion := Version
	selfUpdateAPIBase = ts.URL
	selfUpdateDownloadBase = ts.URL
	selfUpdateHTTPClient = ts.Client()
	selfUpdateExecutablePath = func() (string, error) { return exePath, nil }
	selfUpdateGOOS = "linux"
	selfUpdateGOARCH = "amd64"
	Version = "0.1.0"
	defer func() {
		selfUpdateAPIBase = oldAPI
		selfUpdateDownloadBase = oldDL
		selfUpdateHTTPClient = oldHTTP
		selfUpdateExecutablePath = oldExe
		selfUpdateGOOS = oldOS
		selfUpdateGOARCH = oldArch
		Version = oldVersion
	}()

	stdout, stderr := captureOutput(t, func() {
		if code := RunUpdate([]string{"--yes", "--repo", "test/repo"}); code != 0 {
			t.Fatalf("RunUpdate returned %d", code)
		}
	})
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "Updated pacto") {
		t.Fatalf("expected updated output, got %q", stdout)
	}
	got, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(newBinary) {
		t.Fatalf("expected binary replaced, got %q", string(got))
	}
}

func buildTarGz(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		hdr := &tar.Header{Name: name, Mode: 0o755, Size: int64(len(content))}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected path to exist: %s (%v)", path, err)
	}
}
