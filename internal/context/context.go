package context

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Overlap struct {
	Domain string
	Plans  []string
}

var reDomainItem = regexp.MustCompile(`^\s*[-*]\s+(.+?)\s*$`)

func ExtractDomains(specPath string) []string {
	b, err := os.ReadFile(specPath)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(b), "\n")
	inSection := false
	domains := make([]string, 0, 4)
	seen := map[string]struct{}{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isDomainsHeading(trimmed) {
			inSection = true
			continue
		}
		if !inSection {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			break
		}
		if trimmed == "" {
			continue
		}
		m := reDomainItem.FindStringSubmatch(line)
		if len(m) != 2 {
			continue
		}
		raw := strings.TrimSpace(strings.Trim(m[1], "`"))
		if isPlaceholderDomain(raw) {
			continue
		}
		slug := NormalizeDomainSlug(raw)
		if slug == "" {
			continue
		}
		if _, ok := seen[slug]; ok {
			continue
		}
		seen[slug] = struct{}{}
		domains = append(domains, slug)
	}

	return domains
}

func NormalizeDomainSlug(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" {
		return ""
	}

	var b strings.Builder
	lastDash := false
	for _, r := range v {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			lastDash = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if b.Len() == 0 || lastDash {
				continue
			}
			b.WriteByte('-')
			lastDash = true
		}
	}

	return strings.Trim(b.String(), "-")
}

func ReadContextDomains(contextDir string) []string {
	domainDir := filepath.Join(contextDir, "domains")
	entries, err := os.ReadDir(domainDir)
	if err != nil {
		return nil
	}

	domains := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		domains = append(domains, entry.Name())
	}
	sort.Strings(domains)
	return domains
}

func EnsureDomainFolders(contextDir string, domains []string) error {
	domainDir := filepath.Join(contextDir, "domains")
	if err := os.MkdirAll(domainDir, 0o775); err != nil {
		return fmt.Errorf("create context domains dir: %w", err)
	}

	unique := map[string]struct{}{}
	for _, domain := range domains {
		slug := NormalizeDomainSlug(domain)
		if slug == "" {
			continue
		}
		if _, ok := unique[slug]; ok {
			continue
		}
		unique[slug] = struct{}{}

		folderPath := filepath.Join(domainDir, slug)
		if err := ensureDomainFolder(folderPath, slug); err != nil {
			return err
		}
	}
	return nil
}

func DetectOverlaps(planDomains map[string][]string) []Overlap {
	domainPlans := map[string]map[string]struct{}{}
	for planRef, domains := range planDomains {
		for _, domain := range domains {
			slug := NormalizeDomainSlug(domain)
			if slug == "" {
				continue
			}
			if _, ok := domainPlans[slug]; !ok {
				domainPlans[slug] = map[string]struct{}{}
			}
			domainPlans[slug][planRef] = struct{}{}
		}
	}

	overlaps := make([]Overlap, 0)
	for domain, plans := range domainPlans {
		if len(plans) < 2 {
			continue
		}
		refs := make([]string, 0, len(plans))
		for ref := range plans {
			refs = append(refs, ref)
		}
		sort.Strings(refs)
		overlaps = append(overlaps, Overlap{Domain: domain, Plans: refs})
	}
	sort.Slice(overlaps, func(i, j int) bool {
		return overlaps[i].Domain < overlaps[j].Domain
	})
	return overlaps
}

func InitContext(plansRoot string) error {
	contextDir := ContextDirFromPlansRoot(plansRoot)
	if err := os.MkdirAll(filepath.Join(contextDir, "domains"), 0o775); err != nil {
		return fmt.Errorf("create context workspace: %w", err)
	}

	readmePath := filepath.Join(contextDir, "README.md")
	if _, err := os.Stat(readmePath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat context README: %w", err)
	}

	if err := os.WriteFile(readmePath, []byte(defaultContextReadme), 0o664); err != nil {
		return fmt.Errorf("write context README: %w", err)
	}
	return nil
}

func ContextDirFromPlansRoot(plansRoot string) string {
	return filepath.Join(filepath.Dir(filepath.Clean(plansRoot)), "context")
}

func DomainFolderPath(contextDir, domain string) string {
	return filepath.Join(contextDir, "domains", NormalizeDomainSlug(domain))
}

const defaultContextReadme = `# System Context

This directory holds the system source of truth for pacto.

## Structure

- ` + "`domains/`" + ` contains one folder per domain, for example ` + "`auth/`" + ` or ` + "`billing/`" + `
- each domain folder has ` + "`context.md`" + ` (bounded context) and ` + "`decisions.md`" + ` (ADRs)
- pacto creates domain folders mechanically when plans complete
- humans or agents enrich the files — pacto never overwrites them after creation

## Conventions

- Use lowercase dash-separated domain folder names
- ` + "`context.md`" + `: current-state snapshot — purpose, boundary, key terms, rules, collaborators
- ` + "`decisions.md`" + `: append-only history of architectural decisions (why things are the way they are)
`

func ensureDomainFolder(folderPath, slug string) error {
	if _, err := os.Stat(folderPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat domain folder %q: %w", folderPath, err)
	}

	if err := os.MkdirAll(folderPath, 0o775); err != nil {
		return fmt.Errorf("create domain folder %q: %w", folderPath, err)
	}

	stubs := []struct{ name, content string }{
		{"context.md", fmt.Sprintf("# %s\n\n<!-- Bounded context: purpose, boundary, key terms, rules, collaborators. -->\n", slug)},
		{"decisions.md", fmt.Sprintf("# %s – Decisions\n\n<!-- Architectural decisions (ADR-style): date, status, context, consequence. -->\n", slug)},
	}
	for _, f := range stubs {
		if err := os.WriteFile(filepath.Join(folderPath, f.name), []byte(f.content), 0o664); err != nil {
			return fmt.Errorf("write %s for domain %q: %w", f.name, slug, err)
		}
	}
	return nil
}

func isDomainsHeading(line string) bool {
	v := strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(line)), " "))
	return v == "## domains affected" || v == "## dominios afectados"
}

func isPlaceholderDomain(raw string) bool {
	v := strings.TrimSpace(strings.ToLower(raw))
	switch v {
	case "<domain>", "<dominio>", "<affected domain>", "<dominio afectado>":
		return true
	default:
		return false
	}
}
