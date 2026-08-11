package retrieval

import (
	"os"
	"strings"
	"testing"
	"unicode/utf8"
)

// liveHeader is the boilerplate as it appears on a real hit once whitespace is
// collapsed — copied from a 2026-08-11 sigma-index response, not invented.
const liveHeader = "project agentos echo chain 2026 08 11 > **Mirrored from Claude Code memory** " +
	"· source `memory/project_agentos_echo_chain_2026-08-11.md` > > "

// liveNote reproduces the note that produced the headline failure: the fact the
// agent got wrong sits at char 318 of the collapsed text, and the old window
// closed at 280. It was handed the right note, cut 38 characters short of the
// answer, and asserted SHA-256 with "Uncertainty? None."
func liveNote() string {
	body := "DEPLOYED 2026-08-11 (e89aa23 + b8674d6); recording is OFF pending one owner action " +
		"(sudo bash /tmp/enable-echo.sh): internal/echo implements the Ouroboros v1.19 §6.3 hash " +
		"chain over agent authority decisions with H(x) = BLAKE2b-256, u32be length prefixes, " +
		"domain separators and a genesis anchor. " + strings.Repeat("padding to note length. ", 200)
	return liveHeader + body
}

func TestStripMirrorHeaderRemovesLiveBoilerplate(t *testing.T) {
	got := stripMirrorHeader(liveNote())
	if strings.Contains(got, "Mirrored from Claude Code memory") {
		t.Fatalf("header survived stripping")
	}
	if !strings.HasPrefix(got, "DEPLOYED 2026-08-11") {
		t.Fatalf("body start wrong: %.60q", got)
	}
	if n := len(liveNote()) - len(got); n < 118 || n > 158 {
		t.Fatalf("stripped %d chars, outside the 118-158 measured across 40 live hits", n)
	}
}

// TestStripMirrorHeaderHandlesLiveVariants uses headers copied verbatim from
// sigma-index responses, not headers I composed. The second form — no trailing
// "> >" because the body starts on a heading — is the one that defeated the
// first pattern in production while every unit test passed.
func TestStripMirrorHeaderHandlesLiveVariants(t *testing.T) {
	for _, tc := range []struct{ name, in, wantPrefix string }{
		{"body opens on a quote",
			"project agentos failclosed 2026 08 11 > **Mirrored from Claude Code memory** " +
				"· source `memory/project_agentos_failclosed_2026-08-11.md` > > DEPLOYED (2a32dc6): audit is fail-OPEN",
			"DEPLOYED (2a32dc6)"},
		{"body opens on a heading (MEMORY.md)",
			"memory > **Mirrored from Claude Code memory** · source `memory/MEMORY.md` " +
				"# MEMORY INDEX ## READ THIS BEFORE ANYTHING ELSE",
			"# MEMORY INDEX"},
	} {
		got := stripMirrorHeader(tc.in)
		if strings.Contains(got, "Mirrored from Claude Code memory") {
			t.Errorf("%s: header survived: %.90q", tc.name, got)
		}
		if !strings.HasPrefix(got, tc.wantPrefix) {
			t.Errorf("%s: got %.60q, want prefix %q", tc.name, got, tc.wantPrefix)
		}
	}
}

func TestStripMirrorHeaderLeavesOtherTextAlone(t *testing.T) {
	// A note that is not a Claude Code mirror must pass through untouched, and
	// so must text that merely looks header-ish. Stripping content would be a
	// silent data loss the caller cannot detect.
	for _, s := range []string{
		"plain note text with no header at all",
		"some slug > source `memory/foo.md` > > body", // no Mirrored marker
		"a > **Mirrored from somewhere else** · source `memory/x.md` > > body",
		strings.Repeat("x", 200) + " > **Mirrored from Claude Code memory** · source `memory/y.md` > > body",
	} {
		if got := stripMirrorHeader(s); got != s {
			t.Errorf("modified text it should not have:\n in: %.80q\nout: %.80q", s, got)
		}
	}
}

// TestExcerptReachesTheFactTheOldWindowCut is the regression pin for the actual
// defect. It asserts BOTH directions: the new policy reaches BLAKE2b and the
// old one does not. Without the second assertion the test could pass on a
// build where excerpt() returns the whole note, proving nothing.
func TestExcerptReachesTheFactTheOldWindowCut(t *testing.T) {
	collapsed := strings.Join(strings.Fields(liveNote()), " ")

	oldPolicy := string([]rune(collapsed)[:280]) // header kept, single 280 window
	if strings.Contains(oldPolicy, "BLAKE2b") {
		t.Fatalf("control failed: the OLD policy already showed the fact, so this " +
			"test cannot demonstrate the fix")
	}

	got := excerpt(stripMirrorHeader(collapsed), "What hash function does the Echo chain use?", 350, 350)
	if !strings.Contains(got, "BLAKE2b") {
		t.Fatalf("fact still not in the block:\n%s", got)
	}
}

// TestExcerptAlwaysContainsTheHead pins the monotonicity property. A candidate
// that slid a single window to the query's best match scored better overall but
// LOST two facts a plain head already delivered; the head is therefore taken
// first and unconditionally. This test fails the day someone "optimises" that
// away.
func TestExcerptAlwaysContainsTheHead(t *testing.T) {
	text := strings.Repeat("alpha beta gamma delta ", 400)
	for _, head := range []int{1, 50, 350, 1000} {
		got := excerpt(text, "gamma delta epsilon", head, 350)
		want := string([]rune(text)[:head])
		if !strings.HasPrefix(got, want) {
			t.Errorf("head=%d: block does not start with the head window", head)
		}
	}
}

func TestCentredWindowFindsABuriedFact(t *testing.T) {
	// Fact placed far past any plausible head window: only query-centring reaches it.
	text := strings.Repeat("unrelated filler sentence. ", 60) +
		"the environment variable EAC_ECHO_PATH enables recording. " +
		strings.Repeat("more unrelated filler. ", 60)
	got := excerpt(text, "Which environment variable enables the Echo audit chain?", 350, 350)
	if !strings.Contains(got, "EAC_ECHO_PATH") {
		t.Fatalf("centred window missed the fact it was pointed at:\n%s", got)
	}
	if strings.Contains(string([]rune(text)[:350]), "EAC_ECHO_PATH") {
		t.Fatal("control failed: the fact was inside the head, so centring proved nothing")
	}
}

func TestCentredWindowDegradesToHeadWithoutUsableTerms(t *testing.T) {
	text := strings.Repeat("some note body text here. ", 100)
	got := excerpt(text, "is it a b c?", 100, 100) // every term shorter than 4 chars
	if !strings.HasPrefix(got, string([]rune(text)[:100])) {
		t.Fatal("lost the head when the query had no usable terms")
	}
	if utf8.RuneCountInString(got) > 100+3+100 {
		t.Fatalf("block exceeded head+excerpt budget: %d runes", utf8.RuneCountInString(got))
	}
}

func TestFormatBlockStripsBoilerplateAndBoundsSize(t *testing.T) {
	hits := []indexHit{
		{ID: "note-a", Score: 0.03, Text: liveNote()},
		{ID: "note-b", Score: 0.02, Text: liveNote()},
	}
	block := formatBlock(hits, "What hash function does the Echo chain use?", 350, 350)

	if strings.Contains(block, "Mirrored from Claude Code memory") {
		t.Error("boilerplate reached the model — 138 chars/hit of the budget wasted")
	}
	if !strings.Contains(block, "BLAKE2b") {
		t.Error("the fact did not reach the block")
	}
	if !strings.Contains(block, "note-a") || !strings.Contains(block, "note-b") {
		t.Error("note IDs missing — the model cannot cite what it cannot name")
	}
	// Budget: 2 hits x (350 + 3 + 350) + per-line framing. Generous ceiling; the
	// point is that it is BOUNDED, since every character is ~0.05s of prefill.
	if n := utf8.RuneCountInString(block); n > 2*(350+3+350)+300 {
		t.Errorf("block is %d runes — unbounded growth would eat the answer's time budget", n)
	}
}

func TestExcerptIsRuneSafe(t *testing.T) {
	// Notes carry '·', '§', '→' and emoji. Slicing by byte would split a rune
	// and hand the model replacement characters.
	text := "§ 6.3 échó · chain → BLAKE2b-256 ✅ " + strings.Repeat("ünïcödé pädding ", 200)
	got := excerpt(text, "which hash does the écho chain use", 40, 60)
	if !utf8.ValidString(got) {
		t.Fatal("produced invalid UTF-8")
	}
	got = excerpt("İ"+strings.Repeat("a", 500), "aaaa bbbb", 10, 20)
	if !utf8.ValidString(got) {
		t.Fatal("invalid UTF-8 with a special-casing rune present")
	}
}

func TestEnvIntRejectsUnusableValues(t *testing.T) {
	// A zero or negative window would silently retrieve nothing; a typo must
	// not be able to disable retrieval without saying so.
	for _, v := range []string{"", "0", "-3", "abc", " "} {
		os.Setenv("EAC_TEST_WINDOW", v)
		if got := envInt("EAC_TEST_WINDOW", 350); got != 350 {
			t.Errorf("envInt(%q) = %d, want the 350 default", v, got)
		}
	}
	os.Setenv("EAC_TEST_WINDOW", "7")
	if got := envInt("EAC_TEST_WINDOW", 350); got != 7 {
		t.Errorf("envInt(\"7\") = %d, want 7", got)
	}
	os.Unsetenv("EAC_TEST_WINDOW")
}

func TestDefaultsMatchTheMeasuredPolicy(t *testing.T) {
	// The defaults ARE the live configuration: the deployed unit sets only
	// EAC_RETRIEVER=1. If someone changes them, they should have to change this
	// test too, and read why 2/350/350 was chosen.
	for _, k := range []string{"EAC_RETRIEVER_K", "EAC_RETRIEVER_HEAD", "EAC_RETRIEVER_EXCERPT"} {
		os.Unsetenv(k)
	}
	os.Setenv("EAC_RETRIEVER", "1")
	defer os.Unsetenv("EAC_RETRIEVER")
	r := NewSigmaIndexRetriever()
	if r == nil {
		t.Fatal("retriever disabled with EAC_RETRIEVER=1")
	}
	if r.k != 2 || r.head != 350 || r.excerptLen != 350 {
		t.Fatalf("defaults are k=%d head=%d excerpt=%d; measured best was 2/350/350 "+
			"(7 of 8 facts at 1406 chars, vs 3 of 8 at 1400 for the old k=5/280)",
			r.k, r.head, r.excerptLen)
	}
}

// TestCentredWindowSurvivesRuneCountChange exercises the guard that the first
// mutation run showed was unpinned. 'İ' lowercases to TWO runes, so the
// lowercased copy is longer than the original and every index computed on it is
// shifted. When the best-scoring window is the last one, applying that index to
// the shorter original slices past its end.
//
// The İ must sit AFTER the head window or the centred path never sees it — the
// first version of this test put it at index 0 and proved nothing.
func TestCentredWindowSurvivesRuneCountChange(t *testing.T) {
	text := strings.Repeat("a", 60) + "İ" + strings.Repeat(" filler words here ", 40) +
		"the distinctive marker phrase sits at the very end"
	got := excerpt(text, "distinctive marker phrase", 20, 40)
	if !utf8.ValidString(got) {
		t.Fatal("invalid UTF-8")
	}
	if !strings.HasPrefix(got, string([]rune(text)[:20])) {
		t.Fatal("lost the head")
	}
}

// TestShortQueryWordsDoNotDrownTheDistinctiveTerm pins what the 4-char floor in
// queryTerms is for: a question is mostly common short words, and the window
// must be chosen by the rare term that identifies the fact.
func TestShortQueryWordsDoNotDrownTheDistinctiveTerm(t *testing.T) {
	text := strings.Repeat("is it in an on at to be by we do so no up ", 40) +
		"the answer is EAC_ECHO_PATH here " +
		strings.Repeat("is it in an on at to be by we do so no up ", 40)
	got := excerpt(text, "is it in an on so do we say which EAC_ECHO_PATH", 50, 120)
	if !strings.Contains(got, "EAC_ECHO_PATH") {
		t.Fatalf("centred on the common short words instead of the rare term:\n%s", got)
	}
}

// TestToLowerPreservesRuneCount pins the invariant centredWindow depends on:
// it computes window offsets on the lowercased copy and applies them to the
// original. Go's strings.ToLower maps rune to rune, so the two always line up.
// This is the check that replaced an unreachable defensive branch — a guard
// that can never trigger cannot be shown to work, and mutation testing simply
// reports it as dead. If this invariant ever changes, restore the guard.
func TestToLowerPreservesRuneCount(t *testing.T) {
	for _, s := range []string{"İ", "ẞ", "Σ", "ǅ", "ﬁ", "ⅩⅩ", "ÅÄÖ", "🙂", "§ 6.3 échó → ✅"} {
		if a, b := utf8.RuneCountInString(s), utf8.RuneCountInString(strings.ToLower(s)); a != b {
			t.Fatalf("ToLower(%q) changed rune count %d -> %d; centredWindow's index "+
				"arithmetic is no longer sound and needs its guard back", s, a, b)
		}
	}
}

// TestQueryTermCapKeepsTheDistinctiveTerm asserts the cap cannot evict the term
// that identifies the fact, whatever the question's length or word order. This
// now holds by construction (longest-first), so it is a property test, not a
// guard against one specific ordering.
func TestQueryTermCapKeepsTheDistinctiveTerm(t *testing.T) {
	query := "is it in an on at to be by we do so no up if or as of my me he she " +
		"you the and but not for was are had has him her its our out who why " +
		"which environment variable is EAC_ECHO_PATH"
	text := strings.Repeat("common filler words that appear everywhere. ", 40) +
		"the variable is EAC_ECHO_PATH and it enables recording " +
		strings.Repeat("more common filler words appearing everywhere. ", 40)

	got := excerpt(text, query, 50, 150)
	if !strings.Contains(got, "EAC_ECHO_PATH") {
		t.Fatalf("the rare term was crowded out of the term list by short words:\n%s", got)
	}
}

// TestQueryTermsAreLongestFirstAndCapped pins the ordering directly, so the
// property does not depend on a particular sample of prose.
func TestQueryTermsAreLongestFirstAndCapped(t *testing.T) {
	var long []string
	for i := 0; i < 40; i++ {
		long = append(long, "word"+strings.Repeat("x", i%3))
	}
	q := strings.Join(long, " ") + " EAC_ECHO_PATH_DISTINCTIVE"
	terms := queryTerms(q)

	if len(terms) > maxQueryTerms {
		t.Fatalf("returned %d terms, cap is %d", len(terms), maxQueryTerms)
	}
	for i := 1; i < len(terms); i++ {
		if len(terms[i]) > len(terms[i-1]) {
			t.Fatalf("not longest-first at %d: %q then %q", i, terms[i-1], terms[i])
		}
	}
	var found bool
	for _, x := range terms {
		if strings.Contains(x, "eac_echo_path_distinctive") {
			found = true
		}
	}
	if !found {
		t.Fatal("the longest, most distinctive term was evicted by the cap")
	}
}
