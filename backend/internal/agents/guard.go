package agents

import (
	"context"
	"encoding/binary"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/echo"
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

// echoChain is the audit chain every ladder decision is recorded to. Optional by design:
// when nil, guard behaves exactly as before. Injection mirrors the codebase's existing
// optional-dependency pattern (SetOllama, SetMemory, WithCrews).
var (
	echoMu    sync.RWMutex
	echoChain *echo.Chain
)

// SetEchoChain installs (or clears, with nil) the audit chain.
func SetEchoChain(c *echo.Chain) {
	echoMu.Lock()
	defer echoMu.Unlock()
	echoChain = c
}

// ladderPayloadSchema names the payload shape hashed into each event. Changing the field
// list REQUIRES a new schema name: the schema is hashed first, so a silent field change
// would produce digests that cannot be told apart from the old shape's.
const ladderPayloadSchema = "ladder.decision.v1"

// recordDecision appends one Echo event for a ladder decision.
//
// WHAT IS HASHED: agent codename, the delegation chain, and the three levels — all of them
// enumerated, low-cardinality values. The request itself is NEVER touched: no prompt, no
// message text, no user content reaches this path, which is what keeps the append-only log
// safe to leak (§6.3). Note the digest is a COMMITMENT, not encryption — these fields have
// little entropy and are recoverable by brute force. That is acceptable precisely because
// they are already non-sensitive.
//
// FAILURE POLICY — fail-OPEN, deliberately: a logging failure does not block the request.
// Today every path is L0/L1, a text answer with no side effect, so refusing to answer
// because a disk write failed would convert a logging problem into an outage for no
// integrity gain. ⚠️ This must be revisited the moment L3 acquires a real tool surface:
// once an agent can ACT, "if it cannot be logged it must not happen" becomes the correct
// stance. Failures are counted in eac_echo_append_failures_total so the choice is visible
// rather than silent.
func recordDecision(d ladder.Decision, chain []models.Agent) {
	echoMu.RLock()
	c := echoChain
	echoMu.RUnlock()
	if c == nil {
		return
	}

	names := make([]string, 0, len(chain))
	for _, a := range chain {
		names = append(names, a.Codename)
	}

	payload := ladderPayload(d, names)

	trigger := "ladder.allow"
	if d.Denied {
		trigger = "ladder.deny"
	}
	if _, err := c.Append(trigger, payload); err != nil {
		metrics.IncEchoFailure()
		log.Printf("echo: failed to record %s for %s: %v", trigger, d.Agent, err)
	}
}

// ladderPayload computes the digest committed to the chain for one decision.
//
// Extracted so it can be pinned by a known-answer vector. The field list IS the schema: any
// addition, removal or reordering changes every digest, which is why ladderPayloadSchema must
// be bumped alongside it. TestLadderPayloadKnownAnswer fails on any silent change — including
// an innocent-looking extra field, which would otherwise alter the meaning of every historical
// record without renaming the schema.
func ladderPayload(d ladder.Decision, chain []string) []byte {
	return echo.AgentPayloadHash(ladderPayloadSchema,
		[]byte(d.Agent),
		[]byte(strings.Join(chain, ">")),
		lvlByte(d.Requested), lvlByte(d.Ceiling), lvlByte(d.Granted),
		boolByte(d.Denied),
	)
}

func lvlByte(l ladder.Level) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(l))
	return b
}

func boolByte(v bool) []byte {
	if v {
		return []byte{1}
	}
	return []byte{0}
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
	// Record BOTH outcomes. An audit chain holding only denials cannot answer "what did
	// this agent do?" — and a log that records only refusals makes a compromised allow
	// path invisible, which is the wrong half to keep.
	recordDecision(d, ladder.CallersFrom(next))
	return next, d
}
