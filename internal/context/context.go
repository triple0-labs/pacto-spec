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
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".md" {
			continue
		}
		domains = append(domains, strings.TrimSuffix(name, ".md"))
	}
	sort.Strings(domains)
	return domains
}

func EnsureDomainDocs(contextDir string, domains []string, planRef string) error {
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

		path := filepath.Join(domainDir, slug+".md")
		if err := ensureDomainDoc(path, slug, planRef); err != nil {
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

func DomainDocPath(contextDir, domain string) string {
	return filepath.Join(contextDir, "domains", NormalizeDomainSlug(domain)+".md")
}

const defaultContextReadme = `# System Context

This directory holds the system source of truth for pacto.

## Structure

- ` + "`domains/`" + ` contains one markdown file per domain, for example ` + "`auth.md`" + ` or ` + "`billing.md`" + `
- pacto creates domain docs mechanically from completed plans
- humans or agents enrich those docs with decisions, constraints, and notable references

## Conventions

- Use lowercase dash-separated domain filenames
- Keep one domain per file
- Preserve manual notes; pacto should only create missing scaffolds or append missing references safely
`

func ensureDomainDoc(path, slug, planRef string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.WriteFile(path, []byte(renderDomainDoc(slug, planRef)), 0o664)
	} else if err != nil {
		return fmt.Errorf("stat domain doc %q: %w", path, err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read domain doc %q: %w", path, err)
	}

	updated := ensureDomainSections(string(b), slug, planRef)
	if updated == string(b) {
		return nil
	}
	if err := os.WriteFile(path, []byte(updated), 0o664); err != nil {
		return fmt.Errorf("write domain doc %q: %w", path, err)
	}
	return nil
}

func renderDomainDoc(slug, planRef string) string {
	return fmt.Sprintf(`# Domain: %s

## Summary

<!-- What this domain owns. -->

## Related Plans

- %s

## Decisions

<!-- Add key architectural choices for this domain. -->

## Constraints

<!-- Add rules future plans must respect. -->
`, slug, planRef)
}

func ensureDomainSections(text, slug, planRef string) string {
	updated := strings.ReplaceAll(text, "\r\n", "\n")
	if strings.TrimSpace(updated) == "" {
		return renderDomainDoc(slug, planRef)
	}
	if !strings.Contains(updated, "# Domain:") {
		updated = strings.TrimSpace(renderDomainDoc(slug, planRef)) + "\n\n" + strings.TrimSpace(updated)
	}
	updated = ensureSection(updated, "## Summary", "<!-- What this domain owns. -->")
	updated = ensureRelatedPlan(updated, planRef)
	updated = ensureSection(updated, "## Decisions", "<!-- Add key architectural choices for this domain. -->")
	updated = ensureSection(updated, "## Constraints", "<!-- Add rules future plans must respect. -->")
	if !strings.HasSuffix(updated, "\n") {
		updated += "\n"
	}
	return updated
}

func ensureSection(text, heading, body string) string {
	if strings.Contains(text, heading) {
		return text
	}
	trimmed := strings.TrimRight(text, "\n")
	return trimmed + "\n\n" + heading + "\n\n" + body + "\n"
}

func ensureRelatedPlan(text, planRef string) string {
	heading := "## Related Plans"
	if !strings.Contains(text, heading) {
		trimmed := strings.TrimRight(text, "\n")
		return trimmed + "\n\n" + heading + "\n\n- " + planRef + "\n"
	}

	lines := strings.Split(text, "\n")
	start := -1
	end := len(lines)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == heading {
			start = i
			continue
		}
		if start >= 0 && i > start && strings.HasPrefix(trimmed, "## ") {
			end = i
			break
		}
	}
	if start < 0 {
		return text
	}

	needle := "- " + planRef
	for i := start + 1; i < end; i++ {
		if strings.TrimSpace(lines[i]) == needle {
			return text
		}
	}

	insertAt := end
	for i := end - 1; i > start; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			insertAt = i + 1
			break
		}
	}
	if insertAt < start+1 {
		insertAt = start + 1
	}

	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertAt]...)
	if insertAt == start+1 && strings.TrimSpace(lines[start+1]) != "" {
		newLines = append(newLines, "")
	}
	newLines = append(newLines, needle)
	newLines = append(newLines, lines[insertAt:]...)
	return strings.Join(newLines, "\n")
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
