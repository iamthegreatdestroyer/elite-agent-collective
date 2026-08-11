package agents

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/echo"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/ladder"
)

// auditPolicy is a pure function, so the whole policy is testable without reaching L3 through the
// registry — which currently cannot happen at all, because the grant table is empty. Testing the
// policy directly is what stops it from being unverified code that only runs the day it matters.
func TestAuditPolicy(t *testing.T) {
	boom := errors.New("disk on fire")

	cases := []struct {
		name     string
		d        ladder.Decision
		err      error
		wantDeny bool
	}{
		{"L1 + audit ok -> allowed", ladder.Decision{Requested: ladder.L1}, nil, false},
		{"L1 + audit FAILED -> still allowed (fail-open)", ladder.Decision{Requested: ladder.L1}, boom, false},
		{"L2 + audit FAILED -> still allowed (fail-open)", ladder.Decision{Requested: ladder.L2}, boom, false},
		{"L3 + audit ok -> allowed", ladder.Decision{Requested: ladder.L3}, nil, false},
		{"L3 + audit FAILED -> REFUSED (fail-closed)", ladder.Decision{Requested: ladder.L3}, boom, true},
		{"L3 + NO CHAIN -> REFUSED", ladder.Decision{Requested: ladder.L3}, errNoChain, true},
		{"L4 + audit FAILED -> REFUSED", ladder.Decision{Requested: ladder.L4}, boom, true},
		// An already-denied request is not re-denied, and its original reason is preserved:
		// an operator must see WHY it was refused, not a second refusal layered on top.
		{"already denied + audit failed -> unchanged", ladder.Decision{Requested: ladder.L3, Denied: true, Reason: "no capability"}, boom, false},
	}

	for _, c := range cases {
		got, changed := auditPolicy(c.d, c.err)
		if changed != c.wantDeny {
			t.Errorf("%s: changed=%v, want %v", c.name, changed, c.wantDeny)
		}
		if c.wantDeny {
			if !got.Denied || got.Granted != ladder.L0 {
				t.Errorf("%s: denied=%v granted=%v", c.name, got.Denied, got.Granted)
			}
			if got.Reason == "" {
				t.Errorf("%s: refusal carried no reason", c.name)
			}
		}
	}
}

// The boundary is exactly L3 — the level at which an agent acts.
func TestFailClosedBoundaryIsL3(t *testing.T) {
	boom := errors.New("x")
	for _, lvl := range []ladder.Level{ladder.L0, ladder.L1, ladder.L2} {
		if _, changed := auditPolicy(ladder.Decision{Requested: lvl}, boom); changed {
			t.Errorf("%v was failed closed; below L3 nothing has an effect, so a lost record must not block", lvl)
		}
	}
	for _, lvl := range []ladder.Level{ladder.L3, ladder.L4} {
		if _, changed := auditPolicy(ladder.Decision{Requested: lvl}, boom); !changed {
			t.Errorf("%v was NOT failed closed — an action would proceed unrecorded", lvl)
		}
	}
	if ladder.CapabilityRequiredFrom != ladder.L3 {
		t.Errorf("the capability threshold moved to %v; the audit boundary must move with it",
			ladder.CapabilityRequiredFrom)
	}
}

// End-to-end at a level that IS reachable: a broken chain must not break ordinary traffic.
func TestBrokenChainDoesNotBlockL1(t *testing.T) {
	path := filepath.Join(t.TempDir(), "echo.log")
	c, err := echo.Open(path, []byte("v"), time.Unix(0, 1700000000000000000))
	if err != nil {
		t.Fatal(err)
	}
	SetEchoChain(c)
	t.Cleanup(func() { SetEchoChain(nil); c.Close() })

	// Make every append fail from here on.
	c.Close()
	os.Remove(path)

	_, r := setupTestHandler()
	if w := invoke(r, "SCRIBE", ""); w.Code == http.StatusForbidden {
		t.Errorf("an unlabelled (L1) request was refused because the audit chain broke: %s", w.Body.String())
	}
}

// With no chain configured, L1 still works — auditing is opt-in below L3.
func TestNoChainDoesNotBlockL1(t *testing.T) {
	SetEchoChain(nil)
	_, r := setupTestHandler()
	if w := invoke(r, "SCRIBE", ""); w.Code == http.StatusForbidden {
		t.Error("an L1 request was refused with no audit chain configured")
	}
}

// recordDecision's RETURN is what the policy consumes, so it needs its own assertions —
// otherwise the policy is tested and its input is not. (Mutation M4 turned errNoChain into nil
// and every test still passed.)
func TestRecordDecisionReportsFailures(t *testing.T) {
	d := ladder.Decision{Agent: "SCRIBE", Requested: ladder.L1, Granted: ladder.L1}

	// 1. No chain configured.
	SetEchoChain(nil)
	if err := recordDecision(d, nil); !errors.Is(err, errNoChain) {
		t.Errorf("no chain: got %v, want errNoChain — 'auditing off' must be distinguishable from 'audit ok'", err)
	}

	// 2. A working chain records and reports success.
	path := filepath.Join(t.TempDir(), "echo.log")
	c, err := echo.Open(path, []byte("v"), time.Unix(0, 1700000000000000000))
	if err != nil {
		t.Fatal(err)
	}
	SetEchoChain(c)
	t.Cleanup(func() { SetEchoChain(nil) })
	if err := recordDecision(d, nil); err != nil {
		t.Errorf("working chain: got %v, want nil", err)
	}

	// 3. A broken chain reports the failure rather than swallowing it.
	c.Close()
	os.Remove(path)
	if err := recordDecision(d, nil); err == nil {
		t.Error("broken chain: got nil — an append failure was swallowed")
	}
}

// THE FAIL-CLOSED PATH IS CURRENTLY UNREACHABLE END-TO-END, and this pins that.
//
// auditPolicy only fires on an ALLOWED request at L3+, and the grant table is empty, so no L3
// request is ever allowed — the join between recordDecision's error and the policy cannot be
// exercised through the HTTP path today. The policy itself is tested directly above; what cannot
// be tested is the wiring.
//
// This test FAILS when a capability grant appears, which is exactly when an end-to-end fail-closed
// test becomes both possible and necessary. Discovering that on the day L3 ships is the point: a
// policy whose wiring has never executed is not a policy anyone should rely on.
func TestFailClosedWiringIsNotYetReachable(t *testing.T) {
	_, r := setupTestHandler()
	registry := DefaultRegistry()

	for _, a := range registry.List() {
		if len(ladder.Grants(a)) != 0 {
			t.Fatalf("%s now holds a capability grant — L3 is reachable, so the fail-closed WIRING "+
				"must now be covered by an end-to-end test (allowed L3 request + broken chain => 403). "+
				"Write it, then delete this test.", a.Codename)
		}
	}

	// Corroborate through the real handler rather than trusting the table alone.
	if w := invoke(r, "FORGE", "L3"); w.Code != http.StatusForbidden {
		t.Errorf("FORGE at L3 returned %d — L3 is reachable and this test's premise is stale", w.Code)
	}
}
