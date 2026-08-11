package proposal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Store persists proposals as one JSON file each.
//
// Durability follows the same discipline as the Echo chain (Ouroboros R30-3): write to a temp
// file, fsync, rename, fsync the directory, then READ THE RECORD BACK and verify it before
// reporting success. A store that reports a write it cannot re-read is the defect R30-3 describes
// — marked current without being read back.
type Store struct {
	mu  sync.RWMutex
	dir string
	// readBack is the post-write verification step, injectable so the "Put does not report
	// success until the record re-reads correctly" guarantee can actually be falsified by a
	// test. Without a seam that promise is only a comment — and echo's durability bug was
	// found the moment its equivalent promise became testable.
	readBack func(*Store, string) (*Proposal, error)
}

// NewStore opens (creating if needed) a proposal store rooted at dir.
func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}
	return &Store{dir: dir, readBack: (*Store).getLocked}, nil
}

// idPattern guards the filesystem boundary: an ID is always 16 lowercase hex characters derived
// from the commitment, so anything else is either corruption or a path-traversal attempt. Deriving
// IDs from content rather than accepting them from callers is what makes this check total.
func validID(id string) bool {
	if len(id) != 16 {
		return false
	}
	for _, c := range id {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return false
		}
	}
	return true
}

func (s *Store) path(id string) string { return filepath.Join(s.dir, id+".json") }

// Put writes a proposal, then re-reads and verifies it before returning.
func (s *Store) Put(p *Proposal) error {
	if !validID(p.ID) {
		return fmt.Errorf("%w: id %q is not a 16-char hex digest prefix", ErrInvalid, p.ID)
	}
	if err := p.VerifyCommitment(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	body, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(s.dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := func() { tmp.Close(); os.Remove(tmpName) }

	if _, err := tmp.Write(body); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("proposal: fsync failed, not stored: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, s.path(p.ID)); err != nil {
		os.Remove(tmpName)
		return err
	}
	// fsync the directory so the rename itself is durable, not just the file's contents.
	if d, err := os.Open(s.dir); err == nil {
		_ = d.Sync()
		d.Close()
	}

	// READ-BACK (R30-3): confirm what landed is what we wrote, before reporting success.
	got, err := s.readBack(s, p.ID)
	if err != nil {
		return fmt.Errorf("proposal: %s did not read back after write: %w", p.ID, err)
	}
	if got.Commitment != p.Commitment || got.Status != p.Status {
		return fmt.Errorf("proposal: %s read back different from what was written", p.ID)
	}
	return nil
}

func (s *Store) getLocked(id string) (*Proposal, error) {
	if !validID(id) {
		return nil, fmt.Errorf("%w: malformed id", ErrNotFound)
	}
	body, err := os.ReadFile(s.path(id))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
		}
		return nil, err
	}
	var p Proposal
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, fmt.Errorf("%w: %s is unreadable: %v", ErrInvalid, id, err)
	}
	// Verify on every read, not only on write. A stored proposal is an artifact a human rules
	// on; if its content no longer matches its commitment, the honest answer is an error, not
	// the altered content.
	if err := p.VerifyCommitment(); err != nil {
		return nil, err
	}
	return &p, nil
}

// Get returns one proposal, verifying its commitment.
func (s *Store) Get(id string) (*Proposal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getLocked(id)
}

// List returns all proposals, newest first. Unreadable or tampered entries are returned as errors
// rather than skipped: silently omitting a corrupted proposal would make the list look healthy.
func (s *Store) List() ([]*Proposal, []error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, []error{err}
	}
	var out []*Proposal
	var errs []error
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".json") {
			continue
		}
		p, err := s.getLocked(strings.TrimSuffix(name, ".json"))
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
			continue
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, errs
}
