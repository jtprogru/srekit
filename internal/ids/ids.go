package ids

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

func UUID() string {
	return uuid.New().String()
}

// Short returns the first n characters of u, or u unchanged if it is
// already shorter than n. Used for human-friendly short identifiers
// derived from a full UUID (e.g. RFC-12345678).
func Short(u string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(u) <= n {
		return u
	}
	return u[:n]
}

var slugCleaner = regexp.MustCompile(`[^a-z0-9]+`)

func Slug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugCleaner.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "untitled"
	}
	return s
}
