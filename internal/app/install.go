package app

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"pacto/internal/integrations"
	"pacto/internal/ui"
)

const defaultUpdateRepo = "triple0-labs/pacto-spec"

var (
	selfUpdateAPIBase        = "https://api.github.com"
	selfUpdateDownloadBase   = "https://github.com"
	selfUpdateHTTPClient     = http.DefaultClient
	selfUpdateExecutablePath = os.Executable
	selfUpdateVersion        = func() string { return Version }
	selfUpdateGOOS           = runtime.GOOS
	selfUpdateGOARCH         = runtime.GOARCH
)

func RunInstall(args []string) int {
	return runInstallLike("install", args)
}

func RunUpdate(args []string) int {
	return runUpdateCommand(args)
}

func runInstallLike(cmd string, args []string) int {
	return runToolArtifactsCommand(cmd, args)
}

func runToolArtifactsCommand(cmd string, args []string) int {
	lang := effectiveLanguage("")
	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintf(os.Stderr, "  pacto %s [--tools <all|none|csv>] [--force]\n", cmd)
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		fs.PrintDefaults()
	}

	toolsArg := fs.String("tools", "", "Tools to configure: all, none, or comma-separated list (codex,cursor,claude,opencode)")
	force := fs.Bool("force", false, "Overwrite unmanaged existing files")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.Usage()
			return 0
		}
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr(lang, "parse flags", "error parseando flags"), err)
		return 2
	}
	if len(fs.Args()) > 0 {
		fs.Usage()
		return 2
	}

	cwd, err := filepath.Abs(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve cwd: %v\n", err)
		return 2
	}

	tools := make([]string, 0)
	if strings.TrimSpace(*toolsArg) != "" {
		parsed, err := integrations.ParseToolsArg(*toolsArg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 2
		}
		tools = parsed
	} else {
		detected, err := integrations.DetectTools(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", tr(lang, "detect tools", "detectar herramientas"), err)
			return 2
		}
		if len(detected) == 0 {
			fmt.Fprintf(os.Stderr, "%s --tools (%s)\n", tr(lang, "no tools detected. Use", "no se detectaron herramientas. Usa"), strings.Join(integrations.SupportedTools(), ","))
			return 2
		}
		tools = detected
	}

	if len(tools) == 0 {
		fmt.Println(ui.Dim(tr(lang, "No tools selected; nothing to do.", "No se seleccionaron herramientas; nada por hacer.")))
		return 0
	}

	fmt.Println(ui.ActionHeader(tr(lang, "Running "+cmd, "Ejecutando "+cmd), strings.Join(tools, ", ")))
	created := 0
	updated := 0
	skipped := 0
	failed := 0

	for _, toolID := range tools {
		results := integrations.GenerateForTool(cwd, toolID, *force)
		for _, r := range results {
			if r.Err != nil {
				failed++
				fmt.Fprintf(os.Stderr, "%s: tool=%s kind=%s workflow=%s: %v\n", tr(lang, "error", "error"), r.Tool, r.Kind, r.WorkflowID, r.Err)
				continue
			}
			switch r.Outcome {
			case integrations.OutcomeCreated:
				created++
				fmt.Println(pathLine("created", r.Path))
			case integrations.OutcomeUpdated:
				updated++
				fmt.Println(pathLine("updated", r.Path))
			case integrations.OutcomeSkipped:
				skipped++
				if r.Reason == "unmanaged_exists" {
					fmt.Fprintf(os.Stderr, "%s: %s: %s\n", tr(lang, "warning", "advertencia"), tr(lang, "skipped unmanaged file (use --force)", "archivo no gestionado omitido (usa --force)"), displayPath(r.Path))
				} else {
					fmt.Println(pathLine("skipped", r.Path))
				}
			}
		}
	}

	fmt.Printf("%s: %d  %s: %d  %s: %d  %s: %d\n", tr(lang, "Created", "Creado"), created, tr(lang, "Updated", "Actualizado"), updated, tr(lang, "Skipped", "Omitido"), skipped, tr(lang, "Failed", "Fallido"), failed)
	if failed > 0 {
		return 3
	}
	return 0
}

func runUpdateCommand(args []string) int {
	lang := effectiveLanguage("")
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  pacto update [--check] [--yes] [--version <vX.Y.Z>] [--repo <owner/repo>]")
		fmt.Fprintln(os.Stderr, "  pacto update --artifacts [--tools <all|none|csv>] [--force]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		fs.PrintDefaults()
	}

	artifacts := fs.Bool("artifacts", false, "Use legacy mode: refresh installed tool artifacts instead of updating pacto binary")
	toolsArg := fs.String("tools", "", "Legacy artifacts mode only: tools to configure")
	force := fs.Bool("force", false, "Legacy artifacts mode only: overwrite unmanaged files")
	checkOnly := fs.Bool("check", false, "Check latest release and report status without installing")
	yes := fs.Bool("yes", false, "Auto-confirm binary replacement")
	versionArg := fs.String("version", "", "Install a specific release version (for example: v0.1.15)")
	repoArg := fs.String("repo", defaultUpdateRepo, "GitHub repo for releases (owner/name)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.Usage()
			return 0
		}
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr(lang, "parse flags", "error parseando flags"), err)
		return 2
	}
	if len(fs.Args()) > 0 {
		fs.Usage()
		return 2
	}

	legacyMode := *artifacts || strings.TrimSpace(*toolsArg) != "" || *force
	if legacyMode {
		legacyArgs := make([]string, 0, 4)
		if strings.TrimSpace(*toolsArg) != "" {
			legacyArgs = append(legacyArgs, "--tools", strings.TrimSpace(*toolsArg))
		}
		if *force {
			legacyArgs = append(legacyArgs, "--force")
		}
		return runToolArtifactsCommand("update", legacyArgs)
	}

	opts := selfUpdateOptions{
		Repo:    strings.TrimSpace(*repoArg),
		Version: strings.TrimSpace(*versionArg),
		Check:   *checkOnly,
		Yes:     *yes,
	}
	return runSelfUpdate(opts)
}

type selfUpdateOptions struct {
	Repo    string
	Version string
	Check   bool
	Yes     bool
}

type releaseMeta struct {
	TagName string `json:"tag_name"`
}

func runSelfUpdate(opts selfUpdateOptions) int {
	lang := effectiveLanguage("")
	repo := strings.TrimSpace(opts.Repo)
	if repo == "" {
		repo = defaultUpdateRepo
	}
	current := normalizeReleaseVersion(selfUpdateVersion())
	target := normalizeReleaseVersion(opts.Version)
	var err error
	if target == "" {
		target, err = resolveLatestReleaseVersion(repo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", tr(lang, "resolve latest release", "resolver última versión"), err)
			return 3
		}
	}

	if opts.Check {
		if current == target {
			fmt.Printf("%s %s (%s)\n", tr(lang, "Already up to date:", "Ya está actualizado:"), target, repo)
		} else {
			fmt.Printf("%s %s  %s %s\n", tr(lang, "Update available:", "Actualización disponible:"), target, tr(lang, "current:", "actual:"), current)
		}
		return 0
	}
	if current == target {
		fmt.Printf("%s %s\n", tr(lang, "Already up to date:", "Ya está actualizado:"), target)
		return 0
	}

	exePath, err := selfUpdateExecutablePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr(lang, "resolve executable path", "resolver ruta del ejecutable"), err)
		return 3
	}
	if !opts.Yes {
		fmt.Printf("%s %s -> %s ? [Y/n]: ", tr(lang, "Update pacto binary", "Actualizar binario de pacto"), current, target)
		if !promptYesNo(true) {
			fmt.Fprintln(os.Stderr, tr(lang, "update cancelled", "actualización cancelada"))
			return 2
		}
	}

	artifactName := fmt.Sprintf("pacto_%s_%s_%s.tar.gz", target, selfUpdateGOOS, selfUpdateGOARCH)
	tag := "v" + target
	downloadURL := fmt.Sprintf("%s/%s/releases/download/%s/%s", strings.TrimRight(selfUpdateDownloadBase, "/"), repo, tag, artifactName)
	checksumURL := fmt.Sprintf("%s/%s/releases/download/%s/checksums.txt", strings.TrimRight(selfUpdateDownloadBase, "/"), repo, tag)

	archiveData, err := httpGetBytes(downloadURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr(lang, "download artifact", "descargar artefacto"), err)
		return 3
	}
	checksumData, err := httpGetBytes(checksumURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr(lang, "download checksums", "descargar checksums"), err)
		return 3
	}
	if err := verifyChecksum(artifactName, archiveData, checksumData); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr(lang, "verify checksum", "verificar checksum"), err)
		return 3
	}
	newBinary, err := extractPactoBinary(archiveData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr(lang, "extract artifact", "extraer artefacto"), err)
		return 3
	}
	if err := replaceBinaryAtomic(exePath, newBinary); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr(lang, "replace binary", "reemplazar binario"), err)
		return 3
	}

	fmt.Println(ui.ActionHeader(tr(lang, "Updated pacto", "Pacto actualizado"), target))
	fmt.Println(pathLine("updated", exePath))
	return 0
}

func resolveLatestReleaseVersion(repo string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/releases/latest", strings.TrimRight(selfUpdateAPIBase, "/"), repo)
	b, err := httpGetBytes(url)
	if err != nil {
		return "", err
	}
	var meta releaseMeta
	if err := json.Unmarshal(b, &meta); err != nil {
		return "", err
	}
	v := normalizeReleaseVersion(meta.TagName)
	if v == "" {
		return "", fmt.Errorf("missing tag_name in latest release response")
	}
	return v, nil
}

func normalizeReleaseVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(strings.ToLower(v), "v")
	return v
}

func httpGetBytes(url string) ([]byte, error) {
	resp, err := selfUpdateHTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return io.ReadAll(resp.Body)
}

func verifyChecksum(artifactName string, archiveData, checksumsData []byte) error {
	want := ""
	sc := bufio.NewScanner(strings.NewReader(string(checksumsData)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		name := strings.TrimPrefix(parts[1], "*")
		if name == artifactName {
			want = strings.ToLower(parts[0])
			break
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	if want == "" {
		return fmt.Errorf("checksum entry not found for %s", artifactName)
	}
	sum := sha256.Sum256(archiveData)
	got := hex.EncodeToString(sum[:])
	if got != want {
		return fmt.Errorf("checksum mismatch")
	}
	return nil
}

func extractPactoBinary(archiveData []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archiveData))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		base := filepath.Base(hdr.Name)
		if base != "pacto" {
			continue
		}
		return io.ReadAll(tr)
	}
	return nil, fmt.Errorf("pacto binary not found in archive")
}

func replaceBinaryAtomic(exePath string, newBinary []byte) error {
	dir := filepath.Dir(exePath)
	base := filepath.Base(exePath)
	tmpPath := filepath.Join(dir, base+".new")
	bakPath := filepath.Join(dir, base+".bak")
	if err := os.WriteFile(tmpPath, newBinary, 0o775); err != nil {
		return err
	}
	_ = os.Remove(bakPath)
	if err := os.Rename(exePath, bakPath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, exePath); err != nil {
		_ = os.Rename(bakPath, exePath)
		_ = os.Remove(tmpPath)
		return err
	}
	_ = os.Remove(bakPath)
	return nil
}
