package artifact

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// Write must not report success until the bytes re-read with the SAME digest. Recomputing is the
// point: comparing the stored digest against itself is a check that cannot fail, and a mutation
// doing exactly that passed every other test in this package.
func TestWriteFailsWhenReadBackDiffers(t *testing.T) {
	s, _ := newStore(t)

	// The write lands, but the verification read returns different bytes.
	s.readFile = func(string) ([]byte, error) { return []byte("not what was written"), nil }
	if _, err := s.Write("A", "f.txt", []byte("original")); err == nil {
		t.Error("Write reported success although the file read back with a different digest")
	} else if !strings.Contains(err.Error(), "digest") {
		t.Errorf("error does not name the digest mismatch: %v", err)
	}

	// A verification read that fails outright is also fatal to the write.
	s.readFile = func(string) ([]byte, error) { return nil, errors.New("io error") }
	if _, err := s.Write("A", "g.txt", []byte("x")); err == nil {
		t.Error("Write reported success although the file could not be read back")
	}

	// And with the real reader it succeeds again — the seam is not masking a broken path.
	s.readFile = os.ReadFile
	if _, err := s.Write("A", "h.txt", []byte("x")); err != nil {
		t.Errorf("Write failed with the real reader: %v", err)
	}
}
