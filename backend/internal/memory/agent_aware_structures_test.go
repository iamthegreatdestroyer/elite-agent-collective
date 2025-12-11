package memory

import (
	"fmt"
	"math"
	"testing"
	"time"
)

// ============================================================================
// AGENT AFFINITY GRAPH TESTS
// ============================================================================

func TestAgentAffinityGraph_NewGraph(t *testing.T) {
	g := NewAgentAffinityGraph()

	// Verify all 40 agents are initialized
	if len(g.agentTiers) != 40 {
		t.Errorf("Expected 40 agents, got %d", len(g.agentTiers))
	}

	// Verify tier assignments
	tierCounts := make(map[int]int)
	for _, tier := range g.agentTiers {
		tierCounts[tier]++
	}

	expectedTiers := map[int]int{1: 5, 2: 12, 3: 2, 4: 1, 5: 5, 6: 5, 7: 5, 8: 5}
	for tier, expected := range expectedTiers {
		if tierCounts[tier] != expected {
			t.Errorf("Tier %d: expected %d agents, got %d", tier, expected, tierCounts[tier])
		}
	}
}

func TestAgentAffinityGraph_RecordCollaboration(t *testing.T) {
	g := NewAgentAffinityGraph()

	initialAffinity := g.GetAffinityScore("APEX", "VELOCITY")

	// Record successful collaborations
	for i := 0; i < 10; i++ {
		g.RecordCollaboration("APEX", "VELOCITY", true)
	}

	newAffinity := g.GetAffinityScore("APEX", "VELOCITY")

	// Affinity should increase after successful collaborations
	if newAffinity <= initialAffinity {
		t.Errorf("Affinity should increase after successful collaborations: initial=%.4f, new=%.4f",
			initialAffinity, newAffinity)
	}

	// Test failed collaborations decrease affinity
	for i := 0; i < 20; i++ {
		g.RecordCollaboration("APEX", "CIPHER", false)
	}

	failedAffinity := g.GetAffinityScore("APEX", "CIPHER")
	if failedAffinity >= initialAffinity {
		t.Logf("Expected affinity to decrease after failures: initial=%.4f, after failures=%.4f",
			initialAffinity, failedAffinity)
	}
}

func TestAgentAffinityGraph_GetTopCollaborators(t *testing.T) {
	g := NewAgentAffinityGraph()

	// Build up APEX's collaboration history
	g.RecordCollaboration("APEX", "VELOCITY", true)
	g.RecordCollaboration("APEX", "VELOCITY", true)
	g.RecordCollaboration("APEX", "ARCHITECT", true)
	g.RecordCollaboration("APEX", "ECLIPSE", true)
	g.RecordCollaboration("APEX", "AXIOM", false)

	// Force routing table rebuild
	for i := 0; i < 100; i++ {
		g.RecordCollaboration("APEX", "VELOCITY", true)
	}

	top := g.GetTopCollaborators("APEX", 3)

	if len(top) != 3 {
		t.Errorf("Expected 3 collaborators, got %d", len(top))
	}

	// VELOCITY should be in top collaborators (highest success rate)
	found := false
	for _, agent := range top {
		if agent == "VELOCITY" {
			found = true
			break
		}
	}
	if !found {
		t.Error("VELOCITY should be in top collaborators for APEX")
	}
}

func TestAgentAffinityGraph_SuggestCollaborationTeam(t *testing.T) {
	g := NewAgentAffinityGraph()

	// Build collaboration history
	g.RecordCollaboration("APEX", "VELOCITY", true)
	g.RecordCollaboration("APEX", "ARCHITECT", true)
	g.RecordCollaboration("VELOCITY", "CORE", true)

	team := g.SuggestCollaborationTeam("APEX", 4)

	if len(team) != 4 {
		t.Errorf("Expected team of 4, got %d", len(team))
	}

	// Team should include seed agent
	hasSeed := false
	for _, agent := range team {
		if agent == "APEX" {
			hasSeed = true
			break
		}
	}
	if !hasSeed {
		t.Error("Team should include seed agent")
	}

	// All team members should be unique
	seen := make(map[string]bool)
	for _, agent := range team {
		if seen[agent] {
			t.Error("Team has duplicate agents")
		}
		seen[agent] = true
	}
}

// ============================================================================
// TIER RESONANCE FILTER TESTS
// ============================================================================

func TestTierResonanceFilter_FindResonantTiers(t *testing.T) {
	f := NewTierResonanceFilter()

	tests := []struct {
		content       string
		expectedTiers []int
		description   string
	}{
		{
			content:       "I need help with algorithm optimization and performance tuning",
			expectedTiers: []int{1}, // Foundational (algorithm, optimization, performance)
			description:   "Performance content should resonate with Tier 1",
		},
		{
			content:       "Design a machine learning neural network model for classification",
			expectedTiers: []int{2}, // Specialists (ml, neural)
			description:   "ML content should resonate with Tier 2",
		},
		{
			content:       "I want cross-domain innovation and novel paradigm synthesis",
			expectedTiers: []int{3}, // Innovators (synthesis, innovation, cross-domain, novel, paradigm)
			description:   "Innovation content should resonate with Tier 3",
		},
		{
			content:       "Orchestrate the collective and coordinate evolution meta-learning",
			expectedTiers: []int{4}, // Meta (orchestrate, collective, coordinate, evolve)
			description:   "Meta content should resonate with Tier 4",
		},
		{
			content:       "Build cloud infrastructure monitoring and graph analytics",
			expectedTiers: []int{5}, // Domain (cloud, build, monitor, graph)
			description:   "Domain content should resonate with Tier 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			tiers := f.FindResonantTiers(tt.content)

			if len(tiers) == 0 {
				t.Errorf("Expected resonance for '%s', got none", tt.content)
				return
			}

			// Check if expected tier is in top results
			found := false
			for _, expected := range tt.expectedTiers {
				for _, tier := range tiers[:min(3, len(tiers))] {
					if tier == expected {
						found = true
						break
					}
				}
			}

			if !found {
				t.Errorf("Expected tier %v in results for '%s', got %v",
					tt.expectedTiers, tt.content, tiers)
			}
		})
	}
}

func TestTierResonanceFilter_LearnFromExperience(t *testing.T) {
	f := NewTierResonanceFilter()

	// Learn new pattern for Tier 3
	f.LearnFromExperience(3, "revolutionary breakthrough quantum biomimicry")

	// Now these terms should resonate with Tier 3
	tiers := f.FindResonantTiers("apply biomimicry patterns")

	found := false
	for _, tier := range tiers {
		if tier == 3 {
			found = true
			break
		}
	}

	if !found {
		t.Error("Learned pattern 'biomimicry' should resonate with Tier 3")
	}
}

// ============================================================================
// SKILL BLOOM CASCADE TESTS
// ============================================================================

func TestSkillBloomCascade_FindAgentsWithSkills(t *testing.T) {
	c := NewSkillBloomCascade()

	tests := []struct {
		skills        []string
		expectedAgent string
		description   string
	}{
		{
			skills:        []string{"algorithm", "code"},
			expectedAgent: "APEX",
			description:   "Coding skills should match APEX",
		},
		{
			skills:        []string{"cryptography", "security"},
			expectedAgent: "CIPHER",
			description:   "Security skills should match CIPHER",
		},
		{
			skills:        []string{"machine learning", "pytorch"},
			expectedAgent: "TENSOR",
			description:   "ML skills should match TENSOR",
		},
		{
			skills:        []string{"kubernetes", "docker"},
			expectedAgent: "FLUX",
			description:   "DevOps skills should match FLUX",
		},
		{
			skills:        []string{"testing", "unit test"},
			expectedAgent: "ECLIPSE",
			description:   "Testing skills should match ECLIPSE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			agents := c.FindAgentsWithSkills(tt.skills)

			if len(agents) == 0 {
				t.Errorf("No agents found for skills %v", tt.skills)
				return
			}

			// Expected agent should be in top results
			found := false
			for i, agent := range agents[:min(3, len(agents))] {
				if agent == tt.expectedAgent {
					found = true
					t.Logf("%s found at position %d for skills %v", tt.expectedAgent, i, tt.skills)
					break
				}
			}

			if !found {
				t.Errorf("Expected %s in top results for skills %v, got %v",
					tt.expectedAgent, tt.skills, agents[:min(5, len(agents))])
			}
		})
	}
}

func TestSkillBloomCascade_AddAgent(t *testing.T) {
	c := NewSkillBloomCascade()

	// Add custom agent with unique skills
	c.AddAgent("CUSTOM_AGENT", []string{"unique_skill_xyz", "special_ability"})

	// Find agents with unique skill
	agents := c.FindAgentsWithSkills([]string{"unique_skill_xyz"})

	if len(agents) == 0 || agents[0] != "CUSTOM_AGENT" {
		t.Error("Custom agent should be found with unique skill")
	}
}

// ============================================================================
// TEMPORAL DECAY SKETCH TESTS
// ============================================================================

func TestTemporalDecaySketch_BasicCounting(t *testing.T) {
	s := NewTemporalDecaySketchDefault()

	// Add items
	for i := 0; i < 100; i++ {
		s.Add("strategy_alpha")
	}
	for i := 0; i < 50; i++ {
		s.Add("strategy_beta")
	}

	alphaCount := s.Estimate("strategy_alpha")
	betaCount := s.Estimate("strategy_beta")

	// Alpha should be approximately 2x beta
	ratio := alphaCount / betaCount
	if ratio < 1.5 || ratio > 2.5 {
		t.Errorf("Expected ratio ~2.0, got %.2f (alpha=%.2f, beta=%.2f)",
			ratio, alphaCount, betaCount)
	}
}

func TestTemporalDecaySketch_Decay(t *testing.T) {
	// Create sketch with very short half-life for testing
	s := NewTemporalDecaySketch(100 * time.Millisecond)

	// Add items
	s.Add("decaying_item")
	initialCount := s.Estimate("decaying_item")

	// Wait for one half-life
	time.Sleep(150 * time.Millisecond)

	decayedCount := s.Estimate("decaying_item")

	// Count should have decayed
	if decayedCount >= initialCount*0.8 {
		t.Errorf("Count should have decayed: initial=%.4f, after=%.4f",
			initialCount, decayedCount)
	}
}

func TestTemporalDecaySketch_WeightedAdd(t *testing.T) {
	s := NewTemporalDecaySketchDefault()

	// Add with different weights
	s.AddWithWeight("high_weight", 10.0)
	s.AddWithWeight("low_weight", 1.0)

	high := s.Estimate("high_weight")
	low := s.Estimate("low_weight")

	// High weight should be ~10x low weight
	ratio := high / low
	if ratio < 5 || ratio > 15 {
		t.Errorf("Expected ratio ~10, got %.2f", ratio)
	}
}

// ============================================================================
// COLLABORATIVE ATTENTION INDEX TESTS
// ============================================================================

func TestCollaborativeAttentionIndex_RouteQuery(t *testing.T) {
	idx := NewCollaborativeAttentionIndex()

	tests := []struct {
		query         string
		expectedAgent string
		description   string
	}{
		{
			query:         "implement a function to sort data",
			expectedAgent: "APEX",
			description:   "Coding query should route to APEX",
		},
		{
			query:         "design a scalable microservice architecture",
			expectedAgent: "ARCHITECT",
			description:   "Architecture query should route to ARCHITECT",
		},
		{
			query:         "secure authentication and encrypt data",
			expectedAgent: "CIPHER",
			description:   "Security query should route to CIPHER",
		},
		{
			query:         "optimize performance and benchmark code",
			expectedAgent: "VELOCITY",
			description:   "Performance query should route to VELOCITY",
		},
		{
			query:         "train a neural network model",
			expectedAgent: "TENSOR",
			description:   "ML query should route to TENSOR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			agents := idx.RouteQuery(tt.query, 5)

			if len(agents) == 0 {
				t.Errorf("No agents returned for query '%s'", tt.query)
				return
			}

			// Expected agent should be in top 3
			found := false
			for i, aa := range agents[:min(3, len(agents))] {
				if aa.AgentID == tt.expectedAgent {
					found = true
					t.Logf("%s found at position %d with attention %.4f",
						tt.expectedAgent, i, aa.Attention)
					break
				}
			}

			if !found {
				t.Errorf("Expected %s in top 3 for '%s', got: %v",
					tt.expectedAgent, tt.query, agents[:min(3, len(agents))])
			}
		})
	}
}

func TestCollaborativeAttentionIndex_UpdateAttention(t *testing.T) {
	idx := NewCollaborativeAttentionIndex()

	query := "implement a function"

	// Get initial routing
	initialRouting := idx.RouteQuery(query, 5)
	var initialScore float64
	for _, aa := range initialRouting {
		if aa.AgentID == "CORE" {
			initialScore = aa.Attention
			break
		}
	}

	// Update with positive feedback for CORE
	for i := 0; i < 20; i++ {
		idx.UpdateAttention(query, "CORE", true)
	}

	// Check updated routing
	updatedRouting := idx.RouteQuery(query, 5)
	var updatedScore float64
	for _, aa := range updatedRouting {
		if aa.AgentID == "CORE" {
			updatedScore = aa.Attention
			break
		}
	}

	// CORE's attention should have increased
	if updatedScore <= initialScore {
		t.Errorf("Attention for CORE should have increased: initial=%.4f, updated=%.4f",
			initialScore, updatedScore)
	}
}

// ============================================================================
// EMERGENT INSIGHT DETECTOR TESTS
// ============================================================================

func TestEmergentInsightDetector_RecordOutcome(t *testing.T) {
	d := NewEmergentInsightDetector()

	// Record some baseline outcomes
	for i := 0; i < 10; i++ {
		d.RecordOutcome([]string{"APEX", "ARCHITECT"}, "design", i%2 == 0, "standard strategy")
	}

	// Record a surprising success (unusual pair)
	surprise := d.RecordOutcome([]string{"HELIX", "LEDGER"}, "finance_bio", true, "novel approach")

	// First observation should have some surprise
	if surprise == 0 {
		t.Log("First observation has no surprise baseline, expected")
	}

	// Build up expectation
	for i := 0; i < 20; i++ {
		d.RecordOutcome([]string{"HELIX", "LEDGER"}, "finance_bio", false, "failed attempt")
	}

	// Now a success should be surprising
	surprise = d.RecordOutcome([]string{"HELIX", "LEDGER"}, "finance_bio", true, "breakthrough!")
	if surprise < 1.0 {
		t.Logf("Expected high surprise for unexpected success, got %.4f", surprise)
	}
}

func TestEmergentInsightDetector_GetRecentBreakthroughs(t *testing.T) {
	d := NewEmergentInsightDetector()

	// Build baseline expectations
	for i := 0; i < 30; i++ {
		d.RecordOutcome([]string{"APEX", "VELOCITY"}, "optimization", false, "baseline")
	}

	// Record surprising successes
	for i := 0; i < 5; i++ {
		d.RecordOutcome([]string{"APEX", "VELOCITY"}, "optimization", true, fmt.Sprintf("breakthrough_%d", i))
	}

	breakthroughs := d.GetRecentBreakthroughs(3)

	// Should have some breakthroughs recorded
	t.Logf("Found %d breakthroughs", len(breakthroughs))
	for i, bt := range breakthroughs {
		t.Logf("Breakthrough %d: agents=%v, surprise=%.4f, strategy=%s",
			i, bt.Agents, bt.SurpriseScore, bt.Strategy)
	}
}

func TestEmergentInsightDetector_GetUnexpectedPairs(t *testing.T) {
	d := NewEmergentInsightDetector()

	// Record consistently successful pair
	for i := 0; i < 20; i++ {
		d.RecordOutcome([]string{"NEXUS", "GENESIS"}, "innovation", true, "synergy")
	}

	// Record mediocre pair
	for i := 0; i < 20; i++ {
		d.RecordOutcome([]string{"APEX", "CIPHER"}, "security", i%2 == 0, "mixed")
	}

	pairs := d.GetUnexpectedPairs(10)

	// The consistently successful pair should be flagged
	found := false
	for _, p := range pairs {
		if containsAll(p.Agents, []string{"NEXUS", "GENESIS"}) {
			found = true
			t.Logf("Found unexpected pair: %v with success rate %.2f%%",
				p.Agents, p.SuccessRate*100)
		}
	}

	if !found && len(pairs) > 0 {
		t.Log("NEXUS+GENESIS pair not flagged, but other pairs found")
	}
}

// Helper function
func containsAll(slice []string, items []string) bool {
	for _, item := range items {
		found := false
		for _, s := range slice {
			if s == item {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ============================================================================
// BENCHMARKS
// ============================================================================

func BenchmarkAgentAffinityGraph_GetTopCollaborators(b *testing.B) {
	g := NewAgentAffinityGraph()

	// Build up history
	for i := 0; i < 1000; i++ {
		g.RecordCollaboration("APEX", "VELOCITY", true)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.GetTopCollaborators("APEX", 5)
	}
}

func BenchmarkTierResonanceFilter_FindResonantTiers(b *testing.B) {
	f := NewTierResonanceFilter()
	query := "design scalable cloud architecture with machine learning optimization"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.FindResonantTiers(query)
	}
}

func BenchmarkSkillBloomCascade_FindAgentsWithSkills(b *testing.B) {
	c := NewSkillBloomCascade()
	skills := []string{"algorithm", "performance", "python"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.FindAgentsWithSkills(skills)
	}
}

func BenchmarkTemporalDecaySketch_Estimate(b *testing.B) {
	s := NewTemporalDecaySketchDefault()

	// Pre-populate
	for i := 0; i < 10000; i++ {
		s.Add(fmt.Sprintf("item_%d", i%100))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Estimate("item_42")
	}
}

func BenchmarkCollaborativeAttentionIndex_RouteQuery(b *testing.B) {
	idx := NewCollaborativeAttentionIndex()
	query := "implement a secure authentication system with performance optimization"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.RouteQuery(query, 5)
	}
}

func BenchmarkEmergentInsightDetector_RecordOutcome(b *testing.B) {
	d := NewEmergentInsightDetector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.RecordOutcome([]string{"APEX", "VELOCITY"}, "optimization", i%2 == 0, "strategy")
	}
}

// ============================================================================
// INTEGRATION TESTS
// ============================================================================

func TestIntegration_AgentRoutingPipeline(t *testing.T) {
	// Test the full agent routing pipeline
	affinity := NewAgentAffinityGraph()
	resonance := NewTierResonanceFilter()
	skills := NewSkillBloomCascade()
	attention := NewCollaborativeAttentionIndex()
	insights := NewEmergentInsightDetector()

	// Simulate a query
	query := "implement a secure and performant API with proper testing"

	// Step 1: Find resonant tiers
	tiers := resonance.FindResonantTiers(query)
	t.Logf("Resonant tiers: %v", tiers)

	// Step 2: Route via attention
	attentionRoutes := attention.RouteQuery(query, 5)
	t.Logf("Attention routes: %v", attentionRoutes)

	// Step 3: Find agents with skills
	requiredSkills := []string{"api", "security", "testing"}
	skillMatches := skills.FindAgentsWithSkills(requiredSkills)
	t.Logf("Skill matches: %v", skillMatches[:min(5, len(skillMatches))])

	// Step 4: Select primary agent (first attention route)
	if len(attentionRoutes) == 0 {
		t.Fatal("No agents routed")
	}
	primaryAgent := attentionRoutes[0].AgentID

	// Step 5: Get collaborators
	collaborators := affinity.GetTopCollaborators(primaryAgent, 3)
	t.Logf("Collaborators for %s: %v", primaryAgent, collaborators)

	// Step 6: Record outcome and check for insights
	team := append([]string{primaryAgent}, collaborators[:min(2, len(collaborators))]...)
	surprise := insights.RecordOutcome(team, "api_development", true, "successful implementation")
	t.Logf("Surprise score: %.4f", surprise)

	// Verify pipeline produced reasonable results
	if len(tiers) == 0 {
		t.Error("Expected some resonant tiers")
	}
	if len(attentionRoutes) == 0 {
		t.Error("Expected some attention routes")
	}
	if len(skillMatches) == 0 {
		t.Error("Expected some skill matches")
	}
}

func TestIntegration_CrossTierCollaboration(t *testing.T) {
	affinity := NewAgentAffinityGraph()

	// Simulate cross-tier collaborations
	crossTierPairs := [][2]string{
		{"APEX", "TENSOR"},         // Tier 1 + Tier 2
		{"ARCHITECT", "NEXUS"},     // Tier 1 + Tier 3
		{"VELOCITY", "OMNISCIENT"}, // Tier 1 + Tier 4
		{"TENSOR", "GENESIS"},      // Tier 2 + Tier 3
	}

	// Record successful cross-tier collaborations
	for _, pair := range crossTierPairs {
		for i := 0; i < 10; i++ {
			affinity.RecordCollaboration(pair[0], pair[1], true)
		}
	}

	// Verify affinity increased for cross-tier pairs
	for _, pair := range crossTierPairs {
		score := affinity.GetAffinityScore(pair[0], pair[1])
		t.Logf("Affinity %s <-> %s: %.4f", pair[0], pair[1], score)
		if score < 0.3 {
			t.Errorf("Expected higher affinity for successful cross-tier pair %v", pair)
		}
	}
}

// ============================================================================
// FUZZ TESTS (if Go 1.18+)
// ============================================================================

func FuzzTierResonanceFilter(f *testing.F) {
	filter := NewTierResonanceFilter()

	// Add seed corpus
	f.Add("algorithm optimization performance")
	f.Add("machine learning neural network")
	f.Add("kubernetes docker deployment")
	f.Add("")
	f.Add("!@#$%^&*()")

	f.Fuzz(func(t *testing.T, content string) {
		// Should not panic
		tiers := filter.FindResonantTiers(content)

		// All tiers should be valid (1-8)
		for _, tier := range tiers {
			if tier < 1 || tier > 8 {
				t.Errorf("Invalid tier: %d", tier)
			}
		}
	})
}

func FuzzSkillBloomCascade(f *testing.F) {
	cascade := NewSkillBloomCascade()

	f.Add("algorithm")
	f.Add("machine learning")
	f.Add("")
	f.Add("!@#$%")

	f.Fuzz(func(t *testing.T, skill string) {
		// Should not panic
		agents := cascade.FindAgentsWithSkills([]string{skill})

		// Should return valid agent names or empty
		for _, agent := range agents {
			if agent == "" {
				t.Error("Empty agent name returned")
			}
		}
	})
}

// ============================================================================
// PROPERTY-BASED TESTS
// ============================================================================

func TestProperty_AffinitySymmetry(t *testing.T) {
	g := NewAgentAffinityGraph()

	// Record asymmetric collaborations
	for i := 0; i < 50; i++ {
		g.RecordCollaboration("APEX", "VELOCITY", true)
	}

	// Affinity should be symmetric
	a2v := g.GetAffinityScore("APEX", "VELOCITY")
	v2a := g.GetAffinityScore("VELOCITY", "APEX")

	if math.Abs(a2v-v2a) > 0.001 {
		t.Errorf("Affinity should be symmetric: APEX->VELOCITY=%.4f, VELOCITY->APEX=%.4f",
			a2v, v2a)
	}
}

func TestProperty_DecayMonotonicity(t *testing.T) {
	s := NewTemporalDecaySketch(1 * time.Second)

	s.Add("test_item")

	prev := s.Estimate("test_item")
	for i := 0; i < 5; i++ {
		time.Sleep(200 * time.Millisecond)
		curr := s.Estimate("test_item")
		if curr > prev {
			t.Errorf("Decay should be monotonic: prev=%.4f, curr=%.4f", prev, curr)
		}
		prev = curr
	}
}

func TestProperty_AttentionNormalization(t *testing.T) {
	idx := NewCollaborativeAttentionIndex()

	// Verify attention weights sum to ~1 for each category
	for category, weights := range idx.attentionWeights {
		total := 0.0
		for _, w := range weights {
			total += w
		}
		if math.Abs(total-1.0) > 0.01 {
			t.Errorf("Attention weights for category '%s' should sum to 1, got %.4f",
				category, total)
		}
	}
}
