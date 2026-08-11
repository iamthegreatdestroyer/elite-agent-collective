package retrieval

import (
	"context"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

// The window policy in this package was chosen from a measurement written in
// Python: it simulated header-stripping and windowing over live sigma-index
// results and reported 7 of 8 facts reached at 1,406 characters, against 3 of 8
// for the old k=5/head-280 policy.
//
// A simulation of the code is not the code. The Go implementation tokenises
// differently (longest-first, capped), steps the scan differently, and collapses
// whitespace in a different order. Shipping on the strength of the Python number
// would be trusting a model of the system instead of the system.
//
// This test runs the REAL retriever against the REAL index and re-derives the
// number. It is skipped unless EAC_LIVE_RETRIEVAL_CHECK=1, because it needs the
// gateway and sigma-index up and costs a few seconds of embedding:
//
//	EAC_LIVE_RETRIEVAL_CHECK=1 go test ./internal/retrieval/ -run Live -v
func TestLiveRetrievalReachesTheFacts(t *testing.T) {
	if os.Getenv("EAC_LIVE_RETRIEVAL_CHECK") == "" {
		t.Skip("set EAC_LIVE_RETRIEVAL_CHECK=1 to run against the live gateway and index")
	}

	cases := []struct {
		id, question, fact string
		wasWrong           bool // how the 2026-08-11 accuracy eval graded the agent
	}{
		{"dispatch_sites", "How many agent dispatch sites does the authority ladder gate? Reply with just the number.", `\bsix\b`, false},
		{"artifact_holders", "Which agents currently hold the artifact.write capability? If none hold it, say 'none'.", `granted to nobody|deliberately empty`, false},
		{"approval_executes", "Does approving an L2 proposal execute anything in this system? Answer yes or no, then one sentence.", `executes nothing|no tool surface`, false},
		{"chain_contents", "Does the Echo chain store the text of user prompts? Answer yes or no.", `hashes and metadata|canary`, false},
		{"tier4_ceiling", "What is the authority ceiling for a tier 4 agent? Answer with one of: L0, L1, L2, L3, L4.", `tier[- ]?4[^.]{0,40}L2|tier 4 .{0,20}L2|OMNISCIENT.{0,60}L2`, true},
		{"echo_hash", "What hash function does the Echo chain use? Name it.", `BLAKE2b`, true},
		{"failclosed_level", "Is the audit policy fail-open or fail-closed for L3 requests? Answer with one word.", `fail-CLOSED from L3|fail-closed from L3`, true},
		{"echo_env", "Which environment variable enables the Echo audit chain? Give just the variable name.", `EAC_ECHO_PATH`, true},
	}

	os.Setenv("EAC_RETRIEVER", "1")
	r := NewSigmaIndexRetriever()
	if r == nil || !r.Enabled() {
		t.Skip("retriever disabled or sigma-index unreachable")
	}

	var reached, chars int
	for _, c := range cases {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		block, err := r.Retrieve(ctx, "Elite Computer Science Engineering", c.question)
		cancel()
		if err != nil {
			t.Errorf("%-18s retrieve failed: %v", c.id, err)
			continue
		}
		if strings.Contains(block, "Mirrored from Claude Code memory") {
			t.Errorf("%-18s boilerplate reached the model", c.id)
		}
		ok := regexp.MustCompile("(?i)" + c.fact).MatchString(block)
		if ok {
			reached++
		}
		chars += len(block)
		status := "-"
		if ok {
			status = "FACT REACHED"
		}
		t.Logf("%-18s eval=%-9s %-13s block=%d chars",
			c.id, map[bool]string{true: "WRONG", false: "correct"}[c.wasWrong], status, len(block))
	}

	t.Logf("TOTAL %d of %d facts reached, %d chars/question average", reached, len(cases), chars/len(cases))

	// The Python measurement said 7 of 8. Allow one less for index drift and
	// embedding nondeterminism, but a large shortfall means the Go policy does
	// not do what was measured, and the default should not stand on that number.
	if reached < 6 {
		t.Errorf("only %d of %d facts reached; the policy was chosen on a simulated 7 of 8, "+
			"so either the simulation or this implementation is wrong", reached, len(cases))
	}
	// Cost is the other half of the decision: ~1400 chars was the old policy's
	// spend, and the new one was chosen for reaching more facts WITHOUT costing
	// more. At ~20 prefill tokens/sec, every 80 chars is another second of the
	// 50-second budget.
	if avg := chars / len(cases); avg > 2200 {
		t.Errorf("block averages %d chars — the policy was justified at ~1400; "+
			"the extra context is charged against the answer's time budget", avg)
	}
}
