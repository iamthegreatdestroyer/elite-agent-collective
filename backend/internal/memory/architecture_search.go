// Package memory implements the MNEMONIC memory system for the Elite Agent Collective.
// This file implements Architecture Search for finding optimal agent team compositions.
package memory

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// =============================================================================
// ARCHITECTURE SEARCH FOR AGENT TEAMS
// =============================================================================

// TeamArchitectureSearch implements neural architecture search for agent team composition
type TeamArchitectureSearch struct {
	mu sync.RWMutex

	// Search space definition
	searchSpace *AgentSearchSpace

	// Population of architectures
	population []*TeamArchitecture

	// Best architectures discovered
	eliteArchitectures []*TeamArchitecture

	// Performance history
	performanceHistory map[string]*ArchitecturePerformance

	// Configuration
	config *ArchitectureSearchConfig

	// Metrics
	metrics *SearchMetrics
}

// AgentSearchSpace defines the space of possible team configurations
type AgentSearchSpace struct {
	AvailableAgents   []string            // All available agent IDs
	AgentCapabilities map[string][]string // Agent -> capabilities
	AgentTiers        map[string]int      // Agent -> tier level
	MaxTeamSize       int
	MinTeamSize       int
	RequiredRoles     []string                      // Roles that must be filled
	ForbiddenPairs    [][2]string                   // Agent pairs that shouldn't be together
	SynergyMatrix     map[string]map[string]float64 // Synergy scores
}

// TeamArchitecture represents a specific team configuration
type TeamArchitecture struct {
	ID               string
	Agents           []string          // Ordered list of agents
	Roles            map[string]string // Role -> AgentID assignment
	Structure        TeamStructure     // How team is organized
	Fitness          float64
	ValidationScore  float64
	GenerationNumber int
	ParentIDs        []string
	Mutations        []string
	CreatedAt        time.Time
}

// TeamStructure defines how a team is organized
type TeamStructure int

const (
	StructureHierarchical TeamStructure = iota // One lead, others support
	StructureFlat                              // All equal
	StructureSpecialized                       // Each has distinct role
	StructureHybrid                            // Mix of above
)

// ArchitecturePerformance tracks how an architecture performs
type ArchitecturePerformance struct {
	ArchitectureID string
	TaskCount      int
	SuccessRate    float64
	AverageQuality float64
	AverageLatency time.Duration
	StrengthAreas  []string
	WeaknessAreas  []string
	LastEvaluation time.Time
}

// ArchitectureSearchConfig holds configuration for the search
type ArchitectureSearchConfig struct {
	PopulationSize    int
	EliteCount        int
	MutationRate      float64
	CrossoverRate     float64
	MaxGenerations    int
	TournamentSize    int
	ConvergenceThresh float64
	EvaluationTimeout time.Duration
	DiversityWeight   float64
}

// DefaultArchitectureSearchConfig returns sensible defaults
func DefaultArchitectureSearchConfig() *ArchitectureSearchConfig {
	return &ArchitectureSearchConfig{
		PopulationSize:    50,
		EliteCount:        5,
		MutationRate:      0.2,
		CrossoverRate:     0.7,
		MaxGenerations:    100,
		TournamentSize:    5,
		ConvergenceThresh: 0.001,
		EvaluationTimeout: 10 * time.Second,
		DiversityWeight:   0.1,
	}
}

// SearchMetrics tracks search progress
type SearchMetrics struct {
	mu sync.RWMutex

	TotalGenerations    int
	TotalEvaluations    int
	BestFitness         float64
	AverageFitness      float64
	DiversityScore      float64
	ConvergenceProgress float64
	SearchDuration      time.Duration
}

// NewTeamArchitectureSearch creates a new architecture search instance
func NewTeamArchitectureSearch(searchSpace *AgentSearchSpace, config *ArchitectureSearchConfig) *TeamArchitectureSearch {
	if config == nil {
		config = DefaultArchitectureSearchConfig()
	}

	return &TeamArchitectureSearch{
		searchSpace:        searchSpace,
		population:         make([]*TeamArchitecture, 0, config.PopulationSize),
		eliteArchitectures: make([]*TeamArchitecture, 0, config.EliteCount),
		performanceHistory: make(map[string]*ArchitecturePerformance),
		config:             config,
		metrics:            &SearchMetrics{},
	}
}

// InitializePopulation creates initial random population
func (s *TeamArchitectureSearch) InitializePopulation() error {
	if s.searchSpace == nil || len(s.searchSpace.AvailableAgents) == 0 {
		return errors.New("search space not properly configured")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.population = make([]*TeamArchitecture, 0, s.config.PopulationSize)

	for i := 0; i < s.config.PopulationSize; i++ {
		arch := s.generateRandomArchitecture(i)
		s.population = append(s.population, arch)
	}

	return nil
}

// generateRandomArchitecture creates a random valid architecture
func (s *TeamArchitectureSearch) generateRandomArchitecture(generation int) *TeamArchitecture {
	// Determine team size
	minSize := s.searchSpace.MinTeamSize
	maxSize := s.searchSpace.MaxTeamSize
	if minSize < 1 {
		minSize = 1
	}
	if maxSize < minSize {
		maxSize = minSize
	}
	if maxSize > len(s.searchSpace.AvailableAgents) {
		maxSize = len(s.searchSpace.AvailableAgents)
	}

	teamSize := minSize + rand.Intn(maxSize-minSize+1)

	// Shuffle and select agents
	agents := make([]string, len(s.searchSpace.AvailableAgents))
	copy(agents, s.searchSpace.AvailableAgents)
	rand.Shuffle(len(agents), func(i, j int) {
		agents[i], agents[j] = agents[j], agents[i]
	})

	selectedAgents := agents[:teamSize]

	// Validate no forbidden pairs
	selectedAgents = s.removeForbiddenPairs(selectedAgents)

	// Assign random structure
	structures := []TeamStructure{StructureHierarchical, StructureFlat, StructureSpecialized, StructureHybrid}
	structure := structures[rand.Intn(len(structures))]

	return &TeamArchitecture{
		ID:               generateArchitectureID(),
		Agents:           selectedAgents,
		Roles:            make(map[string]string),
		Structure:        structure,
		Fitness:          0,
		GenerationNumber: generation,
		CreatedAt:        time.Now(),
	}
}

func generateArchitectureID() string {
	return time.Now().Format("20060102150405") + "-" + archRandomString(6)
}

func archRandomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	result := make([]rune, n)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

// removeForbiddenPairs removes agents that form forbidden pairs
func (s *TeamArchitectureSearch) removeForbiddenPairs(agents []string) []string {
	if len(s.searchSpace.ForbiddenPairs) == 0 {
		return agents
	}

	agentSet := make(map[string]bool)
	for _, a := range agents {
		agentSet[a] = true
	}

	for _, pair := range s.searchSpace.ForbiddenPairs {
		if agentSet[pair[0]] && agentSet[pair[1]] {
			// Remove one of the pair randomly
			if rand.Float32() < 0.5 {
				delete(agentSet, pair[0])
			} else {
				delete(agentSet, pair[1])
			}
		}
	}

	result := make([]string, 0, len(agentSet))
	for agent := range agentSet {
		result = append(result, agent)
	}
	return result
}

// EvaluateArchitecture evaluates a single architecture
func (s *TeamArchitectureSearch) EvaluateArchitecture(ctx context.Context, arch *TeamArchitecture, evaluator ArchitectureEvaluator) error {
	if evaluator == nil {
		// Use default fitness function
		arch.Fitness = s.computeDefaultFitness(arch)
		return nil
	}

	result, err := evaluator.Evaluate(ctx, arch)
	if err != nil {
		return err
	}

	arch.Fitness = result.Fitness
	arch.ValidationScore = result.ValidationScore

	// Store performance (already under lock in RunGeneration context)
	s.performanceHistory[arch.ID] = &ArchitecturePerformance{
		ArchitectureID: arch.ID,
		TaskCount:      result.TaskCount,
		SuccessRate:    result.SuccessRate,
		AverageQuality: result.AverageQuality,
		LastEvaluation: time.Now(),
	}
	s.metrics.TotalEvaluations++

	return nil
}

// ArchitectureEvaluator interface for evaluating architectures
type ArchitectureEvaluator interface {
	Evaluate(ctx context.Context, arch *TeamArchitecture) (*EvaluationResult, error)
}

// EvaluationResult holds the result of architecture evaluation
type EvaluationResult struct {
	Fitness         float64
	ValidationScore float64
	TaskCount       int
	SuccessRate     float64
	AverageQuality  float64
	Strengths       []string
	Weaknesses      []string
}

// computeDefaultFitness computes fitness using heuristics
func (s *TeamArchitectureSearch) computeDefaultFitness(arch *TeamArchitecture) float64 {
	fitness := 0.0

	// Factor 1: Team size appropriateness (prefer moderate sizes)
	sizeScore := 1.0 - math.Abs(float64(len(arch.Agents))-3.0)/5.0
	if sizeScore < 0 {
		sizeScore = 0.1
	}
	fitness += sizeScore * 0.2

	// Factor 2: Tier diversity (prefer mix of tiers)
	tierDiversity := s.computeTierDiversity(arch)
	fitness += tierDiversity * 0.2

	// Factor 3: Capability coverage
	capabilityCoverage := s.computeCapabilityCoverage(arch)
	fitness += capabilityCoverage * 0.3

	// Factor 4: Synergy between agents
	synergy := s.computeSynergyScore(arch)
	fitness += synergy * 0.3

	return fitness
}

// computeTierDiversity measures tier distribution
func (s *TeamArchitectureSearch) computeTierDiversity(arch *TeamArchitecture) float64 {
	tierCounts := make(map[int]int)
	for _, agent := range arch.Agents {
		if tier, ok := s.searchSpace.AgentTiers[agent]; ok {
			tierCounts[tier]++
		}
	}

	if len(tierCounts) == 0 {
		return 0.5
	}

	// Shannon entropy normalized
	total := float64(len(arch.Agents))
	entropy := 0.0
	for _, count := range tierCounts {
		p := float64(count) / total
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	// Normalize by max possible entropy
	maxEntropy := math.Log2(float64(len(tierCounts)))
	if maxEntropy == 0 {
		return 0.5
	}

	return entropy / maxEntropy
}

// computeCapabilityCoverage measures how many capabilities are covered
func (s *TeamArchitectureSearch) computeCapabilityCoverage(arch *TeamArchitecture) float64 {
	coveredCapabilities := make(map[string]bool)

	for _, agent := range arch.Agents {
		if caps, ok := s.searchSpace.AgentCapabilities[agent]; ok {
			for _, cap := range caps {
				coveredCapabilities[cap] = true
			}
		}
	}

	// Count total unique capabilities in search space
	allCaps := make(map[string]bool)
	for _, caps := range s.searchSpace.AgentCapabilities {
		for _, cap := range caps {
			allCaps[cap] = true
		}
	}

	if len(allCaps) == 0 {
		return 1.0
	}

	return float64(len(coveredCapabilities)) / float64(len(allCaps))
}

// computeSynergyScore measures synergy between team members
func (s *TeamArchitectureSearch) computeSynergyScore(arch *TeamArchitecture) float64 {
	if s.searchSpace.SynergyMatrix == nil || len(arch.Agents) < 2 {
		return 0.5
	}

	totalSynergy := 0.0
	pairCount := 0

	for i := 0; i < len(arch.Agents); i++ {
		for j := i + 1; j < len(arch.Agents); j++ {
			agent1 := arch.Agents[i]
			agent2 := arch.Agents[j]

			if synergies, ok := s.searchSpace.SynergyMatrix[agent1]; ok {
				if synergy, ok := synergies[agent2]; ok {
					totalSynergy += synergy
					pairCount++
				}
			}
			// Check reverse direction
			if synergies, ok := s.searchSpace.SynergyMatrix[agent2]; ok {
				if synergy, ok := synergies[agent1]; ok {
					totalSynergy += synergy
					pairCount++
				}
			}
		}
	}

	if pairCount == 0 {
		return 0.5
	}

	return totalSynergy / float64(pairCount)
}

// RunGeneration runs one generation of the evolutionary search
func (s *TeamArchitectureSearch) RunGeneration(ctx context.Context, evaluator ArchitectureEvaluator) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Evaluate all architectures
	for _, arch := range s.population {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if arch.Fitness == 0 {
			if err := s.EvaluateArchitecture(ctx, arch, evaluator); err != nil {
				// Continue with other evaluations
				arch.Fitness = 0.01 // Minimal fitness for failed evaluations
			}
		}
	}

	// Sort by fitness
	sort.Slice(s.population, func(i, j int) bool {
		return s.population[i].Fitness > s.population[j].Fitness
	})

	// Update elite
	eliteCount := s.config.EliteCount
	if eliteCount > len(s.population) {
		eliteCount = len(s.population)
	}
	s.eliteArchitectures = make([]*TeamArchitecture, eliteCount)
	for i := 0; i < eliteCount; i++ {
		s.eliteArchitectures[i] = s.cloneArchitecture(s.population[i])
	}

	// Create next generation
	nextGen := make([]*TeamArchitecture, 0, s.config.PopulationSize)

	// Keep elites
	for _, elite := range s.eliteArchitectures {
		nextGen = append(nextGen, elite)
	}

	// Fill rest with crossover and mutation
	for len(nextGen) < s.config.PopulationSize {
		// Tournament selection
		parent1 := s.tournamentSelect()
		parent2 := s.tournamentSelect()

		var child *TeamArchitecture

		if rand.Float64() < s.config.CrossoverRate {
			child = s.crossover(parent1, parent2)
		} else {
			child = s.cloneArchitecture(parent1)
		}

		if rand.Float64() < s.config.MutationRate {
			child = s.mutate(child)
		}

		child.GenerationNumber = s.metrics.TotalGenerations + 1
		nextGen = append(nextGen, child)
	}

	s.population = nextGen
	s.metrics.TotalGenerations++

	// Update metrics
	if len(s.eliteArchitectures) > 0 {
		s.metrics.BestFitness = s.eliteArchitectures[0].Fitness
	}
	s.updateAverageFitness()
	s.updateDiversity()

	return nil
}

// tournamentSelect performs tournament selection
func (s *TeamArchitectureSearch) tournamentSelect() *TeamArchitecture {
	bestIdx := rand.Intn(len(s.population))
	best := s.population[bestIdx]

	for i := 1; i < s.config.TournamentSize; i++ {
		idx := rand.Intn(len(s.population))
		if s.population[idx].Fitness > best.Fitness {
			best = s.population[idx]
		}
	}

	return best
}

// crossover creates a child from two parents
func (s *TeamArchitectureSearch) crossover(parent1, parent2 *TeamArchitecture) *TeamArchitecture {
	child := &TeamArchitecture{
		ID:        generateArchitectureID(),
		Roles:     make(map[string]string),
		ParentIDs: []string{parent1.ID, parent2.ID},
		CreatedAt: time.Now(),
	}

	// Combine agents from both parents
	agentSet := make(map[string]bool)
	for _, a := range parent1.Agents {
		agentSet[a] = true
	}
	for _, a := range parent2.Agents {
		agentSet[a] = true
	}

	// Select subset
	allAgents := make([]string, 0, len(agentSet))
	for a := range agentSet {
		allAgents = append(allAgents, a)
	}
	rand.Shuffle(len(allAgents), func(i, j int) {
		allAgents[i], allAgents[j] = allAgents[j], allAgents[i]
	})

	// Size is average of parents
	targetSize := (len(parent1.Agents) + len(parent2.Agents)) / 2
	if targetSize < s.searchSpace.MinTeamSize {
		targetSize = s.searchSpace.MinTeamSize
	}
	if targetSize > len(allAgents) {
		targetSize = len(allAgents)
	}

	child.Agents = s.removeForbiddenPairs(allAgents[:targetSize])

	// Inherit structure from fitter parent
	if parent1.Fitness > parent2.Fitness {
		child.Structure = parent1.Structure
	} else {
		child.Structure = parent2.Structure
	}

	return child
}

// mutate applies random mutations to an architecture
func (s *TeamArchitectureSearch) mutate(arch *TeamArchitecture) *TeamArchitecture {
	mutated := s.cloneArchitecture(arch)
	mutated.ID = generateArchitectureID()
	mutated.Mutations = append(mutated.Mutations, arch.ID)
	mutated.Fitness = 0 // Will need re-evaluation

	mutationType := rand.Intn(4)
	switch mutationType {
	case 0:
		// Add an agent
		if len(mutated.Agents) < s.searchSpace.MaxTeamSize {
			available := s.getAvailableAgents(mutated.Agents)
			if len(available) > 0 {
				newAgent := available[rand.Intn(len(available))]
				mutated.Agents = append(mutated.Agents, newAgent)
			}
		}
	case 1:
		// Remove an agent
		if len(mutated.Agents) > s.searchSpace.MinTeamSize {
			idx := rand.Intn(len(mutated.Agents))
			mutated.Agents = append(mutated.Agents[:idx], mutated.Agents[idx+1:]...)
		}
	case 2:
		// Swap an agent
		if len(mutated.Agents) > 0 {
			idx := rand.Intn(len(mutated.Agents))
			available := s.getAvailableAgents(mutated.Agents)
			if len(available) > 0 {
				mutated.Agents[idx] = available[rand.Intn(len(available))]
			}
		}
	case 3:
		// Change structure
		structures := []TeamStructure{StructureHierarchical, StructureFlat, StructureSpecialized, StructureHybrid}
		mutated.Structure = structures[rand.Intn(len(structures))]
	}

	mutated.Agents = s.removeForbiddenPairs(mutated.Agents)
	return mutated
}

// getAvailableAgents returns agents not in the current team
func (s *TeamArchitectureSearch) getAvailableAgents(currentTeam []string) []string {
	inTeam := make(map[string]bool)
	for _, a := range currentTeam {
		inTeam[a] = true
	}

	available := make([]string, 0)
	for _, a := range s.searchSpace.AvailableAgents {
		if !inTeam[a] {
			available = append(available, a)
		}
	}
	return available
}

// cloneArchitecture creates a deep copy of an architecture
func (s *TeamArchitectureSearch) cloneArchitecture(arch *TeamArchitecture) *TeamArchitecture {
	clone := &TeamArchitecture{
		ID:               arch.ID,
		Agents:           make([]string, len(arch.Agents)),
		Roles:            make(map[string]string),
		Structure:        arch.Structure,
		Fitness:          arch.Fitness,
		ValidationScore:  arch.ValidationScore,
		GenerationNumber: arch.GenerationNumber,
		ParentIDs:        append([]string{}, arch.ParentIDs...),
		Mutations:        append([]string{}, arch.Mutations...),
		CreatedAt:        arch.CreatedAt,
	}
	copy(clone.Agents, arch.Agents)
	for k, v := range arch.Roles {
		clone.Roles[k] = v
	}
	return clone
}

// updateAverageFitness updates the average fitness metric
func (s *TeamArchitectureSearch) updateAverageFitness() {
	if len(s.population) == 0 {
		return
	}

	total := 0.0
	for _, arch := range s.population {
		total += arch.Fitness
	}
	s.metrics.AverageFitness = total / float64(len(s.population))
}

// updateDiversity computes population diversity
func (s *TeamArchitectureSearch) updateDiversity() {
	if len(s.population) < 2 {
		s.metrics.DiversityScore = 0
		return
	}

	// Compute diversity as average pairwise distance
	totalDistance := 0.0
	comparisons := 0

	for i := 0; i < len(s.population); i++ {
		for j := i + 1; j < len(s.population); j++ {
			distance := s.architectureDistance(s.population[i], s.population[j])
			totalDistance += distance
			comparisons++
		}
	}

	if comparisons > 0 {
		s.metrics.DiversityScore = totalDistance / float64(comparisons)
	}
}

// architectureDistance computes distance between two architectures
func (s *TeamArchitectureSearch) architectureDistance(a1, a2 *TeamArchitecture) float64 {
	// Jaccard distance based on agent sets
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, a := range a1.Agents {
		set1[a] = true
	}
	for _, a := range a2.Agents {
		set2[a] = true
	}

	intersection := 0
	union := 0

	for a := range set1 {
		if set2[a] {
			intersection++
		}
		union++
	}
	for a := range set2 {
		if !set1[a] {
			union++
		}
	}

	if union == 0 {
		return 0
	}

	return 1.0 - float64(intersection)/float64(union)
}

// GetBestArchitecture returns the best architecture found
func (s *TeamArchitectureSearch) GetBestArchitecture() (*TeamArchitecture, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.eliteArchitectures) == 0 {
		return nil, errors.New("no architectures evaluated yet")
	}

	return s.cloneArchitecture(s.eliteArchitectures[0]), nil
}

// GetTopArchitectures returns the top N architectures
func (s *TeamArchitectureSearch) GetTopArchitectures(n int) []*TeamArchitecture {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n > len(s.eliteArchitectures) {
		n = len(s.eliteArchitectures)
	}

	result := make([]*TeamArchitecture, n)
	for i := 0; i < n; i++ {
		result[i] = s.cloneArchitecture(s.eliteArchitectures[i])
	}
	return result
}

// GetMetrics returns search metrics
func (s *TeamArchitectureSearch) GetMetrics() *SearchMetrics {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	return &SearchMetrics{
		TotalGenerations:    s.metrics.TotalGenerations,
		TotalEvaluations:    s.metrics.TotalEvaluations,
		BestFitness:         s.metrics.BestFitness,
		AverageFitness:      s.metrics.AverageFitness,
		DiversityScore:      s.metrics.DiversityScore,
		ConvergenceProgress: s.metrics.ConvergenceProgress,
		SearchDuration:      s.metrics.SearchDuration,
	}
}

// Search runs the full evolutionary search
func (s *TeamArchitectureSearch) Search(ctx context.Context, evaluator ArchitectureEvaluator) (*TeamArchitecture, error) {
	startTime := time.Now()

	// Initialize population
	if err := s.InitializePopulation(); err != nil {
		return nil, err
	}

	lastBestFitness := 0.0
	stagnationCount := 0

	for gen := 0; gen < s.config.MaxGenerations; gen++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if err := s.RunGeneration(ctx, evaluator); err != nil {
			return nil, err
		}

		// Check convergence
		s.mu.RLock()
		currentBest := s.metrics.BestFitness
		s.mu.RUnlock()

		if math.Abs(currentBest-lastBestFitness) < s.config.ConvergenceThresh {
			stagnationCount++
			if stagnationCount > 10 {
				break // Converged
			}
		} else {
			stagnationCount = 0
		}
		lastBestFitness = currentBest
	}

	s.metrics.mu.Lock()
	s.metrics.SearchDuration = time.Since(startTime)
	s.metrics.mu.Unlock()

	return s.GetBestArchitecture()
}
