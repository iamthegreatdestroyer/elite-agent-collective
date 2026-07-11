// Package agents provides the agent registry and HTTP handlers.
package agents

import (
	"strings"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/memory"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// routerStopwords are common function words that carry no routing signal.
// Domain terms (code, api, data, test, ...) are deliberately NOT included.
var routerStopwords = map[string]bool{
	"the": true, "and": true, "for": true, "with": true, "this": true,
	"that": true, "you": true, "your": true, "can": true, "help": true,
	"please": true, "how": true, "what": true, "why": true, "where": true,
	"when": true, "make": true, "need": true, "want": true, "have": true,
	"has": true, "get": true, "got": true, "use": true, "using": true,
	"from": true, "into": true, "about": true, "some": true, "any": true,
	"all": true, "out": true, "are": true, "was": true, "were": true,
	"will": true, "would": true, "should": true, "could": true, "here": true,
	"there": true, "then": true, "than": true, "our": true, "its": true,
	"them": true, "they": true, "but": true, "not": true, "now": true,
	"one": true, "let": true, "give": true,
}

// tokenizeText lowercases s, splits it on non-alphanumeric runes, drops
// stopwords and tokens shorter than 3 characters, and de-duplicates. Shared
// by the semantic router and the agent loader's keyword extraction.
func tokenizeText(s string) []string {
	fields := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})
	seen := make(map[string]bool, len(fields))
	out := make([]string, 0, len(fields))
	for _, w := range fields {
		if len(w) < 3 || routerStopwords[w] {
			continue
		}
		if seen[w] {
			continue
		}
		seen[w] = true
		out = append(out, w)
	}
	return out
}

// SemanticRouter picks a relevant agent for a bare (un-@mentioned) message
// using the dormant SkillBloomCascade — O(k) Bloom-filter skill matching,
// CPU-only, no network. It exists to replace the blind APEX default with a
// semantically chosen specialist.
type SemanticRouter struct {
	cascade *memory.SkillBloomCascade
}

// NewSemanticRouter builds a router seeded with the cascade's curated
// 40-agent skill map, then reconciles it with the live registry: any
// registered agent that is not already in the curated map is added with
// skills derived from its own metadata (so a roster change stays routable).
func NewSemanticRouter(registry *Registry) *SemanticRouter {
	cascade := memory.NewSkillBloomCascade()

	if registry != nil {
		known := make(map[string]bool)
		for _, id := range cascade.KnownAgents() {
			known[strings.ToUpper(id)] = true
		}
		for _, a := range registry.List() {
			if a.Codename == "" || known[strings.ToUpper(a.Codename)] {
				continue
			}
			if skills := deriveSkills(a); len(skills) > 0 {
				cascade.AddAgent(a.Codename, skills)
			}
		}
	}

	return &SemanticRouter{cascade: cascade}
}

// deriveSkills builds a skill token set for an agent from its own metadata,
// used only for agents absent from the cascade's curated map.
func deriveSkills(a models.Agent) []string {
	parts := []string{a.Specialty, a.Name, a.Category}
	parts = append(parts, a.Keywords...)
	parts = append(parts, a.Directives...)
	return tokenizeText(strings.Join(parts, " "))
}

// Route returns the single best-matching agent codename for a free-text
// message, or nil if nothing resonates. Callers use it only when the message
// carries no explicit @mention or @CREW.
func (sr *SemanticRouter) Route(message string) []string {
	if sr == nil || sr.cascade == nil {
		return nil
	}
	tokens := tokenizeText(message)
	if len(tokens) == 0 {
		return nil
	}
	matches := sr.cascade.FindAgentsWithSkills(tokens)
	if len(matches) == 0 {
		return nil
	}
	return []string{matches[0]}
}
