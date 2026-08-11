package agents

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/ladder"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/metrics"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// readCounter scrapes one labeled counter out of the Prometheus text exposition.
// Reading the REAL metrics endpoint rather than a test double is the point: the acceptance
// criterion is about what an operator would actually see on /metrics.
func readCounter(t *testing.T, name, label string) int64 {
	t.Helper()
	w := httptest.NewRecorder()
	metrics.Handler(w, httptest.NewRequest("GET", "/metrics", nil))
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(name) + `\{[a-z]+="` + regexp.QuoteMeta(label) + `"\} (\d+)$`)
	m := re.FindStringSubmatch(w.Body.String())
	if m == nil {
		return 0 // counter not yet emitted for this label
	}
	n, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		t.Fatalf("unparseable counter %s{%s}: %v", name, label, err)
	}
	return n
}

func invoke(r http.Handler, codename, level string) *httptest.ResponseRecorder {
	body := bytes.NewBufferString(`{"messages":[{"role":"user","content":"ping"}],"stream":false}`)
	req := httptest.NewRequest("POST", "/agents/"+codename+"/invoke", body)
	req.Header.Set("Content-Type", "application/json")
	if level != "" {
		req.Header.Set(LevelHeader, level)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// THE ACCEPTANCE TEST for the wiring.
//
// It is a NEGATIVE assertion on purpose. Denying the request is easy; denying it ABOVE
// metrics.IncAgent is the part that can silently regress. Every agent implementation calls
// IncAgent as its first action, so a check placed even one line too low would still return
// 403 while counting the denial as a dispatch — and `eac_requests_by_agent_total` is exactly
// the signal used to tell "an agent ran" from "an agent did not". A test that only asserted
// the 403 would pass in that broken world.
func TestDeniedRequestDoesNotCountAsDispatch(t *testing.T) {
	_, r := setupTestHandler()

	const agent = "OMNISCIENT" // tier 4, ceiling L2
	before := readCounter(t, "eac_requests_by_agent_total", agent)
	denialsBefore := readCounter(t, "eac_ladder_denials_total", agent)

	w := invoke(r, agent, "L3")

	if w.Code != http.StatusForbidden {
		t.Fatalf("L3 to tier-4 %s: status %d, want 403. body=%s", agent, w.Code, w.Body.String())
	}
	if after := readCounter(t, "eac_requests_by_agent_total", agent); after != before {
		t.Errorf("DENIED request incremented eac_requests_by_agent_total{agent=%q}: %d -> %d.\n"+
			"The ladder check is running BELOW metrics.IncAgent, so a refusal is "+
			"indistinguishable from a real invocation.", agent, before, after)
	}
	if after := readCounter(t, "eac_ladder_denials_total", agent); after != denialsBefore+1 {
		t.Errorf("eac_ladder_denials_total{agent=%q} = %d, want %d — the refusal was not counted anywhere",
			agent, after, denialsBefore+1)
	}
}

// The default path must be unchanged: no header => L1 => everything dispatches as before.
// This is the regression guard for "wiring the ladder in broke normal traffic".
func TestUnlabelledRequestIsUnaffected(t *testing.T) {
	_, r := setupTestHandler()
	for _, agent := range []string{"OMNISCIENT", "APEX", "FORGE"} {
		before := readCounter(t, "eac_requests_by_agent_total", agent)
		w := invoke(r, agent, "") // no X-Agent-Level
		if w.Code == http.StatusForbidden {
			t.Errorf("%s: unlabelled request was refused: %s", agent, w.Body.String())
		}
		if after := readCounter(t, "eac_requests_by_agent_total", agent); after != before+1 {
			t.Errorf("%s: dispatch not counted (%d -> %d) — normal traffic regressed", agent, before, after)
		}
	}
}

func TestLevelsAtOrBelowCeilingAreAllowed(t *testing.T) {
	_, r := setupTestHandler()
	cases := []struct{ agent, level string }{
		{"OMNISCIENT", "L2"}, {"OMNISCIENT", "L0"}, // tier 4, ceiling L2
		{"FORGE", "L3"}, {"FORGE", "propose"}, // tier 5, ceiling L3
		{"SCRIBE", "L2"}, // tier 7, ceiling L2
	}
	for _, c := range cases {
		if w := invoke(r, c.agent, c.level); w.Code == http.StatusForbidden {
			t.Errorf("%s at %s was refused: %s", c.agent, c.level, w.Body.String())
		}
	}
}

func TestLevelsAboveCeilingAreRefused(t *testing.T) {
	_, r := setupTestHandler()
	cases := []struct{ agent, level string }{
		{"OMNISCIENT", "L3"}, {"OMNISCIENT", "L4"}, // tier 4 capped at L2
		{"FORGE", "L4"},                    // tier 5 capped at L3
		{"SCRIBE", "L3"}, {"SCRIBE", "L4"}, // tier 7 capped at L2
	}
	for _, c := range cases {
		if w := invoke(r, c.agent, c.level); w.Code != http.StatusForbidden {
			t.Errorf("%s at %s: status %d, want 403", c.agent, c.level, w.Code)
		}
	}
}

// An unparseable level must be rejected, never quietly treated as the default. A caller that
// asks for a rung it cannot name is the caller that should be stopped.
func TestUnparseableLevelIsRejected(t *testing.T) {
	_, r := setupTestHandler()
	for _, bad := range []string{"L9", "root", "admin", "sudo", "-1"} {
		w := invoke(r, "APEX", bad)
		if w.Code != http.StatusBadRequest {
			t.Errorf("level %q: status %d, want 400 (must not fall back to a default)", bad, w.Code)
		}
	}
}

// The delegation hole, end to end through the real handler: OMNISCIENT is capped at L2, and
// must not reach L3 by dispatching to a tier-5 agent that is individually allowed L3.
func TestDelegationCannotEscalate(t *testing.T) {
	forge := models.Agent{Codename: "FORGE", Tier: 5}
	omni := models.Agent{Codename: "OMNISCIENT", Tier: 4}

	// Directly, FORGE may act.
	if d := ladder.CheckChain(ladder.L3, forge); d.Denied {
		t.Fatal("precondition: FORGE alone should be allowed L3")
	}
	// Through OMNISCIENT, it may not — and Authorize (what guard uses) must agree.
	ctx := ladder.WithLevel(context.Background(), ladder.L3)
	ctx, d1 := ladder.Authorize(ctx, omni)
	if !d1.Denied {
		t.Error("OMNISCIENT itself was allowed L3")
	}
	if _, d2 := ladder.Authorize(ctx, forge); !d2.Denied {
		t.Errorf("OMNISCIENT -> FORGE reached L3 by proxy: %s", d2.Reason)
	}
}

// A sibling delegation must not corrupt another's chain via a shared backing array.
func TestSiblingDelegationsDoNotAliasChains(t *testing.T) {
	base := ladder.WithCaller(context.Background(), models.Agent{Codename: "OMNISCIENT", Tier: 4})
	a := ladder.WithCaller(base, models.Agent{Codename: "FORGE", Tier: 5})
	b := ladder.WithCaller(base, models.Agent{Codename: "SCRIBE", Tier: 7})

	got := func(ctx context.Context) string {
		var n []string
		for _, x := range ladder.CallersFrom(ctx) {
			n = append(n, x.Codename)
		}
		return strings.Join(n, "->")
	}
	if want := "OMNISCIENT->FORGE"; got(a) != want {
		t.Errorf("chain A = %s, want %s", got(a), want)
	}
	if want := "OMNISCIENT->SCRIBE"; got(b) != want {
		t.Errorf("chain B = %s, want %s (sibling overwrote it)", got(b), want)
	}
	if want := "OMNISCIENT"; got(base) != want {
		t.Errorf("parent chain mutated to %s, want %s", got(base), want)
	}
}

// The pipeline door and the /copilot door must be gated too — gating one entry point and
// missing another is the recurring failure this test exists to prevent.
func TestCopilotDoorIsGated(t *testing.T) {
	_, r := setupTestHandler()

	body := bytes.NewBufferString(`{"messages":[{"role":"user","content":"@OMNISCIENT ping"}],"stream":false}`)
	req := httptest.NewRequest("POST", "/copilot", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(LevelHeader, "L4")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("/copilot at L4 to OMNISCIENT: status %d, want 403 — this door is ungated", w.Code)
	}
	if w.Code == http.StatusForbidden && !strings.Contains(w.Body.String(), "ceiling") {
		t.Errorf("refusal did not explain itself: %s", w.Body.String())
	}
}
