package artifact

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/ladder"
)

func newStore(t *testing.T) (*Store, string) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "artifacts")
	s, err := NewStore(root)
	if err != nil {
		t.Fatal(err)
	}
	return s, s.root
}

func TestWriteReadBackAndList(t *testing.T) {
	s, root := newStore(t)
	a, err := s.Write("FORGE", "analysis.md", []byte("# findings\n"))
	if err != nil {
		t.Fatal(err)
	}
	if a.Size != 11 || a.Digest == "" {
		t.Errorf("bad artifact record: %+v", a)
	}
	// It landed exactly where the agent's own subtree is, and nowhere else.
	want := filepath.Join(root, "FORGE", "analysis.md")
	if a.Path != want {
		t.Errorf("path = %s, want %s", a.Path, want)
	}
	body, err := os.ReadFile(want)
	if err != nil || string(body) != "# findings\n" {
		t.Errorf("content on disk = %q (%v)", body, err)
	}
	names, err := s.List("FORGE")
	if err != nil || len(names) != 1 || names[0] != "analysis.md" {
		t.Errorf("List = %v (%v)", names, err)
	}
}

// REVERSIBILITY, demonstrated rather than declared. This is what lets Describe() report
// Reversible=true as a fact.
func TestDeleteReversesWrite(t *testing.T) {
	s, root := newStore(t)
	if _, err := s.Write("FORGE", "tmp.txt", []byte("x")); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete("FORGE", "tmp.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "FORGE", "tmp.txt")); !os.IsNotExist(err) {
		t.Error("the artifact survived its reversal")
	}
	if names, _ := s.List("FORGE"); len(names) != 0 {
		t.Errorf("List still reports %v", names)
	}
	// Reversing twice is an error, not a silent success — "already gone" and "removed" are
	// different outcomes and a reversal log must be able to tell them apart.
	if err := s.Delete("FORGE", "tmp.txt"); !errors.Is(err, ErrNotFound) {
		t.Errorf("second delete: %v, want ErrNotFound", err)
	}
}

// The descriptor must not be able to lie about the implementation.
func TestDescriptorMatchesImplementation(t *testing.T) {
	d := Describe()
	if d.Name != Name {
		t.Errorf("descriptor name %q != %q", d.Name, Name)
	}
	if !d.Reversible {
		t.Error("descriptor claims irreversible, but Delete exists and is tested")
	}
	// Prove the claim: a write followed by its reversal leaves no trace.
	s, _ := newStore(t)
	if _, err := s.Write("A", "f", []byte("v")); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete("A", "f"); err != nil {
		t.Fatalf("descriptor claims Reversible=true but reversal failed: %v", err)
	}
	if names, _ := s.List("A"); len(names) != 0 {
		t.Error("descriptor claims Reversible=true but the artifact remains")
	}
}

// Every effect must land inside the root — the property that makes the unit's ReadWritePaths a
// second, independent bound on this capability.
func TestEffectsStayInsideTheRoot(t *testing.T) {
	s, root := newStore(t)
	outside := filepath.Join(filepath.Dir(root), "ESCAPED")

	bad := []struct{ agent, name string }{
		{"FORGE", "../ESCAPED"}, {"FORGE", "../../ESCAPED"}, {"FORGE", "/etc/passwd"},
		{"FORGE", ".."}, {"FORGE", "."}, {"FORGE", ""}, {"FORGE", ".hidden"},
		{"FORGE", "a/b"}, {"FORGE", "a\b"}, {"FORGE", "a\x00b"}, {"FORGE", "a..b"},
		{"../OTHER", "f"}, {"..", "f"}, {"/abs", "f"}, {"", "f"},
		{"FORGE", strings.Repeat("x", 101)},
	}
	for _, c := range bad {
		if _, err := s.Write(c.agent, c.name, []byte("pwned")); err == nil {
			t.Errorf("Write(%q,%q) succeeded", c.agent, c.name)
		}
		if err := s.Delete(c.agent, c.name); err == nil {
			t.Errorf("Delete(%q,%q) succeeded", c.agent, c.name)
		}
	}
	if _, err := os.Stat(outside); !os.IsNotExist(err) {
		t.Fatal("something was written outside the artifact root")
	}
	// Nothing leaked into the root itself either — writes belong in a per-agent subdirectory.
	entries, _ := os.ReadDir(root)
	for _, e := range entries {
		if !e.IsDir() {
			t.Errorf("stray file at the artifact root: %s", e.Name())
		}
	}
}

// Agents cannot reach each other's directories.
func TestAgentsAreIsolated(t *testing.T) {
	s, _ := newStore(t)
	if _, err := s.Write("FORGE", "secret.txt", []byte("forge only")); err != nil {
		t.Fatal(err)
	}
	names, err := s.List("SCRIBE")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Errorf("SCRIBE sees FORGE's artifacts: %v", names)
	}
	if err := s.Delete("SCRIBE", "secret.txt"); !errors.Is(err, ErrNotFound) {
		t.Errorf("SCRIBE deleted from FORGE's directory: %v", err)
	}
}

func TestSizeLimit(t *testing.T) {
	s, _ := newStore(t)
	if _, err := s.Write("A", "big", bytes.Repeat([]byte{'x'}, MaxSize+1)); !errors.Is(err, ErrTooLarge) {
		t.Errorf("oversize write: %v, want ErrTooLarge", err)
	}
	if _, err := s.Write("A", "ok", bytes.Repeat([]byte{'x'}, MaxSize)); err != nil {
		t.Errorf("write at exactly the limit was refused: %v", err)
	}
}

// A rewrite replaces atomically and leaves no temp files behind.
func TestRewriteIsAtomicAndLeavesNoTemps(t *testing.T) {
	s, root := newStore(t)
	for i := 0; i < 3; i++ {
		if _, err := s.Write("A", "f.txt", []byte(strings.Repeat("v", i+1))); err != nil {
			t.Fatal(err)
		}
	}
	body, _ := os.ReadFile(filepath.Join(root, "A", "f.txt"))
	if string(body) != "vvv" {
		t.Errorf("content = %q, want %q", body, "vvv")
	}
	entries, _ := os.ReadDir(filepath.Join(root, "A"))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
	if names, _ := s.List("A"); len(names) != 1 {
		t.Errorf("List = %v, want one entry", names)
	}
}

// THE SWITCH IS OFF. The capability exists and is granted to nobody, so no request can reach it.
// This is deliberate: approval is unauthenticated while OIDC_CLIENT_ID is unset, and an
// unauthenticated approval must not cause an action.
//
// This test fails the day a grant appears, which is exactly when the end-to-end path (propose ->
// approve -> act) must be built and tested.
func TestCapabilityIsGrantedToNobody(t *testing.T) {
	for _, codename := range []string{"FORGE", "SENTRY", "VERTEX", "STREAM", "ATLAS", "OMNISCIENT", "APEX", "SCRIBE"} {
		a := modelsAgent(codename)
		if ladder.HasGrant(a, Name) {
			t.Errorf("%s holds %s — L3 is now reachable. Build and test the propose->approve->act "+
				"path, confirm approval is AUTHENTICATED, then update this test deliberately.", codename, Name)
		}
	}
}
