package changelog

import (
	"errors"
	"fmt"
	"strings"
)

// ReleaseOptions describes the release to cut.
type ReleaseOptions struct {
	// Version is the version being released, without a tag prefix.
	Version string
	// Date is the release date in YYYY-MM-DD form.
	Date string
	// Yanked appends the ` [YANKED]` marker to the new heading.
	Yanked bool
	// Slug is OWNER/REPO, used only to build a link block for a document
	// that has none. It may be empty when the document already has one.
	Slug string
}

// Errors the caller distinguishes. Each corresponds to a precondition that
// leaves the file untouched.
var (
	// ErrNoUnreleased is returned for a document with no `## [Unreleased]`
	// heading. The scanner refuses rather than guessing an insertion point:
	// guessing is how a rewriter destroys history.
	ErrNoUnreleased = errors.New("no `## [Unreleased]` heading found")
	// ErrNothingToRelease is returned when `[Unreleased]` holds nothing but
	// blank lines and scaffold placeholders.
	ErrNothingToRelease = errors.New("`[Unreleased]` has no entries to release")
)

// AlreadyReleasedError reports a version that already has a heading.
type AlreadyReleasedError struct {
	Version string
	Line    int
}

// MixedVocabularyError reports a document whose change types come from
// more than one language. Releasing it would file entries under headings
// that duplicate each other in two languages, so the rewrite refuses and
// the file is left exactly as it was.
type MixedVocabularyError struct {
	// Detail names the offending headings and their lines.
	Detail string
}

func (e *MixedVocabularyError) Error() string {
	return fmt.Sprintf("document mixes change-type vocabularies (%s); a changelog must use one language throughout", e.Detail)
}

func (e *AlreadyReleasedError) Error() string {
	return fmt.Sprintf("version %s is already released (line %d); re-running a release would file entries accumulated since under a version that shipped without them", e.Version, e.Line)
}

// Release produces the rewritten document. It never writes; the caller owns
// output routing. Every failure returns before any bytes are produced, so a
// caller that writes only on success cannot half-edit a file.
func Release(doc *Document, opts ReleaseOptions) ([]byte, error) {
	if doc.Mixed() {
		return nil, &MixedVocabularyError{Detail: doc.DescribeMix()}
	}
	if doc.Unreleased == nil {
		return nil, ErrNoUnreleased
	}
	if s := doc.FindVersion(opts.Version); s != nil {
		return nil, &AlreadyReleasedError{Version: opts.Version, Line: s.Line}
	}

	body := releasedBody(doc)
	if body == "" {
		return nil, ErrNothingToRelease
	}

	conv, ok := doc.DeriveConvention()
	if !ok {
		if opts.Slug == "" {
			return nil, errors.New("document has no link reference block and no repository slug was resolved")
		}
		conv = githubConvention(opts.Slug)
	}

	var b strings.Builder
	// Everything above `## [Unreleased]` is copied byte for byte.
	b.Write(doc.Src[:doc.Unreleased.Region.Start])

	// The `[Unreleased]` heading keeps its own bytes and is left with no
	// entries. It is not refilled with the six-type skeleton: Keep a
	// Changelog's own example keeps an empty `[Unreleased]`, and refilling
	// would put six headings and six placeholders into every release diff
	// only for the next release to strip them again.
	b.WriteString(strings.TrimRight(doc.Unreleased.HeadingRegion.Text(doc.Src), "\n"))
	b.WriteString("\n\n")

	b.WriteString(versionHeading(opts))
	b.WriteString("\n\n")
	b.WriteString(body)
	b.WriteString("\n\n")

	// Everything from the next section heading to the link block is copied
	// byte for byte: previously released versions are untouched.
	mid := doc.Src[doc.Unreleased.Region.End:doc.LinksRegion.Start]
	b.WriteString(strings.TrimLeft(string(mid), "\n"))

	writeLinkBlock(&b, doc, opts, conv)
	return []byte(b.String()), nil
}

func versionHeading(opts ReleaseOptions) string {
	h := fmt.Sprintf("## [%s] - %s", opts.Version, opts.Date)
	if opts.Yanked {
		h += " [YANKED]"
	}
	return h
}

// releasedBody assembles the new version's body out of `[Unreleased]`:
// any lead prose, then each change-type subsection that actually has
// entries. Subsections that are empty or hold only the scaffold's bare `-`
// are dropped, so a released version never ships an empty `### Deprecated`.
func releasedBody(doc *Document) string {
	u := doc.Unreleased
	var parts []string

	if lead := u.Lead.Text(doc.Src); hasContent(lead) {
		parts = append(parts, strings.Trim(lead, "\n"))
	}
	for _, c := range u.Changes {
		if !c.HasEntries(doc.Src) {
			continue
		}
		parts = append(parts, strings.Trim(c.Region.Text(doc.Src), "\n"))
	}
	return strings.Join(parts, "\n\n")
}

// writeLinkBlock emits the trailing link block and whatever follows it.
// `[Unreleased]` is repointed at the newly released tag and the new
// version's definition is inserted directly beneath it, which is where the
// previously newest definition used to sit.
func writeLinkBlock(b *strings.Builder, doc *Document, opts ReleaseOptions, conv Convention) {
	newTag := conv.Tag(opts.Version)
	unreleasedDef := LinkDef{Label: "Unreleased", URL: conv.CompareURL(newTag, "HEAD")}

	// The new version compares against the previously newest release, or
	// points at its own tag when it is the first one.
	versionDef := LinkDef{Label: opts.Version, URL: conv.TagURL(opts.Version)}
	if prev := previousNewest(doc); prev != "" {
		versionDef.URL = conv.CompareURL(conv.Tag(prev), newTag)
	}

	defs := make([]LinkDef, 0, len(doc.Links)+2)
	inserted := false
	for _, d := range doc.Links {
		if strings.EqualFold(d.Label, "unreleased") {
			defs = append(defs, unreleasedDef, versionDef)
			inserted = true
			continue
		}
		defs = append(defs, d)
	}
	if !inserted {
		defs = append([]LinkDef{unreleasedDef, versionDef}, defs...)
	}

	if !doc.LinksPresent {
		// No block to replace: append one, separated by a blank line.
		cur := strings.TrimRight(b.String(), "\n")
		b.Reset()
		b.WriteString(cur)
		b.WriteString("\n\n")
		b.WriteString(renderLinkBlock(defs))
		return
	}

	b.WriteString(renderLinkBlock(defs))
	b.WriteString(doc.Trailing.Text(doc.Src))
}

// previousNewest returns the highest released version already in the
// document, or "" when there is none. Headings whose version does not parse
// are skipped rather than ordered — `validate` is where they are reported.
func previousNewest(doc *Document) string {
	best := ""
	for _, s := range doc.Versions {
		if s.Version == "" {
			continue
		}
		if _, ok := ParseVersion(s.Version); !ok {
			continue
		}
		if best == "" {
			best = s.Version
			continue
		}
		if c, ok := CompareVersions(s.Version, best); ok && c > 0 {
			best = s.Version
		}
	}
	return best
}
