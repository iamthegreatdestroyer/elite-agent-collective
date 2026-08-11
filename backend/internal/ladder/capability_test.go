package ladder

import (
	"context"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

func agentT(name string, tier int) models.Agent { return models.Agent{Codename: name, Tier: tier} }

// THE POINT OF THE WHOLE MECHANISM: with no capability implemented, the grant table must be empty,
// which makes L3 unreachable by construction rather than by nobody happening to ask.
//
// This test fails the moment a grant appears — so granting power cannot ride along quietly with
// unrelated work. When the first capability is genuinely added, this test is updated in the same
// reviewed commit, deliberately.
func TestGrantTableIsEmpty(t *testing.T) {
	if len(grants) != 0 {
		t.Errorf("the grant table has %d entries: %v\n"+
			"If a capability was added deliberately, update this test in the same commit — "+
			"it exists so that granting power is never quiet.", len(grants), grants)
	}
}

// §2.2: "per-agent enumerated grants; nothing ambient." A tier-5 agent may REACH L3 by ceiling and
// must still be refused, because reaching a rung is not the same as holding a capability.
func TestTierFiveCannotActWithoutAGrant(t *testing.T) {
	forge := agentT("FORGE", 5) // ceiling L3

	// The ceiling alone allows it — this is the ambient authority the spec forbids.
	if d := CheckChain(L3, forge); d.Denied {
		t.Fatal("precondition: tier 5's CEILING should permit L3")
	}

	// Through Authorize, which applies the grant rule, it must be refused.
	ctx := WithLevel(context.Background(), L3)
	_, d := Authorize(ctx, forge)
	if !d.Denied {
		t.Error("tier-5 agent was granted L3 with no capability — authority is ambient")
	}
	if d.Granted != L0 {
		t.Errorf("denied request granted %v, want L0", d.Granted)
	}
}

// A capability must be NAMED at L3; an unnamed one is not a wildcard.
func TestUnnamedCapabilityIsRefusedAtL3(t *testing.T) {
	forge := agentT("FORGE", 5)
	ctx := WithLevel(context.Background(), L3) // no WithCapability
	if _, d := Authorize(ctx, forge); !d.Denied {
		t.Error("L3 with no named capability was allowed")
	}
	// And naming one that is not granted is also refused.
	ctx = WithCapability(WithLevel(context.Background(), L3), "artifact.write")
	if _, d := Authorize(ctx, forge); !d.Denied {
		t.Error("an ungranted capability was allowed")
	}
}

// Below L3 nothing names a capability and nothing needs to — the rule must not break L0–L2.
func TestLevelsBelowL3AreUnaffected(t *testing.T) {
	for _, a := range []models.Agent{agentT("OMNISCIENT", 4), agentT("FORGE", 5), agentT("SCRIBE", 7)} {
		for _, lvl := range []Level{L0, L1, L2} {
			ctx := WithLevel(context.Background(), lvl)
			if _, d := Authorize(ctx, a); d.Denied {
				t.Errorf("%s at %v was denied: %s", a.Codename, lvl, d.Reason)
			}
		}
	}
}

// Capability must narrow along a chain exactly as the ceiling does: a caller cannot borrow a
// callee's grant. Verified against a temporary grant so the empty-table default is not what passes.
func TestCapabilityDoesNotWidenByDelegation(t *testing.T) {
	const cap = Capability("artifact.write")
	orig := grants
	grants = map[string][]Capability{"FORGE": {cap}} // callee granted, caller NOT
	defer func() { grants = orig }()

	omni := agentT("OMNISCIENT", 4)
	forge := agentT("FORGE", 5)

	// Directly, FORGE holds it — but its own ceiling still governs the level.
	if !HasGrant(forge, cap) {
		t.Fatal("precondition: FORGE should hold the capability")
	}
	if HasGrant(omni, cap) {
		t.Fatal("precondition: OMNISCIENT should NOT hold it")
	}

	ok, why := checkCapability(L3, cap, []models.Agent{omni, forge})
	if ok {
		t.Error("OMNISCIENT->FORGE exercised a capability the caller does not hold")
	} else if why == "" {
		t.Error("refusal carried no reason")
	}

	// And the callee alone passes the capability check, proving the chain is what refused it.
	if ok, _ := checkCapability(L3, cap, []models.Agent{forge}); !ok {
		t.Error("FORGE alone was refused a capability it holds")
	}
}

func TestGrantsAreCopied(t *testing.T) {
	orig := grants
	grants = map[string][]Capability{"FORGE": {"artifact.write"}}
	defer func() { grants = orig }()

	g := Grants(agentT("FORGE", 5))
	if len(g) != 1 {
		t.Fatalf("got %d grants", len(g))
	}
	g[0] = "root.everything" // must not affect the table
	if !HasGrant(agentT("FORGE", 5), "artifact.write") {
		t.Error("mutating the returned slice corrupted the grant table")
	}
}
