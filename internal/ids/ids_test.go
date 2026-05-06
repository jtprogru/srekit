package ids

import "testing"

func TestUUIDIsRFC4122(t *testing.T) {
	u := UUID()
	if len(u) != 36 || u[8] != '-' || u[13] != '-' || u[18] != '-' || u[23] != '-' {
		t.Fatalf("not a UUID: %q", u)
	}
}

func TestSlug(t *testing.T) {
	cases := map[string]string{
		"Hello World":                 "hello-world",
		"  Tail Latency на api-gw  ":  "tail-latency-api-gw",
		"---":                         "untitled",
		"":                            "untitled",
		"Foo / Bar :: Baz!":           "foo-bar-baz",
		"already-good-slug":           "already-good-slug",
	}
	for in, want := range cases {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
}
