package specsbaseline

import "strings"

// ParseRequirementsText extracts Requirements from a plan-local
// `## Requirements` section in the given markdown text. Requirements are
// headings at the next level deeper than the section heading; Scenarios are
// one level deeper than the Requirement.
func ParseRequirementsText(text string) ([]Requirement, error) {
	lines := strings.Split(text, "\n")
	// Find the Requirements section.
	sectionLevel := 0
	startIdx := -1
	for i, line := range lines {
		lvl, head, ok := headingLevel(line)
		if !ok {
			continue
		}
		if isRequirementsSection(head) {
			sectionLevel = lvl
			startIdx = i + 1
			break
		}
	}
	if startIdx == -1 {
		return nil, nil
	}
	endIdx := len(lines)
	for i := startIdx; i < len(lines); i++ {
		lvl, _, ok := headingLevel(lines[i])
		if ok && lvl <= sectionLevel {
			endIdx = i
			break
		}
	}
	return parseRequirementBlocks(lines[startIdx:endIdx], sectionLevel+1)
}

// parseRequirementBlocks parses one or more Requirement blocks at reqLevel.
// Scenarios live at reqLevel+1.
func parseRequirementBlocks(lines []string, reqLevel int) ([]Requirement, error) {
	var out []Requirement
	seenNames := map[string]bool{}

	i := 0
	for i < len(lines) {
		lvl, head, ok := headingLevel(lines[i])
		if !ok {
			i++
			continue
		}
		if lvl < reqLevel {
			break
		}
		if lvl != reqLevel {
			// stray heading deeper than reqLevel without a parent Requirement
			i++
			continue
		}
		name, isReq := reqHeading(head)
		if !isReq {
			// Not a requirement heading at this level; skip.
			i++
			continue
		}
		if seenNames[strings.ToLower(name)] {
			return nil, &ParseError{Msg: "duplicate Requirement name: " + name, Line: i + 1}
		}
		seenNames[strings.ToLower(name)] = true

		// Collect body until next heading at lvl<=reqLevel (excluding scenarios).
		body := []string{}
		scenarios := []Scenario{}
		seenScn := map[string]bool{}
		j := i + 1
		for j < len(lines) {
			lvl2, head2, ok2 := headingLevel(lines[j])
			if ok2 && lvl2 <= reqLevel {
				break
			}
			if ok2 && lvl2 == reqLevel+1 {
				scnName, isScn := scnHeading(head2)
				if !isScn {
					return nil, &ParseError{Msg: "expected Scenario heading at level " + itoa(reqLevel+1) + ", got: " + head2, Line: j + 1}
				}
				if seenScn[strings.ToLower(scnName)] {
					return nil, &ParseError{Msg: "duplicate Scenario name: " + scnName, Line: j + 1}
				}
				seenScn[strings.ToLower(scnName)] = true
				// collect scenario body
				k := j + 1
				scnBody := []string{}
				for k < len(lines) {
					lvl3, _, ok3 := headingLevel(lines[k])
					if ok3 && lvl3 <= reqLevel+1 {
						break
					}
					scnBody = append(scnBody, lines[k])
					k++
				}
				scenarios = append(scenarios, Scenario{
					ID:    assignScnID(len(scenarios), extractIDOverride(scnBody)),
					Name:  scnName,
					Lines: scnBody,
				})
				j = k
				continue
			}
			if ok2 && lvl2 > reqLevel+1 {
				return nil, &ParseError{Msg: "scenario heading too deep at level " + itoa(lvl2), Line: j + 1}
			}
			body = append(body, lines[j])
			j++
		}
		req := Requirement{
			ID:         assignReqID(len(out), extractIDOverride(body)),
			Name:       name,
			Body:       body,
			Scenarios:  scenarios,
			HeaderLine: i + 1,
		}
		out = append(out, req)
		i = j
	}
	return out, nil
}

// ParseDeltasText scans the document for `## Capability: <slug>` blocks and
// nested delta op subsections.
func ParseDeltasText(text string) ([]Capability, error) {
	lines := strings.Split(text, "\n")
	var caps []Capability
	seenCaps := map[string]bool{}

	i := 0
	for i < len(lines) {
		lvl, head, ok := headingLevel(lines[i])
		if !ok {
			i++
			continue
		}
		slug, isCap := capHeading(head)
		if !isCap {
			i++
			continue
		}
		slug = strings.TrimSpace(slug)
		if slug == "" {
			return nil, &ParseError{Msg: "Capability heading missing slug", Line: i + 1}
		}
		if seenCaps[slug] {
			return nil, &ParseError{Msg: "duplicate Capability block: " + slug, Line: i + 1}
		}
		seenCaps[slug] = true

		capLevel := lvl
		// find end of this capability block
		end := len(lines)
		for j := i + 1; j < len(lines); j++ {
			lvl2, _, ok2 := headingLevel(lines[j])
			if ok2 && lvl2 <= capLevel {
				end = j
				break
			}
		}
		deltas, err := parseDeltaBlocks(lines[i+1:end], capLevel+1)
		if err != nil {
			return nil, err
		}
		caps = append(caps, Capability{Slug: slug, Deltas: deltas})
		i = end
	}
	return caps, nil
}

// parseDeltaBlocks parses `### ADDED Requirements` style subsections at
// opLevel within a Capability block.
func parseDeltaBlocks(lines []string, opLevel int) ([]Delta, error) {
	var out []Delta
	i := 0
	for i < len(lines) {
		lvl, head, ok := headingLevel(lines[i])
		if !ok {
			i++
			continue
		}
		if lvl < opLevel {
			break
		}
		if lvl != opLevel {
			i++
			continue
		}
		op, isOp := deltaOpHeading(head)
		if !isOp {
			return nil, &ParseError{Msg: "unknown delta op heading: " + head, Line: i + 1}
		}
		// find end of this delta op block
		end := len(lines)
		for j := i + 1; j < len(lines); j++ {
			lvl2, _, ok2 := headingLevel(lines[j])
			if ok2 && lvl2 <= opLevel {
				end = j
				break
			}
		}
		reqs, err := parseRequirementBlocks(lines[i+1:end], opLevel+1)
		if err != nil {
			return nil, err
		}
		if op == DeltaRenamed {
			if err := decorateRenamed(reqs); err != nil {
				return nil, err
			}
		}
		out = append(out, Delta{Op: op, Requirements: reqs})
		i = end
	}
	return out, nil
}

// decorateRenamed extracts the new name from RENAMED requirement bodies.
// Convention: a body line `- to: <new name>` (or `From: ... → To: ...` block)
// declares the new name. We accept the simple `to:` form.
func decorateRenamed(reqs []Requirement) error {
	for i := range reqs {
		newName := ""
		for _, l := range reqs[i].Body {
			t := strings.TrimSpace(l)
			t = strings.TrimPrefix(t, "-")
			t = strings.TrimPrefix(t, "*")
			t = strings.TrimSpace(t)
			low := strings.ToLower(t)
			if strings.HasPrefix(low, "to:") {
				newName = strings.TrimSpace(t[3:])
				break
			}
		}
		if newName == "" {
			return &ParseError{Msg: "RENAMED requirement missing `- to: <new name>` line: " + reqs[i].Name, Line: reqs[i].HeaderLine}
		}
		reqs[i].NewName = newName
	}
	return nil
}

// ParseError carries a human-readable message and the 1-based line where
// the problem was detected.
type ParseError struct {
	Msg  string
	Line int
}

func (e *ParseError) Error() string {
	if e.Line > 0 {
		return "specsbaseline: line " + itoa(e.Line) + ": " + e.Msg
	}
	return "specsbaseline: " + e.Msg
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
