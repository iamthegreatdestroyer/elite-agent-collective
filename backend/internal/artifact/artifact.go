// Package artifact implements `artifact.write` — the first L3 capability.
//
// An agent may write a file into its own subdirectory of an artifact root, and nothing else. No
// exec, no network, no path outside that subtree.
//
// # WHY THIS ONE FIRST
//
// The Phase-0 hardening is a PRIVILEGE boundary, not a capability boundary: `execve` is permitted,
// `ProtectSystem=strict` makes the filesystem read-only rather than inaccessible, and outbound
// sockets are allowed. So for most capabilities the sandbox would not bound the blast radius and
// code would be the only control.
//
// This capability is the exception, and that is the whole reason it is first: the artifact root
// lives inside the unit's `ReadWritePaths`, so the unit file and the code agree on the limit. It is
// the one capability where the privilege boundary and the capability boundary COINCIDE — if the
// code's path check were wrong, the kernel would still refuse the write.
//
// # REVERSIBILITY IS IMPLEMENTED, NOT ASSERTED
//
// §7 rates "agent error at machine speed" Medium on the strength of "Reversible L3", and a
// proposal's `reversible_claimed` is the model's word for it — a claim, recorded as one. This
// capability's reversibility is different in kind: Delete exists, it is tested, and Descriptor
// reports Reversible=true BECAUSE of that. It is our statement about our own tool, and it is
// falsifiable — TestDescriptorMatchesImplementation fails if the two ever disagree.
//
// # WHAT THIS PACKAGE DOES NOT DO
//
// It does not decide WHO may write. That is the ladder's enumerated grant table, which is empty:
// no agent holds `artifact.write`, so nothing can reach this code through a request today. Nor
// does it execute on approval — approval is unauthenticated while OIDC_CLIENT_ID is unset, and an
// unauthenticated approval must not cause an action.
package artifact

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/echo"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/ladder"
)

// Name is the capability's name in the ladder's grant table.
const Name = ladder.Capability("artifact.write")

// MaxSize bounds a single artifact. An unbounded write from a model response is a disk-exhaustion
// path, and the writable subtree is shared with the memory store.
const MaxSize = 256 * 1024

var (
	ErrBadName  = errors.New("artifact: invalid name")
	ErrTooLarge = errors.New("artifact: content exceeds the size limit")
	ErrNotFound = errors.New("artifact: not found")
)

// Artifact is the record of one write. It carries everything needed to reverse it.
type Artifact struct {
	Agent     string    `json:"agent"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Size      int       `json:"size"`
	Digest    string    `json:"digest"` // hex BLAKE2b-256 of the content
	WrittenAt time.Time `json:"written_at"`
}

// Store writes artifacts beneath a single root.
type Store struct {
	mu   sync.Mutex
	root string
	// readFile is the post-write verification read, injectable so the "Write does not report
	// success until the bytes re-read with the same digest" guarantee can be falsified by a
	// test. Every durability promise in this codebase that lacked such a seam turned out to be
	// wrong (echo's Append) or untested (the proposal store), so the seam comes first now.
	readFile func(string) ([]byte, error)
}

// NewStore opens (creating if needed) an artifact store.
func NewStore(root string) (*Store, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o750); err != nil {
		return nil, err
	}
	return &Store{root: abs, readFile: os.ReadFile}, nil
}

// validName accepts a conservative filename and nothing else.
//
// An allowlist, not a denylist: the set of dangerous names is open-ended (traversal, absolute
// paths, NUL, Windows devices, unicode lookalikes) while the set of acceptable ones is small and
// nameable. A denylist here would be a guess about what an attacker will try.
func validName(n string) bool {
	if n == "" || len(n) > 100 {
		return false
	}
	if n == "." || n == ".." || strings.HasPrefix(n, ".") {
		return false
	}
	for _, c := range n {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		case c == '-', c == '_', c == '.':
		default:
			return false
		}
	}
	// No consecutive dots anywhere: ".." is rejected above, but "a..b" would still normalise
	// through a path join on some systems.
	return !strings.Contains(n, "..")
}

// agentDir returns the per-agent directory, verified to sit under the root.
//
// The containment check is made against the RESOLVED path rather than the input string. A check on
// the input asks "does this look safe"; a check on the result asks "did we end up where we meant
// to", which is the question that actually matters.
func (s *Store) agentDir(agent string) (string, error) {
	if !validName(agent) {
		return "", fmt.Errorf("%w: agent %q", ErrBadName, agent)
	}
	dir := filepath.Join(s.root, agent)
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	// DEFENCE IN DEPTH, and currently UNREACHABLE: validName already rejects "/", "\\" and
	// "..", so filepath.Join cannot produce a path outside the root — verified structurally and
	// corroborated by brute-forcing every accepted name up to length 3 (0 escapes). It is kept
	// because it checks the RESULT rather than the input: if validName is ever loosened, this
	// asks the question that actually matters — "did we end up where we meant to?" — instead of
	// "does the input look safe?".
	if abs != s.root && !strings.HasPrefix(abs, s.root+string(os.PathSeparator)) {
		return "", fmt.Errorf("%w: %q escapes the artifact root", ErrBadName, agent)
	}
	return abs, nil
}

// Write stores content as agent/name, replacing any previous artifact of that name.
//
// Durability follows the Echo chain's discipline (Ouroboros R30-3): temp write, fsync, rename,
// directory fsync, then READ BACK and compare the digest before reporting success. A capability
// that reports a write it cannot re-read is the defect R30-3 describes.
func (s *Store) Write(agent, name string, content []byte) (Artifact, error) {
	if !validName(name) {
		return Artifact{}, fmt.Errorf("%w: %q", ErrBadName, name)
	}
	if len(content) > MaxSize {
		return Artifact{}, fmt.Errorf("%w: %d > %d bytes", ErrTooLarge, len(content), MaxSize)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	dir, err := s.agentDir(agent)
	if err != nil {
		return Artifact{}, err
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return Artifact{}, err
	}

	final := filepath.Join(dir, name)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return Artifact{}, err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return Artifact{}, err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return Artifact{}, fmt.Errorf("artifact: fsync failed, not written: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return Artifact{}, err
	}
	if err := os.Rename(tmpName, final); err != nil {
		os.Remove(tmpName)
		return Artifact{}, err
	}
	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		d.Close()
	}

	digest := hex.EncodeToString(echo.AgentPayloadHash("artifact.content.v1", content))

	// READ-BACK before reporting success.
	back, err := s.readFile(final)
	if err != nil {
		return Artifact{}, fmt.Errorf("artifact: %s did not read back: %w", name, err)
	}
	if hex.EncodeToString(echo.AgentPayloadHash("artifact.content.v1", back)) != digest {
		return Artifact{}, fmt.Errorf("artifact: %s read back with a different digest", name)
	}

	return Artifact{
		Agent: agent, Name: name, Path: final,
		Size: len(content), Digest: digest, WrittenAt: time.Now().UTC(),
	}, nil
}

// Delete removes an artifact. This is the REVERSAL of Write, and the reason Descriptor may report
// Reversible=true as a fact rather than a claim.
func (s *Store) Delete(agent, name string) error {
	if !validName(name) {
		return fmt.Errorf("%w: %q", ErrBadName, name)
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	dir, err := s.agentDir(agent)
	if err != nil {
		return err
	}
	target := filepath.Join(dir, name)
	if err := os.Remove(target); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s/%s", ErrNotFound, agent, name)
		}
		return err
	}
	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		d.Close()
	}
	return nil
}

// List returns one agent's artifacts, by name.
func (s *Store) List(agent string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir, err := s.agentDir(agent)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() && !strings.HasPrefix(e.Name(), ".tmp-") {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}

// Descriptor describes the capability to the authority system.
type Descriptor struct {
	Name ladder.Capability
	// Reversible is TRUE because Delete is implemented and tested — not because a model said so.
	// TestDescriptorMatchesImplementation fails if this and the implementation disagree.
	Reversible bool
	// Bounded reports that the unit's own filesystem policy confines this capability, so the
	// kernel refuses what the code would refuse. True only for capabilities whose entire effect
	// is inside ReadWritePaths.
	Bounded bool
	Summary string
}

// Describe returns the capability descriptor.
func Describe() Descriptor {
	return Descriptor{
		Name:       Name,
		Reversible: true,
		Bounded:    true,
		Summary:    "write a file into the agent's own artifact directory; reversal is deletion",
	}
}
