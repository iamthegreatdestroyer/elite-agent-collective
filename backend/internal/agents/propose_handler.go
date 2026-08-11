package agents

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/auth"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/copilot"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/echo"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/ladder"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/metrics"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/proposal"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// proposalStore is optional, like the Echo chain: unset means the L2 endpoints return 503 and
// nothing else changes. A feature that cannot be disabled cannot be rolled back.
var (
	propMu    sync.RWMutex
	propStore *proposal.Store
)

// SetProposalStore installs (or clears, with nil) the L2 proposal store.
func SetProposalStore(s *proposal.Store) {
	propMu.Lock()
	defer propMu.Unlock()
	propStore = s
}

func getProposalStore() *proposal.Store {
	propMu.RLock()
	defer propMu.RUnlock()
	return propStore
}

// proposalInstruction is appended to the agent's prompt for an L2 request.
//
// The model is asked for JSON and nothing else. If it does not comply, the request FAILS — see
// ProposeAgent. There is deliberately no salvage path that wraps prose in a proposal-shaped
// envelope, because that would make L2 a label on an L1 answer.
const proposalInstruction = `

You are being invoked at authority level L2 (PROPOSE). Do not answer in prose.
Reply with a single JSON object and nothing else — no commentary, no code fence:
{
  "title": "<one line, what you propose>",
  "rationale": "<why, grounded in what you were told>",
  "steps": [{"order": 1, "action": "<verb phrase>", "target": "<what it acts on>", "reversible": true}],
  "risks": ["<what could go wrong>"]
}
Replace every bracketed placeholder with real content. Do not return the template.
You are proposing, not acting. Nothing you write here is executed.`

// templatePlaceholders are the exact strings the instruction above uses as placeholders.
//
// WHY THIS EXISTS — a defect my own test caught: the instruction contains a *valid instance of the
// schema*, so an agent that simply echoes the prompt back produces a proposal that parses, passes
// every structural check (the placeholders are non-empty strings of reasonable length), and is
// semantically empty. "Structured" was satisfied while "proposal" was not.
//
// Matching is EXACT, never a substring or a heuristic on angle brackets: a real proposal may
// legitimately contain "<" or the word "verb phrase", and a guard that fires on those would reject
// good proposals. Only a verbatim echo of a placeholder is rejected.
//
// These strings are duplicated from the instruction on purpose and TestPlaceholdersMatchInstruction
// fails if they drift — the guard is worthless if it names placeholders the prompt no longer uses.
var templatePlaceholders = []string{
	"<one line, what you propose>",
	"<why, grounded in what you were told>",
	"<verb phrase>",
	"<what it acts on>",
	"<what could go wrong>",
}

// echoesTemplate reports whether any field is a verbatim placeholder — i.e. the agent returned the
// instruction rather than a proposal.
func echoesTemplate(p *proposal.Proposal) bool {
	isPlaceholder := func(v string) bool {
		v = strings.TrimSpace(v)
		for _, ph := range templatePlaceholders {
			if v == ph {
				return true
			}
		}
		return false
	}
	if isPlaceholder(p.Title) || isPlaceholder(p.Rationale) {
		return true
	}
	for _, st := range p.Steps {
		if isPlaceholder(st.Action) || isPlaceholder(st.Target) {
			return true
		}
	}
	for _, r := range p.Risks {
		if isPlaceholder(r) {
			return true
		}
	}
	return false
}

// extractJSON pulls the first balanced top-level JSON object out of a model response.
//
// Models wrap JSON in prose or fences despite instructions. Scanning for a balanced object is
// tolerant of that WITHOUT being tolerant of the failure that matters: if there is no parseable
// object, the caller gets an error and no proposal is created.
func extractJSON(s string) (string, bool) {
	start := strings.IndexByte(s, '{')
	if start < 0 {
		return "", false
	}
	depth, inStr, esc := 0, false, false
	for i := start; i < len(s); i++ {
		c := s[i]
		switch {
		case esc:
			esc = false
		case c == '\\' && inStr:
			esc = true
		case c == '"':
			inStr = !inStr
		case inStr:
			// nothing
		case c == '{':
			depth++
		case c == '}':
			depth--
			if depth == 0 {
				return s[start : i+1], true
			}
		}
	}
	return "", false
}

// ProposeAgent handles POST /agents/{codename}/propose — the L2 path.
func (h *Handler) ProposeAgent(w http.ResponseWriter, r *http.Request) {
	store := getProposalStore()
	if store == nil {
		copilot.WriteError(w, "proposals are not enabled (EAC_PROPOSAL_DIR unset)", http.StatusServiceUnavailable)
		return
	}

	codename := chi.URLParam(r, "codename")
	agent, err := h.registry.Get(codename)
	if err != nil {
		copilot.WriteError(w, err.Error(), http.StatusNotFound)
		return
	}

	req, err := copilot.ParseRequest(r)
	if err != nil {
		copilot.WriteError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// This endpoint IS an L2 request by definition. A header, if present, must not claim less
	// than it is doing — and must still pass the ladder, which is what caps tier-5 agents here
	// exactly as it does on /invoke.
	ctx, dec := guard(ladder.WithLevel(r.Context(), ladder.L2), agent)
	if dec.Denied {
		copilot.WriteError(w, dec.Reason, http.StatusForbidden)
		return
	}

	// Append the proposal instruction to the last user message.
	shaped := shapeForProposal(req)
	resp, err := agent.Handle(ctx, shaped)
	if err != nil || resp == nil || len(resp.Choices) == 0 {
		copilot.WriteError(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	raw, ok := extractJSON(resp.Choices[0].Message.Content)
	if !ok {
		// 422: the request was well-formed, the agent simply did not produce a proposal.
		copilot.WriteError(w, "agent did not return a proposal (no JSON object in the response)",
			http.StatusUnprocessableEntity)
		return
	}

	var p proposal.Proposal
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		copilot.WriteError(w, fmt.Sprintf("agent returned malformed proposal JSON: %v", err),
			http.StatusUnprocessableEntity)
		return
	}

	// Fields the AGENT does not get to choose. Taking the codename from the route rather than
	// the model's output means a proposal cannot claim to come from an agent that did not make it.
	p.Agent = agent.GetInfo().Codename
	p.CreatedAt = time.Now().UTC()
	p.Status = proposal.StatusPending
	p.Decision = nil

	if echoesTemplate(&p) {
		copilot.WriteError(w, "agent returned the proposal template rather than a proposal",
			http.StatusUnprocessableEntity)
		return
	}
	if err := p.Validate(); err != nil {
		copilot.WriteError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	p.Commit()

	if err := store.Put(&p); err != nil {
		log.Printf("proposal: store failed for %s: %v", p.Agent, err)
		copilot.WriteError(w, "could not store proposal", http.StatusInternalServerError)
		return
	}
	recordProposalEvent("proposal.created", &p)

	writeJSON(w, http.StatusCreated, &p)
}

// shapeForProposal returns a copy of req with the proposal instruction appended to the last user
// message. The original request is not mutated — the caller's object is theirs.
func shapeForProposal(req *models.CopilotRequest) *models.CopilotRequest {
	out := *req
	out.Messages = append([]models.Message(nil), req.Messages...)
	for i := len(out.Messages) - 1; i >= 0; i-- {
		if out.Messages[i].Role == "user" {
			out.Messages[i].Content += proposalInstruction
			return &out
		}
	}
	out.Messages = append(out.Messages, models.Message{Role: "user", Content: proposalInstruction})
	return &out
}

// ListProposals handles GET /proposals.
func (h *Handler) ListProposals(w http.ResponseWriter, r *http.Request) {
	store := getProposalStore()
	if store == nil {
		copilot.WriteError(w, "proposals are not enabled", http.StatusServiceUnavailable)
		return
	}
	items, errs := store.List()
	// Unreadable proposals are REPORTED, not omitted: a list that silently drops a corrupted
	// entry looks healthy while hiding the one thing worth knowing.
	out := map[string]any{"proposals": items, "count": len(items)}
	if len(errs) > 0 {
		msgs := make([]string, 0, len(errs))
		for _, e := range errs {
			msgs = append(msgs, e.Error())
		}
		out["unreadable"] = msgs
	}
	writeJSON(w, http.StatusOK, out)
}

// GetProposal handles GET /proposals/{id}.
func (h *Handler) GetProposal(w http.ResponseWriter, r *http.Request) {
	store := getProposalStore()
	if store == nil {
		copilot.WriteError(w, "proposals are not enabled", http.StatusServiceUnavailable)
		return
	}
	p, err := store.Get(chi.URLParam(r, "id"))
	if err != nil {
		copilot.WriteError(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

type decisionRequest struct {
	Decision string `json:"decision"` // "approve" | "reject"
	Note     string `json:"note"`
}

// DecideProposal handles POST /proposals/{id}/decision — the human ruling.
//
// ⚠️ It records a decision. It executes NOTHING: there is no tool surface in this service, and L3
// is not implemented. And it cannot currently verify the decider is human — auth is enabled only
// when OIDC_CLIENT_ID is set. The stored Decision carries `authenticated: false` in that case, so
// the record states its own weakness rather than relying on a reader knowing the configuration.
func (h *Handler) DecideProposal(w http.ResponseWriter, r *http.Request) {
	store := getProposalStore()
	if store == nil {
		copilot.WriteError(w, "proposals are not enabled", http.StatusServiceUnavailable)
		return
	}

	var body decisionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		copilot.WriteError(w, "invalid decision body", http.StatusBadRequest)
		return
	}

	var status proposal.Status
	switch strings.ToLower(strings.TrimSpace(body.Decision)) {
	case "approve", "approved":
		status = proposal.StatusApproved
	case "reject", "rejected":
		status = proposal.StatusRejected
	default:
		copilot.WriteError(w, `decision must be "approve" or "reject"`, http.StatusBadRequest)
		return
	}

	p, err := store.Get(chi.URLParam(r, "id"))
	if err != nil {
		copilot.WriteError(w, err.Error(), http.StatusNotFound)
		return
	}

	approver, authenticated := approverFrom(r)
	if err := p.Decide(proposal.Decision{
		Status: status, At: time.Now().UTC(),
		Approver: approver, Authenticated: authenticated, Note: body.Note,
	}); err != nil {
		copilot.WriteError(w, err.Error(), http.StatusConflict)
		return
	}
	if err := store.Put(p); err != nil {
		copilot.WriteError(w, "could not store decision", http.StatusInternalServerError)
		return
	}
	recordProposalEvent("proposal."+string(status), p)

	writeJSON(w, http.StatusOK, p)
}

// approverFrom reports who decided, and whether that identity was actually verified.
//
// The second return value is the honest part: with auth disabled the middleware passes every
// request through, so any identity here is a claim, not a fact.
func approverFrom(r *http.Request) (string, bool) {
	if c, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims); ok && c != nil && c.Subject != "" {
		return c.Subject, true
	}
	return "unauthenticated", false
}

// recordProposalEvent appends the proposal's COMMITMENT to the audit chain — never its content.
// The chain holds hashes and metadata only; the proposal itself lives in the store.
func recordProposalEvent(trigger string, p *proposal.Proposal) {
	echoMu.RLock()
	c := echoChain
	echoMu.RUnlock()
	if c == nil {
		return
	}
	payload := echo.AgentPayloadHash("proposal.event.v1",
		[]byte(trigger), []byte(p.Agent), []byte(p.ID), []byte(p.Commitment), []byte(p.Status))
	if _, err := c.Append(trigger, payload); err != nil {
		metrics.IncEchoFailure()
		log.Printf("echo: failed to record %s for %s: %v", trigger, p.ID, err)
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("proposal: encode failed: %v", err)
	}
}
