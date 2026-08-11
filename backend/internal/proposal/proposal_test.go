package proposal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const tsFixed = int64(1700000000000000000)

func sample() *Proposal {
	p := &Proposal{
		Agent:     "FORGE",
		Title:     "Rebuild the index",
		Rationale: "The index is stale and blocks retrieval.",
		CreatedAt: time.Unix(0, tsFixed).UTC(),
		Steps: []Step{
			{Order: 1, Action: "rebuild", Target: "sigma-index", ReversibleClaimed: true},
			{Order: 2, Action: "restart", Target: "eac-backend", ReversibleClaimed: true},
		},
		Risks:  []string{"brief downtime during restart"},
		Status: StatusPending,
	}
	p.Commit()
	return p
}

// KNOWN-ANSWER vector, computed by a separate implementation from the documented encoding
// (BLAKE2b-256 over "sigma-agentos-payload-v1" ‖ LP("proposal.v2") ‖ LP(field)…).
//
// The vector CHANGED when Step.Reversible became Step.ReversibleClaimed — the field list
// moved, so every commitment moved, so Schema was bumped v1 -> v2. That is the mechanism
// working: the rename could not be made quietly.
//
// This is what makes Schema a promise: any change to the field list or its order changes every
// commitment and breaks this test, forcing the schema name to be bumped with it.
func TestCommitmentKnownAnswer(t *testing.T) {
	p := sample()
	const want = "a18ae55ed64948f5229f3a227bfcf2c9494134d4d7bd963949ba4e421ba9c5ba"
	if p.Commitment != want {
		t.Errorf("commitment\n got  %s\n want %s\n(if intentional, bump Schema)", p.Commitment, want)
	}
	if p.ID != want[:16] {
		t.Errorf("id = %s, want %s", p.ID, want[:16])
	}
}

// The counts of the variable-length lists must be hashed, or two different shapes present the
// same field sequence. Without them (steps=[a,b], risks=[]) and (steps=[a], risks=[b]) collide.
func TestListCountsAreCommitted(t *testing.T) {
	a := &Proposal{Agent: "X", Title: "T", Rationale: "R", CreatedAt: time.Unix(0, tsFixed),
		Steps: []Step{{Order: 1, Action: "one"}, {Order: 2, Action: "two"}}}
	b := &Proposal{Agent: "X", Title: "T", Rationale: "R", CreatedAt: time.Unix(0, tsFixed),
		Steps: []Step{{Order: 1, Action: "one"}}, Risks: []string{"two"}}
	a.Commit()
	b.Commit()
	if a.Commitment == b.Commitment {
		t.Error("a two-step proposal and a one-step-one-risk proposal share a commitment — list counts are not hashed")
	}
}

func TestVerifyCommitmentDetectsTampering(t *testing.T) {
	p := sample()
	if err := p.VerifyCommitment(); err != nil {
		t.Fatalf("fresh proposal failed to verify: %v", err)
	}
	p.Steps[0].Target = "something-else"
	if err := p.VerifyCommitment(); err == nil {
		t.Error("an altered step verified successfully")
	}
}

// A decision must NOT invalidate the commitment: the content stays verifiable across the whole
// lifecycle, which is the point of excluding Status and Decision from the hash.
func TestDecisionDoesNotInvalidateCommitment(t *testing.T) {
	p := sample()
	before := p.Commitment
	if err := p.Decide(Decision{Status: StatusApproved, At: time.Unix(0, tsFixed), Approver: "steve", Authenticated: true}); err != nil {
		t.Fatal(err)
	}
	if p.Commitment != before {
		t.Error("deciding changed the commitment")
	}
	if err := p.VerifyCommitment(); err != nil {
		t.Errorf("decided proposal no longer verifies: %v", err)
	}
	if p.Status != StatusApproved {
		t.Errorf("status = %s, want approved", p.Status)
	}
}

func TestDecideIsOnceOnly(t *testing.T) {
	p := sample()
	if err := p.Decide(Decision{Status: StatusApproved}); err != nil {
		t.Fatal(err)
	}
	if err := p.Decide(Decision{Status: StatusRejected}); err == nil {
		t.Error("a second decision was accepted — a ruling is not editable")
	}
}

// An unauthenticated decision must SAY SO in the record. Provenance travels with the artifact, so
// a later reader can tell a verified approval from an unverified one without knowing how the
// service was configured at the time.
func TestUnauthenticatedDecisionIsMarked(t *testing.T) {
	p := sample()
	if err := p.Decide(Decision{Status: StatusApproved, At: time.Now()}); err != nil {
		t.Fatal(err)
	}
	if p.Decision.Authenticated {
		t.Error("decision claims to be authenticated by default")
	}
	if p.Decision.Approver != "unauthenticated" {
		t.Errorf("approver = %q, want %q", p.Decision.Approver, "unauthenticated")
	}
}

// A model that cannot produce a valid proposal must produce NO proposal — there is deliberately
// no path that wraps prose in a proposal-shaped envelope.
func TestValidateRejectsNonProposals(t *testing.T) {
	base := sample()
	cases := map[string]func(*Proposal){
		"no agent":     func(p *Proposal) { p.Agent = "" },
		"no title":     func(p *Proposal) { p.Title = "  " },
		"no rationale": func(p *Proposal) { p.Rationale = "" },
		"no steps":     func(p *Proposal) { p.Steps = nil },
		"empty action": func(p *Proposal) { p.Steps = []Step{{Order: 1, Action: " "}} },
		"empty risk":   func(p *Proposal) { p.Risks = []string{""} },
		"huge title":   func(p *Proposal) { p.Title = strings.Repeat("x", MaxTitle+1) },
		"too many steps": func(p *Proposal) {
			p.Steps = make([]Step, MaxSteps+1)
			for i := range p.Steps {
				p.Steps[i] = Step{Order: i, Action: "a"}
			}
		},
	}
	for name, mutate := range cases {
		p := *base
		p.Steps = append([]Step(nil), base.Steps...)
		p.Risks = append([]string(nil), base.Risks...)
		mutate(&p)
		if err := p.Validate(); err == nil {
			t.Errorf("%s: accepted as a valid proposal", name)
		}
	}
	if err := base.Validate(); err != nil {
		t.Errorf("the valid sample was rejected: %v", err)
	}
}

// ── store ────────────────────────────────────────────────────────────────────

func TestStoreRoundTrip(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	p := sample()
	if err := s.Put(p); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Commitment != p.Commitment || got.Title != p.Title || len(got.Steps) != len(p.Steps) {
		t.Errorf("round trip differs:\n put %+v\n got %+v", p, got)
	}
}

// A stored proposal whose content is edited on disk must fail to load, not load altered. It is an
// artifact a human rules on; returning the altered content would be the worst outcome.
func TestStoreRejectsTamperedFile(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	p := sample()
	if err := s.Put(p); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, p.ID+".json")
	body, _ := os.ReadFile(path)
	var raw map[string]any
	json.Unmarshal(body, &raw)
	raw["title"] = "Delete everything" // content changed, commitment left intact
	edited, _ := json.Marshal(raw)
	os.WriteFile(path, edited, 0o600)

	if _, err := s.Get(p.ID); err == nil {
		t.Error("a tampered proposal loaded successfully")
	}
	_, errs := s.List()
	if len(errs) == 0 {
		t.Error("List silently omitted the tampered proposal instead of reporting it")
	}
}

// IDs come from content, never from callers — so a traversal attempt cannot reach outside the dir.
//
// ⚠️ The obvious version of this test passes for the WRONG REASON: with the ID guard removed,
// Get("../../etc/passwd") still errors, because that file does not exist as a proposal — absence,
// not rejection. Mutation testing caught it. So this plants a REAL, VALID proposal at a path a
// traversal would reach, and asserts Get still cannot load it. Now the only way to pass is to
// reject the ID.
func TestStoreRejectsMalformedIDs(t *testing.T) {
	root := t.TempDir()
	inner := filepath.Join(root, "store")
	s, err := NewStore(inner)
	if err != nil {
		t.Fatal(err)
	}

	// A genuine, loadable proposal one directory ABOVE the store.
	bait := sample()
	body, _ := json.MarshalIndent(bait, "", "  ")
	if err := os.WriteFile(filepath.Join(root, bait.ID+".json"), body, 0o600); err != nil {
		t.Fatal(err)
	}
	// Sanity: it really is loadable by a store rooted where it sits — so if the traversal
	// worked, Get WOULD return it, and this test would fail loudly instead of vacuously.
	if outer, err := NewStore(root); err != nil {
		t.Fatal(err)
	} else if _, err := outer.Get(bait.ID); err != nil {
		t.Fatalf("precondition: the planted proposal is not loadable: %v", err)
	}

	if _, err := s.Get("../" + bait.ID); err == nil {
		t.Error("traversal reached a proposal outside the store directory")
	}

	for _, bad := range []string{"..", "", "ZZZZZZZZZZZZZZZZ", "abc", strings.Repeat("a", 17), "../../etc/passwd"} {
		if _, err := s.Get(bad); err == nil {
			t.Errorf("Get(%q) succeeded", bad)
		}
	}
}

// Put must not report success until the record re-reads correctly (Ouroboros R30-3 applied to the
// proposal store, as it is to the echo chain).
func TestPutFailsWhenReadBackDisagrees(t *testing.T) {
	s, _ := NewStore(t.TempDir())
	p := sample()

	// The write itself succeeds; the read-back returns something else.
	s.readBack = func(_ *Store, _ string) (*Proposal, error) {
		other := sample()
		other.Title = "not what was written"
		other.Commit()
		return other, nil
	}
	if err := s.Put(p); err == nil {
		t.Error("Put reported success although the record read back different")
	}

	// And a read-back that fails outright is also fatal to the write.
	s.readBack = func(_ *Store, _ string) (*Proposal, error) { return nil, ErrNotFound }
	if err := s.Put(p); err == nil {
		t.Error("Put reported success although the record could not be read back")
	}
}

func TestListIsNewestFirst(t *testing.T) {
	s, _ := NewStore(t.TempDir())
	for i := 0; i < 3; i++ {
		p := sample()
		p.Title = strings.Repeat("t", i+1)
		p.CreatedAt = time.Unix(0, tsFixed+int64(i)*1e9).UTC()
		p.Commit()
		if err := s.Put(p); err != nil {
			t.Fatal(err)
		}
	}
	got, errs := s.List()
	if len(errs) != 0 {
		t.Fatalf("errors: %v", errs)
	}
	if len(got) != 3 {
		t.Fatalf("got %d proposals, want 3", len(got))
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].CreatedAt.Before(got[i].CreatedAt) {
			t.Error("List is not newest-first")
		}
	}
}
