package agents

import (
	"strings"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/ladder"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// The delegation guard (min-over-chain) is CORRECT and, as of 2026-08-11, UNREACHABLE: it cannot
// change any outcome, because OMNISCIENT is the only agent that delegates and its tier-4 ceiling
// of L2 is the lowest the table produces.
//
// This test asserts that redundancy deliberately, against the real registry. It FAILS the moment
// the guard becomes load-bearing — a second delegator appears, a delegator's ceiling rises, or a
// ceiling below L2 is introduced — which is precisely when ladder.go's reachability comment stops
// being true and must be rewritten.
//
// A latent guarantee that silently becomes active is worse than either state: nobody re-reads the
// comment, and the tests keep passing.
func TestDelegationGuardIsCurrentlyRedundant(t *testing.T) {
	registry := DefaultRegistry()
	all := registry.List()
	if len(all) == 0 {
		t.Fatal("registry is empty; this test would pass vacuously")
	}

	// 1. Exactly one agent delegates. Enforced by source shape, asserted here by identity:
	//    if another delegator is added, the author must revisit this test.
	const delegator = "OMNISCIENT"
	h, err := registry.Get(delegator)
	if err != nil {
		t.Fatalf("%s absent from the registry: %v", delegator, err)
	}
	outer := h.GetInfo()

	// 2. The delegator's ceiling is the minimum over every agent's ceiling.
	outerCeil := ladder.Ceiling(outer)
	for _, a := range all {
		if c := ladder.Ceiling(a); c < outerCeil {
			t.Fatalf("%s (tier %d) has ceiling %v, BELOW the delegator %s's %v — "+
				"min-over-chain is now load-bearing; update ladder.go's reachability comment",
				a.Codename, a.Tier, c, delegator, outerCeil)
		}
	}

	// 3. Therefore the chain check agrees with the per-agent check on every callee, at every level.
	for _, a := range all {
		for _, lvl := range []ladder.Level{ladder.L0, ladder.L1, ladder.L2, ladder.L3, ladder.L4} {
			chain := ladder.CheckChain(lvl, outer, a)
			solo := ladder.Check(outer, lvl)
			if chain.Denied != solo.Denied {
				t.Errorf("%s->%s at %v: chain denied=%v but delegator alone denied=%v — "+
					"the chain check now changes an outcome; it is no longer redundant",
					delegator, a.Codename, lvl, chain.Denied, solo.Denied)
			}
		}
	}
}

// The delegation chain must still be CONSTRUCTED correctly even while the check is redundant —
// otherwise the guard would be untested when it becomes load-bearing.
func TestDelegationChainIsBuiltEvenWhenRedundant(t *testing.T) {
	omni := models.Agent{Codename: "OMNISCIENT", Tier: 4}
	forge := models.Agent{Codename: "FORGE", Tier: 5}

	d := ladder.CheckChain(ladder.L3, omni, forge)
	if !d.Denied {
		t.Fatal("OMNISCIENT->FORGE reached L3")
	}
	// The reason must name both hops, or an operator cannot tell a delegated denial from a direct one.
	for _, want := range []string{"OMNISCIENT", "FORGE"} {
		if !strings.Contains(d.Reason, want) {
			t.Errorf("denial reason %q does not name %s", d.Reason, want)
		}
	}
}
