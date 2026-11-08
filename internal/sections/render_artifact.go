package sections

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	yaml "go.yaml.in/yaml/v3"
)

// RenderArtifact composes the full markdown document for an Artifact:
//
//	---
//	<frontmatter, with scalar values evaluated through the FuncMap>
//	---
//
//	# <title rendered>
//
//	- <meta_bullet rendered>
//	- ...
//
//	<header_body rendered, if non-empty>
//
//	## <section.title>
//
//	<section.body>
//
//	...
//
//	<footer_body rendered, if non-empty>
//
// `rendered` is the section list already produced by sections.Merge
// (overrides applied, defaults expanded). `ctx` is the data root every
// template string is evaluated against — typically `struct{Meta T}{...}`.
//
// Empty fields are skipped — no leading `---\n---` block when frontmatter
// is empty, no bare `# ` when title is empty, no blank stanza when
// header_body or footer_body is whitespace. Frontmatter key order is
// preserved from the source YAML.
//
// Every block is opened through startBlock, which owns the separation
// invariant: exactly one blank line between adjacent blocks, in every
// combination of present and absent elements. Blocks therefore write only
// their own content and never their surrounding separators — the previous
// arrangement, where each block padded itself, is what produced a double
// blank line wherever a trailing and a leading separator met.
func RenderArtifact(a *Artifact, rendered []RenderedSection, ctx any) ([]byte, error) {
	if a == nil {
		return nil, errors.New("artifact is nil")
	}

	var b bytes.Buffer

	// 1. Frontmatter block.
	if a.Frontmatter.Kind == yaml.MappingNode && len(a.Frontmatter.Content) > 0 {
		fm := copyNode(&a.Frontmatter)
		if err := evalMappingValues(fm, ctx); err != nil {
			return nil, fmt.Errorf("frontmatter: %w", err)
		}
		var fmBuf bytes.Buffer
		enc := yaml.NewEncoder(&fmBuf)
		enc.SetIndent(2)
		if err := enc.Encode(fm); err != nil {
			return nil, fmt.Errorf("encode frontmatter: %w", err)
		}
		if err := enc.Close(); err != nil {
			return nil, fmt.Errorf("close frontmatter encoder: %w", err)
		}
		startBlock(&b)
		b.WriteString("---\n")
		b.Write(fmBuf.Bytes())
		b.WriteString("---\n")
	}

	// 2. H1 title.
	if strings.TrimSpace(a.Title) != "" {
		title, err := renderTemplate(a.Title, ctx)
		if err != nil {
			return nil, fmt.Errorf("title: %w", err)
		}
		startBlock(&b)
		b.WriteString("# ")
		b.WriteString(title)
		b.WriteString("\n")
	}

	// 3. Meta bullets.
	if len(a.MetaBullets) > 0 {
		startBlock(&b)
		for _, mb := range a.MetaBullets {
			line, err := renderTemplate(mb, ctx)
			if err != nil {
				return nil, fmt.Errorf("meta_bullets: %w", err)
			}
			b.WriteString("- ")
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	// 4. Header body (freeform Markdown escape hatch).
	if strings.TrimSpace(a.HeaderBody) != "" {
		body, err := renderTemplate(a.HeaderBody, ctx)
		if err != nil {
			return nil, fmt.Errorf("header_body: %w", err)
		}
		startBlock(&b)
		b.WriteString(strings.TrimRight(body, "\n"))
		b.WriteString("\n")
	}

	// 5. Sections. Section titles were already template-evaluated in
	// sections.Merge so the JSON contract and the rendered markdown agree.
	// The heading and its body are two blocks, so the same invariant puts
	// the blank line between them.
	for _, s := range rendered {
		startBlock(&b)
		b.WriteString("## ")
		b.WriteString(s.Title)
		b.WriteString("\n")
		startBlock(&b)
		b.WriteString(strings.TrimRight(s.Body, "\n"))
		b.WriteString("\n")
	}

	// 6. Footer body (trailing document-level material, e.g. link
	// reference definitions).
	if strings.TrimSpace(a.FooterBody) != "" {
		footer, err := renderTemplate(a.FooterBody, ctx)
		if err != nil {
			return nil, fmt.Errorf("footer_body: %w", err)
		}
		startBlock(&b)
		b.WriteString(strings.TrimRight(footer, "\n"))
		b.WriteString("\n")
	}

	// Single trailing newline.
	out := strings.TrimRight(b.String(), "\n") + "\n"
	return []byte(out), nil
}

// startBlock prepares the buffer to receive a new block by ensuring it
// ends in exactly one blank line. It is a no-op on an empty buffer, so
// the first block present — whichever it turns out to be — never gets a
// leading separator.
//
// Normalizing here rather than padding at each write site is what makes
// "exactly one blank line between blocks" hold for every combination of
// present and absent elements, instead of being a property each call site
// has to re-derive from what it guesses came before it.
func startBlock(b *bytes.Buffer) {
	if b.Len() == 0 {
		return
	}
	trimmed := strings.TrimRight(b.String(), "\n")
	b.Reset()
	b.WriteString(trimmed)
	b.WriteString("\n\n")
}

// evalMappingValues walks a YAML mapping node and runs each scalar value
// through the template engine. Keys are not evaluated (would be surprising;
// authors expect frontmatter keys to be literal). Nested mappings and
// sequences are recursed into.
func evalMappingValues(n *yaml.Node, ctx any) error {
	if n == nil {
		return nil
	}
	switch n.Kind {
	case yaml.MappingNode:
		// Content is [key0, val0, key1, val1, ...]; only walk values.
		for i := 1; i < len(n.Content); i += 2 {
			if err := evalMappingValues(n.Content[i], ctx); err != nil {
				return err
			}
		}
	case yaml.SequenceNode:
		for _, c := range n.Content {
			if err := evalMappingValues(c, ctx); err != nil {
				return err
			}
		}
	case yaml.ScalarNode:
		if !strings.Contains(n.Value, "{{") {
			return nil
		}
		rendered, err := renderTemplate(n.Value, ctx)
		if err != nil {
			return err
		}
		n.Value = rendered
	case yaml.DocumentNode, yaml.AliasNode:
		// Documents wrap a single child; aliases are by-reference (skip).
		for _, c := range n.Content {
			if err := evalMappingValues(c, ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// copyNode produces a deep copy of a yaml.Node so RenderArtifact never
// mutates the caller's Artifact (allows render-then-render-again).
func copyNode(n *yaml.Node) *yaml.Node {
	if n == nil {
		return nil
	}
	out := &yaml.Node{
		Kind:        n.Kind,
		Style:       n.Style,
		Tag:         n.Tag,
		Value:       n.Value,
		Anchor:      n.Anchor,
		Alias:       n.Alias,
		HeadComment: n.HeadComment,
		LineComment: n.LineComment,
		FootComment: n.FootComment,
		Line:        n.Line,
		Column:      n.Column,
	}
	if len(n.Content) > 0 {
		out.Content = make([]*yaml.Node, len(n.Content))
		for i, c := range n.Content {
			out.Content[i] = copyNode(c)
		}
	}
	return out
}
