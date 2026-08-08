package sections

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Payload is the structured input a generator accepts through `--from`:
// a JSON object with optional `meta` and `sections` maps. Both are
// optional, so a caller can supply only the section bodies it cares
// about and let the rest fall back to the artifact's defaults.
//
// It lives here rather than in a command file because its shape is
// artifact-shaped, not command-shaped — `{meta, sections}` means the
// same thing for every artifact, and more than one command reads it.
type Payload struct {
	Meta     map[string]string `json:"meta"`
	Sections map[string]string `json:"sections"`
}

// ReadPayload loads structured input named by path. An empty path
// returns a zero-value payload, which is how a generator expresses
// "no --from given: render every section from the artifact defaults".
// A path of "-" reads stdin, which the caller supplies so the command's
// own input stream (and therefore its tests) stay in control.
//
// An empty file is treated as no input rather than as an error: a
// pipeline stage that produced nothing should not look like a syntax
// error. Malformed JSON fails with the file named, because the whole
// point of the message is telling the caller which file to open.
func ReadPayload(stdin io.Reader, path string) (Payload, error) {
	var p Payload
	if path == "" {
		return p, nil
	}
	var raw []byte
	var err error
	if path == "-" {
		raw, err = io.ReadAll(stdin)
	} else {
		raw, err = os.ReadFile(path)
	}
	if err != nil {
		return p, fmt.Errorf("read --from %s: %w", path, err)
	}
	if len(raw) == 0 {
		return p, nil
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return p, fmt.Errorf("parse --from %s: %w", path, err)
	}
	return p, nil
}

// PickNonEmpty returns the first non-empty string from values. It is how
// generators express the flag-beats-file-beats-derived precedence: pass
// the flag first, the payload's meta value second, the derived value
// last.
func PickNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
