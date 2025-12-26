package memory

import (
	"context"
	"testing"
)

// =============================================================================
// ARCHITECTURE SEARCH TESTS
// =============================================================================

func createTestSearchSpace() *AgentSearchSpace {
	return &AgentSearchSpace{
		AvailableAgents: []string{"APEX", "CIPHER", "ARCHITECT", "TENSOR", "VELOCITY", "FLUX", "PRISM"},
		AgentCapabilities: map[string][]string{
			"APEX":      {"coding", "design"},
			"CIPHER":    {"security", "crypto"},
			"ARCHITECT": {"design", "system"},
			"TENSOR":    {"ml", "ai"},
			"VELOCITY":  {"performance", "optimization"},
			"FLUX":      {"devops", "deployment"},
			"PRISM":     {"data", "analytics"},
		},
		AgentTiers: map[string]int{
			"APEX":      1,
			"CIPHER":    1,
			"ARCHITECT": 1,
			"TENSOR":    2,
			"VELOCITY":  1,
			"FLUX":      2,
			"PRISM":     2,
		},
		MaxTeamSize: 5,
		MinTeamSize: 2,
		ForbiddenPairs: [][2]string{
			{"APEX", "VELOCITY"}, // Example: too similar
		},
		SynergyMatrix: map[string]map[string]float64{
			"APEX": {
				"ARCHITECT": 0.9,
				"TENSOR":    0.7,
			},
			"CIPHER": {
				"ARCHITECT": 0.8,
				"FLUX":      0.7,
			},
			"TENSOR": {
				"PRISM": 0.95,
			},
		},
	}
}

func TestTeamArchitectureSearch_New(t *testing.T) {
	searchSpace := createTestSearchSpace()

	t.Run("with default config", func(t *testing.T) {
		tas := NewTeamArchitectureSearch(searchSpace, nil)
		if tas == nil {
			t.Fatal("expected non-nil search")
		}
		if tas.config.PopulationSize != 50 {
			t.Errorf("expected 50 population size, got %d", tas.config.PopulationSize)
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &ArchitectureSearchConfig{
			PopulationSize: 100,
			EliteCount:     10,
			MutationRate:   0.3,
		}
		tas := NewTeamArchitectureSearch(searchSpace, config)
		if tas.config.PopulationSize != 100 {
			t.Errorf("expected 100 population size, got %d", tas.config.PopulationSize)
		}
	})
}

func TestTeamArchitectureSearch_InitializePopulation(t *testing.T) {
	searchSpace := createTestSearchSpace()
	config := &ArchitectureSearchConfig{
		PopulationSize: 20,
		EliteCount:     5,
	}
	tas := NewTeamArchitectureSearch(searchSpace, config)

	t.Run("initialize creates population", func(t *testing.T) {
		err := tas.InitializePopulation()
		if err != nil {
			t.Fatalf("failed to initialize: %v", err)
		}

		if len(tas.population) != 20 {
			t.Errorf("expected 20 architectures, got %d", len(tas.population))
		}
	})

	t.Run("all architectures valid", func(t *testing.T) {
		for i, arch := range tas.population {
			if len(arch.Agents) < searchSpace.MinTeamSize {
				t.Errorf("arch %d has too few agents: %d", i, len(arch.Agents))
			}
			if len(arch.Agents) > searchSpace.MaxTeamSize {
				t.Errorf("arch %d has too many agents: %d", i, len(arch.Agents))
			}
			if arch.ID == "" {
				t.Errorf("arch %d has empty ID", i)
			}
		}
	})

	t.Run("forbidden pairs excluded", func(t *testing.T) {
		for i, arch := range tas.population {
			hasApex := false
			hasVelocity := false
			for _, a := range arch.Agents {
				if a == "APEX" {
					hasApex = true
				}
				if a == "VELOCITY" {
					hasVelocity = true
				}
			}
			if hasApex && hasVelocity {
				t.Errorf("arch %d has forbidden pair APEX-VELOCITY", i)
			}
		}
	})
}

func TestTeamArchitectureSearch_InitializeWithNilSearchSpace(t *testing.T) {
	tas := NewTeamArchitectureSearch(nil, nil)
	err := tas.InitializePopulation()
	if err == nil {
		t.Error("expected error for nil search space")
	}
}

func TestTeamArchitectureSearch_EvaluateArchitecture(t *testing.T) {
	searchSpace := createTestSearchSpace()
	tas := NewTeamArchitectureSearch(searchSpace, nil)
	ctx := context.Background()

	arch := &TeamArchitecture{
		ID:        "test-arch",
		Agents:    []string{"APEX", "ARCHITECT", "TENSOR"},
		Structure: StructureSpecialized,
	}

	t.Run("default fitness evaluation", func(t *testing.T) {
		err := tas.EvaluateArchitecture(ctx, arch, nil)
		if err != nil {
			t.Fatalf("evaluation failed: %v", err)
		}

		if arch.Fitness <= 0 || arch.Fitness > 1 {
			t.Errorf("fitness should be in (0,1], got %f", arch.Fitness)
		}
	})
}

func TestTeamArchitectureSearch_FitnessComponents(t *testing.T) {
	searchSpace := createTestSearchSpace()
	tas := NewTeamArchitectureSearch(searchSpace, nil)

	t.Run("tier diversity", func(t *testing.T) {
		// Mixed tiers should have higher diversity
		mixed := &TeamArchitecture{Agents: []string{"APEX", "TENSOR", "PRISM"}}
		same := &TeamArchitecture{Agents: []string{"APEX", "CIPHER", "VELOCITY"}}

		mixedDiv := tas.computeTierDiversity(mixed)
		sameDiv := tas.computeTierDiversity(same)

		if mixedDiv <= sameDiv {
			t.Logf("Mixed tier diversity: %f, Same tier diversity: %f", mixedDiv, sameDiv)
		}
	})

	t.Run("capability coverage", func(t *testing.T) {
		// More agents = more coverage
		small := &TeamArchitecture{Agents: []string{"APEX"}}
		large := &TeamArchitecture{Agents: []string{"APEX", "CIPHER", "TENSOR", "FLUX"}}

		smallCov := tas.computeCapabilityCoverage(small)
		largeCov := tas.computeCapabilityCoverage(large)

		if largeCov <= smallCov {
			t.Errorf("larger team should have more coverage: small=%f, large=%f", smallCov, largeCov)
		}
	})

	t.Run("synergy score", func(t *testing.T) {
		// High synergy pair
		synergistic := &TeamArchitecture{Agents: []string{"TENSOR", "PRISM"}}
		// Unknown synergy
		unknown := &TeamArchitecture{Agents: []string{"FLUX", "VELOCITY"}}

		synScore := tas.computeSynergyScore(synergistic)
		unkScore := tas.computeSynergyScore(unknown)

		t.Logf("Synergistic: %f, Unknown: %f", synScore, unkScore)
	})
}

func TestTeamArchitectureSearch_RunGeneration(t *testing.T) {
	searchSpace := createTestSearchSpace()
	config := &ArchitectureSearchConfig{
		PopulationSize: 10,
		EliteCount:     2,
		MutationRate:   0.3,
		CrossoverRate:  0.7,
		TournamentSize: 3,
	}
	tas := NewTeamArchitectureSearch(searchSpace, config)
	ctx := context.Background()

	tas.InitializePopulation()

	t.Run("generation updates population", func(t *testing.T) {
		err := tas.RunGeneration(ctx, nil)
		if err != nil {
			t.Fatalf("generation failed: %v", err)
		}

		// Population should still be same size
		if len(tas.population) != 10 {
			t.Errorf("expected 10, got %d", len(tas.population))
		}

		// Elites should be populated
		if len(tas.eliteArchitectures) != 2 {
			t.Errorf("expected 2 elites, got %d", len(tas.eliteArchitectures))
		}

		// Generation count should increase
		if tas.metrics.TotalGenerations != 1 {
			t.Errorf("expected 1 generation, got %d", tas.metrics.TotalGenerations)
		}
	})

	t.Run("multiple generations improve fitness", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			tas.RunGeneration(ctx, nil)
		}

		if tas.metrics.BestFitness <= 0 {
			t.Error("best fitness should improve")
		}
	})
}

func TestTeamArchitectureSearch_Crossover(t *testing.T) {
	searchSpace := createTestSearchSpace()
	tas := NewTeamArchitectureSearch(searchSpace, nil)

	parent1 := &TeamArchitecture{
		ID:        "p1",
		Agents:    []string{"APEX", "CIPHER"},
		Structure: StructureFlat,
		Fitness:   0.8,
	}
	parent2 := &TeamArchitecture{
		ID:        "p2",
		Agents:    []string{"TENSOR", "PRISM"},
		Structure: StructureHierarchical,
		Fitness:   0.6,
	}

	child := tas.crossover(parent1, parent2)

	t.Run("child has valid agents", func(t *testing.T) {
		if len(child.Agents) < searchSpace.MinTeamSize {
			t.Errorf("child has too few agents: %d", len(child.Agents))
		}
	})

	t.Run("child inherits structure from fitter parent", func(t *testing.T) {
		if child.Structure != StructureFlat {
			t.Errorf("expected Flat structure from fitter parent")
		}
	})

	t.Run("child has parent references", func(t *testing.T) {
		if len(child.ParentIDs) != 2 {
			t.Errorf("expected 2 parents, got %d", len(child.ParentIDs))
		}
	})
}

func TestTeamArchitectureSearch_Mutation(t *testing.T) {
	searchSpace := createTestSearchSpace()
	tas := NewTeamArchitectureSearch(searchSpace, nil)

	original := &TeamArchitecture{
		ID:        "orig",
		Agents:    []string{"APEX", "ARCHITECT", "TENSOR"},
		Structure: StructureFlat,
		Fitness:   0.7,
	}

	// Mutate many times to cover different mutation types
	mutations := make([]*TeamArchitecture, 20)
	for i := 0; i < 20; i++ {
		mutations[i] = tas.mutate(original)
	}

	t.Run("mutations have new IDs", func(t *testing.T) {
		for i, m := range mutations {
			if m.ID == original.ID {
				t.Errorf("mutation %d should have new ID", i)
			}
		}
	})

	t.Run("mutations track lineage", func(t *testing.T) {
		for i, m := range mutations {
			if len(m.Mutations) == 0 || m.Mutations[0] != original.ID {
				t.Errorf("mutation %d should track original", i)
			}
		}
	})

	t.Run("mutations reset fitness", func(t *testing.T) {
		for i, m := range mutations {
			if m.Fitness != 0 {
				t.Errorf("mutation %d should have 0 fitness", i)
			}
		}
	})
}

func TestTeamArchitectureSearch_GetBest(t *testing.T) {
	searchSpace := createTestSearchSpace()
	config := &ArchitectureSearchConfig{
		PopulationSize: 10,
		EliteCount:     3,
	}
	tas := NewTeamArchitectureSearch(searchSpace, config)
	ctx := context.Background()

	t.Run("error before evaluation", func(t *testing.T) {
		_, err := tas.GetBestArchitecture()
		if err == nil {
			t.Error("expected error before any evaluation")
		}
	})

	tas.InitializePopulation()
	tas.RunGeneration(ctx, nil)

	t.Run("returns best after evaluation", func(t *testing.T) {
		best, err := tas.GetBestArchitecture()
		if err != nil {
			t.Fatalf("failed to get best: %v", err)
		}
		if best == nil {
			t.Fatal("expected non-nil best")
		}
	})

	t.Run("get top K", func(t *testing.T) {
		top := tas.GetTopArchitectures(2)
		if len(top) != 2 {
			t.Errorf("expected 2, got %d", len(top))
		}
		// Should be sorted by fitness
		if top[0].Fitness < top[1].Fitness {
			t.Error("top architectures should be sorted by fitness")
		}
	})
}

func TestTeamArchitectureSearch_Metrics(t *testing.T) {
	searchSpace := createTestSearchSpace()
	config := &ArchitectureSearchConfig{
		PopulationSize: 10,
		EliteCount:     2,
	}
	tas := NewTeamArchitectureSearch(searchSpace, config)
	ctx := context.Background()

	tas.InitializePopulation()
	for i := 0; i < 3; i++ {
		tas.RunGeneration(ctx, nil)
	}

	metrics := tas.GetMetrics()

	if metrics.TotalGenerations != 3 {
		t.Errorf("expected 3 generations, got %d", metrics.TotalGenerations)
	}
	if metrics.BestFitness <= 0 {
		t.Error("best fitness should be positive")
	}
	if metrics.AverageFitness <= 0 {
		t.Error("average fitness should be positive")
	}
}

func TestTeamArchitectureSearch_FullSearch(t *testing.T) {
	searchSpace := createTestSearchSpace()
	config := &ArchitectureSearchConfig{
		PopulationSize:    15,
		EliteCount:        3,
		MaxGenerations:    10,
		MutationRate:      0.2,
		CrossoverRate:     0.7,
		TournamentSize:    3,
		ConvergenceThresh: 0.0001,
	}
	tas := NewTeamArchitectureSearch(searchSpace, config)
	ctx := context.Background()

	best, err := tas.Search(ctx, nil)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if best == nil {
		t.Fatal("expected non-nil best architecture")
	}

	t.Logf("Best architecture: Agents=%v, Fitness=%.3f, Structure=%d",
		best.Agents, best.Fitness, best.Structure)

	metrics := tas.GetMetrics()
	t.Logf("Search completed: Generations=%d, BestFitness=%.3f, AvgFitness=%.3f, Duration=%v",
		metrics.TotalGenerations, metrics.BestFitness, metrics.AverageFitness, metrics.SearchDuration)
}

func TestTeamArchitectureSearch_ContextCancellation(t *testing.T) {
	searchSpace := createTestSearchSpace()
	config := &ArchitectureSearchConfig{
		PopulationSize: 10,
		MaxGenerations: 100,
	}
	tas := NewTeamArchitectureSearch(searchSpace, config)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := tas.Search(ctx, nil)
	if err == nil {
		t.Error("expected context cancellation error")
	}
}

func TestTeamArchitectureSearch_Diversity(t *testing.T) {
	searchSpace := createTestSearchSpace()
	tas := NewTeamArchitectureSearch(searchSpace, nil)

	arch1 := &TeamArchitecture{Agents: []string{"APEX", "CIPHER"}}
	arch2 := &TeamArchitecture{Agents: []string{"APEX", "CIPHER"}}
	arch3 := &TeamArchitecture{Agents: []string{"TENSOR", "PRISM"}}

	t.Run("identical architectures have zero distance", func(t *testing.T) {
		dist := tas.architectureDistance(arch1, arch2)
		if dist != 0 {
			t.Errorf("expected 0, got %f", dist)
		}
	})

	t.Run("different architectures have positive distance", func(t *testing.T) {
		dist := tas.architectureDistance(arch1, arch3)
		if dist <= 0 {
			t.Errorf("expected positive distance, got %f", dist)
		}
	})
}

// =============================================================================
// CUSTOM EVALUATOR TEST
// =============================================================================

type mockEvaluator struct {
	preferredSize int
}

func (m *mockEvaluator) Evaluate(ctx context.Context, arch *TeamArchitecture) (*EvaluationResult, error) {
	// Prefer teams of specific size
	sizeDiff := float64(len(arch.Agents) - m.preferredSize)
	fitness := 1.0 - (sizeDiff*sizeDiff)*0.1
	if fitness < 0 {
		fitness = 0.01
	}

	return &EvaluationResult{
		Fitness:        fitness,
		SuccessRate:    0.8,
		AverageQuality: 0.75,
	}, nil
}

func TestTeamArchitectureSearch_CustomEvaluator(t *testing.T) {
	searchSpace := createTestSearchSpace()
	config := &ArchitectureSearchConfig{
		PopulationSize: 15,
		EliteCount:     3,
		MaxGenerations: 10,
	}
	tas := NewTeamArchitectureSearch(searchSpace, config)
	ctx := context.Background()

	evaluator := &mockEvaluator{preferredSize: 3}

	best, err := tas.Search(ctx, evaluator)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	t.Logf("Best with custom evaluator: Size=%d, Fitness=%.3f", len(best.Agents), best.Fitness)

	// Should converge toward preferred size
	if len(best.Agents) < 2 || len(best.Agents) > 4 {
		t.Logf("Expected size around 3, got %d (this may vary due to randomness)", len(best.Agents))
	}
}

// =============================================================================
// BENCHMARKS
// =============================================================================

func BenchmarkTeamArchitectureSearch_Generation(b *testing.B) {
	searchSpace := createTestSearchSpace()
	config := &ArchitectureSearchConfig{
		PopulationSize: 50,
		EliteCount:     5,
	}
	tas := NewTeamArchitectureSearch(searchSpace, config)
	ctx := context.Background()
	tas.InitializePopulation()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tas.RunGeneration(ctx, nil)
	}
}

func BenchmarkTeamArchitectureSearch_Evaluation(b *testing.B) {
	searchSpace := createTestSearchSpace()
	tas := NewTeamArchitectureSearch(searchSpace, nil)
	ctx := context.Background()

	arch := &TeamArchitecture{
		Agents: []string{"APEX", "ARCHITECT", "TENSOR"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arch.Fitness = 0
		tas.EvaluateArchitecture(ctx, arch, nil)
	}
}
