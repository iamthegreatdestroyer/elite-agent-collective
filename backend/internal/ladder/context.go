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
	d := CheckChain(LevelFrom(ctx), CallersFrom(next)...)
	return next, d
}
