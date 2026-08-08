package changelog

// The JSON shape emitted by `changelog release --json`. Keys are camelCase,
// matching every other JSON surface in srekit (the YAML artifacts are
// snake_case because they are hand-edited; JSON is machine contract).

// Description is the parsed document rendered as data.
type Description struct {
	Path       string              `json:"path"`
	Vocabulary string              `json:"vocabulary"`
	Unreleased *SectionDescriptor  `json:"unreleased"`
	Versions   []SectionDescriptor `json:"versions"`
	Links      []LinkDef           `json:"links"`
}

// SectionDescriptor is one `## ` section as data.
type SectionDescriptor struct {
	Version    string             `json:"version,omitempty"`
	Date       string             `json:"date,omitempty"`
	Yanked     bool               `json:"yanked"`
	WellFormed bool               `json:"wellFormed"`
	Line       int                `json:"line"`
	Changes    []ChangeDescriptor `json:"changes"`
}

// ChangeDescriptor is one `### ` change-type subsection as data.
type ChangeDescriptor struct {
	Type       string `json:"type"`
	Recognized bool   `json:"recognized"`
	HasEntries bool   `json:"hasEntries"`
	Line       int    `json:"line"`
}

// Describe renders the document as the JSON payload. path is echoed back so
// a consumer piping several files apart can tell them apart.
func (d *Document) Describe(path string) Description {
	out := Description{
		Path:       path,
		Vocabulary: d.Lang,
		Versions:   []SectionDescriptor{},
		Links:      d.Links,
	}
	if out.Links == nil {
		out.Links = []LinkDef{}
	}
	if d.Unreleased != nil {
		s := d.describeSection(d.Unreleased)
		out.Unreleased = &s
	}
	for _, v := range d.Versions {
		out.Versions = append(out.Versions, d.describeSection(v))
	}
	return out
}

func (d *Document) describeSection(s *Section) SectionDescriptor {
	out := SectionDescriptor{
		Version:    s.Version,
		Date:       s.Date,
		Yanked:     s.Yanked,
		WellFormed: s.WellFormed,
		Line:       s.Line,
		Changes:    []ChangeDescriptor{},
	}
	for _, c := range s.Changes {
		out.Changes = append(out.Changes, ChangeDescriptor{
			Type:       c.Heading,
			Recognized: c.Recognized,
			HasEntries: c.HasEntries(d.Src),
			Line:       c.Line,
		})
	}
	return out
}
