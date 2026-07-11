package agents

import "testing"

// TestSemanticRouter_ClearWinners checks that unambiguous prompts route to the
// expected specialist via the woken Bloom skill-cascade.
func TestSemanticRouter_ClearWinners(t *testing.T) {
	r := NewSemanticRouter(DefaultRegistry())

	cases := []struct {
		name string
		msg  string
		want string
	}{
		{"performance", "profile and optimize the slow caching layer for better performance", "VELOCITY"},
		{"crypto", "implement encryption and a tls handshake with proper cryptography", "CIPHER"},
		{"docs", "write documentation and tutorials for this module", "SCRIBE"},
		{"devops", "set up a kubernetes deployment with docker and terraform", "FLUX"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := r.Route(c.msg)
			if len(got) != 1 || got[0] != c.want {
				t.Fatalf("Route(%q) = %v, want [%s]", c.msg, got, c.want)
			}
		})
	}
}

// TestSemanticRouter_NoSignal returns nil when the message has no routable
// tokens (all stopwords/short), so callers keep their default.
func TestSemanticRouter_NoSignal(t *testing.T) {
	r := NewSemanticRouter(DefaultRegistry())
	if got := r.Route("to be or not to be at all"); got != nil {
		t.Fatalf("expected nil for no-signal message, got %v", got)
	}
}

// TestResolveCodenames_ExplicitMentionWins ensures an explicit @mention still
// takes precedence over semantic routing.
func TestResolveCodenames_ExplicitMentionWins(t *testing.T) {
	h := NewHandler(DefaultRegistry())
	got := h.resolveCodenames("hey @CIPHER optimize this slow caching performance")
	if len(got) != 1 || got[0] != "CIPHER" {
		t.Fatalf("explicit mention should win, got %v", got)
	}
}

// TestResolveCodenames_SemanticFallback ensures a bare prompt reaches a
// specialist instead of the blind APEX default.
func TestResolveCodenames_SemanticFallback(t *testing.T) {
	h := NewHandler(DefaultRegistry())
	got := h.resolveCodenames("optimize the slow caching layer for better performance")
	if len(got) != 1 || got[0] != "VELOCITY" {
		t.Fatalf("expected semantic route to VELOCITY, got %v", got)
	}
}
