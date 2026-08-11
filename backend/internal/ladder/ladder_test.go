package ladder

import (
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

func agent(codename string, tier int) models.Agent {
	return models.Agent{Codename: codename, Tier: tier}
}

// The ceiling table from §2.2, checked against the tiers the live registry actually serves
// (verified 2026-08-10: 40 agents; tier 4 = {OMNISCIENT}; tier 5 = 5 agents).
func TestCeilingMatchesSpec(t *testing.T) {
	cases := []struct {
		tier int
		want Level
		why  string
	}{
		{4, L2, "OMNISCIENT — propose, never act"},
		{5, L3, "domain specialists, scoped"},
		{1, L2, "tiers 1-3 default"},
		{2, L2, "tiers 1-3 default"},
		{3, L2, "tiers 1-3 default"},
		{6, L2, "tiers 6-8 default"},
		{7, L2, "tiers 6-8 default"},
		{8, L2, "tiers 6-8 default"},
	}
	for _, c := range cases {
		if got := Ceiling(agent("X", c.tier)); got != c.want {
			t.Errorf("tier %d: ceiling = %v, want %v (%s)", c.tier, got, c.want, c.why)
		}
	}
}

// A ladder that fails OPEN is not a ladder. An unknown or absent tier must land on a
// restrictive rung, never a permissive one.
func TestUnknownTierFailsClosed(t *testing.T) {
	for _, tier := range []int{0, -1, 9, 99} {
		if got := Ceiling(agent("X", tier)); got != L2 {
			t.Errorf("tier %d: ceiling = %v, want L2 (must fail closed, never L3/L4)", tier, got)
		}
		if Ceiling(agent("X", tier)) >= L3 {
			t.Fatalf("tier %d granted >= L3 by default — the ladder fails OPEN", tier)
		}
	}
}

// The headline rule: OMNISCIENT may propose but must never act.
func TestOmniscientCannotAct(t *testing.T) {
	o := agent("OMNISCIENT", 4)

	if d := Check(o, L2); d.Denied {
		t.Errorf("OMNISCIENT should be granted L2 (Propose), got denied: %s", d.Reason)
	}
	for _, lvl := range []Level{L3, L4} {
		d := Check(o, lvl)
		if !d.Denied {
			t.Errorf("OMNISCIENT was granted %v — §2.2 caps tier 4 at L2", lvl)
		}
		if d.Granted != L0 {
			t.Errorf("denied request granted %v; a denial must not silently downgrade to a usable rung", d.Granted)
		}
		if d.Ceiling != L2 {
			t.Errorf("ceiling reported as %v, want L2", d.Ceiling)
		}
	}
}

// A tier-5 domain specialist may act; the same request from a tier-7 agent must not.
func TestTierFiveMayActOthersMayNot(t *testing.T) {
	if d := Check(agent("FORGE", 5), L3); d.Denied {
		t.Errorf("tier-5 FORGE denied L3: %s", d.Reason)
	}
	if d := Check(agent("SCRIBE", 7), L3); !d.Denied {
		t.Error("tier-7 SCRIBE was granted L3 — only tier 5 may act")
	}
}

// A denial must never quietly become a lesser grant: the caller has to be able to tell
// "I acted" from "I only spoke".
func TestDenialIsExplicitNotDowngraded(t *testing.T) {
	d := Check(agent("SCRIBE", 7), L4)
	if !d.Denied {
		t.Fatal("expected denial")
	}
	if d.Granted != L0 {
		t.Errorf("granted %v on a denial; must be L0", d.Granted)
	}
	if d.Reason == "" {
		t.Error("denial carried no reason — an Echo event could not explain it")
	}
	if d.Agent != "SCRIBE" {
		t.Errorf("decision lost the agent identity: %q", d.Agent)
	}
}

// Everything at or below the ceiling is granted unchanged.
func TestGrantsAtOrBelowCeiling(t *testing.T) {
	a := agent("APEX", 1) // ceiling L2
	for _, lvl := range []Level{L0, L1, L2} {
		d := Check(a, lvl)
		if d.Denied {
			t.Errorf("%v denied below ceiling: %s", lvl, d.Reason)
		}
		if d.Granted != lvl {
			t.Errorf("granted %v, want %v", d.Granted, lvl)
		}
	}
}

func TestParse(t *testing.T) {
	ok := map[string]Level{
		"L0": L0, "l1": L1, "2": L2, "L3": L3, "l4": L4,
		"perceive": L0, "Converse": L1, "PROPOSE": L2, "act": L3, "human-gated": L4,
		"": DefaultRequested, "  L2  ": L2,
	}
	for in, want := range ok {
		got, err := Parse(in)
		if err != nil {
			t.Errorf("Parse(%q) errored: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("Parse(%q) = %v, want %v", in, got, want)
		}
	}
	// Unknown input must ERROR, not fall back to something permissive.
	for _, bad := range []string{"L5", "9", "root", "admin", "sudo"} {
		if lvl, err := Parse(bad); err == nil {
			t.Errorf("Parse(%q) returned %v with no error — unknown levels must fail closed", bad, lvl)
		}
	}
}

// The delegation hole: OMNISCIENT is capped at L2, but omniscient.go:145/:166 dispatch to
// another agent. Without a chain-aware check, a tier-5 callee would grant L3 to a tier-4 caller.
func TestOmniscientCannotEscalateByDelegating(t *testing.T) {
	omni := agent("OMNISCIENT", 4) // ceiling L2
	forge := agent("FORGE", 5)     // ceiling L3

	// The bug this guards: checking the callee ALONE would allow it.
	if d := Check(forge, L3); d.Denied {
		t.Fatal("precondition: FORGE alone should be allowed L3")
	}
	// Through the chain it must not be.
	d := CheckChain(L3, omni, forge)
	if !d.Denied {
		t.Errorf("OMNISCIENT -> FORGE was granted L3; authority must narrow along a chain, not widen. %s", d.Reason)
	}
	if d.Ceiling != L2 {
		t.Errorf("effective ceiling = %v, want L2 (min of L2, L3)", d.Ceiling)
	}
	// L2 through the same chain is fine.
	if d := CheckChain(L2, omni, forge); d.Denied {
		t.Errorf("L2 denied through chain: %s", d.Reason)
	}
}

func TestChainCeilingIsMinimum(t *testing.T) {
	if got := ChainCeiling(agent("FORGE", 5), agent("SCRIBE", 7)); got != L2 {
		t.Errorf("min(L3,L2) = %v, want L2", got)
	}
	if got := ChainCeiling(agent("FORGE", 5)); got != L3 {
		t.Errorf("single tier-5 = %v, want L3", got)
	}
	// Order must not matter — min is symmetric.
	a := ChainCeiling(agent("A", 5), agent("B", 7))
	b := ChainCeiling(agent("B", 7), agent("A", 5))
	if a != b {
		t.Errorf("chain ceiling depends on order: %v vs %v", a, b)
	}
	// No identity => no authority.
	if got := ChainCeiling(); got != L0 {
		t.Errorf("empty chain = %v, want L0", got)
	}
	if d := CheckChain(L1); !d.Denied {
		t.Error("empty chain was granted authority")
	}
}

func TestStringAndName(t *testing.T) {
	if L2.String() != "L2" || L2.Name() != "Propose" {
		t.Errorf("L2 renders as %q/%q", L2.String(), L2.Name())
	}
	if Level(9).Valid() {
		t.Error("Level(9) reported valid")
	}
	if d := Check(agent("X", 1), Level(9)); !d.Denied {
		t.Error("an invalid level was not denied")
	}
}
