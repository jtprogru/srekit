package ids

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

func UUID() string {
	return uuid.New().String()
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
