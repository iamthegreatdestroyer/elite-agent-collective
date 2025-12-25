package memory

import (
	"testing"
	"time"
)

// ============================================================================
// Condition Tests
// ============================================================================

func TestCondition_MatchEquals(t *testing.T) {
	item := &WorkingMemoryItem{
		ID:          "item-1",
		ContentType: ContentTypeGoal,
		Content:     "test content",
		Source:      SourcePerception,
	}

	cond := &Condition{
		Type:      ConditionEquals,
		Attribute: "type",
		Value:     "goal",
	}

	if !cond.Match(item) {
		t.Error("Should match item type")
	}

	cond.Value = "task"
	if cond.Match(item) {
		t.Error("Should not match different type")
	}
}

func TestCondition_MatchNumeric(t *testing.T) {
	item := &WorkingMemoryItem{
		ID:         "item-1",
		Activation: 0.75,
	}

	tests := []struct {
		name     string
		cond     *Condition
		expected bool
	}{
		{
			name:     "GreaterThan matches",
			cond:     &Condition{Type: ConditionGreaterThan, Attribute: "activation", Value: 0.5},
			expected: true,
		},
		{
			name:     "GreaterThan fails",
			cond:     &Condition{Type: ConditionGreaterThan, Attribute: "activation", Value: 0.8},
			expected: false,
		},
		{
			name:     "LessThan matches",
			cond:     &Condition{Type: ConditionLessThan, Attribute: "activation", Value: 0.8},
			expected: true,
		},
		{
			name:     "InRange matches",
			cond:     &Condition{Type: ConditionInRange, Attribute: "activation", Value: 0.5, SecondValue: 0.9},
			expected: true,
		},
		{
			name:     "InRange fails",
			cond:     &Condition{Type: ConditionInRange, Attribute: "activation", Value: 0.8, SecondValue: 1.0},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cond.Match(item) != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, !tt.expected)
			}
		})
	}
}

func TestCondition_MatchExists(t *testing.T) {
	item := &WorkingMemoryItem{
		ID:          "item-1",
		ContentType: ContentTypeGoal,
	}

	existsCond := &Condition{Type: ConditionExists, Attribute: "id"}
	if !existsCond.Match(item) {
		t.Error("Should match existing attribute")
	}

	notExistsCond := &Condition{Type: ConditionNotExists, Attribute: "missing"}
	if !notExistsCond.Match(item) {
		t.Error("Should match non-existing attribute")
	}
}

func TestCondition_MatchNegated(t *testing.T) {
	item := &WorkingMemoryItem{
		ID:          "item-1",
		ContentType: ContentTypeGoal,
	}

	cond := &Condition{
		Type:      ConditionEquals,
		Attribute: "type",
		Value:     "goal",
		Negated:   true,
	}

	if cond.Match(item) {
		t.Error("Negated condition should not match")
	}

	cond.Value = "task"
	if !cond.Match(item) {
		t.Error("Negated condition should match when original doesn't")
	}
}

func TestCondition_MatchContains(t *testing.T) {
	item := &WorkingMemoryItem{
		ID:      "item-1",
		Content: "This is a test message",
	}

	cond := &Condition{
		Type:      ConditionContains,
		Attribute: "content",
		Value:     "test",
	}

	if !cond.Match(item) {
		t.Error("Should match substring")
	}

	cond.Value = "missing"
	if cond.Match(item) {
		t.Error("Should not match non-existing substring")
	}
}

func TestCondition_MatchMetadata(t *testing.T) {
	item := &WorkingMemoryItem{
		ID: "item-1",
		Metadata: map[string]interface{}{
			"priority": "high",
			"score":    95,
		},
	}

	cond := &Condition{
		Type:      ConditionEquals,
		Attribute: "priority",
		Value:     "high",
	}

	if !cond.Match(item) {
		t.Error("Should match metadata value")
	}
}

// ============================================================================
// Production Tests
// ============================================================================

func TestProduction_Utility(t *testing.T) {
	prod := &Production{
		Priority:     1.0,
		FireCount:    0,
		SuccessCount: 0,
	}

	// No fires - utility equals priority
	if prod.Utility() != 1.0 {
		t.Errorf("Expected utility 1.0 with no fires, got %f", prod.Utility())
	}

	// 50% success rate
	prod.FireCount = 10
	prod.SuccessCount = 5
	expected := 0.5 * 1.0
	if prod.Utility() != expected {
		t.Errorf("Expected utility %f, got %f", expected, prod.Utility())
	}
}

// ============================================================================
// Production System Tests
// ============================================================================

func TestNewProductionSystem(t *testing.T) {
	ps := NewProductionSystem(nil, nil, nil, nil)

	if ps == nil {
		t.Fatal("NewProductionSystem returned nil")
	}

	if ps.Count() != 0 {
		t.Errorf("Expected 0 productions, got %d", ps.Count())
	}
}

func TestProductionSystem_AddProduction(t *testing.T) {
	ps := NewProductionSystem(nil, nil, nil, nil)

	prod := &Production{
		Name:        "test-prod",
		Description: "Test production",
		Conditions:  []*Condition{{Type: ConditionExists, Attribute: "id"}},
		Actions:     []*Action{{Type: ActionLog, Message: "fired"}},
		Priority:    1.0,
	}

	err := ps.AddProduction(prod)
	if err != nil {
		t.Fatalf("AddProduction failed: %v", err)
	}

	if ps.Count() != 1 {
		t.Errorf("Expected 1 production, got %d", ps.Count())
	}

	// Production should have ID assigned
	if prod.ID == "" {
		t.Error("Production should have ID assigned")
	}

	// Should be enabled by default
	if !prod.Enabled {
		t.Error("Production should be enabled by default")
	}

	t.Logf("Added production: %s", prod.ID)
}

func TestProductionSystem_AddProductionWithTags(t *testing.T) {
	ps := NewProductionSystem(nil, nil, nil, nil)

	prod := &Production{
		Name:       "tagged-prod",
		Conditions: []*Condition{},
		Actions:    []*Action{},
		Tags:       []string{"agent", "routing"},
	}

	ps.AddProduction(prod)

	agentProds := ps.GetByTag("agent")
	if len(agentProds) != 1 {
		t.Errorf("Expected 1 agent-tagged production, got %d", len(agentProds))
	}

	routingProds := ps.GetByTag("routing")
	if len(routingProds) != 1 {
		t.Errorf("Expected 1 routing-tagged production, got %d", len(routingProds))
	}
}

func TestProductionSystem_RemoveProduction(t *testing.T) {
	ps := NewProductionSystem(nil, nil, nil, nil)

	prod := &Production{Name: "to-remove", Tags: []string{"temp"}}
	ps.AddProduction(prod)

	if ps.Count() != 1 {
		t.Fatal("Production not added")
	}

	err := ps.RemoveProduction(prod.ID)
	if err != nil {
		t.Fatalf("RemoveProduction failed: %v", err)
	}

	if ps.Count() != 0 {
		t.Errorf("Expected 0 productions after removal, got %d", ps.Count())
	}

	// Should be removed from tag index
	if len(ps.GetByTag("temp")) != 0 {
		t.Error("Production should be removed from tag index")
	}
}

func TestProductionSystem_EnableDisable(t *testing.T) {
	ps := NewProductionSystem(nil, nil, nil, nil)

	prod := &Production{Name: "toggle-prod"}
	ps.AddProduction(prod)

	// Disable
	err := ps.DisableProduction(prod.ID)
	if err != nil {
		t.Fatalf("DisableProduction failed: %v", err)
	}

	retrieved, _ := ps.GetProduction(prod.ID)
	if retrieved.Enabled {
		t.Error("Production should be disabled")
	}

	// Enable
	err = ps.EnableProduction(prod.ID)
	if err != nil {
		t.Fatalf("EnableProduction failed: %v", err)
	}

	retrieved, _ = ps.GetProduction(prod.ID)
	if !retrieved.Enabled {
		t.Error("Production should be enabled")
	}
}

func TestProductionSystem_Match(t *testing.T) {
	config := DefaultWorkingMemoryConfig()
	wm := NewCognitiveWorkingMemory(config)
	ps := NewProductionSystem(nil, wm, nil, nil)

	// Add item to working memory
	item := &WorkingMemoryItem{
		ID:          "item-1",
		ContentType: ContentTypeGoal,
		Content:     "test task",
		Activation:  0.8,
	}
	wm.Add(item)

	// Add production that should match
	matchingProd := &Production{
		Name: "matching-prod",
		Conditions: []*Condition{
			{Type: ConditionEquals, Attribute: "type", Value: "goal"},
			{Type: ConditionGreaterThan, Attribute: "activation", Value: 0.5},
		},
		Actions:  []*Action{{Type: ActionLog}},
		Priority: 1.0,
	}
	ps.AddProduction(matchingProd)

	// Add production that should NOT match
	nonMatchingProd := &Production{
		Name: "non-matching-prod",
		Conditions: []*Condition{
			{Type: ConditionEquals, Attribute: "type", Value: "experience"},
		},
		Actions:  []*Action{{Type: ActionLog}},
		Priority: 0.5,
	}
	ps.AddProduction(nonMatchingProd)

	matches := ps.Match()

	if len(matches) != 1 {
		t.Fatalf("Expected 1 match, got %d", len(matches))
	}

	if matches[0].Production.Name != "matching-prod" {
		t.Errorf("Wrong production matched: %s", matches[0].Production.Name)
	}
}

func TestProductionSystem_ResolveConflict(t *testing.T) {
	config := DefaultWorkingMemoryConfig()
	wm := NewCognitiveWorkingMemory(config)
	ps := NewProductionSystem(nil, wm, nil, nil)

	// Add item
	wm.Add(&WorkingMemoryItem{ID: "item-1", ContentType: ContentTypeGoal, Content: "task"})

	// Add multiple matching productions with different priorities
	lowPriority := &Production{
		Name:       "low-priority",
		Conditions: []*Condition{{Type: ConditionEquals, Attribute: "type", Value: "goal"}},
		Actions:    []*Action{{Type: ActionLog}},
		Priority:   0.3,
	}
	ps.AddProduction(lowPriority)

	highPriority := &Production{
		Name:       "high-priority",
		Conditions: []*Condition{{Type: ConditionEquals, Attribute: "type", Value: "goal"}},
		Actions:    []*Action{{Type: ActionLog}},
		Priority:   0.9,
	}
	ps.AddProduction(highPriority)

	ps.Match()
	selected, err := ps.ResolveConflict()

	if err != nil {
		t.Fatalf("ResolveConflict failed: %v", err)
	}

	if selected.Production.Name != "high-priority" {
		t.Errorf("Expected high-priority to win, got %s", selected.Production.Name)
	}
}

func TestProductionSystem_ResolveConflictBySpecificity(t *testing.T) {
	config := DefaultWorkingMemoryConfig()
	wm := NewCognitiveWorkingMemory(config)

	psConfig := DefaultProductionSystemConfig()
	psConfig.ConflictStrategies = []ProductionConflictStrategy{ConflictStrategySpecificity}
	ps := NewProductionSystem(psConfig, wm, nil, nil)

	// Add item
	wm.Add(&WorkingMemoryItem{ID: "item-1", ContentType: ContentTypeGoal, Content: "important", Activation: 0.9})

	// Less specific (1 condition)
	general := &Production{
		Name:       "general",
		Conditions: []*Condition{{Type: ConditionEquals, Attribute: "type", Value: "goal"}},
		Actions:    []*Action{{Type: ActionLog}},
		Priority:   1.0,
	}
	ps.AddProduction(general)

	// More specific (2 conditions)
	specific := &Production{
		Name: "specific",
		Conditions: []*Condition{
			{Type: ConditionEquals, Attribute: "type", Value: "goal"},
			{Type: ConditionGreaterThan, Attribute: "activation", Value: 0.5},
		},
		Actions:  []*Action{{Type: ActionLog}},
		Priority: 0.5, // Lower priority but more specific
	}
	ps.AddProduction(specific)

	ps.Match()
	selected, _ := ps.ResolveConflict()

	if selected.Production.Name != "specific" {
		t.Errorf("Expected specific (more conditions) to win, got %s", selected.Production.Name)
	}
}

func TestProductionSystem_Fire(t *testing.T) {
	config := DefaultWorkingMemoryConfig()
	wm := NewCognitiveWorkingMemory(config)
	ps := NewProductionSystem(nil, wm, nil, nil)

	// Add item
	wm.Add(&WorkingMemoryItem{ID: "item-1", ContentType: ContentTypeGoal, Content: "task"})

	// Add production
	prod := &Production{
		Name:       "fire-test",
		Conditions: []*Condition{{Type: ConditionEquals, Attribute: "type", Value: "goal"}},
		Actions:    []*Action{{Type: ActionLog, Message: "fired!"}},
	}
	ps.AddProduction(prod)

	matches := ps.Match()
	if len(matches) == 0 {
		t.Fatal("No matches found")
	}

	err := ps.Fire(matches[0])
	if err != nil {
		t.Fatalf("Fire failed: %v", err)
	}

	// Check production was updated
	retrieved, _ := ps.GetProduction(prod.ID)
	if retrieved.FireCount != 1 {
		t.Errorf("Expected FireCount 1, got %d", retrieved.FireCount)
	}

	if retrieved.LastFiredAt == nil {
		t.Error("LastFiredAt should be set")
	}
}

func TestProductionSystem_Refraction(t *testing.T) {
	config := DefaultWorkingMemoryConfig()
	wm := NewCognitiveWorkingMemory(config)

	psConfig := DefaultProductionSystemConfig()
	psConfig.EnableRefraction = true
	ps := NewProductionSystem(psConfig, wm, nil, nil)

	// Add item
	wm.Add(&WorkingMemoryItem{ID: "item-1", ContentType: ContentTypeGoal, Content: "task"})

	// Add production
	prod := &Production{
		Name:       "refraction-test",
		Conditions: []*Condition{{Type: ConditionEquals, Attribute: "type", Value: "goal"}},
		Actions:    []*Action{{Type: ActionLog}},
	}
	ps.AddProduction(prod)

	// First match should succeed
	matches1 := ps.Match()
	if len(matches1) != 1 {
		t.Fatal("First match should succeed")
	}

	ps.Fire(matches1[0])

	// Second match should be blocked by refraction
	matches2 := ps.Match()
	if len(matches2) != 0 {
		t.Error("Second match should be blocked by refraction")
	}
}

func TestProductionSystem_Cycle(t *testing.T) {
	config := DefaultWorkingMemoryConfig()
	wm := NewCognitiveWorkingMemory(config)

	psConfig := DefaultProductionSystemConfig()
	psConfig.EnableRefraction = false // Allow repeated firing
	ps := NewProductionSystem(psConfig, wm, nil, nil)

	// Add item
	wm.Add(&WorkingMemoryItem{ID: "item-1", ContentType: ContentTypeGoal, Content: "task"})

	// Add production
	ps.AddProduction(&Production{
		Name:       "cycle-test",
		Conditions: []*Condition{{Type: ConditionEquals, Attribute: "type", Value: "goal"}},
		Actions:    []*Action{{Type: ActionLog}},
	})

	result, err := ps.Cycle()
	if err != nil {
		t.Fatalf("Cycle failed: %v", err)
	}

	if result == nil {
		t.Error("Cycle should return result")
	}

	t.Logf("Cycle fired: %s", result.Production.Name)
}

func TestProductionSystem_Run(t *testing.T) {
	config := DefaultWorkingMemoryConfig()
	wm := NewCognitiveWorkingMemory(config)

	psConfig := DefaultProductionSystemConfig()
	psConfig.EnableRefraction = true
	ps := NewProductionSystem(psConfig, wm, nil, nil)

	// Add items
	wm.Add(&WorkingMemoryItem{ID: "item-1", ContentType: ContentTypeGoal})
	wm.Add(&WorkingMemoryItem{ID: "item-2", ContentType: ContentTypeExperience})

	// Add productions
	ps.AddProduction(&Production{
		Name:       "goal-prod",
		Conditions: []*Condition{{Type: ConditionEquals, Attribute: "type", Value: "goal"}},
		Actions:    []*Action{{Type: ActionLog}},
	})

	ps.AddProduction(&Production{
		Name:       "exp-prod",
		Conditions: []*Condition{{Type: ConditionEquals, Attribute: "type", Value: "experience"}},
		Actions:    []*Action{{Type: ActionLog}},
	})

	cycles, err := ps.Run(10)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Should run 2 cycles (one per production before refraction blocks all)
	if cycles != 2 {
		t.Errorf("Expected 2 cycles, got %d", cycles)
	}
}

func TestProductionSystem_Callbacks(t *testing.T) {
	config := DefaultWorkingMemoryConfig()
	wm := NewCognitiveWorkingMemory(config)
	ps := NewProductionSystem(nil, wm, nil, nil)

	firedCount := 0
	conflictCount := 0

	ps.OnProductionFired(func(prod *Production, result *MatchResult) {
		firedCount++
	})

	ps.OnConflict(func(matches []*MatchResult) {
		conflictCount++
	})

	// Add item
	wm.Add(&WorkingMemoryItem{ID: "item-1", ContentType: ContentTypeGoal})

	// Add multiple matching productions
	ps.AddProduction(&Production{
		Name:       "prod-1",
		Conditions: []*Condition{{Type: ConditionEquals, Attribute: "type", Value: "goal"}},
		Actions:    []*Action{{Type: ActionLog}},
	})

	ps.AddProduction(&Production{
		Name:       "prod-2",
		Conditions: []*Condition{{Type: ConditionEquals, Attribute: "type", Value: "goal"}},
		Actions:    []*Action{{Type: ActionLog}},
	})

	ps.Cycle()

	if firedCount != 1 {
		t.Errorf("Expected 1 fire callback, got %d", firedCount)
	}

	if conflictCount != 1 {
		t.Errorf("Expected 1 conflict callback, got %d", conflictCount)
	}
}

func TestProductionSystem_MarkSuccess(t *testing.T) {
	config := DefaultWorkingMemoryConfig()
	wm := NewCognitiveWorkingMemory(config)
	ps := NewProductionSystem(nil, wm, nil, nil)

	wm.Add(&WorkingMemoryItem{ID: "item-1", ContentType: ContentTypeGoal})

	ps.AddProduction(&Production{
		Name:       "success-test",
		Conditions: []*Condition{{Type: ConditionEquals, Attribute: "type", Value: "goal"}},
		Actions:    []*Action{{Type: ActionLog}},
	})

	ps.Cycle()
	ps.MarkSuccess()

	stats := ps.GetStats()
	if stats.SuccessfulFirings != 1 {
		t.Errorf("Expected 1 successful firing, got %d", stats.SuccessfulFirings)
	}
}

func TestProductionSystem_LearnChunk(t *testing.T) {
	config := DefaultWorkingMemoryConfig()
	wm := NewCognitiveWorkingMemory(config)

	psConfig := DefaultProductionSystemConfig()
	psConfig.MinChunkSequence = 2
	ps := NewProductionSystem(psConfig, wm, nil, nil)

	// Add productions with conditions and actions
	prod1 := &Production{
		Name:       "step-1",
		Conditions: []*Condition{{Type: ConditionEquals, Attribute: "type", Value: "goal"}},
		Actions:    []*Action{{Type: ActionLog, Message: "step 1"}},
	}
	ps.AddProduction(prod1)

	prod2 := &Production{
		Name:       "step-2",
		Conditions: []*Condition{{Type: ConditionEquals, Attribute: "type", Value: "experience"}},
		Actions:    []*Action{{Type: ActionEmit, Message: "complete"}},
	}
	ps.AddProduction(prod2)

	// Create firing sequence
	sequence := []*FiringRecord{
		{ProductionID: prod1.ID, FiredAt: time.Now(), Success: true},
		{ProductionID: prod2.ID, FiredAt: time.Now(), Success: true},
	}

	chunk, err := ps.LearnChunk("test-chunk", sequence)
	if err != nil {
		t.Fatalf("LearnChunk failed: %v", err)
	}

	if chunk.Source != "learned" {
		t.Errorf("Expected source 'learned', got '%s'", chunk.Source)
	}

	if len(chunk.Conditions) == 0 {
		t.Error("Chunk should have conditions from first production")
	}

	if len(chunk.Actions) == 0 {
		t.Error("Chunk should have actions from last production")
	}

	// Should have "learned" tag
	hasTag := false
	for _, tag := range chunk.Tags {
		if tag == "learned" {
			hasTag = true
			break
		}
	}
	if !hasTag {
		t.Error("Chunk should have 'learned' tag")
	}

	t.Logf("Learned chunk: %s with %d conditions and %d actions",
		chunk.ID, len(chunk.Conditions), len(chunk.Actions))
}

func TestProductionSystem_GetStats(t *testing.T) {
	config := DefaultWorkingMemoryConfig()
	wm := NewCognitiveWorkingMemory(config)
	ps := NewProductionSystem(nil, wm, nil, nil)

	wm.Add(&WorkingMemoryItem{ID: "item-1", ContentType: ContentTypeGoal})

	ps.AddProduction(&Production{
		Name:       "stats-test",
		Conditions: []*Condition{{Type: ConditionEquals, Attribute: "type", Value: "goal"}},
		Actions:    []*Action{{Type: ActionLog}},
	})

	ps.Cycle()
	ps.MarkSuccess()

	stats := ps.GetStats()

	if stats.TotalProductions != 1 {
		t.Errorf("Expected 1 production, got %d", stats.TotalProductions)
	}

	if stats.TotalFirings != 1 {
		t.Errorf("Expected 1 firing, got %d", stats.TotalFirings)
	}

	if stats.SuccessfulFirings != 1 {
		t.Errorf("Expected 1 successful firing, got %d", stats.SuccessfulFirings)
	}
}

func TestProductionSystem_Snapshot(t *testing.T) {
	ps := NewProductionSystem(nil, nil, nil, nil)

	ps.AddProduction(&Production{Name: "enabled-1"})
	ps.AddProduction(&Production{Name: "enabled-2"})

	prod3 := &Production{Name: "disabled"}
	ps.AddProduction(prod3)
	ps.DisableProduction(prod3.ID)

	snapshot := ps.Snapshot()

	if snapshot.ProductionCount != 3 {
		t.Errorf("Expected 3 productions, got %d", snapshot.ProductionCount)
	}

	if snapshot.EnabledCount != 2 {
		t.Errorf("Expected 2 enabled, got %d", snapshot.EnabledCount)
	}

	if snapshot.DisabledCount != 1 {
		t.Errorf("Expected 1 disabled, got %d", snapshot.DisabledCount)
	}
}

func TestProductionSystem_Clear(t *testing.T) {
	ps := NewProductionSystem(nil, nil, nil, nil)

	ps.AddProduction(&Production{Name: "prod-1"})
	ps.AddProduction(&Production{Name: "prod-2"})

	if ps.Count() != 2 {
		t.Fatal("Productions not added")
	}

	ps.Clear()

	if ps.Count() != 0 {
		t.Errorf("Expected 0 productions after clear, got %d", ps.Count())
	}
}

func TestConditionType_String(t *testing.T) {
	tests := []struct {
		ctype    ConditionType
		expected string
	}{
		{ConditionEquals, "EQUALS"},
		{ConditionNotEquals, "NOT_EQUALS"},
		{ConditionGreaterThan, "GREATER_THAN"},
		{ConditionLessThan, "LESS_THAN"},
		{ConditionContains, "CONTAINS"},
		{ConditionExists, "EXISTS"},
		{ConditionNotExists, "NOT_EXISTS"},
	}

	for _, tt := range tests {
		if tt.ctype.String() != tt.expected {
			t.Errorf("Expected '%s', got '%s'", tt.expected, tt.ctype.String())
		}
	}
}

func TestActionType_String(t *testing.T) {
	tests := []struct {
		atype    ActionType
		expected string
	}{
		{ActionAdd, "ADD"},
		{ActionRemove, "REMOVE"},
		{ActionModify, "MODIFY"},
		{ActionPushGoal, "PUSH_GOAL"},
		{ActionCompleteGoal, "COMPLETE_GOAL"},
		{ActionInvokeAgent, "INVOKE_AGENT"},
		{ActionEmit, "EMIT"},
		{ActionLog, "LOG"},
		{ActionHalt, "HALT"},
	}

	for _, tt := range tests {
		if tt.atype.String() != tt.expected {
			t.Errorf("Expected '%s', got '%s'", tt.expected, tt.atype.String())
		}
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkProductionSystem_Match(b *testing.B) {
	config := DefaultWorkingMemoryConfig()
	wm := NewCognitiveWorkingMemory(config)
	ps := NewProductionSystem(nil, wm, nil, nil)

	// Add items (limited by capacity)
	for i := 0; i < 7; i++ {
		wm.Add(&WorkingMemoryItem{
			ID:          "item-" + string(rune('0'+i)),
			ContentType: ContentTypeGoal,
			Activation:  float64(i) / 10,
		})
	}

	// Add 20 productions
	for i := 0; i < 20; i++ {
		ps.AddProduction(&Production{
			Name:       "prod-" + string(rune('0'+i)),
			Conditions: []*Condition{{Type: ConditionEquals, Attribute: "type", Value: "goal"}},
			Actions:    []*Action{{Type: ActionLog}},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.Match()
	}
}

func BenchmarkProductionSystem_Cycle(b *testing.B) {
	config := DefaultWorkingMemoryConfig()
	wm := NewCognitiveWorkingMemory(config)

	psConfig := DefaultProductionSystemConfig()
	psConfig.EnableRefraction = false
	ps := NewProductionSystem(psConfig, wm, nil, nil)

	wm.Add(&WorkingMemoryItem{ID: "item-1", ContentType: ContentTypeGoal})

	ps.AddProduction(&Production{
		Name:       "bench-prod",
		Conditions: []*Condition{{Type: ConditionEquals, Attribute: "type", Value: "goal"}},
		Actions:    []*Action{{Type: ActionLog}},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.Cycle()
	}
}

func BenchmarkCondition_Match(b *testing.B) {
	item := &WorkingMemoryItem{
		ContentType: ContentTypeGoal,
		Activation:  0.8,
	}

	cond := &Condition{
		Type:      ConditionGreaterThan,
		Attribute: "activation",
		Value:     0.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cond.Match(item)
	}
}
