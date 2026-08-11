package agents

import (
	"bytes"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/echo"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/ladder"
)

// withEchoChain installs a fresh chain for one test and removes it afterwards, so the
// package-level chain never leaks between tests.
func withEchoChain(t *testing.T) (string, *echo.Chain) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "echo.log")
	c, err := echo.Open(path, []byte("test-vault"), time.Unix(0, 1700000000000000000))
	if err != nil {
		t.Fatal(err)
	}
	SetEchoChain(c)
	t.Cleanup(func() { SetEchoChain(nil); c.Close() })
	return path, c
}

func countTrigger(t *testing.T, path, trigger string) int {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, line := range bytes.Split(bytes.TrimSpace(raw), []byte("\n")) {
		if bytes.Contains(line, []byte(`"trigger":"`+trigger+`"`)) {
			n++
		}
	}
	return n
}

// A denial must produce a durable, verifiable audit record — that is the whole reason the
// chain exists. Enforcement without evidence is not reviewable.
func TestDeniedDecisionIsRecorded(t *testing.T) {
	path, c := withEchoChain(t)
	_, r := setupTestHandler()

	if got := countTrigger(t, path, "ladder.deny"); got != 0 {
		t.Fatalf("precondition: chain already holds %d denials", got)
	}

	w := invoke(r, "OMNISCIENT", "L3")
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}

	if got := countTrigger(t, path, "ladder.deny"); got != 1 {
		t.Errorf("ladder.deny events = %d, want 1 — the denial was enforced but not recorded", got)
	}
	if err := c.Verify(); err != nil {
		t.Errorf("chain does not verify after recording a denial: %v", err)
	}
}

// Allows are recorded too. A chain holding only refusals cannot answer "what did this agent
// actually do?", and it makes a compromised allow path invisible.
func TestAllowedDecisionIsRecorded(t *testing.T) {
	path, c := withEchoChain(t)
	_, r := setupTestHandler()

	if w := invoke(r, "SCRIBE", ""); w.Code == http.StatusForbidden {
		t.Fatalf("precondition: unlabelled request was refused")
	}
	if got := countTrigger(t, path, "ladder.allow"); got != 1 {
		t.Errorf("ladder.allow events = %d, want 1", got)
	}
	if err := c.Verify(); err != nil {
		t.Errorf("chain does not verify: %v", err)
	}
}

// THE ONE THAT MATTERS MOST. The chain is append-only: prompt text written by mistake can
// never be removed. This asserts that a distinctive string present in the request body
// reaches no byte of the audit log.
func TestRequestTextNeverReachesTheChain(t *testing.T) {
	path, _ := withEchoChain(t)
	_, r := setupTestHandler()

	const secret = "ZZQX-CANARY-PROMPT-do-not-log-me-9137"
	body := bytes.NewBufferString(`{"messages":[{"role":"user","content":"` + secret + `"}],"stream":false}`)
	req := httptest.NewRequest("POST", "/agents/OMNISCIENT/invoke", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(LevelHeader, "L3") // denied, so it takes the recording path
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{secret, "ZZQX", "CANARY", "do-not-log-me"} {
		if bytes.Contains(raw, []byte(forbidden)) {
			t.Errorf("request text %q reached the append-only audit chain — it cannot be un-logged", forbidden)
		}
	}
}

// Distinct decisions must not collide: the payload digest has to distinguish who was denied
// and at what level, or the chain records that "something was denied" and nothing more.
func TestDecisionPayloadsAreDistinct(t *testing.T) {
	path, _ := withEchoChain(t)
	_, r := setupTestHandler()

	invoke(r, "OMNISCIENT", "L3")
	invoke(r, "SCRIBE", "L3")
	invoke(r, "FORGE", "L4")

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, line := range bytes.Split(bytes.TrimSpace(raw), []byte("\n")) {
		if !bytes.Contains(line, []byte(`"trigger":"ladder.deny"`)) {
			continue
		}
		i := bytes.Index(line, []byte(`"payload_hash":"`))
		if i < 0 {
			t.Fatalf("record has no payload_hash: %s", line)
		}
		rest := line[i+len(`"payload_hash":"`):]
		hash := string(rest[:bytes.IndexByte(rest, '"')])
		if seen[hash] {
			t.Errorf("two different denials share payload hash %s… — the digest does not distinguish them", hash[:12])
		}
		seen[hash] = true
	}
	if len(seen) != 3 {
		t.Errorf("recorded %d distinct denial payloads, want 3", len(seen))
	}
}

// With no chain installed, guard must behave exactly as before — auditing is opt-in and
// must never be able to break a running service by existing.
func TestNoChainInstalledIsHarmless(t *testing.T) {
	SetEchoChain(nil)
	_, r := setupTestHandler()

	if w := invoke(r, "SCRIBE", ""); w.Code == http.StatusForbidden {
		t.Error("unlabelled request refused with no chain installed")
	}
	if w := invoke(r, "OMNISCIENT", "L3"); w.Code != http.StatusForbidden {
		t.Error("ladder stopped enforcing when no chain was installed")
	}
}

// KNOWN-ANSWER vectors for the audit payload, computed by a separate implementation from the
// documented encoding (BLAKE2b-256 over "sigma-agentos-payload-v1" ‖ LP(schema) ‖ LP(fields…)).
//
// This is what makes the schema a promise rather than a comment. Adding, removing or
// reordering a field changes every digest and breaks this test, forcing ladderPayloadSchema
// to be bumped with it. Mutation testing showed why: dropping the three level fields, and
// silently folding an extra field in, both passed the entire behavioural suite.
func TestLadderPayloadKnownAnswer(t *testing.T) {
	cases := []struct {
		name  string
		d     ladder.Decision
		chain []string
		want  string
	}{
		{
			"denied", ladder.Decision{Agent: "OMNISCIENT", Requested: ladder.L3,
				Ceiling: ladder.L2, Granted: ladder.L0, Denied: true},
			[]string{"OMNISCIENT"},
			"84f1a176dde3927d04f44e8d243f610a42b9ff4a6720e53986346f1c6f346ca8",
		},
		{
			"allowed", ladder.Decision{Agent: "SCRIBE", Requested: ladder.L1,
				Ceiling: ladder.L2, Granted: ladder.L1},
			[]string{"SCRIBE"},
			"e21c65c945cbcd90c82e9e0e13da80e77fde89f2f96d51da4f88aa2141b14df7",
		},
		{
			"delegated", ladder.Decision{Agent: "FORGE", Requested: ladder.L3,
				Ceiling: ladder.L2, Granted: ladder.L0, Denied: true},
			[]string{"OMNISCIENT", "FORGE"},
			"fce7bc34eaecd01130d5a5d00f22b68e2b3af61d9be8ee6ac540f18d0b788f7f",
		},
	}
	for _, c := range cases {
		if got := hex.EncodeToString(ladderPayload(c.d, c.chain)); got != c.want {
			t.Errorf("%s payload\n got  %s\n want %s\n(if this change was intentional, bump ladderPayloadSchema)",
				c.name, got, c.want)
		}
	}
}

// The same agent denied at different levels must produce different digests. Otherwise the
// chain records only "this agent was denied" and loses what was actually attempted — the
// behavioural half of what the known-answer vectors pin structurally.
func TestSameAgentDifferentLevelsAreDistinct(t *testing.T) {
	base := ladder.Decision{Agent: "OMNISCIENT", Ceiling: ladder.L2, Granted: ladder.L0, Denied: true}
	l3, l4 := base, base
	l3.Requested = ladder.L3
	l4.Requested = ladder.L4

	if bytes.Equal(ladderPayload(l3, []string{"OMNISCIENT"}), ladderPayload(l4, []string{"OMNISCIENT"})) {
		t.Error("L3 and L4 denials of the same agent share a digest — the attempted level is not recorded")
	}

	// Allowed vs denied at the same level must also differ.
	allowed := ladder.Decision{Agent: "FORGE", Requested: ladder.L3, Ceiling: ladder.L3, Granted: ladder.L3}
	denied := ladder.Decision{Agent: "FORGE", Requested: ladder.L3, Ceiling: ladder.L3, Granted: ladder.L3, Denied: true}
	if bytes.Equal(ladderPayload(allowed, []string{"FORGE"}), ladderPayload(denied, []string{"FORGE"})) {
		t.Error("an allow and a deny share a digest — the outcome is not recorded")
	}
}
