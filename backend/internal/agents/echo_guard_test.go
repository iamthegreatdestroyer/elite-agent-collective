package agents

import (
	"strings"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/proposal"
)

// The guard is worthless if it names placeholders the prompt no longer uses, so pin them to the
// instruction itself. Duplication is deliberate; drift is what this catches.
func TestPlaceholdersMatchInstruction(t *testing.T) {
	for _, ph := range templatePlaceholders {
		if !strings.Contains(proposalInstruction, ph) {
			t.Errorf("placeholder %q is not in the instruction — the guard and the prompt have drifted", ph)
		}
	}
	// And every <...> in the instruction must be covered by the list.
	for _, seg := range strings.Split(proposalInstruction, "<")[1:] {
		end := strings.IndexByte(seg, '>')
		if end < 0 {
			continue
		}
		ph := "<" + seg[:end] + ">"
		found := false
		for _, known := range templatePlaceholders {
			if known == ph {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("instruction placeholder %q is not in templatePlaceholders — an echo of it would pass", ph)
		}
	}
}

// An agent that returns the template must be refused; a real proposal must not be.
func TestEchoesTemplate(t *testing.T) {
	echoed := &proposal.Proposal{
		Title:     "<one line, what you propose>",
		Rationale: "<why, grounded in what you were told>",
		Steps:     []proposal.Step{{Order: 1, Action: "<verb phrase>", Target: "<what it acts on>"}},
		Risks:     []string{"<what could go wrong>"},
	}
	if !echoesTemplate(echoed) {
		t.Error("a verbatim template echo was not detected")
	}

	// Only ONE echoed field is enough — a half-filled template is still not a proposal.
	partial := &proposal.Proposal{
		Title:     "Rebuild the index",
		Rationale: "It is stale and blocks retrieval.",
		Steps:     []proposal.Step{{Order: 1, Action: "rebuild", Target: "<what it acts on>"}},
	}
	if !echoesTemplate(partial) {
		t.Error("a partially echoed template was not detected")
	}

	// And a real proposal that merely CONTAINS angle brackets or similar words must pass —
	// the match is exact, not a heuristic.
	real := &proposal.Proposal{
		Title:     "Handle <script> tags in the sanitiser",
		Rationale: "The verb phrase parser mishandles what it acts on when nested.",
		Steps:     []proposal.Step{{Order: 1, Action: "escape <script>", Target: "sanitiser <input>"}},
		Risks:     []string{"what could go wrong: nothing, it is additive"},
	}
	if echoesTemplate(real) {
		t.Error("a real proposal containing angle brackets was rejected as a template echo")
	}
}
