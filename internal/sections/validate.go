package sections

import "strings"

// ValidationResult is one row in the report from RequiredCheck. Status is
// "OK" or "FAIL"; Reason is populated only when Status == "FAIL".
type ValidationResult struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

// RequiredCheck reports per-required-section whether the input contains a
// non-empty body. Non-required sections are not included in the result.
//
// Unknown-ID typo guarding is left to sections.Merge: callers should
// invoke Merge first (which returns an error on unknown IDs) and only
// then run RequiredCheck if they want the completeness signal.
func (m *Manifest) RequiredCheck(input map[string]string) []ValidationResult {
	var out []ValidationResult
	for _, s := range m.Sections {
		if !s.Required {
			continue
		}
		body := input[s.ID]
		res := ValidationResult{ID: s.ID, Title: s.Title}
		if strings.TrimSpace(body) == "" {
			res.Status = "FAIL"
			res.Reason = "required body is empty"
		} else {
			res.Status = "OK"
		}
		out = append(out, res)
	}
	return out
}
