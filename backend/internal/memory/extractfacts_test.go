package memory

import (
	"strings"
	"testing"
)

// TestExtractFacts_NoPanicOnWideningUnicode guards the ToLower byte-widening
// slice bug: U+023A lowercases to U+2C65 (2 bytes -> 3 bytes), so offsets from
// the lowercased copy can overrun the original message. ExtractFacts must not
// panic (it runs on the request hot path).
func TestExtractFacts_NoPanicOnWideningUnicode(t *testing.T) {
	cases := []string{
		strings.Repeat("Ⱥ", 100) + "i'm using rust",
		strings.Repeat("Ⱥ", 100) + "i'm working on a crypto vault",
		strings.Repeat("Ⱥ", 50) + "we use go with using docker",
	}
	for _, msg := range cases {
		_ = ExtractFacts("APEX", msg) // must not panic
	}
}

// TestExtractFacts_AsciiStillWorks confirms the guards don't regress the
// normal ASCII extraction path.
func TestExtractFacts_AsciiStillWorks(t *testing.T) {
	facts := ExtractFacts("APEX", "hey, i'm using rust for this")
	found := false
	for _, f := range facts {
		if f.Key == "language" && strings.Contains(f.Value, "rust") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a rust language fact, got %+v", facts)
	}
}
