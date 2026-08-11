package ladder

import (
	"context"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// The requested level and the delegation chain ride on the context rather than on a new
// parameter, for one reason: OMNISCIENT delegates through `models.AgentHandler.Handle`, an
// interface implemented in several places. Threading a level through it would change that
// signature everywhere and still miss any future implementation. A context value propagates
// through delegation automatically and cannot be forgotten at a new call site.
//
// This matches the codebase's existing idiom — internal/auth/middleware.go already carries
// claims and the GitHub token this way.

type ctxKey int

const (
	levelKey ctxKey = iota
	chainKey
	capabilityKey
)

// WithLevel attaches the authority level a request is asking for.
func WithLevel(ctx context.Context, l Level) context.Context {
	return context.WithValue(ctx, levelKey, l)
}

// LevelFrom returns the requested level, defaulting to L1 (Converse) when none was set.
//
// L1 is what every invocation does today — answer in plain language, no side effect — so an
// un-annotated request behaves exactly as it did before the ladder existed. The default is
// deliberately the LOWEST useful rung, not the highest: an unmarked request must not inherit
// authority it never asked for.
func LevelFrom(ctx context.Context) Level {
	if l, ok := ctx.Value(levelKey).(Level); ok && l.Valid() {
		return l
	}
	return DefaultRequested
}

// WithCapability attaches the capability a request wants to exercise. Required from L3 up.
func WithCapability(ctx context.Context, c Capability) context.Context {
	return context.WithValue(ctx, capabilityKey, c)
}

// CapabilityFrom returns the requested capability, or "" when none was named.
//
// The zero value is deliberately NOT a wildcard: an unnamed capability is refused at L3 and above,
// so forgetting to set one fails closed rather than granting everything.
func CapabilityFrom(ctx context.Context) Capability {
	if c, ok := ctx.Value(capabilityKey).(Capability); ok {
		return c
	}
	return ""
}

// WithCaller appends an agent to the delegation chain.
//
// The slice is COPIED rather than appended in place. Two sibling delegations from the same
// parent context share a backing array, and `append` would let the second overwrite the
// first's entry — a chain that silently reports the wrong caller is worse than no chain.
func WithCaller(ctx context.Context, a models.Agent) context.Context {
	prev := CallersFrom(ctx)
	next := make([]models.Agent, len(prev), len(prev)+1)
	copy(next, prev)
	next = append(next, a)
	return context.WithValue(ctx, chainKey, next)
}

// CallersFrom returns the delegation chain so far, outermost caller first. Never nil.
func CallersFrom(ctx context.Context) []models.Agent {
	if c, ok := ctx.Value(chainKey).([]models.Agent); ok {
		return c
	}
	return nil
}

// Authorize adjudicates dispatching to `callee` under the level and chain carried by ctx.
// It returns the Decision and a context extended with callee, ready to pass to Handle.
//
// The returned context is only meaningful when the Decision is allowed; callers must check
// Denied before dispatching. It is returned unconditionally so the happy path stays a single
// call at each of the six dispatch sites.
func Authorize(ctx context.Context, callee models.Agent) (context.Context, Decision) {
	next := WithCaller(ctx, callee) // copies; never aliases a sibling's backing array
	chain := CallersFrom(next)
	requested := LevelFrom(ctx)

	d := CheckChain(requested, chain...)
	if d.Denied {
		return next, d
	}

	// §2.2's enumerated-grant rule. Checked AFTER the ceiling so the reported reason is the
	// first thing that actually failed, and applied over the whole chain for the same reason
	// the ceiling is: authority must not widen by delegation.
	cap := CapabilityFrom(ctx)
	if ok, why := checkCapability(requested, cap, chain); !ok {
		d.Denied = true
		d.Granted = L0
		d.Capability = cap
		d.Reason = why
		return next, d
	}
	d.Capability = cap
	return next, d
}
