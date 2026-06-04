package sections

// JSONSchema returns a JSON Schema (draft 2020-12) describing the `--from`
// input payload for this manifest. `title` is used as the schema's `title`
// field; `metaFields` enumerates the meta object's known properties (the
// list comes from the consuming command — e.g. postmortem passes
// {"id","title","severity","start","end","owner","now"}).
//
// Every section body is typed as `string` regardless of section type:
// list and table are rendered into a markdown string at merge time, so
// the on-the-wire input format is always a string.
//
// The returned value is a `map[string]any` ready for json.MarshalIndent.
// Output keys are alphabetically sorted by the encoder, which makes
// schema diffs stable.
func (m *Manifest) JSONSchema(title string, metaFields []string) map[string]any {
	metaProps := make(map[string]any, len(metaFields))
	for _, f := range metaFields {
		metaProps[f] = map[string]any{"type": "string"}
	}

	sectionsProps := make(map[string]any, len(m.Sections))
	var required []string
	for _, s := range m.Sections {
		sectionsProps[s.ID] = map[string]any{
			"type":        "string",
			"description": s.Title + " (type: " + string(s.Type) + ")",
		}
		if s.Required {
			required = append(required, s.ID)
		}
	}

	sectionsSchema := map[string]any{
		"type":                 "object",
		"properties":           sectionsProps,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		sectionsSchema["required"] = required
	}

	return map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"title":   title,
		"type":    "object",
		"properties": map[string]any{
			"meta": map[string]any{
				"type":                 "object",
				"properties":           metaProps,
				"additionalProperties": false,
			},
			"sections": sectionsSchema,
		},
		"additionalProperties": false,
	}
}
