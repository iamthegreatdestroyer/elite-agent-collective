// Package ladder implements the Authority Ladder from sigma-agent-os v1.3 §2.1/§2.2 —
// the constitution that caps what an agent is permitted to do, independent of what it
// is asked to do.
//
// §2.1, verbatim:
//
//	L0 Perceive      Echo chain, node state, logs, §5 matrices — metadata only
//	L1 Converse      Plain-language answers from actual state
//	L2 Propose       Signed, structured, Echo-logged proposals — human-approved
//	L3 Act (scoped)  Narrow pre-authorized capability set; tools, not root
//	L4 Human-gated   Irreversibles — agent-prepared, human-performed (hardware key)
//
// §2.2 ceilings, mapped from the tier the registry ALREADY carries:
//
//	Tier 4 (OMNISCIENT)   -> L2   "the cheapest enforcement in the codebase"
//	Tier 5 (Domain)       -> L3   domain-scoped, per-agent enumerated grants
//	Tiers 1-3, 6-8        -> L2   "L1-L2 default"; a ceiling is a maximum, so L2
//
// Verified against the live registry 2026-08-10: 40 agents, tier 4 = {OMNISCIENT} exactly,
// tier 5 = {SENTRY, VERTEX, STREAM, ATLAS, FORGE}, remainder spread over 1-3 and 6-8.
//
// WHY THIS COSTS NO NEW STRUCT FIELD: models.Agent.Tier (pkg/models/agent.go:5) is already
// populated from .agent.md frontmatter and already served on /agents. The doc calls the
// tier-4 cap "one struct field"; measured, it is a READ of a field that exists.
//
// §2.2 CAVEAT NOW CLOSED: the spec marks the tier-4 privilege budget FALSE — "while served
// by a User=stevo eac-backend, the serving process holds group 114(docker), an unenumerated
// root channel; the ceiling holds, the budget line does not, until the docker-group owner
// action lands". As of 2026-08-10 eac-backend runs as `sigma-agent` (uid 996) with NO
// supplementary groups and therefore no docker path, so the budget line now holds too.
//
// SCOPE, stated honestly: today every agent path is L0/L1 — a text answer with no side
// effect. `exec.Command` appears 0 times in the backend and the only write surface is two
// calls under MEMORY_DATA_DIR. So L3 cannot leak in by accident: there is no tool surface
// to grant. This package exists to make that guarantee explicit and checkable BEFORE a tool
// surface is ever added, not to restrain a capability that already exists.
package ladder

import (
	"fmt"
	"strings"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// Level is a rung of the Authority Ladder. Higher means more permitted.
type Level int

const (
	L0 Level = iota // Perceive
	L1              // Converse
	L2              // Propose
	L3              // Act (scoped)
	L4              // Human-gated
)

// DefaultRequested is the rung assumed when a caller does not name one. L1 (Converse)
// matches what every agent invocation does today: answer in plain language, no side effect.
const DefaultRequested = L1

var names = [...]string{"Perceive", "Converse", "Propose", "Act", "Human-gated"}

// String renders the canonical short form ("L2"), which is what Echo events and metrics use.
func (l Level) String() string {
	if l < L0 || l > L4 {
		return fmt.Sprintf("L?(%d)", int(l))
	}
	return fmt.Sprintf("L%d", int(l))
}

// Name renders the ladder's own word for the rung ("Propose"), for human-facing text.
func (l Level) Name() string {
	if l < L0 || l > L4 {
		return "unknown"
	}
	return names[l]
}

// Valid reports whether l is a defined rung.
func (l Level) Valid() bool { return l >= L0 && l <= L4 }

// Parse accepts "L2", "l2", "2", or the rung's name ("propose"). It is deliberately
// permissive on input format and strict on range: an unknown rung is an error, never a
// silent fallback to something permissive.
func Parse(s string) (Level, error) {
	t := strings.ToLower(strings.TrimSpace(s))
	if t == "" {
		return DefaultRequested, nil
	}
	t = strings.TrimPrefix(t, "l")
	switch t {
	case "0", "perceive":
		return L0, nil
	case "1", "converse":
		return L1, nil
	case "2", "propose":
		return L2, nil
	case "3", "act", "act (scoped)", "act-scoped":
		return L3, nil
	case "4", "human-gated", "human gated", "humangated":
		return L4, nil
	}
	return L0, fmt.Errorf("ladder: unknown level %q (want L0-L4 or perceive/converse/propose/act/human-gated)", s)
}

// Ceiling returns the highest rung an agent may occupy, from its tier (§2.2).
//
// The default is deliberately L2, NOT L4: an agent whose tier is missing, zero, or outside
// the known set must land on a restrictive rung, never a permissive one. A ladder that
// fails open is not a ladder.
func Ceiling(a models.Agent) Level {
	switch a.Tier {
	case 4:
		return L2 // OMNISCIENT — propose, never act
	case 5:
		return L3 // domain specialists, scoped grants
	default:
		return L2 // tiers 1-3, 6-8 — and anything unrecognised
	}
}

// Decision is the outcome of a ladder check. It is designed to be logged whole: an Echo
// event should be able to answer "what was asked, what was allowed, and why" without
// consulting anything else.
type Decision struct {
	Agent     string `json:"agent"`
	Requested Level  `json:"requested"`
	Ceiling   Level  `json:"ceiling"`
	Granted   Level  `json:"granted"`
	Denied    bool   `json:"denied"`
	Reason    string `json:"reason"`
	// Capability is the named ability the request wanted to exercise. Empty below L3, where
	// no capability is named or needed. Recorded so an audit entry can answer "to do what",
	// not merely "how high" — a level alone does not describe an action.
	Capability Capability `json:"capability,omitempty"`
}

// Check adjudicates a requested rung against an agent's ceiling.
//
// A request above the ceiling is DENIED outright rather than silently downgraded. Quietly
// granting L1 to something that asked for L3 would let a caller believe it had acted when
// it had only spoken — the exact "absence presented as a presence" shape this ecosystem
// keeps rediscovering. The caller must see the refusal.
func Check(a models.Agent, requested Level) Decision {
	c := Ceiling(a)
	d := Decision{Agent: a.Codename, Requested: requested, Ceiling: c}

	if !requested.Valid() {
		d.Denied = true
		d.Granted = L0
		d.Reason = fmt.Sprintf("invalid level %v; ladder fails closed", requested)
		return d
	}
	if requested > c {
		d.Denied = true
		d.Granted = L0
		d.Reason = fmt.Sprintf("%s requests %s (%s) but tier %d ceiling is %s (%s)",
			a.Codename, requested, requested.Name(), a.Tier, c, c.Name())
		return d
	}
	d.Granted = requested
	d.Reason = fmt.Sprintf("%s granted %s (%s), within tier %d ceiling %s",
		a.Codename, requested, requested.Name(), a.Tier, c)
	return d
}

// ChainCeiling returns the effective ceiling for a delegation chain, caller first.
//
// WHY THIS EXISTS — a per-agent ceiling alone does not hold. Measured 2026-08-10: there are
// five call sites that actually run an agent, and TWO of them (omniscient.go:145 and :166) are
// OMNISCIENT dispatching to a DIFFERENT agent. So "tier 4 is capped at L2" is not enforced by
// capping tier 4: OMNISCIENT (L2) delegating to a tier-5 agent (L3) would reach L3 by proxy,
// and the per-agent check at the inner site would see only the tier-5 agent and allow it.
//
// The ceiling must therefore be MONOTONIC along the chain — authority can only narrow as it is
// delegated, never widen. Effective ceiling = min over every agent in the chain.
//
// REACHABILITY, stated precisely (measured 2026-08-11): this guard currently cannot change any
// outcome, and the earlier note here understated why. It is not merely that nothing requests above
// L1 — it is that OMNISCIENT is the ONLY agent that delegates (omniscient.go dispatches to a
// specialist and to fallbackAPEX; no other handler calls another handler), and OMNISCIENT's tier-4
// ceiling of L2 is the LOWEST ceiling the table produces. So min(L2, anything) == L2 == the
// caller's own ceiling, and the chain check agrees with a per-agent check on every input.
//
// It becomes load-bearing the moment any of these change:
//   - a second agent gains the ability to delegate, or
//   - a delegator's ceiling rises above the lowest callee ceiling, or
//   - Ceiling gains a value below L2.
//
// TestDelegationGuardIsCurrentlyRedundant pins that, and FAILS when it stops being true — which is
// exactly when this comment must be rewritten. A guarantee that is documented as latent and
// silently becomes active is worse than either state.
func ChainCeiling(chain ...models.Agent) Level {
	if len(chain) == 0 {
		return L0 // no identity, no authority
	}
	eff := L4
	for _, a := range chain {
		if c := Ceiling(a); c < eff {
			eff = c
		}
	}
	return eff
}

// CheckChain adjudicates a request against a delegation chain (caller first, callee last).
// The Decision names the callee, since that is the agent about to run.
func CheckChain(requested Level, chain ...models.Agent) Decision {
	if len(chain) == 0 {
		return Decision{Requested: requested, Ceiling: L0, Granted: L0, Denied: true,
			Reason: "empty delegation chain; no agent identity to authorise"}
	}
	callee := chain[len(chain)-1]
	eff := ChainCeiling(chain...)
	d := Decision{Agent: callee.Codename, Requested: requested, Ceiling: eff}

	names := make([]string, len(chain))
	for i, a := range chain {
		names[i] = fmt.Sprintf("%s(t%d)", a.Codename, a.Tier)
	}
	path := strings.Join(names, " -> ")

	if !requested.Valid() || requested > eff {
		d.Denied = true
		d.Granted = L0
		d.Reason = fmt.Sprintf("chain %s requests %s but effective ceiling is %s (min over chain)",
			path, requested, eff)
		return d
	}
	d.Granted = requested
	d.Reason = fmt.Sprintf("chain %s granted %s, within effective ceiling %s", path, requested, eff)
	return d
}
