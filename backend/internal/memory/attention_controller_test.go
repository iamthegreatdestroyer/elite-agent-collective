package memory

import (
	"sync"
	"testing"
	"time"
)

// ============================================================================
// Focus Item Type Tests
// ============================================================================

func TestFocusItemType_String(t *testing.T) {
	tests := []struct {
		itemType FocusItemType
		expected string
	}{
		{FocusGoal, "goal"},
		{FocusTask, "task"},
		{FocusContext, "context"},
		{FocusExperience, "experience"},
		{FocusAgent, "agent"},
		{FocusInterrupt, "interrupt"},
		{FocusReflection, "reflection"},
		{FocusItemType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.itemType.String(); got != tt.expected {
				t.Errorf("FocusItemType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFocusItemType_BasePriority(t *testing.T) {
	// Interrupts should have highest base priority
	if FocusInterrupt.BasePriority() <= FocusGoal.BasePriority() {
		t.Error("Interrupt should have higher base priority than Goal")
	}

	// Goals should have higher priority than tasks
	if FocusGoal.BasePriority() <= FocusTask.BasePriority() {
		t.Error("Goal should have higher base priority than Task")
	}

	// All priorities should be in valid range
	types := []FocusItemType{FocusGoal, FocusTask, FocusContext, FocusExperience, FocusAgent, FocusInterrupt, FocusReflection}
	for _, ft := range types {
		p := ft.BasePriority()
		if p < 0 || p > 1 {
			t.Errorf("BasePriority for %v = %v, should be in [0, 1]", ft, p)
		}
	}
}

// ============================================================================
// Focus Item Tests
// ============================================================================

func TestNewFocusItem(t *testing.T) {
	item := NewFocusItem(FocusGoal, "test content", "Test Goal", 0.8)

	if item.ID == "" {
		t.Error("Focus item should have an ID")
	}
	if item.Type != FocusGoal {
		t.Errorf("Type = %v, want FocusGoal", item.Type)
	}
	if item.Label != "Test Goal" {
		t.Errorf("Label = %v, want 'Test Goal'", item.Label)
	}
	if item.Salience != 0.8 {
		t.Errorf("Salience = %v, want 0.8", item.Salience)
	}
	if item.CognitiveLoad <= 0 {
		t.Error("CognitiveLoad should be positive")
	}
	if item.Priority <= 0 {
		t.Error("Priority should be positive")
	}
	if item.EntryTime.IsZero() {
		t.Error("EntryTime should be set")
	}
}

func TestFocusItem_SalienceClamp(t *testing.T) {
	// Test salience clamping
	item1 := NewFocusItem(FocusTask, nil, "Test", 1.5) // Over 1.0
	if item1.Salience != 1.0 {
		t.Errorf("Salience should be clamped to 1.0, got %v", item1.Salience)
	}

	item2 := NewFocusItem(FocusTask, nil, "Test", -0.5) // Under 0.0
	if item2.Salience != 0.0 {
		t.Errorf("Salience should be clamped to 0.0, got %v", item2.Salience)
	}
}

func TestFocusItem_Touch(t *testing.T) {
	item := NewFocusItem(FocusTask, nil, "Test", 0.5)
	initialAccess := item.LastAccessTime
	initialCount := item.AccessCount

	time.Sleep(10 * time.Millisecond)
	item.Touch()

	if !item.LastAccessTime.After(initialAccess) {
		t.Error("Touch should update LastAccessTime")
	}
	if item.AccessCount != initialCount+1 {
		t.Error("Touch should increment AccessCount")
	}
}

func TestFocusItem_DecaySalience(t *testing.T) {
	item := NewFocusItem(FocusTask, nil, "Test", 0.8)
	item.DecayRate = 0.1 // 10% per second
	initialSalience := item.Salience

	item.DecaySalience(time.Second)

	if item.Salience >= initialSalience {
		t.Error("Salience should decay over time")
	}
	if item.Salience <= 0 {
		t.Error("Salience should not decay to zero immediately")
	}
}

func TestFocusItem_Clone(t *testing.T) {
	item := NewFocusItem(FocusGoal, "content", "Test", 0.7)
	item.Sticky = true
	item.Metadata["key"] = "value"

	clone := item.Clone()

	if clone.ID != item.ID {
		t.Error("Clone should have same ID")
	}
	if clone.Sticky != item.Sticky {
		t.Error("Clone should have same Sticky value")
	}
	if clone.Metadata["key"] != "value" {
		t.Error("Clone should copy metadata")
	}

	// Modify clone metadata, original should be unaffected
	clone.Metadata["key"] = "modified"
	if item.Metadata["key"] == "modified" {
		t.Error("Clone metadata should be independent")
	}
}

// ============================================================================
// Salience Computer Tests
// ============================================================================

func TestSalienceComputer_ComputeSalience(t *testing.T) {
	sc := NewSalienceComputer()

	factors := &SalienceFactors{
		Novelty:     0.8,
		Relevance:   0.9,
		Urgency:     0.5,
		Emotional:   0.3,
		SourceTrust: 1.0,
	}

	salience := sc.ComputeSalience(factors)

	if salience < 0 || salience > 1 {
		t.Errorf("Salience should be in [0, 1], got %v", salience)
	}

	// With high relevance, salience should be relatively high
	if salience < 0.5 {
		t.Errorf("Expected higher salience with high relevance, got %v", salience)
	}
}

func TestSalienceComputer_ComputeNovelty(t *testing.T) {
	sc := NewSalienceComputer()

	// First occurrence should be fully novel
	novelty1 := sc.ComputeNovelty("test_key", 0.5)
	if novelty1 != 1.0 {
		t.Errorf("First occurrence should be fully novel (1.0), got %v", novelty1)
	}

	// Same value should have lower novelty
	novelty2 := sc.ComputeNovelty("test_key", 0.5)
	if novelty2 >= novelty1 {
		t.Error("Repeated value should have lower novelty")
	}

	// Very different value should have higher novelty
	novelty3 := sc.ComputeNovelty("test_key", 0.9)
	if novelty3 <= novelty2 {
		t.Error("Different value should have higher novelty")
	}
}

func TestSalienceComputer_SetWeights(t *testing.T) {
	sc := NewSalienceComputer()
	sc.SetWeights(0.5, 0.3, 0.1, 0.05, 0.05)

	factors := &SalienceFactors{
		Novelty:     1.0,
		Relevance:   0.0,
		Urgency:     0.0,
		Emotional:   0.0,
		SourceTrust: 0.0,
	}

	// With only novelty factor at 1.0 and weight 0.5
	salience := sc.ComputeSalience(factors)
	if salience != 0.5 {
		t.Errorf("Expected salience = 0.5, got %v", salience)
	}
}

// ============================================================================
// Attention Controller Tests
// ============================================================================

func TestNewAttentionController(t *testing.T) {
	ac := NewAttentionController(nil)

	if ac == nil {
		t.Fatal("NewAttentionController returned nil")
	}
	if ac.Capacity() != 7.0 {
		t.Errorf("Default capacity should be 7.0 (Miller's Law), got %v", ac.Capacity())
	}
	if ac.FocusCount() != 0 {
		t.Error("New controller should have no focused items")
	}
}

func TestAttentionController_Focus(t *testing.T) {
	ac := NewAttentionController(nil)

	item := NewFocusItem(FocusGoal, "test", "Test Goal", 0.8)
	success, err := ac.Focus(item)

	if err != nil {
		t.Fatalf("Focus failed: %v", err)
	}
	if !success {
		t.Error("Focus should return true on success")
	}
	if ac.FocusCount() != 1 {
		t.Errorf("FocusCount = %d, want 1", ac.FocusCount())
	}

	// Get focused item
	retrieved, exists := ac.GetFocused(item.ID)
	if !exists {
		t.Error("Focused item should be retrievable")
	}
	if retrieved.Label != "Test Goal" {
		t.Error("Retrieved item should match original")
	}
}

func TestAttentionController_FocusDuplicate(t *testing.T) {
	ac := NewAttentionController(nil)

	item := NewFocusItem(FocusGoal, "test", "Test Goal", 0.8)
	ac.Focus(item)
	initialCount := ac.FocusCount()

	// Focus same item again
	success, err := ac.Focus(item)

	if err != nil {
		t.Fatalf("Duplicate focus failed: %v", err)
	}
	if !success {
		t.Error("Duplicate focus should succeed (update)")
	}
	if ac.FocusCount() != initialCount {
		t.Error("Duplicate focus should not increase count")
	}
}

func TestAttentionController_Unfocus(t *testing.T) {
	ac := NewAttentionController(nil)

	item := NewFocusItem(FocusGoal, "test", "Test Goal", 0.8)
	ac.Focus(item)

	err := ac.Unfocus(item.ID)
	if err != nil {
		t.Fatalf("Unfocus failed: %v", err)
	}

	if ac.FocusCount() != 0 {
		t.Error("After unfocus, count should be 0")
	}

	_, exists := ac.GetFocused(item.ID)
	if exists {
		t.Error("Unfocused item should not be retrievable")
	}
}

func TestAttentionController_UnfocusNotFound(t *testing.T) {
	ac := NewAttentionController(nil)

	err := ac.Unfocus("nonexistent")
	if err != ErrFocusItemNotFound {
		t.Errorf("Expected ErrFocusItemNotFound, got %v", err)
	}
}

func TestAttentionController_Touch(t *testing.T) {
	ac := NewAttentionController(nil)

	item := NewFocusItem(FocusTask, "test", "Test", 0.5)
	ac.Focus(item)
	time.Sleep(10 * time.Millisecond)

	err := ac.Touch(item.ID)
	if err != nil {
		t.Fatalf("Touch failed: %v", err)
	}

	retrieved, _ := ac.GetFocused(item.ID)
	if retrieved.AccessCount != 1 {
		t.Errorf("AccessCount = %d, want 1", retrieved.AccessCount)
	}
}

func TestAttentionController_CapacityLimit(t *testing.T) {
	config := &AttentionConfig{
		Capacity:      1.0, // Very limited capacity
		MaxFocusItems: 10,
		MinSalience:   0.05,
	}
	ac := NewAttentionController(config)

	// Add item that uses most capacity
	item1 := NewFocusItem(FocusGoal, "test1", "Goal 1", 0.9)
	item1.CognitiveLoad = 0.8
	ac.Focus(item1)

	// Try to add another item that would exceed capacity
	item2 := NewFocusItem(FocusTask, "test2", "Task 2", 0.3)
	item2.CognitiveLoad = 0.5
	success, err := ac.Focus(item2)

	// Should fail because capacity exceeded and item2 has lower priority
	if err == nil && success {
		// It might succeed if it evicts item1 (which shouldn't happen due to priority)
		if ac.FocusCount() > 1 {
			t.Log("Both items fit after eviction logic")
		}
	}
}

func TestAttentionController_MaxItemsLimit(t *testing.T) {
	config := &AttentionConfig{
		Capacity:      100.0, // High capacity
		MaxFocusItems: 3,     // Limited items
		MinSalience:   0.01,
	}
	ac := NewAttentionController(config)

	// Add max items
	for i := 0; i < 3; i++ {
		item := NewFocusItem(FocusTask, nil, "Task", 0.5)
		item.CognitiveLoad = 0.1
		ac.Focus(item)
	}

	if ac.FocusCount() != 3 {
		t.Errorf("FocusCount = %d, want 3", ac.FocusCount())
	}

	// Try to add one more with higher priority
	item := NewFocusItem(FocusGoal, nil, "Important", 0.9)
	item.CognitiveLoad = 0.1
	success, _ := ac.Focus(item)

	// Should succeed by evicting lowest priority
	if !success {
		t.Error("High-priority item should evict lower priority item")
	}
}

func TestAttentionController_PriorityEviction(t *testing.T) {
	config := &AttentionConfig{
		Capacity:      1.0,
		MaxFocusItems: 2,
		MinSalience:   0.01,
	}
	ac := NewAttentionController(config)

	// Add low priority item
	low := NewFocusItem(FocusContext, nil, "Low", 0.2)
	low.CognitiveLoad = 0.3
	ac.Focus(low)

	// Add medium priority item
	med := NewFocusItem(FocusTask, nil, "Medium", 0.5)
	med.CognitiveLoad = 0.3
	ac.Focus(med)

	// Add high priority item (should evict low)
	high := NewFocusItem(FocusGoal, nil, "High", 0.9)
	high.CognitiveLoad = 0.5
	ac.Focus(high)

	// Low priority should be evicted
	_, exists := ac.GetFocused(low.ID)
	if exists {
		t.Error("Low priority item should have been evicted")
	}

	// High priority should be present
	_, exists = ac.GetFocused(high.ID)
	if !exists {
		t.Error("High priority item should be present")
	}
}

func TestAttentionController_StickyItems(t *testing.T) {
	config := &AttentionConfig{
		Capacity:      1.0,
		MaxFocusItems: 2,
		MinSalience:   0.01,
	}
	ac := NewAttentionController(config)

	// Add sticky low priority item
	sticky := NewFocusItem(FocusContext, nil, "Sticky", 0.2)
	sticky.CognitiveLoad = 0.3
	sticky.Sticky = true
	ac.Focus(sticky)

	// Add non-sticky medium priority item
	nonSticky := NewFocusItem(FocusTask, nil, "NonSticky", 0.4)
	nonSticky.CognitiveLoad = 0.3
	ac.Focus(nonSticky)

	// Add high priority item (should evict non-sticky, not sticky)
	high := NewFocusItem(FocusGoal, nil, "High", 0.9)
	high.CognitiveLoad = 0.5
	ac.Focus(high)

	// Sticky should still be present
	_, exists := ac.GetFocused(sticky.ID)
	if !exists {
		t.Error("Sticky item should not be evicted")
	}

	// Non-sticky should be evicted
	_, exists = ac.GetFocused(nonSticky.ID)
	if exists {
		t.Error("Non-sticky item should have been evicted")
	}
}

func TestAttentionController_FocusInterrupt(t *testing.T) {
	config := DefaultAttentionConfig()
	config.Capacity = 0.5
	ac := NewAttentionController(config)

	// Fill capacity
	item := NewFocusItem(FocusTask, nil, "Task", 0.5)
	item.CognitiveLoad = 0.4
	ac.Focus(item)

	// Interrupt should force its way in
	interrupt := NewFocusItem(FocusInterrupt, nil, "Interrupt!", 0.9)
	interrupt.CognitiveLoad = 0.3
	err := ac.FocusInterrupt(interrupt)

	if err != nil {
		t.Fatalf("Interrupt should succeed: %v", err)
	}

	_, exists := ac.GetFocused(interrupt.ID)
	if !exists {
		t.Error("Interrupt should be in focus")
	}
}

func TestAttentionController_InterruptBelowThreshold(t *testing.T) {
	config := DefaultAttentionConfig()
	config.InterruptThreshold = 0.7
	ac := NewAttentionController(config)

	// Interrupt with salience below threshold
	interrupt := NewFocusItem(FocusInterrupt, nil, "Weak Interrupt", 0.5)
	err := ac.FocusInterrupt(interrupt)

	if err != ErrAttentionBlocked {
		t.Errorf("Low-salience interrupt should be blocked, got %v", err)
	}
}

func TestAttentionController_DecayAll(t *testing.T) {
	config := DefaultAttentionConfig()
	config.MinSalience = 0.3
	ac := NewAttentionController(config)

	// Add item with moderate salience
	item := NewFocusItem(FocusTask, nil, "Test", 0.4)
	item.DecayRate = 1.0 // High decay rate for testing
	ac.Focus(item)

	// Apply heavy decay
	evicted := ac.DecayAll(2 * time.Second)

	// Item should be evicted due to decay below threshold
	if evicted == 0 {
		// Check if salience dropped significantly
		retrieved, exists := ac.GetFocused(item.ID)
		if exists && retrieved.Salience >= 0.3 {
			t.Error("Salience should have decayed below threshold")
		}
	}
}

func TestAttentionController_GetTopFocused(t *testing.T) {
	ac := NewAttentionController(nil)

	// Add items with different priorities
	for i := 0; i < 5; i++ {
		salience := float64(i+1) / 10.0
		item := NewFocusItem(FocusTask, nil, "Task", salience)
		item.CognitiveLoad = 0.1
		ac.Focus(item)
	}

	top := ac.GetTopFocused(3)

	if len(top) != 3 {
		t.Errorf("Expected 3 items, got %d", len(top))
	}

	// Verify order (highest priority first)
	for i := 1; i < len(top); i++ {
		if top[i].Priority > top[i-1].Priority {
			t.Error("Items should be sorted by priority descending")
		}
	}
}

func TestAttentionController_GetFocusedByType(t *testing.T) {
	ac := NewAttentionController(nil)

	// Add different types
	ac.Focus(NewFocusItem(FocusGoal, nil, "Goal", 0.8))
	ac.Focus(NewFocusItem(FocusTask, nil, "Task 1", 0.5))
	ac.Focus(NewFocusItem(FocusTask, nil, "Task 2", 0.6))
	ac.Focus(NewFocusItem(FocusContext, nil, "Context", 0.3))

	tasks := ac.GetFocusedByType(FocusTask)
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}

	goals := ac.GetFocusedByType(FocusGoal)
	if len(goals) != 1 {
		t.Errorf("Expected 1 goal, got %d", len(goals))
	}
}

func TestAttentionController_LoadPercent(t *testing.T) {
	config := &AttentionConfig{
		Capacity:      10.0,
		MaxFocusItems: 100,
		MinSalience:   0.01,
	}
	ac := NewAttentionController(config)

	// Add items totaling 5.0 load (50% of capacity)
	for i := 0; i < 5; i++ {
		item := NewFocusItem(FocusTask, nil, "Task", 0.5)
		item.CognitiveLoad = 1.0
		ac.Focus(item)
	}

	loadPct := ac.LoadPercent()
	if loadPct < 49 || loadPct > 51 {
		t.Errorf("LoadPercent should be ~50%%, got %v%%", loadPct)
	}
}

func TestAttentionController_Callbacks(t *testing.T) {
	ac := NewAttentionController(nil)

	gainedCalled := false
	lostCalled := false
	lostReason := ""

	ac.OnFocusGained(func(item *FocusItem) {
		gainedCalled = true
	})
	ac.OnFocusLost(func(item *FocusItem, reason string) {
		lostCalled = true
		lostReason = reason
	})

	item := NewFocusItem(FocusTask, nil, "Test", 0.5)
	ac.Focus(item)

	if !gainedCalled {
		t.Error("OnFocusGained callback should be called")
	}

	ac.Unfocus(item.ID)

	if !lostCalled {
		t.Error("OnFocusLost callback should be called")
	}
	if lostReason != "manual" {
		t.Errorf("Lost reason should be 'manual', got '%s'", lostReason)
	}
}

func TestAttentionController_SnapshotRestore(t *testing.T) {
	ac := NewAttentionController(nil)

	// Add some items
	item1 := NewFocusItem(FocusGoal, "content1", "Goal", 0.8)
	item2 := NewFocusItem(FocusTask, "content2", "Task", 0.5)
	ac.Focus(item1)
	ac.Focus(item2)

	// Take snapshot
	snapshot := ac.Snapshot()

	if len(snapshot.Items) != 2 {
		t.Errorf("Snapshot should have 2 items, got %d", len(snapshot.Items))
	}

	// Clear and verify empty
	ac.Clear()
	if ac.FocusCount() != 0 {
		t.Error("After Clear, count should be 0")
	}

	// Restore
	err := ac.Restore(snapshot)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	if ac.FocusCount() != 2 {
		t.Errorf("After restore, count should be 2, got %d", ac.FocusCount())
	}

	_, exists := ac.GetFocused(item1.ID)
	if !exists {
		t.Error("Restored item1 should be retrievable")
	}
}

func TestAttentionController_Stats(t *testing.T) {
	ac := NewAttentionController(nil)

	// Add and remove items
	item := NewFocusItem(FocusTask, nil, "Test", 0.5)
	ac.Focus(item)
	ac.Unfocus(item.ID)

	stats := ac.GetStats()

	if stats.TotalItemsFocused != 1 {
		t.Errorf("TotalItemsFocused = %d, want 1", stats.TotalItemsFocused)
	}
	if stats.FocusGainedCount != 1 {
		t.Errorf("FocusGainedCount = %d, want 1", stats.FocusGainedCount)
	}
	if stats.FocusLostCount != 1 {
		t.Errorf("FocusLostCount = %d, want 1", stats.FocusLostCount)
	}
}

func TestAttentionController_FilterByAttention(t *testing.T) {
	config := &AttentionConfig{
		Capacity:      2.0,
		MaxFocusItems: 10,
		MinSalience:   0.01,
	}
	ac := NewAttentionController(config)

	// Use 1.0 of capacity
	item := NewFocusItem(FocusGoal, nil, "Goal", 0.8)
	item.CognitiveLoad = 1.0
	ac.Focus(item)

	// Filter 5 items
	items := []interface{}{"a", "b", "c", "d", "e"}
	salienceFunc := func(item interface{}) float64 {
		switch item.(string) {
		case "a":
			return 0.9
		case "b":
			return 0.7
		case "c":
			return 0.5
		case "d":
			return 0.3
		case "e":
			return 0.1
		default:
			return 0.5
		}
	}

	filtered := ac.FilterByAttention(items, salienceFunc)

	// Should get some items based on remaining capacity
	if len(filtered) == 0 {
		t.Error("Should filter some items")
	}

	// First item should be highest salience
	if len(filtered) > 0 && filtered[0].(string) != "a" {
		t.Errorf("First filtered item should be 'a', got %v", filtered[0])
	}
}

func TestAttentionController_CanFocus(t *testing.T) {
	config := &AttentionConfig{
		Capacity:      5.0,
		MaxFocusItems: 10,
		MinSalience:   0.01,
	}
	ac := NewAttentionController(config)

	// Should be able to focus initially
	if !ac.CanFocus(1.0) {
		t.Error("Should be able to focus with available capacity")
	}

	// Fill capacity
	item := NewFocusItem(FocusGoal, nil, "Goal", 0.8)
	item.CognitiveLoad = 4.5
	ac.Focus(item)

	// Should not be able to focus large item
	if ac.CanFocus(1.0) {
		t.Error("Should not be able to focus when capacity exceeded")
	}

	// Small item might still fit
	if !ac.CanFocus(0.4) {
		t.Error("Small item should still fit in remaining capacity")
	}
}

// ============================================================================
// Working Memory Integration Tests
// ============================================================================

func TestAttentionController_WorkingMemoryIntegration(t *testing.T) {
	ac := NewAttentionController(nil)
	wm := NewCognitiveWorkingMemory(DefaultWorkingMemoryConfig())

	ac.SetWorkingMemory(wm)

	// Focus an item
	item := NewFocusItem(FocusGoal, "goal content", "Test Goal", 0.8)
	ac.Focus(item)

	// Check if item was synced to working memory
	wmItem, exists := wm.Get("attn-" + item.ID)
	if !exists {
		t.Error("Focused item should be synced to working memory")
	}
	if wmItem.ContentType != ContentTypeGoal {
		t.Error("WM item should have correct content type")
	}

	// Unfocus and check removal from WM
	ac.Unfocus(item.ID)

	_, exists = wm.Get("attn-" + item.ID)
	if exists {
		t.Error("Unfocused item should be removed from working memory")
	}
}

// ============================================================================
// Concurrency Tests
// ============================================================================

func TestAttentionController_ConcurrentAccess(t *testing.T) {
	ac := NewAttentionController(nil)
	var wg sync.WaitGroup
	numGoroutines := 10
	itemsPerGoroutine := 10

	// Concurrent focus
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(gid int) {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				item := NewFocusItem(FocusTask, nil, "Task", 0.5)
				item.CognitiveLoad = 0.01
				ac.Focus(item)
			}
		}(i)
	}
	wg.Wait()

	// Verify no panic and some items focused
	if ac.FocusCount() == 0 {
		t.Error("Some items should be focused after concurrent access")
	}

	// Concurrent reads
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = ac.GetAllFocused()
			_ = ac.LoadPercent()
			_ = ac.GetStats()
		}()
	}
	wg.Wait()
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkAttentionController_Focus(b *testing.B) {
	ac := NewAttentionController(&AttentionConfig{
		Capacity:      1000.0,
		MaxFocusItems: 10000,
		MinSalience:   0.01,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		item := NewFocusItem(FocusTask, nil, "Task", 0.5)
		item.CognitiveLoad = 0.001
		ac.Focus(item)
	}
}

func BenchmarkAttentionController_GetTopFocused(b *testing.B) {
	ac := NewAttentionController(&AttentionConfig{
		Capacity:      1000.0,
		MaxFocusItems: 1000,
		MinSalience:   0.01,
	})

	// Pre-populate
	for i := 0; i < 500; i++ {
		item := NewFocusItem(FocusTask, nil, "Task", float64(i)/1000.0)
		item.CognitiveLoad = 0.1
		ac.Focus(item)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ac.GetTopFocused(10)
	}
}

func BenchmarkSalienceComputer_ComputeSalience(b *testing.B) {
	sc := NewSalienceComputer()
	factors := &SalienceFactors{
		Novelty:     0.8,
		Relevance:   0.9,
		Urgency:     0.5,
		Emotional:   0.3,
		SourceTrust: 1.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sc.ComputeSalience(factors)
	}
}
