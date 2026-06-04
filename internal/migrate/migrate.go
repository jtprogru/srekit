// Package migrate converts legacy `.tmpl` files (and their optional
// `.sections.yaml` sidecars) into the v1 single-file `<name>.yaml`
// artifact format introduced in v0.14.0. It powers
// `srekit templates migrate` and is intentionally heuristic rather than
// a full Go-template parser — sections we can't simplify confidently are
// preserved verbatim inside `git merge`-style diff markers so a human
// can resolve them.
package migrate

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"

	yaml "go.yaml.in/yaml/v3"

	"github.com/jtprogru/srekit/internal/sections"
)

// Convert produces v1 artifact YAML from a legacy .tmpl source body and
// (optionally) the body of a sibling `.sections.yaml`.
//
// When sectionsManifest is non-empty and parseable, its section list is
// copied into the result verbatim — this is the postmortem-style v0.13.x
// → v1 path where the sidecar already has the canonical section layout
// and the .tmpl was only carrying the header.
//
// Otherwise sections are detected by splitting on `^## ` boundaries.
// Each section's type is inferred:
//   - body contains a GFM table → type:table (columns + rows extracted),
//   - body contains Go-template control flow (`{{ if|range|with }}`) →
//     type:text wrapped in diff markers (human review required),
//   - everything else → type:text with default_body verbatim.
//
// Section IDs are slugified from the English portion of a "Russian
// (English)" heading if present, otherwise from the whole heading.
func Convert(tmplBody, sectionsManifest []byte) ([]byte, error) {
	src := string(tmplBody)
	fm, src := extractFrontmatter(src)
	title, src := extractH1(src)
	bullets, src := extractMetaBullets(src)
	headerBody, sectionsSrc := splitOnFirstH2(src)

	var secs []sections.Section
	if len(sectionsManifest) > 0 {
		m, err := sections.ParseManifest(sectionsManifest)
		if err == nil {
			secs = m.Sections
			// When the sidecar provides sections, the .tmpl body between
			// header and the first `##` was the `{{ range .Sections }}`
			// loop that rendered them — Go-template code, not prose.
			// Keeping it in header_body would corrupt the v1 output.
			headerBody = ""
		}
	}
	if len(secs) == 0 {
		secs = parseSectionsFromTmpl(sectionsSrc)
	}
	if len(secs) == 0 {
		return nil, errors.New("no `## ` sections detected in source; nothing to migrate")
	}

	art := sections.Artifact{
		Version:     1,
		Frontmatter: fm,
		Title:       title,
		MetaBullets: bullets,
		HeaderBody:  strings.TrimSpace(headerBody),
		Sections:    secs,
	}
	return encodeArtifactYAML(&art)
}

// encodeArtifactYAML serializes an Artifact to YAML with deterministic
// quoting choices. Plain `yaml.Marshal(art)` won't do — template-syntax
// values (`{{ .X }}`) need double quotes or yaml.v3 either errors on
// parse-back or emits ambiguous output. We hand-assemble the top-level
// blocks and only delegate the more complex `sections` to yaml.Encoder
// (which handles their nested types cleanly because we don't mix bare
// template expressions there).
func encodeArtifactYAML(art *sections.Artifact) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("version: 1\n")

	if art.Frontmatter.Kind == yaml.MappingNode && len(art.Frontmatter.Content) > 0 {
		buf.WriteString("\nfrontmatter:\n")
		fmBytes, err := marshalNodeIndented(&art.Frontmatter, 2)
		if err != nil {
			return nil, fmt.Errorf("encode frontmatter: %w", err)
		}
		buf.Write(fmBytes)
	}

	if art.Title != "" {
		buf.WriteString("\ntitle: ")
		buf.WriteString(quoteForYAML(art.Title))
		buf.WriteString("\n")
	}

	if len(art.MetaBullets) > 0 {
		buf.WriteString("\nmeta_bullets:\n")
		for _, b := range art.MetaBullets {
			buf.WriteString("  - ")
			buf.WriteString(quoteForYAML(b))
			buf.WriteString("\n")
		}
	}

	if art.HeaderBody != "" {
		buf.WriteString("\nheader_body: |\n")
		for _, line := range strings.Split(strings.TrimRight(art.HeaderBody, "\n"), "\n") {
			buf.WriteString("  ")
			buf.WriteString(line)
			buf.WriteString("\n")
		}
	}

	if len(art.Sections) > 0 {
		var secBuf bytes.Buffer
		enc := yaml.NewEncoder(&secBuf)
		enc.SetIndent(2)
		if err := enc.Encode(map[string]any{"sections": art.Sections}); err != nil {
			return nil, fmt.Errorf("encode sections: %w", err)
		}
		if err := enc.Close(); err != nil {
			return nil, fmt.Errorf("close sections encoder: %w", err)
		}
		buf.WriteString("\n")
		buf.Write(secBuf.Bytes())
	}

	return buf.Bytes(), nil
}

// marshalNodeIndented serializes a yaml.Node and indents every output
// line by `indent` spaces so the result can be nested under a parent key
// without losing yaml.v3's per-node quoting/style decisions.
func marshalNodeIndented(n *yaml.Node, indent int) ([]byte, error) {
	var raw bytes.Buffer
	enc := yaml.NewEncoder(&raw)
	enc.SetIndent(2)
	if err := enc.Encode(n); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	pad := strings.Repeat(" ", indent)
	var out bytes.Buffer
	for _, line := range strings.Split(strings.TrimRight(raw.String(), "\n"), "\n") {
		out.WriteString(pad)
		out.WriteString(line)
		out.WriteString("\n")
	}
	return out.Bytes(), nil
}

// yamlNeedsQuoteRE matches values containing characters that YAML plain
// scalars don't reliably handle: `{`, `}` (flow-mapping delimiters),
// leading `-`/`?`/`*`, leading whitespace, or a `:` followed by space
// (which would parse as a key-value pair).
var yamlNeedsQuoteRE = regexp.MustCompile(`(\{|\}|\[|\])|^[\s\-?*]|^$|:\s`)

// quoteForYAML returns a string ready to drop into a YAML value position.
// Strings that look risky (template syntax, leading control chars, etc.)
// get double-quoted with internal `"` escaped; safe strings pass through
// verbatim so simple values stay readable.
func quoteForYAML(s string) string {
	if yamlNeedsQuoteRE.MatchString(s) {
		return `"` + strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`) + `"`
	}
	return s
}

// extractFrontmatter pulls the YAML block between leading `---` / `---`
// markers and parses it into a yaml.Node (preserving author key order).
// Returns the node + the rest of the source. If the source has no
// frontmatter block, returns a zero-value Node + the original source.
//
// Legacy .tmpl files have bare Go-template expressions in YAML values
// (`id: {{ .ID }}`), which yaml.v3 would mis-parse as flow-mapping
// objects. preQuoteTemplateValues wraps such values in double quotes
// before parsing so they round-trip cleanly.
func extractFrontmatter(src string) (yaml.Node, string) {
	if !strings.HasPrefix(src, "---\n") {
		return yaml.Node{}, src
	}
	rest := src[len("---\n"):]
	closeIdx := strings.Index(rest, "\n---\n")
	if closeIdx < 0 {
		return yaml.Node{}, src
	}
	fmStr := preQuoteTemplateValues(rest[:closeIdx])
	after := rest[closeIdx+len("\n---\n"):]

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(fmStr), &node); err != nil {
		return yaml.Node{}, src
	}
	// yaml.Unmarshal wraps the root in a DocumentNode; unwrap to the
	// underlying MappingNode so it round-trips cleanly through encode.
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = *node.Content[0]
	}
	return node, after
}

// bareTemplateValueRE matches `key: {{ ... }}` lines where the value is
// an unquoted Go-template expression. These need to be wrapped in quotes
// before yaml.v3 parses them, or the parser interprets `{{ }}` as a
// flow-mapping (object) and produces nonsense.
var bareTemplateValueRE = regexp.MustCompile(`(?m)^(\s*[\w_-]+:[ \t]+)(\{\{[^\n]+?\}\})([ \t]*)$`)

func preQuoteTemplateValues(src string) string {
	return bareTemplateValueRE.ReplaceAllString(src, `${1}"${2}"${3}`)
}

// extractH1 returns the text of the first `# Title` line and the rest of
// the source after that line.
func extractH1(src string) (string, string) {
	src = strings.TrimLeft(src, "\n")
	if !strings.HasPrefix(src, "# ") {
		return "", src
	}
	end := strings.Index(src, "\n")
	if end < 0 {
		return strings.TrimSpace(strings.TrimPrefix(src, "# ")), ""
	}
	title := strings.TrimSpace(strings.TrimPrefix(src[:end], "# "))
	return title, src[end:]
}

// extractMetaBullets walks consecutive `- **X:**` lines after the title
// (skipping blank lines first) and returns them stripped of the `- ` prefix.
// Stops at the first line that doesn't match. Anything after stays in rest.
func extractMetaBullets(src string) ([]string, string) {
	src = strings.TrimLeft(src, "\n")
	var bullets []string
	rest := src
	for {
		nl := strings.Index(rest, "\n")
		line := rest
		if nl >= 0 {
			line = rest[:nl]
		}
		if !strings.HasPrefix(line, "- **") {
			break
		}
		bullets = append(bullets, strings.TrimPrefix(line, "- "))
		if nl < 0 {
			rest = ""
			break
		}
		rest = rest[nl+1:]
	}
	return bullets, rest
}

// splitOnFirstH2 splits the source at the first `^## ` line. The part
// before it is the header_body (any freeform Markdown after meta_bullets
// and before the sections); the part starting from `## ` is the sections
// block.
func splitOnFirstH2(src string) (header, sectionsSrc string) {
	idx := indexLine(src, "## ")
	if idx < 0 {
		return src, ""
	}
	return src[:idx], src[idx:]
}

// indexLine returns the byte offset of the first line that starts with
// prefix, or -1 if no such line exists. Searches the source treating each
// `\n`-delimited segment as a line.
func indexLine(src, prefix string) int {
	if strings.HasPrefix(src, prefix) {
		return 0
	}
	needle := "\n" + prefix
	idx := strings.Index(src, needle)
	if idx < 0 {
		return -1
	}
	return idx + 1
}

// parseSectionsFromTmpl splits the sections block on `^## ` boundaries
// and converts each chunk to a sections.Section. Type inference is
// minimal but content-preserving — see Convert's docstring.
func parseSectionsFromTmpl(src string) []sections.Section {
	src = strings.TrimSpace(src)
	if !strings.HasPrefix(src, "## ") {
		return nil
	}
	chunks := splitH2Blocks(src)
	out := make([]sections.Section, 0, len(chunks))
	for _, chunk := range chunks {
		nl := strings.Index(chunk, "\n")
		if nl < 0 {
			// Heading-only section with no body.
			title := strings.TrimSpace(strings.TrimPrefix(chunk, "## "))
			out = append(out, sections.Section{
				ID:    slugifyHeading(title),
				Title: title,
				Type:  sections.TypeText,
			})
			continue
		}
		title := strings.TrimSpace(strings.TrimPrefix(chunk[:nl], "## "))
		body := strings.TrimSpace(chunk[nl+1:])
		out = append(out, classifySection(title, body))
	}
	return out
}

// splitH2Blocks returns chunks that each start with `## Title\n...` and
// extend up to (but not including) the next `## ` line.
func splitH2Blocks(src string) []string {
	var out []string
	for src != "" {
		// src starts with `## `
		next := indexLine(src[1:], "## ")
		if next < 0 {
			out = append(out, src)
			return out
		}
		// +1 because we searched from src[1:].
		out = append(out, src[:next+1])
		src = src[next+1:]
	}
	return out
}

// classifySection produces a Section from a heading + body. Heuristics:
//   - GFM table (header row + |---| separator) → type:table.
//   - Body contains Go-template control flow → type:text wrapped in diff
//     markers (caller intent: don't auto-translate, ask human).
//   - Otherwise → type:text with default_body verbatim.
func classifySection(title, body string) sections.Section {
	id := slugifyHeading(title)
	switch {
	case hasGFMTable(body):
		cols, rows := parseGFMTable(body)
		return sections.Section{
			ID:      id,
			Title:   title,
			Type:    sections.TypeTable,
			Columns: cols,
			Rows:    rows,
		}
	case hasControlFlow(body):
		return sections.Section{
			ID:          id,
			Title:       title,
			Type:        sections.TypeText,
			DefaultBody: wrapWithDiffMarkers(body),
		}
	default:
		return sections.Section{
			ID:          id,
			Title:       title,
			Type:        sections.TypeText,
			DefaultBody: body + "\n",
		}
	}
}

// englishInParensRE captures the English portion of a "Кириллица (English)"
// heading. Used to produce stable, lowercase, underscore-joined section IDs.
var englishInParensRE = regexp.MustCompile(`\(([A-Za-z][A-Za-z0-9 \-/]*)\)`)

// slugifyHeading produces a stable section ID from a heading. Bilingual
// "Кириллица (English)" headings use the English portion (lowercased,
// underscored); fully-Latin headings use the whole heading; other headings
// fall back to a best-effort slug of the whole title.
func slugifyHeading(title string) string {
	m := englishInParensRE.FindStringSubmatch(title)
	if m != nil {
		return slugify(m[1])
	}
	return slugify(title)
}

// slugify lowercases and replaces non-alphanumerics with underscores.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevUnderscore := false
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevUnderscore = false
		case !prevUnderscore && b.Len() > 0:
			b.WriteRune('_')
			prevUnderscore = true
		}
	}
	out := strings.TrimRight(b.String(), "_")
	if out == "" {
		return "section"
	}
	return out
}

// gfmTableSeparatorRE matches a markdown table separator row (e.g.
// `|---|---|` or `| :---: | --- |`). Used to detect tables.
var gfmTableSeparatorRE = regexp.MustCompile(`^\|[\s:\-|]+\|\s*$`)

// hasGFMTable returns true when body contains a markdown table (header
// row + separator row pattern).
func hasGFMTable(body string) bool {
	lines := strings.Split(body, "\n")
	for i := range len(lines) - 1 {
		if strings.HasPrefix(lines[i], "|") && gfmTableSeparatorRE.MatchString(lines[i+1]) {
			return true
		}
	}
	return false
}

// parseGFMTable extracts columns + rows from the first markdown table
// found in body. Returns column names and data rows (header + separator
// rows excluded). Body before/after the table is dropped — callers that
// want to preserve preamble should detect it before calling this.
func parseGFMTable(body string) ([]string, [][]string) {
	lines := strings.Split(body, "\n")
	startIdx := -1
	for i := range len(lines) - 1 {
		if strings.HasPrefix(lines[i], "|") && gfmTableSeparatorRE.MatchString(lines[i+1]) {
			startIdx = i
			break
		}
	}
	if startIdx < 0 {
		return nil, nil
	}
	cols := splitTableRow(lines[startIdx])
	var rows [][]string
	for i := startIdx + 2; i < len(lines); i++ {
		if !strings.HasPrefix(lines[i], "|") {
			break
		}
		rows = append(rows, splitTableRow(lines[i]))
	}
	return cols, rows
}

// splitTableRow tokenizes a `| a | b | c |` row into ["a", "b", "c"].
func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	cells := strings.Split(line, "|")
	out := make([]string, len(cells))
	for i, c := range cells {
		out[i] = strings.TrimSpace(c)
	}
	return out
}

// controlFlowRE catches the three Go-template control-flow openers. When
// present in a section body the converter doesn't try to simplify the
// body; it wraps the whole thing in diff markers for human review.
var controlFlowRE = regexp.MustCompile(`{{[\s-]*(if|range|with)\b`)

func hasControlFlow(body string) bool {
	return controlFlowRE.MatchString(body)
}

// wrapWithDiffMarkers wraps a body in `git merge`-style markers so the
// user sees the original content in the v1 yaml and resolves it manually.
// The middle (`=======`) is empty by intent — the user replaces it with
// the v1-friendly rendering they want.
func wrapWithDiffMarkers(body string) string {
	return strings.Join([]string{
		"<<<<<<< srekit migrate: this section contains Go-template control flow",
		"that the converter couldn't simplify deterministically. The original",
		"content is preserved below; replace with v1-friendly markdown or a",
		"typed list/table and remove the markers.",
		body,
		"=======",
		"TODO: write a clean replacement here.",
		">>>>>>>",
		"",
	}, "\n")
}
