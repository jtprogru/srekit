package changelog

import (
	"fmt"
	"strings"
)

// Convention is how a document builds its link reference URLs: the base a
// comparison hangs off, the base a release tag hangs off, and whether tags
// carry a `v` prefix.
//
// It is read back out of the document rather than re-derived from the git
// remote. The existing `[Unreleased]` definition already encodes the host,
// the repository path, the comparison URL shape and the tag scheme, and
// reading it is both more accurate and more portable: it handles a
// self-hosted GitLab, a changelog pointing at a mirror, and tag schemes
// this tool never anticipated. Git is consulted only when there is no link
// block at all.
type Convention struct {
	// CompareBase is everything before `/compare/`, e.g.
	// `https://github.com/acme/api` or `https://git.example.com/g/p/-`.
	CompareBase string
	// TagBase is everything before the tag in a release-tag URL, e.g.
	// `https://github.com/acme/api/releases/tag`.
	TagBase string
	// TagPrefix is the prefix tags carry, `v` or empty.
	TagPrefix string
}

// githubConvention builds the convention srekit's own scaffold emits.
func githubConvention(slug string) Convention {
	return Convention{
		CompareBase: "https://github.com/" + slug,
		TagBase:     "https://github.com/" + slug + "/releases/tag",
		TagPrefix:   "v",
	}
}

// Tag renders a version as the tag name this document uses.
func (c Convention) Tag(version string) string { return c.TagPrefix + version }

// CompareURL builds a comparison URL between two refs.
func (c Convention) CompareURL(from, to string) string {
	return fmt.Sprintf("%s/compare/%s...%s", c.CompareBase, from, to)
}

// TagURL builds a release-tag URL for a version.
func (c Convention) TagURL(version string) string {
	return c.TagBase + "/" + c.Tag(version)
}

// DeriveConvention reads the URL conventions out of the document's own link
// block. ok is false when the block carries nothing the conventions can be
// read from, in which case the caller falls back to the resolved repository
// slug.
func (d *Document) DeriveConvention() (conv Convention, ok bool) {
	if !d.LinksPresent {
		return Convention{}, false
	}

	// The comparison shape and tag prefix come from any definition whose URL
	// contains `/compare/`; `[Unreleased]` is the one that always has it, so
	// it is tried first.
	for _, def := range d.orderedForDerivation() {
		base, from, _, found := splitCompareURL(def.URL)
		if !found {
			continue
		}
		conv.CompareBase = base
		conv.TagPrefix = tagPrefixOf(from)
		ok = true
		break
	}

	// The release-tag shape is learned separately, because a document whose
	// first release predates any comparison still defines one.
	for _, def := range d.Links {
		if i := strings.LastIndex(def.URL, "/releases/tag/"); i >= 0 {
			conv.TagBase = def.URL[:i] + "/releases/tag"
			if !ok {
				conv.CompareBase = def.URL[:i]
				conv.TagPrefix = tagPrefixOf(def.URL[i+len("/releases/tag/"):])
				ok = true
			}
			break
		}
	}
	if ok && conv.TagBase == "" {
		conv.TagBase = conv.CompareBase + "/releases/tag"
	}
	return conv, ok
}

// orderedForDerivation puts the `[Unreleased]` definition first: it is the
// one guaranteed to compare against HEAD, so its `from` ref is the most
// reliable place to read the tag prefix.
func (d *Document) orderedForDerivation() []LinkDef {
	out := make([]LinkDef, 0, len(d.Links))
	for _, def := range d.Links {
		if strings.EqualFold(def.Label, "unreleased") {
			out = append(out, def)
		}
	}
	for _, def := range d.Links {
		if !strings.EqualFold(def.Label, "unreleased") {
			out = append(out, def)
		}
	}
	return out
}

func splitCompareURL(u string) (base, from, to string, ok bool) {
	i := strings.LastIndex(u, "/compare/")
	if i < 0 {
		return "", "", "", false
	}
	base = u[:i]
	spec := u[i+len("/compare/"):]
	j := strings.Index(spec, "...")
	if j < 0 {
		// Some hosts use a two-dot range; treat anything else as the whole ref.
		return base, spec, "", true
	}
	return base, spec[:j], spec[j+3:], true
}

// tagPrefixOf returns the non-numeric prefix a tag carries, so `v1.1.0`
// yields `v` and `1.1.0` yields the empty string. Only a leading `v` is
// recognized; anything else is treated as part of the ref and produces no
// prefix, because inventing one would rename a tag that exists.
func tagPrefixOf(ref string) string {
	if strings.HasPrefix(ref, "v") {
		if _, ok := ParseVersion(strings.TrimPrefix(ref, "v")); ok {
			return "v"
		}
	}
	return ""
}

// FindLink returns the URL defined for label, case-insensitively.
func (d *Document) FindLink(label string) (string, bool) {
	for _, def := range d.Links {
		if strings.EqualFold(def.Label, label) {
			return def.URL, true
		}
	}
	return "", false
}

// renderLinkBlock serializes definitions back to `[label]: url` lines.
func renderLinkBlock(defs []LinkDef) string {
	var b strings.Builder
	for _, d := range defs {
		b.WriteString("[")
		b.WriteString(d.Label)
		b.WriteString("]: ")
		b.WriteString(d.URL)
		b.WriteString("\n")
	}
	return b.String()
}
