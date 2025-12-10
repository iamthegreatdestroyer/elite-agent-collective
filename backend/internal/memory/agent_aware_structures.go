// Package memory provides advanced sub-linear data structures for the Elite Agent Collective.
// This file implements AGENT-AWARE sub-linear structures that understand the collective's
// 40 agents, 8 tiers, and collaboration patterns.
//
// 🧠 Innovation Philosophy:
// @TENSOR: Apply ML-inspired gradient and attention mechanisms to agent interactions
// @NEXUS: Cross-pollinate ideas from neuroscience, ecology, and social networks
// @GENESIS: Invent entirely new paradigms for multi-agent knowledge systems
//
// ============================================================================
// INNOVATIONS IMPLEMENTED:
// ============================================================================
// 1. AgentAffinityGraph: O(1) amortized agent collaboration recommendation
//    - Inspired by: Graph neural networks + collaborative filtering
//    - Purpose: Instantly suggest which agents should collaborate on a task
//
// 2. TierResonanceFilter: O(1) tier-aware experience propagation
//    - Inspired by: Resonance theory in neuroscience + signal processing
//    - Purpose: Experiences "resonate" to appropriate tiers without linear scan
//
// 3. SkillBloomCascade: O(k) multi-skill matching with cascading filters
//    - Inspired by: Cascade classifiers in computer vision
//    - Purpose: Match tasks to agents by skill intersection
//
// 4. TemporalDecaySketch: O(1) time-aware frequency with automatic forgetting
//    - Inspired by: Memory consolidation in human brain
//    - Purpose: Recent strategies weighted higher, old ones fade naturally
//
// 5. CollaborativeAttentionIndex: O(1) attention-based agent routing
//    - Inspired by: Transformer attention mechanisms
//    - Purpose: Route queries to agents using learned attention weights
//
// 6. EmergentInsightDetector: O(1) detection of cross-domain breakthroughs
//    - Inspired by: Anomaly detection + serendipity in scientific discovery
//    - Purpose: Identify when agent combinations produce unexpected insights
// ============================================================================

package memory

import (
	"encoding/binary"
	"hash/fnv"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// 1. AGENT AFFINITY GRAPH
// ============================================================================
// O(1) amortized agent collaboration recommendation using a probabilistic
// graph structure. Inspired by collaborative filtering + random walks.
//
// Innovation: Instead of computing shortest paths (O(V+E)), we maintain
// "affinity scores" that update incrementally on each successful collaboration.

// AgentAffinityGraph tracks collaboration success patterns between agents.
type AgentAffinityGraph struct {
	// Adjacency matrix storing collaboration affinity scores
	// Higher scores = more successful collaborations
	affinity map[string]map[string]float64

	// Tier membership for tier-aware recommendations
	agentTiers map[string]int

	// Quick lookup tables using probabilistic routing
	routingTable map[string][]string // Pre-computed top collaborators

	// Success count for computing affinity
	successCount map[string]map[string]int
	totalCount   map[string]map[string]int

	// Decay factor for temporal relevance
	decayFactor float64

	mu sync.RWMutex
}

// NewAgentAffinityGraph creates a new affinity graph for all 40 agents.
func NewAgentAffinityGraph() *AgentAffinityGraph {
	g := &AgentAffinityGraph{
		affinity:     make(map[string]map[string]float64),
		agentTiers:   make(map[string]int),
		routingTable: make(map[string][]string),
		successCount: make(map[string]map[string]int),
		totalCount:   make(map[string]map[string]int),
		decayFactor:  0.95, // 5% decay per update cycle
	}

	// Initialize with the 40 Elite Agents and their tiers
	agents := map[string]int{
		// Tier 1: Foundational
		"APEX": 1, "CIPHER": 1, "ARCHITECT": 1, "AXIOM": 1, "VELOCITY": 1,
		// Tier 2: Specialists
		"QUANTUM": 2, "TENSOR": 2, "FORTRESS": 2, "NEURAL": 2, "CRYPTO": 2,
		"FLUX": 2, "PRISM": 2, "SYNAPSE": 2, "CORE": 2, "HELIX": 2,
		"VANGUARD": 2, "ECLIPSE": 2,
		// Tier 3: Innovators
		"NEXUS": 3, "GENESIS": 3,
		// Tier 4: Meta
		"OMNISCIENT": 4,
		// Tier 5: Domain Specialists
		"ATLAS": 5, "FORGE": 5, "SENTRY": 5, "VERTEX": 5, "STREAM": 5,
		// Tier 6: Emerging Tech
		"PHOTON": 6, "LATTICE": 6, "MORPH": 6, "PHANTOM": 6, "ORBIT": 6,
		// Tier 7: Human-Centric
		"CANVAS": 7, "LINGUA": 7, "SCRIBE": 7, "MENTOR": 7, "BRIDGE": 7,
		// Tier 8: Enterprise
		"AEGIS": 8, "LEDGER": 8, "PULSE": 8, "ARBITER": 8, "ORACLE": 8,
	}

	for agent, tier := range agents {
		g.agentTiers[agent] = tier
		g.affinity[agent] = make(map[string]float64)
		g.successCount[agent] = make(map[string]int)
		g.totalCount[agent] = make(map[string]int)

		// Initialize with prior affinities based on tier proximity
		for other, otherTier := range agents {
			if agent != other {
				// Same tier = higher initial affinity
				tierDist := math.Abs(float64(tier - otherTier))
				g.affinity[agent][other] = 1.0 / (1.0 + tierDist)
			}
		}
	}

	// Initialize routing tables
	g.rebuildRoutingTables()

	return g
}

// RecordCollaboration records a collaboration outcome between agents.
// This is O(1) for the record, amortized O(agents) for routing table updates.
func (g *AgentAffinityGraph) RecordCollaboration(agent1, agent2 string, success bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Update counts
	if g.totalCount[agent1] == nil {
		g.totalCount[agent1] = make(map[string]int)
		g.successCount[agent1] = make(map[string]int)
	}
	g.totalCount[agent1][agent2]++
	if success {
		g.successCount[agent1][agent2]++
	}

	// Symmetric update
	if g.totalCount[agent2] == nil {
		g.totalCount[agent2] = make(map[string]int)
		g.successCount[agent2] = make(map[string]int)
	}
	g.totalCount[agent2][agent1]++
	if success {
		g.successCount[agent2][agent1]++
	}

	// Update affinity using exponential moving average with success boost
	// Success rate based on recent collaborations
	successRate := float64(g.successCount[agent1][agent2]) / float64(g.totalCount[agent1][agent2])

	// Current affinity decays slightly, success boosts it above 1.0 potential
	// This ensures that successful collaborations always increase affinity
	baseAffinity := g.affinity[agent1][agent2]
	if success {
		// Successful collaboration increases affinity with diminishing returns
		boost := 0.1 * (2.0 - baseAffinity) // Larger boost when affinity is lower
		g.affinity[agent1][agent2] = math.Min(2.0, baseAffinity+boost)
	} else {
		// Failed collaboration decays affinity
		g.affinity[agent1][agent2] = math.Max(0.1, baseAffinity*0.95)
	}
	g.affinity[agent2][agent1] = g.affinity[agent1][agent2]

	// Also maintain long-term success rate as a separate metric
	_ = successRate // Stored in successCount/totalCount for advanced analytics

	// Lazy routing table rebuild (every 100 updates or on demand)
	if (g.totalCount[agent1][agent2]+g.totalCount[agent2][agent1])%100 == 0 {
		g.rebuildRoutingTables()
	}
}

// GetTopCollaborators returns the top k collaborators for an agent in O(1).
func (g *AgentAffinityGraph) GetTopCollaborators(agent string, k int) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if routes, ok := g.routingTable[agent]; ok {
		if k > len(routes) {
			k = len(routes)
		}
		result := make([]string, k)
		copy(result, routes[:k])
		return result
	}
	return nil
}

// GetAffinityScore returns the collaboration affinity between two agents in O(1).
func (g *AgentAffinityGraph) GetAffinityScore(agent1, agent2 string) float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if aff, ok := g.affinity[agent1]; ok {
		return aff[agent2]
	}
	return 0
}

// SuggestCollaborationTeam suggests a team of agents for a task.
// Uses random walk with restart to find a cohesive team.
func (g *AgentAffinityGraph) SuggestCollaborationTeam(seedAgent string, teamSize int) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	team := make(map[string]bool)
	team[seedAgent] = true

	current := seedAgent
	restartProb := 0.15 // Probability to restart at seed

	for len(team) < teamSize {
		if rand.Float64() < restartProb {
			current = seedAgent
		} else {
			// Walk to a neighbor with probability proportional to affinity
			neighbors := g.affinity[current]
			totalAff := 0.0
			for other, aff := range neighbors {
				if !team[other] {
					totalAff += aff
				}
			}

			if totalAff == 0 {
				current = seedAgent
				continue
			}

			r := rand.Float64() * totalAff
			cumulative := 0.0
			for other, aff := range neighbors {
				if !team[other] {
					cumulative += aff
					if cumulative >= r {
						team[other] = true
						current = other
						break
					}
				}
			}
		}
	}

	result := make([]string, 0, len(team))
	for agent := range team {
		result = append(result, agent)
	}
	return result
}

// rebuildRoutingTables rebuilds the pre-computed routing tables.
func (g *AgentAffinityGraph) rebuildRoutingTables() {
	for agent, affinities := range g.affinity {
		// Sort by affinity descending
		type agentAff struct {
			agent    string
			affinity float64
		}
		sorted := make([]agentAff, 0, len(affinities))
		for other, aff := range affinities {
			sorted = append(sorted, agentAff{other, aff})
		}
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].affinity > sorted[j].affinity
		})

		routes := make([]string, len(sorted))
		for i, aa := range sorted {
			routes[i] = aa.agent
		}
		g.routingTable[agent] = routes
	}
}

// ============================================================================
// 2. TIER RESONANCE FILTER
// ============================================================================
// O(1) tier-aware experience propagation using "resonance" patterns.
// Experiences naturally propagate to appropriate tiers based on content signature.
//
// Innovation: Uses hierarchical Bloom filters with tier-specific hash functions.
// An experience "resonates" with a tier if its signature matches that tier's filter.

// TierResonanceFilter enables O(1) tier-appropriate experience routing.
type TierResonanceFilter struct {
	// Each tier has its own Bloom filter trained on tier-specific patterns
	tierFilters [9][]uint64 // Tiers 1-8 (index 0 unused)

	// Filter parameters
	filterSize   int
	numHashFuncs int

	// Tier signature patterns (learned from successful experiences)
	tierKeywords [9][]string

	mu sync.RWMutex
}

// NewTierResonanceFilter creates a new tier resonance filter.
func NewTierResonanceFilter() *TierResonanceFilter {
	f := &TierResonanceFilter{
		filterSize:   8192,
		numHashFuncs: 5,
	}

	// Initialize filters
	for tier := 1; tier <= 8; tier++ {
		f.tierFilters[tier] = make([]uint64, f.filterSize/64)
	}

	// Initialize tier keywords (domain knowledge)
	f.tierKeywords = [9][]string{
		{}, // Index 0 unused
		{"algorithm", "code", "architecture", "security", "performance", "optimization"}, // Tier 1: Foundational
		{"quantum", "ml", "neural", "blockchain", "devops", "data", "api", "compiler", "bio", "research", "testing"},  // Tier 2: Specialists
		{"synthesis", "innovation", "cross-domain", "novel", "paradigm"},                   // Tier 3: Innovators
		{"orchestrate", "collective", "meta", "coordinate", "evolve"},                      // Tier 4: Meta
		{"cloud", "build", "monitor", "graph", "stream"},                                   // Tier 5: Domain
		{"edge", "iot", "consensus", "migration", "reverse", "embedded"},                   // Tier 6: Emerging
		{"ui", "ux", "nlp", "documentation", "education", "mobile"},                        // Tier 7: Human-Centric
		{"compliance", "finance", "healthcare", "merge", "analytics"},                      // Tier 8: Enterprise
	}

	// Seed the filters with initial keywords
	for tier := 1; tier <= 8; tier++ {
		for _, kw := range f.tierKeywords[tier] {
			f.addToTierFilter(tier, kw)
		}
	}

	return f
}

// addToTierFilter adds a pattern to a tier's filter.
func (f *TierResonanceFilter) addToTierFilter(tier int, pattern string) {
	for i := 0; i < f.numHashFuncs; i++ {
		hash := f.tierHash(pattern, i)
		idx := hash % uint64(f.filterSize)
		f.tierFilters[tier][idx/64] |= 1 << (idx % 64)
	}
}

// tierHash computes a hash for a pattern with a seed.
func (f *TierResonanceFilter) tierHash(pattern string, seed int) uint64 {
	h := fnv.New64a()
	binary.Write(h, binary.LittleEndian, uint32(seed))
	h.Write([]byte(strings.ToLower(pattern)))
	return h.Sum64()
}

// FindResonantTiers finds which tiers an experience resonates with.
// Returns tiers in order of resonance strength. O(8) = O(1) constant time.
func (f *TierResonanceFilter) FindResonantTiers(content string) []int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Tokenize content
	tokens := tokenizeForResonance(content)

	// Calculate resonance score for each tier
	type tierScore struct {
		tier  int
		score float64
	}
	scores := make([]tierScore, 0, 8)

	for tier := 1; tier <= 8; tier++ {
		matches := 0
		for _, token := range tokens {
			if f.checkTierFilter(tier, token) {
				matches++
			}
		}
		if matches > 0 {
			scores = append(scores, tierScore{tier, float64(matches) / float64(len(tokens)+1)})
		}
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	result := make([]int, len(scores))
	for i, ts := range scores {
		result[i] = ts.tier
	}
	return result
}

// checkTierFilter checks if a token matches a tier's filter.
func (f *TierResonanceFilter) checkTierFilter(tier int, token string) bool {
	for i := 0; i < f.numHashFuncs; i++ {
		hash := f.tierHash(token, i)
		idx := hash % uint64(f.filterSize)
		if f.tierFilters[tier][idx/64]&(1<<(idx%64)) == 0 {
			return false
		}
	}
	return true
}

// LearnFromExperience updates tier filters based on successful experience.
func (f *TierResonanceFilter) LearnFromExperience(tier int, content string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	tokens := tokenizeForResonance(content)
	for _, token := range tokens {
		if len(token) > 3 { // Only learn meaningful tokens
			f.addToTierFilter(tier, token)
		}
	}
}

// tokenizeForResonance splits content into tokens for resonance matching.
func tokenizeForResonance(content string) []string {
	words := strings.FieldsFunc(strings.ToLower(content), func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})
	return words
}

// ============================================================================
// 3. SKILL BLOOM CASCADE
// ============================================================================
// O(k) multi-skill matching using cascading Bloom filters.
// Inspired by cascade classifiers in computer vision (Viola-Jones).
//
// Innovation: Each agent has a "skill signature" stored in a Bloom filter.
// Query with required skills cascades through filters, eliminating non-matches.

// SkillBloomCascade enables fast multi-skill agent matching.
type SkillBloomCascade struct {
	// Per-agent skill filters
	agentFilters map[string]*SkillFilter

	// Inverted index: skill -> agents (for cascade optimization)
	skillIndex map[string][]string

	mu sync.RWMutex
}

// SkillFilter is a Bloom filter for an agent's skills.
type SkillFilter struct {
	bits     []uint64
	size     int
	numHash  int
	skills   []string
	agentID  string
}

// NewSkillBloomCascade creates a new skill cascade.
func NewSkillBloomCascade() *SkillBloomCascade {
	c := &SkillBloomCascade{
		agentFilters: make(map[string]*SkillFilter),
		skillIndex:   make(map[string][]string),
	}

	// Initialize with known agent skills
	agentSkills := map[string][]string{
		"APEX":      {"algorithm", "code", "system design", "clean code", "patterns", "python", "go", "rust", "java"},
		"CIPHER":    {"cryptography", "encryption", "security", "tls", "pki", "authentication", "zero-knowledge"},
		"ARCHITECT": {"architecture", "microservices", "ddd", "cqrs", "scalability", "distributed systems"},
		"AXIOM":     {"mathematics", "proofs", "complexity", "algorithms", "formal verification", "logic"},
		"VELOCITY":  {"performance", "optimization", "profiling", "caching", "simd", "concurrency", "sub-linear"},
		"TENSOR":    {"machine learning", "deep learning", "neural networks", "pytorch", "tensorflow", "mlops"},
		"QUANTUM":   {"quantum computing", "qiskit", "quantum algorithms", "error correction"},
		"FORTRESS":  {"penetration testing", "security audit", "red team", "forensics", "incident response"},
		"NEURAL":    {"agi", "cognitive", "neurosymbolic", "meta-learning", "alignment"},
		"CRYPTO":    {"blockchain", "smart contracts", "solidity", "defi", "consensus"},
		"FLUX":      {"devops", "kubernetes", "docker", "terraform", "ci/cd", "gitops"},
		"PRISM":     {"data science", "statistics", "bayesian", "a/b testing", "visualization"},
		"SYNAPSE":   {"api design", "rest", "graphql", "grpc", "integration", "kafka"},
		"CORE":      {"systems", "compiler", "assembly", "kernel", "memory management"},
		"HELIX":     {"bioinformatics", "genomics", "protein", "alphafold", "drug discovery"},
		"VANGUARD":  {"research", "literature review", "meta-analysis", "academic writing"},
		"ECLIPSE":   {"testing", "unit test", "integration test", "fuzzing", "formal methods", "tla+"},
		"NEXUS":     {"synthesis", "cross-domain", "innovation", "paradigm", "biomimicry"},
		"GENESIS":   {"first principles", "novel", "invention", "discovery", "zero-to-one"},
		"OMNISCIENT": {"orchestration", "coordination", "collective", "meta-learning", "evolution"},
		"ATLAS":     {"cloud", "aws", "azure", "gcp", "multi-cloud", "infrastructure"},
		"FORGE":     {"build systems", "bazel", "cmake", "gradle", "monorepo"},
		"SENTRY":    {"observability", "monitoring", "logging", "prometheus", "grafana", "tracing"},
		"VERTEX":    {"graph database", "neo4j", "cypher", "knowledge graph", "network analysis"},
		"STREAM":    {"streaming", "kafka", "flink", "event sourcing", "real-time"},
		"PHOTON":    {"edge computing", "iot", "mqtt", "embedded", "tinyml"},
		"LATTICE":   {"consensus", "crdt", "distributed", "byzantine", "raft", "paxos"},
		"MORPH":     {"migration", "legacy", "modernization", "refactoring", "monolith"},
		"PHANTOM":   {"reverse engineering", "binary analysis", "malware", "ghidra", "ida"},
		"ORBIT":     {"satellite", "rtos", "space", "fault tolerance", "radiation"},
		"CANVAS":    {"ui", "ux", "design systems", "accessibility", "css", "frontend"},
		"LINGUA":    {"nlp", "llm", "fine-tuning", "rag", "embeddings", "transformers"},
		"SCRIBE":    {"documentation", "technical writing", "api docs", "tutorials"},
		"MENTOR":    {"code review", "education", "teaching", "mentorship", "interview"},
		"BRIDGE":    {"cross-platform", "mobile", "react native", "flutter", "electron"},
		"AEGIS":     {"compliance", "gdpr", "soc2", "iso27001", "audit"},
		"LEDGER":    {"finance", "accounting", "payment", "fintech", "trading"},
		"PULSE":     {"healthcare", "hipaa", "hl7", "fhir", "medical devices"},
		"ARBITER":   {"merge", "conflict", "git", "branching", "collaboration"},
		"ORACLE":    {"analytics", "forecasting", "time series", "prediction", "business intelligence"},
	}

	for agent, skills := range agentSkills {
		c.AddAgent(agent, skills)
	}

	return c
}

// AddAgent adds an agent with their skills.
func (c *SkillBloomCascade) AddAgent(agentID string, skills []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	filter := &SkillFilter{
		size:    512,
		numHash: 4,
		skills:  skills,
		agentID: agentID,
	}
	filter.bits = make([]uint64, filter.size/64)

	for _, skill := range skills {
		normalizedSkill := strings.ToLower(skill)
		filter.addSkill(normalizedSkill)

		// Update inverted index
		if c.skillIndex[normalizedSkill] == nil {
			c.skillIndex[normalizedSkill] = []string{}
		}
		c.skillIndex[normalizedSkill] = append(c.skillIndex[normalizedSkill], agentID)
	}

	c.agentFilters[agentID] = filter
}

// addSkill adds a skill to the filter.
func (f *SkillFilter) addSkill(skill string) {
	for i := 0; i < f.numHash; i++ {
		h := fnv.New64a()
		binary.Write(h, binary.LittleEndian, uint32(i))
		h.Write([]byte(skill))
		idx := h.Sum64() % uint64(f.size)
		f.bits[idx/64] |= 1 << (idx % 64)
	}
}

// hasSkill checks if a skill might be present.
func (f *SkillFilter) hasSkill(skill string) bool {
	for i := 0; i < f.numHash; i++ {
		h := fnv.New64a()
		binary.Write(h, binary.LittleEndian, uint32(i))
		h.Write([]byte(skill))
		idx := h.Sum64() % uint64(f.size)
		if f.bits[idx/64]&(1<<(idx%64)) == 0 {
			return false
		}
	}
	return true
}

// FindAgentsWithSkills finds agents matching required skills using cascade.
// Returns agents in order of skill match count. O(k) where k = skills checked.
func (c *SkillBloomCascade) FindAgentsWithSkills(requiredSkills []string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Start with agents matching first skill (using inverted index)
	if len(requiredSkills) == 0 {
		return nil
	}

	// Cascade through skills, eliminating non-matches
	normalizedSkills := make([]string, len(requiredSkills))
	for i, s := range requiredSkills {
		normalizedSkills[i] = strings.ToLower(s)
	}

	// Score each agent by skill match
	type agentScore struct {
		agent string
		score int
	}
	scores := []agentScore{}

	for agentID, filter := range c.agentFilters {
		matchCount := 0
		for _, skill := range normalizedSkills {
			if filter.hasSkill(skill) {
				matchCount++
			}
		}
		if matchCount > 0 {
			scores = append(scores, agentScore{agentID, matchCount})
		}
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	result := make([]string, len(scores))
	for i, as := range scores {
		result[i] = as.agent
	}
	return result
}

// ============================================================================
// 4. TEMPORAL DECAY SKETCH
// ============================================================================
// O(1) time-aware frequency counting with automatic forgetting.
// Inspired by memory consolidation in the human brain.
//
// Innovation: Combines Count-Min Sketch with exponential time decay.
// Recent events weighted higher, old events fade automatically.

// TemporalDecaySketch tracks frequency with time decay.
type TemporalDecaySketch struct {
	// Count-Min matrix with timestamps
	counts     [][]float64
	timestamps [][]time.Time

	width    int
	depth    int
	halfLife time.Duration // Time for count to decay to half

	mu sync.RWMutex
}

// NewTemporalDecaySketch creates a new temporal decay sketch.
func NewTemporalDecaySketch(halfLife time.Duration) *TemporalDecaySketch {
	width := 1024
	depth := 4

	s := &TemporalDecaySketch{
		counts:     make([][]float64, depth),
		timestamps: make([][]time.Time, depth),
		width:      width,
		depth:      depth,
		halfLife:   halfLife,
	}

	now := time.Now()
	for i := 0; i < depth; i++ {
		s.counts[i] = make([]float64, width)
		s.timestamps[i] = make([]time.Time, width)
		for j := 0; j < width; j++ {
			s.timestamps[i][j] = now
		}
	}

	return s
}

// NewTemporalDecaySketchDefault creates a sketch with 24-hour half-life.
func NewTemporalDecaySketchDefault() *TemporalDecaySketch {
	return NewTemporalDecaySketch(24 * time.Hour)
}

// hash computes hash for position in matrix.
func (s *TemporalDecaySketch) hash(key string, seed int) int {
	h := fnv.New64a()
	binary.Write(h, binary.LittleEndian, uint32(seed))
	h.Write([]byte(key))
	return int(h.Sum64() % uint64(s.width))
}

// decay computes the decay factor based on time elapsed.
func (s *TemporalDecaySketch) decay(lastUpdate time.Time) float64 {
	elapsed := time.Since(lastUpdate)
	// Exponential decay: count * 2^(-elapsed/halfLife)
	return math.Pow(2, -float64(elapsed)/float64(s.halfLife))
}

// Add adds an observation with weight 1.
func (s *TemporalDecaySketch) Add(key string) {
	s.AddWithWeight(key, 1.0)
}

// AddWithWeight adds an observation with custom weight.
func (s *TemporalDecaySketch) AddWithWeight(key string, weight float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for i := 0; i < s.depth; i++ {
		idx := s.hash(key, i)

		// Decay existing count
		decayedCount := s.counts[i][idx] * s.decay(s.timestamps[i][idx])

		// Add new weight
		s.counts[i][idx] = decayedCount + weight
		s.timestamps[i][idx] = now
	}
}

// Estimate returns the decayed frequency estimate.
func (s *TemporalDecaySketch) Estimate(key string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	minCount := math.MaxFloat64
	for i := 0; i < s.depth; i++ {
		idx := s.hash(key, i)
		decayedCount := s.counts[i][idx] * s.decay(s.timestamps[i][idx])
		if decayedCount < minCount {
			minCount = decayedCount
		}
	}

	if minCount == math.MaxFloat64 {
		return 0
	}
	return minCount
}

// EstimateRecent returns frequency weighted by recency.
func (s *TemporalDecaySketch) EstimateRecent(key string, window time.Duration) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cutoff := time.Now().Add(-window)
	minCount := math.MaxFloat64

	for i := 0; i < s.depth; i++ {
		idx := s.hash(key, i)
		if s.timestamps[i][idx].After(cutoff) {
			decayedCount := s.counts[i][idx] * s.decay(s.timestamps[i][idx])
			if decayedCount < minCount {
				minCount = decayedCount
			}
		}
	}

	if minCount == math.MaxFloat64 {
		return 0
	}
	return minCount
}

// ============================================================================
// 5. COLLABORATIVE ATTENTION INDEX
// ============================================================================
// O(1) attention-based agent routing using learned attention weights.
// Inspired by Transformer attention mechanisms.
//
// Innovation: Pre-compute "attention scores" between query patterns and agents.
// Queries are routed to agents with highest attention scores.

// CollaborativeAttentionIndex routes queries to agents using attention.
type CollaborativeAttentionIndex struct {
	// Learned attention weights: pattern category -> agent -> weight
	attentionWeights map[string]map[string]float64

	// Pattern embeddings (simplified as keyword sets)
	patternCategories map[string][]string

	// Agent capability vectors
	agentCapabilities map[string][]float64

	// Learning rate for weight updates
	learningRate float64

	mu sync.RWMutex
}

// NewCollaborativeAttentionIndex creates a new attention index.
func NewCollaborativeAttentionIndex() *CollaborativeAttentionIndex {
	idx := &CollaborativeAttentionIndex{
		attentionWeights:  make(map[string]map[string]float64),
		patternCategories: make(map[string][]string),
		agentCapabilities: make(map[string][]float64),
		learningRate:      0.1,
	}

	// Initialize pattern categories
	idx.patternCategories = map[string][]string{
		"coding":       {"implement", "code", "function", "class", "algorithm", "fix", "debug"},
		"architecture": {"design", "system", "architecture", "scale", "microservice", "pattern"},
		"security":     {"secure", "encrypt", "vulnerability", "authentication", "audit"},
		"performance":  {"optimize", "performance", "fast", "efficient", "cache", "benchmark"},
		"ml":           {"model", "train", "neural", "learning", "predict", "classify"},
		"data":         {"data", "analyze", "statistics", "query", "database", "sql"},
		"devops":       {"deploy", "kubernetes", "docker", "ci/cd", "infrastructure"},
		"testing":      {"test", "verify", "assert", "coverage", "unit", "integration"},
		"documentation": {"document", "explain", "write", "tutorial", "readme"},
		"research":     {"research", "paper", "literature", "survey", "study"},
	}

	// Initialize attention weights uniformly
	allAgents := []string{
		"APEX", "CIPHER", "ARCHITECT", "AXIOM", "VELOCITY",
		"TENSOR", "QUANTUM", "FORTRESS", "NEURAL", "CRYPTO",
		"FLUX", "PRISM", "SYNAPSE", "CORE", "HELIX",
		"VANGUARD", "ECLIPSE", "NEXUS", "GENESIS", "OMNISCIENT",
		"ATLAS", "FORGE", "SENTRY", "VERTEX", "STREAM",
		"PHOTON", "LATTICE", "MORPH", "PHANTOM", "ORBIT",
		"CANVAS", "LINGUA", "SCRIBE", "MENTOR", "BRIDGE",
		"AEGIS", "LEDGER", "PULSE", "ARBITER", "ORACLE",
	}

	for category := range idx.patternCategories {
		idx.attentionWeights[category] = make(map[string]float64)
		for _, agent := range allAgents {
			idx.attentionWeights[category][agent] = 1.0 / float64(len(allAgents))
		}
	}

	// Set prior attention based on agent specializations
	idx.setPriorAttention()

	return idx
}

// setPriorAttention sets initial attention based on agent specializations.
func (idx *CollaborativeAttentionIndex) setPriorAttention() {
	priors := map[string]map[string]float64{
		"coding":        {"APEX": 0.3, "CORE": 0.2, "ECLIPSE": 0.15},
		"architecture":  {"ARCHITECT": 0.4, "APEX": 0.2, "ATLAS": 0.15},
		"security":      {"CIPHER": 0.3, "FORTRESS": 0.3, "AEGIS": 0.15},
		"performance":   {"VELOCITY": 0.4, "APEX": 0.2, "CORE": 0.15},
		"ml":            {"TENSOR": 0.35, "NEURAL": 0.25, "PRISM": 0.15},
		"data":          {"PRISM": 0.3, "VERTEX": 0.2, "STREAM": 0.15},
		"devops":        {"FLUX": 0.35, "ATLAS": 0.25, "SENTRY": 0.15},
		"testing":       {"ECLIPSE": 0.4, "APEX": 0.2, "AXIOM": 0.15},
		"documentation": {"SCRIBE": 0.4, "MENTOR": 0.25, "LINGUA": 0.15},
		"research":      {"VANGUARD": 0.4, "GENESIS": 0.2, "NEXUS": 0.15},
	}

	for category, weights := range priors {
		for agent, weight := range weights {
			idx.attentionWeights[category][agent] = weight
		}
	}

	// Normalize weights
	for category := range idx.attentionWeights {
		total := 0.0
		for _, w := range idx.attentionWeights[category] {
			total += w
		}
		for agent := range idx.attentionWeights[category] {
			idx.attentionWeights[category][agent] /= total
		}
	}
}

// RouteQuery routes a query to the most relevant agents using attention.
// Returns top k agents with their attention scores. O(categories * agents) = O(1)
func (idx *CollaborativeAttentionIndex) RouteQuery(query string, topK int) []AgentAttention {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	queryLower := strings.ToLower(query)

	// Compute attention scores for each agent
	agentScores := make(map[string]float64)

	for category, keywords := range idx.patternCategories {
		// Check if query matches this category
		categoryMatch := 0.0
		for _, kw := range keywords {
			if strings.Contains(queryLower, kw) {
				categoryMatch += 1.0
			}
		}

		if categoryMatch > 0 {
			// Add weighted attention from this category
			categoryWeight := categoryMatch / float64(len(keywords))
			for agent, attention := range idx.attentionWeights[category] {
				agentScores[agent] += categoryWeight * attention
			}
		}
	}

	// Convert to sorted list
	type agentScore struct {
		agent string
		score float64
	}
	scores := make([]agentScore, 0, len(agentScores))
	for agent, score := range agentScores {
		scores = append(scores, agentScore{agent, score})
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Return top k
	if topK > len(scores) {
		topK = len(scores)
	}
	result := make([]AgentAttention, topK)
	for i := 0; i < topK; i++ {
		result[i] = AgentAttention{
			AgentID:   scores[i].agent,
			Attention: scores[i].score,
		}
	}
	return result
}

// AgentAttention represents an agent with attention score.
type AgentAttention struct {
	AgentID   string
	Attention float64
}

// UpdateAttention updates attention weights based on feedback.
// Uses gradient descent-like update.
func (idx *CollaborativeAttentionIndex) UpdateAttention(query string, selectedAgent string, success bool) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	queryLower := strings.ToLower(query)
	reward := -0.5
	if success {
		reward = 1.0
	}

	for category, keywords := range idx.patternCategories {
		categoryMatch := 0.0
		for _, kw := range keywords {
			if strings.Contains(queryLower, kw) {
				categoryMatch += 1.0
			}
		}

		if categoryMatch > 0 {
			// Update attention for selected agent
			currentWeight := idx.attentionWeights[category][selectedAgent]
			newWeight := currentWeight + idx.learningRate*reward*(1-currentWeight)
			if newWeight < 0.01 {
				newWeight = 0.01
			}
			idx.attentionWeights[category][selectedAgent] = newWeight

			// Normalize
			total := 0.0
			for _, w := range idx.attentionWeights[category] {
				total += w
			}
			for agent := range idx.attentionWeights[category] {
				idx.attentionWeights[category][agent] /= total
			}
		}
	}
}

// ============================================================================
// 6. EMERGENT INSIGHT DETECTOR
// ============================================================================
// O(1) detection of cross-domain breakthroughs using anomaly detection.
// Inspired by serendipity in scientific discovery.
//
// Innovation: Track "surprise" when agent combinations produce unexpected results.
// High surprise = potential breakthrough worth propagating.

// EmergentInsightDetector detects unexpected cross-agent discoveries.
type EmergentInsightDetector struct {
	// Expected outcome distributions per agent pair
	// Stores mean and variance of success rates
	expectedOutcomes map[string]*OutcomeDistribution

	// Recent surprising events for analysis
	surpriseBuffer []SurpriseEvent

	// Threshold for detecting breakthrough
	surpriseThreshold float64

	mu sync.RWMutex
}

// OutcomeDistribution tracks outcome statistics for agent combinations.
type OutcomeDistribution struct {
	SuccessCount int
	TotalCount   int
	Mean         float64
	Variance     float64
}

// SurpriseEvent records an unexpectedly successful outcome.
type SurpriseEvent struct {
	Agents       []string
	TaskType     string
	SurpriseScore float64
	Timestamp    time.Time
	Strategy     string
}

// NewEmergentInsightDetector creates a new insight detector.
func NewEmergentInsightDetector() *EmergentInsightDetector {
	return &EmergentInsightDetector{
		expectedOutcomes:  make(map[string]*OutcomeDistribution),
		surpriseBuffer:    make([]SurpriseEvent, 0, 100),
		surpriseThreshold: 2.0, // 2 standard deviations
	}
}

// agentPairKey creates a canonical key for agent pairs.
func agentPairKey(agents []string) string {
	sorted := make([]string, len(agents))
	copy(sorted, agents)
	sort.Strings(sorted)
	return strings.Join(sorted, "+")
}

// RecordOutcome records an outcome for agent combination.
// Returns surprise score (high = potential breakthrough).
func (d *EmergentInsightDetector) RecordOutcome(agents []string, taskType string, success bool, strategy string) float64 {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := agentPairKey(agents) + ":" + taskType
	dist, exists := d.expectedOutcomes[key]
	if !exists {
		dist = &OutcomeDistribution{
			Mean:     0.5, // Prior: 50% success
			Variance: 0.25,
		}
		d.expectedOutcomes[key] = dist
	}

	// Compute surprise score
	outcome := 0.0
	if success {
		outcome = 1.0
	}

	// Surprise = |outcome - expected| / stddev
	stddev := math.Sqrt(dist.Variance + 0.01) // Add small epsilon
	surprise := math.Abs(outcome-dist.Mean) / stddev

	// Update distribution with Welford's algorithm
	dist.TotalCount++
	if success {
		dist.SuccessCount++
	}

	delta := outcome - dist.Mean
	dist.Mean += delta / float64(dist.TotalCount)
	delta2 := outcome - dist.Mean
	if dist.TotalCount > 1 {
		dist.Variance = ((float64(dist.TotalCount-1) * dist.Variance) + delta*delta2) / float64(dist.TotalCount)
	}

	// Record surprise event if above threshold
	if surprise > d.surpriseThreshold && success {
		event := SurpriseEvent{
			Agents:        agents,
			TaskType:      taskType,
			SurpriseScore: surprise,
			Timestamp:     time.Now(),
			Strategy:      strategy,
		}

		// Keep buffer bounded
		if len(d.surpriseBuffer) >= 100 {
			d.surpriseBuffer = d.surpriseBuffer[1:]
		}
		d.surpriseBuffer = append(d.surpriseBuffer, event)
	}

	return surprise
}

// GetRecentBreakthroughs returns recent surprising successes.
func (d *EmergentInsightDetector) GetRecentBreakthroughs(limit int) []SurpriseEvent {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Sort by surprise score descending
	sorted := make([]SurpriseEvent, len(d.surpriseBuffer))
	copy(sorted, d.surpriseBuffer)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].SurpriseScore > sorted[j].SurpriseScore
	})

	if limit > len(sorted) {
		limit = len(sorted)
	}
	return sorted[:limit]
}

// GetUnexpectedPairs returns agent pairs with higher than expected success.
func (d *EmergentInsightDetector) GetUnexpectedPairs(minSamples int) []UnexpectedPair {
	d.mu.RLock()
	defer d.mu.RUnlock()

	pairs := []UnexpectedPair{}
	for key, dist := range d.expectedOutcomes {
		if dist.TotalCount >= minSamples {
			// Check if success rate is surprisingly high
			observedRate := float64(dist.SuccessCount) / float64(dist.TotalCount)
			expectedRate := 0.5 // Prior expectation
			if observedRate > expectedRate+0.2 { // 20% higher than expected
				parts := strings.Split(key, ":")
				agents := strings.Split(parts[0], "+")
				taskType := ""
				if len(parts) > 1 {
					taskType = parts[1]
				}
				pairs = append(pairs, UnexpectedPair{
					Agents:       agents,
					TaskType:     taskType,
					SuccessRate:  observedRate,
					SampleCount:  dist.TotalCount,
				})
			}
		}
	}

	// Sort by success rate descending
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].SuccessRate > pairs[j].SuccessRate
	})

	return pairs
}

// UnexpectedPair represents an unexpectedly successful agent combination.
type UnexpectedPair struct {
	Agents      []string
	TaskType    string
	SuccessRate float64
	SampleCount int
}
