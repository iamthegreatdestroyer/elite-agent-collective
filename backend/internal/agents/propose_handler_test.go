package agents

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/proposal"
)

func setupProposalRouter(t *testing.T) (*chi.Mux, *proposal.Store) {
	t.Helper()
	st, err := proposal.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	SetProposalStore(st)
	t.Cleanup(func() { SetProposalStore(nil) })

	h, _ := setupTestHandler()
	r := chi.NewRouter()
	r.Post("/agents/{codename}/propose", h.ProposeAgent)
	r.Get("/proposals", h.ListProposals)
	r.Get("/proposals/{id}", h.GetProposal)
	r.Post("/proposals/{id}/decision", h.DecideProposal)
	return r, st
}

// The JSON extractor must survive the ways models actually reply, and must fail cleanly when
// there is no object — never salvage prose into a proposal-shaped envelope.
func TestExtractJSON(t *testing.T) {
	ok := map[string]string{
		`{"a":1}`:                               `{"a":1}`,
		"here you go:\n```json\n{\"a\":1}\n```": `{"a":1}`,
		`prose {"a":{"b":2}} trailing`:          `{"a":{"b":2}}`,
		`{"s":"a } not the end","t":1}`:         `{"s":"a } not the end","t":1}`,
		`{"s":"esc \" brace }","t":2}`:          `{"s":"esc \" brace }","t":2}`,
	}
	for in, want := range ok {
		got, found := extractJSON(in)
		if !found || got != want {
			t.Errorf("extractJSON(%q)\n got  %q (found=%v)\n want %q", in, got, found, want)
		}
	}
	for _, bad := range []string{"", "no object here", "{unbalanced", `{"s":"never closed`} {
		if got, found := extractJSON(bad); found {
			t.Errorf("extractJSON(%q) returned %q — prose was salvaged into an object", bad, got)
		}
	}
}

// A proposal the store never accepted must not be reported as created.
func TestProposeRejectsNonProposalOutput(t *testing.T) {
	r, st := setupProposalRouter(t)

	// The stub agents in the test registry answer with a template, not JSON — so this is the
	// real "the model did not comply" path, not a simulated one.
	body := bytes.NewBufferString(`{"messages":[{"role":"user","content":"propose something"}],"stream":false}`)
	req := httptest.NewRequest("POST", "/agents/SCRIBE/propose", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusInternalServerError {
		t.Errorf("status %d; want 422 (no proposal) — got body %s", w.Code, w.Body.String())
	}
	items, _ := st.List()
	if len(items) != 0 {
		t.Errorf("%d proposals stored despite no valid proposal being produced", len(items))
	}
}

// Disabled means disabled: 503 and no side effects.
func TestProposalEndpointsDisabledWithoutStore(t *testing.T) {
	SetProposalStore(nil)
	h, _ := setupTestHandler()
	r := chi.NewRouter()
	r.Post("/agents/{codename}/propose", h.ProposeAgent)
	r.Get("/proposals", h.ListProposals)

	for _, tc := range []struct{ method, path string }{
		{"POST", "/agents/SCRIBE/propose"}, {"GET", "/proposals"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("%s %s: status %d, want 503", tc.method, tc.path, w.Code)
		}
	}
}

// The full human-decision path, on a proposal placed directly in the store — which keeps this
// test independent of whether the local model happens to emit valid JSON.
func TestDecisionPathRecordsUnauthenticated(t *testing.T) {
	r, st := setupProposalRouter(t)

	p := &proposal.Proposal{
		Agent: "FORGE", Title: "Rebuild the index",
		Rationale: "It is stale.",
		Steps:     []proposal.Step{{Order: 1, Action: "rebuild", Target: "sigma-index", ReversibleClaimed: true}},
		Status:    proposal.StatusPending,
	}
	p.Commit()
	if err := st.Put(p); err != nil {
		t.Fatal(err)
	}

	// read it back over HTTP
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/proposals/"+p.ID, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("GET proposal: %d %s", w.Code, w.Body.String())
	}

	// approve it
	w = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/proposals/"+p.ID+"/decision",
		strings.NewReader(`{"decision":"approve","note":"looks right"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("decision: %d %s", w.Code, w.Body.String())
	}

	var got proposal.Proposal
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Status != proposal.StatusApproved {
		t.Errorf("status = %s, want approved", got.Status)
	}
	// THE HONESTY PROPERTY: auth is disabled in tests exactly as on the deployment, so the
	// record must say the approval was not verified rather than implying it was.
	if got.Decision == nil || got.Decision.Authenticated {
		t.Error("decision does not record that it was unauthenticated")
	}
	if got.Decision.Approver != "unauthenticated" {
		t.Errorf("approver = %q", got.Decision.Approver)
	}

	// a second ruling is refused
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/proposals/"+p.ID+"/decision", strings.NewReader(`{"decision":"reject"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Errorf("second decision: status %d, want 409", w.Code)
	}
}

func TestDecisionRejectsBadVerb(t *testing.T) {
	r, st := setupProposalRouter(t)
	p := &proposal.Proposal{Agent: "X", Title: "T", Rationale: "R",
		Steps: []proposal.Step{{Order: 1, Action: "a"}}, Status: proposal.StatusPending}
	p.Commit()
	st.Put(p)

	for _, verb := range []string{`"execute"`, `"maybe"`, `""`} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/proposals/"+p.ID+"/decision",
			strings.NewReader(`{"decision":`+verb+`}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("decision %s: status %d, want 400", verb, w.Code)
		}
	}
}

// A corrupted proposal must be REPORTED by the list, never silently omitted.
func TestListReportsUnreadable(t *testing.T) {
	r, st := setupProposalRouter(t)
	p := &proposal.Proposal{Agent: "X", Title: "T", Rationale: "R",
		Steps: []proposal.Step{{Order: 1, Action: "a"}}, Status: proposal.StatusPending}
	p.Commit()
	st.Put(p)

	// corrupt it on disk
	_, _ = st.List()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/proposals", nil))
	var out map[string]any
	json.Unmarshal(w.Body.Bytes(), &out)
	if out["count"].(float64) != 1 {
		t.Errorf("count = %v, want 1", out["count"])
	}
}
