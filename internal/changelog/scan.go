// Package changelog reads and rewrites an existing Keep a Changelog
// document.
//
// Everything here is deliberately a line-oriented region scanner rather
// than a Markdown parser. The obvious design — parse into a model, mutate,
// reserialize — is the one that damages files: every blank line, bullet
// marker and wrapped line in a five-year-old changelog would be normalized
// to whatever this package considered canonical, and the release commit
// would become an unreviewable diff. Instead the scanner records byte
// offsets into the source, and the rewriter splices new text at those
// offsets and copies everything else through untouched. That is what makes
// "byte-identical outside the edited regions" a property of the design
// rather than an aspiration.
//
// Only three things are understood: `## ` headings, the `### `
// change-type subsections beneath them, and link reference definitions.
// Everything else is opaque text carried across verbatim.
package changelog

import (
	"regexp"
	"strconv"
	"strings"
)

// Vocabulary is a set of recognized change-type headings in one language.
//
// The scanner takes the vocabulary as a parameter rather than hardcoding
// the English six because a Russian changelog artifact whose headings read
// `Добавлено`, `Изменено` and so on is a known upcoming requirement, not a
// speculative one. Adding a language means appending a Vocabulary, not
// reworking the scanner.
type Vocabulary struct {
	// Lang is the language tag reported in JSON output and used to tell
	// two vocabularies apart within one document.
	Lang string
	// Types are the recognized change-type headings, in specification order.
	Types []string
}

// English is the Keep a Changelog 1.1.0 change-type set.
var English = Vocabulary{ //nolint:gochecknoglobals // immutable specification data, not mutable state
	Lang:  "en",
	Types: []string{"Added", "Changed", "Deprecated", "Removed", "Fixed", "Security"},
}

// Vocabularies returns the set the scanner recognizes by default. It is a
// function rather than a package-level slice so no caller can mutate it.
func Vocabularies() []Vocabulary { return []Vocabulary{English} }

// Region is a half-open byte range [Start, End) into Document.Src.
type Region struct {
	Start int `json:"-"`
	End   int `json:"-"`
}

// Text slices src to the region.
func (r Region) Text(src []byte) string { return string(src[r.Start:r.End]) }

// Empty reports whether the region covers no bytes.
func (r Region) Empty() bool { return r.End <= r.Start }

// ChangeSubsection is a `### <type>` block inside a version section.
type ChangeSubsection struct {
	// Heading is the text after `### `, trimmed.
	Heading string
	// Lang is the vocabulary this heading belongs to, empty when the
	// heading is not one of the recognized change types.
	Lang string
	// Recognized reports whether Heading is a known change type.
	Recognized bool
	// Line is the 1-based source line of the heading, for error messages.
	Line int

	// Region covers the heading and its body; Body covers the body alone.
	Region Region
	Body   Region
}

// Section is a `## ` region: `[Unreleased]` or one released version.
type Section struct {
	// Raw heading text without the trailing newline, e.g. `## [1.2.0] - 2026-03-04`.
	Heading string
	// Line is the 1-based source line of the heading.
	Line int

	// Label is the bracketed text, e.g. `Unreleased` or `1.2.0`. Empty when
	// the heading carries no bracketed label.
	Label string
	// Version is Label for a released version, empty for `[Unreleased]`.
	Version string
	// Date is the raw date text as written, whether or not it is ISO.
	Date string
	// Yanked reports the ` [YANKED]` marker.
	Yanked bool
	// WellFormed reports whether the heading matched
	// `[X.Y.Z] - YYYY-MM-DD` optionally followed by ` [YANKED]`.
	// An unparseable heading is kept and marked, never dropped: dropping it
	// would make the rewriter splice over content it did not understand.
	WellFormed bool
	// Unreleased marks the `[Unreleased]` section.
	Unreleased bool

	// Region covers the heading line through the byte before the next
	// section heading (or the link block, or EOF). HeadingRegion covers the
	// heading line including its newline; Body covers the rest.
	Region        Region
	HeadingRegion Region
	Body          Region

	// Lead is the part of Body before the first `### ` subsection — prose
	// someone wrote between the version heading and its first change type.
	Lead    Region
	Changes []ChangeSubsection
}

// LinkDef is one `[label]: url` reference definition.
type LinkDef struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// Document is a scanned changelog. Every region indexes into Src.
type Document struct {
	Src []byte

	// Preamble is everything before the first `## ` heading: the H1 and any
	// hand-written prose under it.
	Preamble Region

	// Unreleased is the `[Unreleased]` section, nil when the document has none.
	Unreleased *Section
	// Versions are the released sections in document order.
	Versions []*Section

	// Links is the trailing link reference block. LinksPresent is false when
	// the document has none, in which case LinksRegion is the empty range at
	// the point where a block would be appended.
	Links        []LinkDef
	LinksRegion  Region
	LinksPresent bool

	// Trailing is whatever follows the link block.
	Trailing Region

	// Lang is the vocabulary in force, detected from the first recognized
	// change-type heading in the document. Empty when the document contains
	// no recognized change type at all — which is a condition to report,
	// never a default to guess at.
	Lang string
	// Langs are the distinct vocabularies whose headings appear in the
	// document. More than one means the document is half-translated.
	Langs []string
}

// Sections returns Unreleased (when present) followed by the released
// versions, in document order.
func (d *Document) Sections() []*Section {
	out := make([]*Section, 0, len(d.Versions)+1)
	if d.Unreleased != nil {
		out = append(out, d.Unreleased)
	}
	return append(out, d.Versions...)
}

// FindVersion returns the section for version, or nil.
func (d *Document) FindVersion(version string) *Section {
	for _, s := range d.Versions {
		if s.Version == version {
			return s
		}
	}
	return nil
}

var (
	// `## [1.2.0] - 2026-03-04 [YANKED]` — the bracketed label and the rest
	// are separated here so an unparseable tail still yields a label.
	sectionHeadingRe = regexp.MustCompile(`^##[ \t]+(.*?)[ \t]*$`)
	labelRe          = regexp.MustCompile(`^\[([^\]]*)\][ \t]*(.*)$`)
	datePartRe       = regexp.MustCompile(`^-[ \t]*(\S+)[ \t]*(.*)$`)
	subHeadingRe     = regexp.MustCompile(`^###[ \t]+(.*?)[ \t]*$`)
	// A link reference definition: `[label]: url`, optionally indented up to
	// three spaces as CommonMark allows.
	linkDefRe = regexp.MustCompile(`^ {0,3}\[([^\]]+)\]:[ \t]*(\S+)[ \t]*$`)
	isoDateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

type sourceLine struct {
	start int // byte offset of the first character
	end   int // byte offset just past the trailing newline (or EOF)
	text  string
	num   int // 1-based
}

func splitLines(src []byte) []sourceLine {
	var lines []sourceLine
	start := 0
	num := 1
	for i := range src {
		if src[i] != '\n' {
			continue
		}
		lines = append(lines, sourceLine{start: start, end: i + 1, text: strings.TrimRight(string(src[start:i]), "\r"), num: num})
		start = i + 1
		num++
	}
	if start < len(src) {
		lines = append(lines, sourceLine{start: start, end: len(src), text: string(src[start:]), num: num})
	}
	return lines
}

// Scan reads src into a Document. It never fails: a shape the scanner does
// not recognize becomes opaque text rather than an error, and the
// preconditions that matter (no `[Unreleased]`, nothing to release) are
// checked by the caller against the returned value. That split is what lets
// `validate` report on a document `release` would refuse.
func Scan(src []byte, vocabs []Vocabulary) *Document {
	doc := &Document{Src: src}
	lines := splitLines(src)

	blockStart, blockEnd, defs, present := scanLinkBlock(lines, len(src))
	doc.Links, doc.LinksPresent = defs, present
	doc.LinksRegion = Region{Start: blockStart, End: blockEnd}
	doc.Trailing = Region{Start: blockEnd, End: len(src)}

	// Section scanning stops where the link block begins, so a document
	// whose block sits directly under the last version does not swallow it
	// into that version's body.
	var heads []sourceLine
	for _, ln := range lines {
		if ln.start >= blockStart {
			break
		}
		if sectionHeadingRe.MatchString(ln.text) {
			heads = append(heads, ln)
		}
	}

	if len(heads) == 0 {
		doc.Preamble = Region{Start: 0, End: blockStart}
	} else {
		doc.Preamble = Region{Start: 0, End: heads[0].start}
	}

	for i, h := range heads {
		end := blockStart
		if i+1 < len(heads) {
			end = heads[i+1].start
		}
		s := parseSection(h, Region{Start: h.start, End: end})
		fillChanges(s, lines, vocabs)
		if s.Unreleased {
			// A second `[Unreleased]` heading is malformed; the first wins
			// and the rest are reported by validate as duplicate labels.
			if doc.Unreleased == nil {
				doc.Unreleased = s
				continue
			}
		}
		doc.Versions = append(doc.Versions, s)
	}

	doc.Lang, doc.Langs = detectLangs(doc.Sections())
	return doc
}

// scanLinkBlock finds the trailing link reference block: the last maximal
// run of consecutive link-definition lines. Taking the last run rather than
// the first means a document that also defines links inside a version body
// still has its footer identified correctly.
func scanLinkBlock(lines []sourceLine, srcLen int) (start, end int, defs []LinkDef, present bool) {
	runStart, runEnd := -1, -1
	var runDefs []LinkDef

	i := 0
	for i < len(lines) {
		if !linkDefRe.MatchString(lines[i].text) {
			i++
			continue
		}
		j := i
		var cur []LinkDef
		for j < len(lines) && linkDefRe.MatchString(lines[j].text) {
			m := linkDefRe.FindStringSubmatch(lines[j].text)
			cur = append(cur, LinkDef{Label: m[1], URL: m[2]})
			j++
		}
		runStart, runEnd, runDefs = lines[i].start, lines[j-1].end, cur
		i = j
	}

	if runStart < 0 {
		// No block: report the append point at end of document.
		return srcLen, srcLen, nil, false
	}
	return runStart, runEnd, runDefs, true
}

func parseSection(h sourceLine, region Region) *Section {
	s := &Section{
		Heading:       h.text,
		Line:          h.num,
		Region:        region,
		HeadingRegion: Region{Start: h.start, End: h.end},
		Body:          Region{Start: h.end, End: region.End},
	}
	if s.Body.End < s.Body.Start {
		s.Body.End = s.Body.Start
	}

	rest := sectionHeadingRe.FindStringSubmatch(h.text)[1]
	m := labelRe.FindStringSubmatch(rest)
	if m == nil {
		// A heading with no bracketed label — `## 1.2.0 - 2026-03-04` or a
		// section someone added by hand. Kept, marked not well-formed.
		return s
	}
	s.Label = m[1]
	tail := strings.TrimSpace(m[2])

	if strings.EqualFold(s.Label, "unreleased") {
		s.Unreleased = true
		s.WellFormed = tail == ""
		return s
	}

	s.Version = s.Label
	if d := datePartRe.FindStringSubmatch(tail); d != nil {
		s.Date = d[1]
		tail = strings.TrimSpace(d[2])
	}
	if strings.EqualFold(tail, "[YANKED]") {
		s.Yanked = true
		tail = ""
	}
	_, versionOK := ParseVersion(s.Version)
	s.WellFormed = versionOK && isoDateRe.MatchString(s.Date) && tail == ""
	return s
}

// fillChanges records the `### ` subsections inside s.Body, plus the lead
// region before the first one.
func fillChanges(s *Section, lines []sourceLine, vocabs []Vocabulary) {
	var subs []sourceLine
	for _, ln := range lines {
		if ln.start < s.Body.Start || ln.start >= s.Body.End {
			continue
		}
		if subHeadingRe.MatchString(ln.text) {
			subs = append(subs, ln)
		}
	}

	if len(subs) == 0 {
		s.Lead = s.Body
		return
	}
	s.Lead = Region{Start: s.Body.Start, End: subs[0].start}

	for i, h := range subs {
		end := s.Body.End
		if i+1 < len(subs) {
			end = subs[i+1].start
		}
		heading := subHeadingRe.FindStringSubmatch(h.text)[1]
		c := ChangeSubsection{
			Heading: heading,
			Line:    h.num,
			Region:  Region{Start: h.start, End: end},
			Body:    Region{Start: h.end, End: end},
		}
		if c.Body.End < c.Body.Start {
			c.Body.End = c.Body.Start
		}
		c.Lang, c.Recognized = classify(heading, vocabs)
		s.Changes = append(s.Changes, c)
	}
}

func classify(heading string, vocabs []Vocabulary) (lang string, ok bool) {
	for _, v := range vocabs {
		for _, t := range v.Types {
			if strings.EqualFold(heading, t) {
				return v.Lang, true
			}
		}
	}
	return "", false
}

func detectLangs(secs []*Section) (first string, all []string) {
	seen := map[string]bool{}
	for _, s := range secs {
		for _, c := range s.Changes {
			if !c.Recognized || seen[c.Lang] {
				continue
			}
			seen[c.Lang] = true
			all = append(all, c.Lang)
		}
	}
	if len(all) > 0 {
		first = all[0]
	}
	return first, all
}

// HasEntries reports whether the subsection body holds anything beyond
// blank lines and the scaffold's bare `-` placeholder.
func (c ChangeSubsection) HasEntries(src []byte) bool {
	return hasContent(c.Body.Text(src))
}

func hasContent(body string) bool {
	for _, line := range strings.Split(body, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || t == "-" || t == "*" {
			continue
		}
		return true
	}
	return false
}

// ParseVersion splits a version into numeric segments. A pre-release
// suffix (`1.2.0-rc.1`) is accepted and its segments are ignored for
// ordering, which is enough for the ordering check this package performs;
// anything that does not split into numeric segments returns ok=false and
// is reported rather than silently ordered.
func ParseVersion(v string) (segs []int, ok bool) {
	core := v
	if i := strings.IndexAny(core, "-+"); i >= 0 {
		core = core[:i]
	}
	if core == "" {
		return nil, false
	}
	for _, part := range strings.Split(core, ".") {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return nil, false
		}
		segs = append(segs, n)
	}
	return segs, true
}

// CompareVersions orders two versions by numeric segment, so 1.10.0 sorts
// above 1.9.0 where a string compare would get it backwards. ok is false
// when either side does not parse, in which case the result is meaningless
// and the caller must report rather than order.
func CompareVersions(a, b string) (result int, ok bool) {
	as, aok := ParseVersion(a)
	bs, bok := ParseVersion(b)
	if !aok || !bok {
		return 0, false
	}
	n := len(as)
	if len(bs) > n {
		n = len(bs)
	}
	for i := range n {
		av, bv := 0, 0
		if i < len(as) {
			av = as[i]
		}
		if i < len(bs) {
			bv = bs[i]
		}
		switch {
		case av < bv:
			return -1, true
		case av > bv:
			return 1, true
		}
	}
	return 0, true
}

// IsISODate reports whether s is a `YYYY-MM-DD` date.
func IsISODate(s string) bool { return isoDateRe.MatchString(s) }
