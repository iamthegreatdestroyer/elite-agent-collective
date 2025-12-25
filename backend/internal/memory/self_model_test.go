package memory

import (
	"fmt"
	"testing"
	"time"
)

func TestNewSelfModel(t *testing.T) {
	sm := NewSelfModel("APEX", 1)
	if sm == nil {
		t.Fatal("Expected non-nil self-model")
	}

	if sm.AgentID != "APEX" {
		t.Errorf("Expected AgentID 'APEX', got '%s'", sm.AgentID)
	}

	if sm.Tier != 1 {
		t.Errorf("Expected Tier 1, got %d", sm.Tier)
	}

	if sm.CapabilityModel == nil {
		t.Error("Expected non-nil capability model")
	}

	if sm.LimitationModel == nil {
		t.Error("Expected non-nil limitation model")
	}
}

func TestCapabilityModel_RegisterAndGet(t *testing.T) {
	cm := NewCapabilityModel()

	cm.Register("coding", 0.9)
	cm.Register("design", 0.8)

	cap := cm.Get("coding")
	if cap == nil {
		t.Fatal("Expected to find 'coding' capability")
	}

	if cap.Confidence != 0.9 {
		t.Errorf("Expected confidence 0.9, got %f", cap.Confidence)
	}

	// Test non-existent
	cap = cm.Get("unknown")
	if cap != nil {
		t.Error("Expected nil for unknown capability")
	}
}

func TestCapabilityModel_Reinforce(t *testing.T) {
	cm := NewCapabilityModel()

	cm.Register("coding", 0.7)

	// Reinforce with high quality
	cm.Reinforce("coding", 1.0)

	cap := cm.Get("coding")
	if cap.Confidence <= 0.7 {
		t.Errorf("Expected confidence to increase, got %f", cap.Confidence)
	}

	if cap.UsageCount != 1 {
		t.Errorf("Expected usage count 1, got %d", cap.UsageCount)
	}

	// Test auto-registration
	cm.Reinforce("new_capability", 0.8)
	newCap := cm.Get("new_capability")
	if newCap == nil {
		t.Error("Expected auto-registered capability")
	}
}

func TestCapabilityModel_Weaken(t *testing.T) {
	cm := NewCapabilityModel()

	cm.Register("coding", 0.9)

	cm.Weaken("coding", 0.8)

	cap := cm.Get("coding")
	if cap.Confidence >= 0.9 {
		t.Errorf("Expected confidence to decrease, got %f", cap.Confidence)
	}

	// Test floor
	for i := 0; i < 100; i++ {
		cm.Weaken("coding", 0.5)
	}

	cap = cm.Get("coding")
	if cap.Confidence < 0.1 {
		t.Errorf("Expected confidence floor at 0.1, got %f", cap.Confidence)
	}
}

func TestCapabilityModel_TopCapabilities(t *testing.T) {
	cm := NewCapabilityModel()

	cm.Register("low", 0.3)
	cm.Register("medium", 0.6)
	cm.Register("high", 0.9)

	top := cm.TopCapabilities(2)
	if len(top) != 2 {
		t.Fatalf("Expected 2 capabilities, got %d", len(top))
	}

	if top[0].Name != "high" {
		t.Errorf("Expected first to be 'high', got '%s'", top[0].Name)
	}
}

func TestLimitationModel_RegisterAndGet(t *testing.T) {
	lm := NewLimitationModel()

	lm.Register("math", "Struggles with complex math", "Use AXIOM agent", 0.7)

	lim := lm.Get("math")
	if lim == nil {
		t.Fatal("Expected to find 'math' limitation")
	}

	if lim.Severity != 0.7 {
		t.Errorf("Expected severity 0.7, got %f", lim.Severity)
	}

	if lim.Workaround != "Use AXIOM agent" {
		t.Errorf("Unexpected workaround: %s", lim.Workaround)
	}
}

func TestLimitationModel_RecordFailure(t *testing.T) {
	lm := NewLimitationModel()

	lm.Register("math", "Math issues", "", 0.5)

	lm.RecordFailure("math")
	lm.RecordFailure("math")

	lim := lm.Get("math")
	if lim.FailureCount != 2 {
		t.Errorf("Expected failure count 2, got %d", lim.FailureCount)
	}

	if lim.LastFailure.IsZero() {
		t.Error("Expected LastFailure to be set")
	}
}

func TestLimitationModel_RegisterPotentialLimitation(t *testing.T) {
	lm := NewLimitationModel()

	lm.RegisterPotentialLimitation("quantum_tasks")

	lim := lm.Get("quantum_tasks_limitation")
	if lim == nil {
		t.Fatal("Expected potential limitation to be registered")
	}

	if lim.Severity != 0.3 {
		t.Errorf("Expected initial severity 0.3, got %f", lim.Severity)
	}

	// Repeated failures should increase severity
	lm.RegisterPotentialLimitation("quantum_tasks")
	lm.RegisterPotentialLimitation("quantum_tasks")

	lim = lm.Get("quantum_tasks_limitation")
	if lim.Severity <= 0.3 {
		t.Error("Expected severity to increase with failures")
	}
}

func TestUncertaintyModel_EstimateUncertainty(t *testing.T) {
	um := NewUncertaintyModel()

	// Unknown domain should have higher uncertainty
	unc := um.EstimateUncertainty("unknown_domain")
	if unc <= 0.2 {
		t.Errorf("Expected higher uncertainty for unknown domain, got %f", unc)
	}

	// Set known domain
	um.SetDomainUncertainty("coding", 0.1)
	unc = um.EstimateUncertainty("coding")
	if unc >= 0.2 {
		t.Errorf("Expected lower uncertainty for known domain, got %f", unc)
	}
}

func TestUncertaintyModel_UpdateCalibration(t *testing.T) {
	um := NewUncertaintyModel()

	initialFactor := um.GetCalibrationFactor()

	// Simulate calibration errors
	um.UpdateCalibration(0.2)
	um.UpdateCalibration(0.3)

	newFactor := um.GetCalibrationFactor()
	if newFactor == initialFactor {
		t.Error("Expected calibration factor to change")
	}
}

func TestPerformanceTracker_Add(t *testing.T) {
	pt := NewPerformanceTracker(100)

	record := &PerformanceRecord{
		TaskType:   "coding",
		Timestamp:  time.Now(),
		Success:    true,
		Quality:    0.9,
		Duration:   100 * time.Millisecond,
		Confidence: 0.8,
		Actual:     0.9,
	}

	pt.Add(record)

	stats := pt.GetStats()
	if stats.TotalTasks != 1 {
		t.Errorf("Expected 1 task, got %d", stats.TotalTasks)
	}
}

func TestPerformanceTracker_MaxRecords(t *testing.T) {
	pt := NewPerformanceTracker(5)

	// Add more than max
	for i := 0; i < 10; i++ {
		pt.Add(&PerformanceRecord{
			TaskType:  "test",
			Timestamp: time.Now(),
			Success:   true,
		})
	}

	stats := pt.GetStats()
	if stats.TotalTasks != 5 {
		t.Errorf("Expected 5 tasks (max), got %d", stats.TotalTasks)
	}
}

func TestPerformanceTracker_GetStats(t *testing.T) {
	pt := NewPerformanceTracker(100)

	// Add mixed results
	for i := 0; i < 10; i++ {
		pt.Add(&PerformanceRecord{
			TaskType:   "test",
			Timestamp:  time.Now(),
			Success:    i < 7, // 7 successes
			Quality:    float64(i) / 10.0,
			Duration:   time.Duration(i) * time.Millisecond,
			Confidence: 0.8,
			Actual:     float64(i) / 10.0,
		})
	}

	stats := pt.GetStats()
	if stats.TotalTasks != 10 {
		t.Errorf("Expected 10 tasks, got %d", stats.TotalTasks)
	}

	if stats.SuccessfulTasks != 7 {
		t.Errorf("Expected 7 successes, got %d", stats.SuccessfulTasks)
	}

	if stats.SuccessRate != 0.7 {
		t.Errorf("Expected success rate 0.7, got %f", stats.SuccessRate)
	}
}

func TestPerformanceTracker_GetRecentRecords(t *testing.T) {
	pt := NewPerformanceTracker(100)

	for i := 0; i < 10; i++ {
		pt.Add(&PerformanceRecord{
			TaskType:  "test",
			Quality:   float64(i),
			Timestamp: time.Now(),
		})
	}

	recent := pt.GetRecentRecords(3)
	if len(recent) != 3 {
		t.Errorf("Expected 3 records, got %d", len(recent))
	}

	// Should be the last 3 (quality 7, 8, 9)
	if recent[0].Quality != 7.0 {
		t.Errorf("Expected first recent to have quality 7, got %f", recent[0].Quality)
	}
}

func TestSelfModel_CanHandle(t *testing.T) {
	sm := NewSelfModel("APEX", 1)

	// Register some capabilities
	sm.CapabilityModel.Register("coding", 0.9)
	sm.CapabilityModel.Register("design", 0.7)

	// Task that should be handleable
	task := &Task{
		ID:   "task-1",
		Name: "coding",
	}

	eval := sm.CanHandle(task)
	if !eval.CanHandle {
		t.Error("Expected to be able to handle coding task")
	}

	if eval.Confidence != 0.9 {
		t.Errorf("Expected confidence 0.9, got %f", eval.Confidence)
	}
}

func TestSelfModel_CanHandleWithLimitation(t *testing.T) {
	sm := NewSelfModel("APEX", 1)

	// Register capability and limitation
	sm.CapabilityModel.Register("quantum", 0.5)
	sm.LimitationModel.Register("quantum", "Struggles with quantum tasks", "Use QUANTUM agent", 0.8)

	task := &Task{
		ID:   "task-1",
		Name: "quantum",
	}

	eval := sm.CanHandle(task)
	if eval.CanHandle {
		t.Error("Expected not to handle task with high severity limitation")
	}

	if len(eval.Limitations) == 0 {
		t.Error("Expected limitations to be reported")
	}
}

func TestSelfModel_RecordPerformance(t *testing.T) {
	sm := NewSelfModel("APEX", 1)

	// Register a capability
	sm.CapabilityModel.Register("coding", 0.7)

	// Record successful performance
	record := &PerformanceRecord{
		TaskType:   "coding",
		Timestamp:  time.Now(),
		Success:    true,
		Quality:    0.95,
		Duration:   100 * time.Millisecond,
		Confidence: 0.8,
		Actual:     0.95,
	}

	sm.RecordPerformance(record)

	// Capability should be reinforced
	cap := sm.CapabilityModel.Get("coding")
	if cap.Confidence <= 0.7 {
		t.Error("Expected capability to be reinforced after success")
	}

	// Stats should be updated
	stats := sm.GetPerformanceStats()
	if stats.TotalTasks != 1 {
		t.Errorf("Expected 1 task in stats, got %d", stats.TotalTasks)
	}
}

func TestSelfModel_OverconfidentFailure(t *testing.T) {
	sm := NewSelfModel("APEX", 1)

	// Record overconfident failure
	record := &PerformanceRecord{
		TaskType:   "new_task",
		Timestamp:  time.Now(),
		Success:    false,
		Quality:    0.2,
		Confidence: 0.9, // High confidence
		Actual:     0.2, // Low actual
	}

	sm.RecordPerformance(record)

	// Should register potential limitation
	lim := sm.LimitationModel.Get("new_task_limitation")
	if lim == nil {
		t.Error("Expected potential limitation to be registered for overconfident failure")
	}
}

func TestSelfModelRegistry(t *testing.T) {
	registry := NewSelfModelRegistry()

	// Register some models
	sm1 := NewSelfModel("APEX", 1)
	sm2 := NewSelfModel("CIPHER", 1)

	registry.Register(sm1)
	registry.Register(sm2)

	// Get
	retrieved := registry.Get("APEX")
	if retrieved == nil {
		t.Error("Expected to find APEX")
	}

	if retrieved.AgentID != "APEX" {
		t.Errorf("Expected APEX, got %s", retrieved.AgentID)
	}

	// GetOrCreate
	sm3 := registry.GetOrCreate("ARCHITECT", 1)
	if sm3 == nil {
		t.Error("Expected new model to be created")
	}

	// Should return same model on second call
	sm3b := registry.GetOrCreate("ARCHITECT", 1)
	if sm3 != sm3b {
		t.Error("Expected same model on second GetOrCreate")
	}

	// GetAllAgents
	agents := registry.GetAllAgents()
	if len(agents) != 3 {
		t.Errorf("Expected 3 agents, got %d", len(agents))
	}
}

func TestTaskEvaluation_Fields(t *testing.T) {
	eval := &TaskEvaluation{
		CanHandle:         true,
		Confidence:        0.85,
		Uncertainty:       0.1,
		WeakestCapability: "design",
		Limitations:       []string{"time_constraint"},
		Recommendation:    "Proceed with caution",
	}

	if !eval.CanHandle {
		t.Error("Expected CanHandle to be true")
	}

	if eval.Confidence != 0.85 {
		t.Errorf("Expected confidence 0.85, got %f", eval.Confidence)
	}
}

// Benchmarks

func BenchmarkSelfModel_CanHandle(b *testing.B) {
	sm := NewSelfModel("APEX", 1)

	// Register capabilities
	for i := 0; i < 20; i++ {
		sm.CapabilityModel.Register(fmt.Sprintf("cap_%d", i), 0.8)
	}

	task := &Task{ID: "task-1", Name: "cap_10"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.CanHandle(task)
	}
}

func BenchmarkCapabilityModel_Reinforce(b *testing.B) {
	cm := NewCapabilityModel()
	cm.Register("test", 0.5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.Reinforce("test", 0.9)
	}
}

func BenchmarkPerformanceTracker_Add(b *testing.B) {
	pt := NewPerformanceTracker(10000)

	record := &PerformanceRecord{
		TaskType:   "test",
		Timestamp:  time.Now(),
		Success:    true,
		Quality:    0.9,
		Duration:   100 * time.Millisecond,
		Confidence: 0.8,
		Actual:     0.9,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pt.Add(record)
	}
}

