package changelog

import (
	"fmt"
	"strings"
)

// CheckResult is one validation check and its outcome. Every check is
// reported, passing or failing: a person repairing a drifted changelog
// wants the whole list in one pass, not the first failure.
type CheckResult struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
}

// Check names, stable enough to grep for in CI output.
const (
	CheckHeadingShape = "heading-shape"
	CheckUnreleased   = "unreleased-section"
	CheckOrder        = "version-order"
	CheckDuplicates   = "no-duplicate-versions"
	CheckChangeTypes  = "change-types"
	CheckVocabulary   = "change-type-language"
	CheckLinks        = "link-definitions"
)

// Validate runs every check against a scanned document.
func Validate(doc *Document, vocabs []Vocabulary) []CheckResult {
	return []CheckResult{
		checkHeadingShape(doc),
		checkUnreleased(doc),
		checkOrder(doc),
		checkDuplicates(doc),
		checkChangeTypes(doc, vocabs),
		checkVocabulary(doc, vocabs),
		checkLinks(doc),
	}
}

func pass(name string) CheckResult { return CheckResult{Name: name, OK: true} }

func passWith(name, format string, a ...any) CheckResult {
	return CheckResult{Name: name, OK: true, Detail: fmt.Sprintf(format, a...)}
}

func fail(name, format string, a ...any) CheckResult {
	return CheckResult{Name: name, Detail: fmt.Sprintf(format, a...)}
}

// checkHeadingShape requires `[X.Y.Z] - YYYY-MM-DD`, optionally followed by
// ` [YANKED]`. This is the check that catches a regional date.
func checkHeadingShape(doc *Document) CheckResult {
	var bad []string
	for _, s := range doc.Sections() {
		if s.WellFormed {
			continue
		}
		bad = append(bad, fmt.Sprintf("line %d: %s", s.Line, s.Heading))
	}
	if len(bad) == 0 {
		return pass(CheckHeadingShape)
	}
	return fail(CheckHeadingShape, "expected `## [X.Y.Z] - YYYY-MM-DD` optionally followed by ` [YANKED]`; got %s", strings.Join(bad, "; "))
}

func checkUnreleased(doc *Document) CheckResult {
	if doc.Unreleased == nil {
		return fail(CheckUnreleased, "no `## [Unreleased]` heading")
	}
	for _, s := range doc.Versions {
		if s.Line < doc.Unreleased.Line {
			return fail(CheckUnreleased, "`[Unreleased]` (line %d) must precede every released version, but %s is at line %d",
				doc.Unreleased.Line, headingLabel(s), s.Line)
		}
	}
	return pass(CheckUnreleased)
}

// checkOrder requires descending version order. A version that does not
// parse is reported here rather than silently ordered.
func checkOrder(doc *Document) CheckResult {
	var problems []string
	prev := ""
	prevLine := 0
	for _, s := range doc.Versions {
		if s.Version == "" {
			continue
		}
		if _, ok := ParseVersion(s.Version); !ok {
			problems = append(problems, fmt.Sprintf("line %d: %s is not a numeric version", s.Line, s.Version))
			continue
		}
		if prev != "" {
			if c, ok := CompareVersions(s.Version, prev); ok && c > 0 {
				problems = append(problems, fmt.Sprintf("%s (line %d) is listed above %s (line %d)", prev, prevLine, s.Version, s.Line))
			}
		}
		prev, prevLine = s.Version, s.Line
	}
	if len(problems) == 0 {
		return pass(CheckOrder)
	}
	return fail(CheckOrder, "versions must appear in descending order: %s", strings.Join(problems, "; "))
}

func checkDuplicates(doc *Document) CheckResult {
	seen := map[string]int{}
	var dups []string
	for _, s := range doc.Sections() {
		label := s.Label
		if label == "" {
			continue
		}
		key := strings.ToLower(label)
		if first, ok := seen[key]; ok {
			dups = append(dups, fmt.Sprintf("%s at lines %d and %d", label, first, s.Line))
			continue
		}
		seen[key] = s.Line
	}
	if len(dups) == 0 {
		return pass(CheckDuplicates)
	}
	return fail(CheckDuplicates, "each version may appear once: %s", strings.Join(dups, "; "))
}

// checkChangeTypes requires every `### ` subsection to be one of the
// recognized change types. The vocabulary is the specification's, not the
// artifact template's: a user who renames headings in a customized template
// gets a failure here, which is the correct answer — the format names those
// types.
func checkChangeTypes(doc *Document, vocabs []Vocabulary) CheckResult {
	var bad []string
	for _, s := range doc.Sections() {
		for _, c := range s.Changes {
			if c.Recognized {
				continue
			}
			bad = append(bad, fmt.Sprintf("line %d: %s", c.Line, c.Heading))
		}
	}
	if len(bad) == 0 {
		return pass(CheckChangeTypes)
	}
	return fail(CheckChangeTypes, "unrecognized change type %s; allowed: %s", strings.Join(bad, "; "), allowedTypes(vocabs))
}

// checkVocabulary reports which change-type vocabulary the document is
// written in, and fails when that question has no single answer.
//
// It reports on a pass as well as a failure, because a user who expected
// one language and got the other is otherwise handed a list of passing
// checks that measured the wrong thing.
func checkVocabulary(doc *Document, vocabs []Vocabulary) CheckResult {
	switch {
	case doc.Mixed():
		return fail(CheckVocabulary, "document mixes change-type vocabularies: %s; a changelog must use one language throughout", doc.DescribeMix())
	case doc.Lang == "":
		return fail(CheckVocabulary, "no recognized change types found; expected one of: %s", allowedTypes(vocabs))
	default:
		return passWith(CheckVocabulary, "%s (%s) change types", LangName(doc.Lang), doc.Lang)
	}
}

func allowedTypes(vocabs []Vocabulary) string {
	var all []string
	for _, v := range vocabs {
		all = append(all, v.Types...)
	}
	return strings.Join(all, ", ")
}

func checkLinks(doc *Document) CheckResult {
	if !doc.LinksPresent {
		return fail(CheckLinks, "no link reference block; every version heading needs a `[X.Y.Z]: <url>` definition")
	}
	var missing []string
	for _, s := range doc.Sections() {
		if s.Label == "" {
			continue
		}
		if _, ok := doc.FindLink(s.Label); !ok {
			missing = append(missing, s.Label)
		}
	}
	if len(missing) == 0 {
		return pass(CheckLinks)
	}
	return fail(CheckLinks, "no link definition for %s", strings.Join(missing, ", "))
}

func headingLabel(s *Section) string {
	if s.Label != "" {
		return "[" + s.Label + "]"
	}
	return s.Heading
}
