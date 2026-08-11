package agents

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/ladder"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/metrics"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// LevelHeader lets a caller request an authority level (§2.1: L0-L4, or a rung name).
//
// A HEADER, not a body field, on purpose: `/agents` and the Copilot request/response shapes
// are a contract other things already depend on, and the persona change was verified by the
// payload staying byte-identical. A header adds the capability without touching that contract.
const LevelHeader = "X-Agent-Level"

// levelFromRequest reads the requested rung from the header.
//
// An UNPARSEABLE level is denied, not defaulted. Silently treating "L9" or "root" as L1 would
// hand a caller a quiet success for something it never got — and a caller that asks for a rung
// it cannot name is exactly the caller that should be stopped.
func levelFromRequest(r *http.Request) (ladder.Level, error) {
	return ladder.Parse(strings.TrimSpace(r.Header.Get(LevelHeader)))
}

// guard adjudicates dispatching to h under the level and delegation chain carried by ctx.
//
// It MUST be called BEFORE Handle, never inside it. Every agent implementation increments
// metrics.IncAgent as its first action (base.go:59, apex.go:70, omniscient.go:70), so a check
// placed below that point would count a denial as an invocation and make the two
// indistinguishable in `eac_requests_by_agent_total` — the very signal the acceptance test
// reads. Denials get their own counter instead.
//
// Returns the context to dispatch with (chain extended by this agent) and the Decision.
func guard(ctx context.Context, h models.AgentHandler) (context.Context, ladder.Decision) {
	info := h.GetInfo()
	next, d := ladder.Authorize(ctx, info)
	if d.Denied {
		metrics.IncLadderDenial(info.Codename)
		log.Printf("ladder: DENIED %s", d.Reason)
	}
	return next, d
}
