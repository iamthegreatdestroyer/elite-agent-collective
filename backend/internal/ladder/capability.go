package ladder

import (
	"fmt"
	"sort"
	"strings"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// Capability is a named, scoped ability an agent may exercise at L3.
//
// §2.2 requires "per-agent enumerated grants; nothing ambient." A LEVEL is not a grant: knowing an
// agent may reach L3 says nothing about what it may do there. Ceiling answers "how high", Grants
// answers "to do what", and both must be satisfied.
type Capability string

// grants is the complete, enumerated capability table.
//
// ⚠️ IT IS DELIBERATELY EMPTY, and that is the correct state today: no capability is implemented,
// there is no tool surface, and an entry here would grant an ability that does not exist. Because
// it is empty, L3 is unreachable BY CONSTRUCTION rather than by the accident of nobody asking —
// every L3 request fails the capability check regardless of tier.
//
// Adding an entry is the act of granting power. It should be a reviewed commit of its own, and
// TestGrantTableIsEmpty fails the moment one appears, so it cannot be added quietly alongside
// unrelated work.
var grants = map[string][]Capability{}

// Grants returns the capabilities enumerated for an agent, or nil.
//
// nil is the safe answer for an unknown agent: an agent nobody has granted anything has been
// granted nothing. There is no wildcard and no inheritance by tier — tier sets the ceiling, never
// the grant, because a tier is a class and a grant is a decision about one agent.
func Grants(a models.Agent) []Capability {
	out := grants[a.Codename]
	if out == nil {
		return nil
	}
	cp := make([]Capability, len(out))
	copy(cp, out)
	return cp
}

// HasGrant reports whether an agent may exercise a specific capability.
func HasGrant(a models.Agent, c Capability) bool {
	if c == "" {
		return false
	}
	for _, g := range grants[a.Codename] {
		if g == c {
			return true
		}
	}
	return false
}

// CapabilityRequiredFrom is the lowest rung that requires a named capability.
//
// L0–L2 are perception, conversation and proposal: they produce words, not effects, so "what may
// it do" has no content there. L3 is where an agent acts, and §2.1 defines it as a "narrow
// pre-authorized capability set" — a set, named in advance, not a rung reached.
const CapabilityRequiredFrom = L3

// checkCapability enforces the enumerated-grant rule over a delegation chain.
//
// Capability is checked with the SAME min-over-chain logic as the level: every agent in the chain
// must hold the capability. A caller cannot borrow a callee's grant, exactly as it cannot borrow a
// callee's ceiling — otherwise "nothing ambient" would hold per-agent and leak through delegation.
func checkCapability(requested Level, c Capability, chain []models.Agent) (bool, string) {
	if requested < CapabilityRequiredFrom {
		return true, "" // levels below L3 name no capability and need none
	}
	if c == "" {
		return false, fmt.Sprintf("%s requires a named capability (§2.2: per-agent enumerated grants, nothing ambient); none was given", requested)
	}
	for _, a := range chain {
		if !HasGrant(a, c) {
			return false, fmt.Sprintf("%s is not granted capability %q (granted: %s)",
				a.Codename, c, describeGrants(a))
		}
	}
	return true, ""
}

func describeGrants(a models.Agent) string {
	g := Grants(a)
	if len(g) == 0 {
		return "none"
	}
	s := make([]string, len(g))
	for i, c := range g {
		s[i] = string(c)
	}
	sort.Strings(s)
	return strings.Join(s, ",")
}
