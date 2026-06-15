package ids

import (
	"fmt"
	"math/rand/v2"
	"regexp"
	"strings"
)

// UUID returns an RFC 4122 version 4 UUID. The identifiers label generated
// artifacts (RFCs, postmortems, …) and are not security-sensitive, so we use
// math/rand/v2 instead of crypto/rand. That keeps the whole crypto stack
// (and the FIPS-140 module the Go toolchain links with it) out of the binary.
func UUID() string {
	var b [16]byte
	for i := range b {
		b[i] = byte(rand.UintN(256)) //nolint:gosec // G404: artifact IDs are not security-sensitive; math/rand/v2 is deliberate (see doc comment)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
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
