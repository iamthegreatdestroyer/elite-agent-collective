package memory

import (
	"testing"
	"time"
)

// ============================================================================
// Cognitive Working Memory Tests
// ============================================================================

func TestNewCognitiveWorkingMemory(t *testing.T) {
	config := DefaultWorkingMemoryConfig()
	wm := NewCognitiveWorkingMemory(config)

	if wm == nil {
		t.Fatal("NewCognitiveWorkingMemory returned nil")
	}

	if wm.Capacity() != DefaultWorkingMemoryCapacity {
		t.Errorf("Expected capacity %d, got %d", DefaultWorkingMemoryCapacity, wm.Capacity())
	}

	if wm.Size() != 0 {
		t.Errorf("Expected size 0, got %d", wm.Size())
	}
}

func TestWorkingMemory_CapacityLimits(t *testing.T) {
	// Test that capacity is bounded to Miller's 7±2
	tests := []struct {
		inputCapacity    int
		expectedCapacity int
	}{
		{3, MinWorkingMemoryCapacity},  // Below min
		{5, 5},                         // At min
		{7, 7},                         // Default
		{9, 9},                         // At max
		{15, MaxWorkingMemoryCapacity}, // Above max
	}

	for _, tt := range tests {
		config := DefaultWorkingMemoryConfig()
		config.Capacity = tt.inputCapacity
		wm := NewCognitiveWorkingMemory(config)

		if wm.Capacity() != tt.expectedCapacity {
			t.Errorf("Input capacity %d: expected %d, got %d",
				tt.inputCapacity, tt.expectedCapacity, wm.Capacity())
		}
	}
}

func TestWorkingMemory_AddAndGet(t *testing.T) {
	wm := NewCognitiveWorkingMemory(DefaultWorkingMemoryConfig())

	item := &WorkingMemoryItem{
		ID:          "test-1",
		Content:     "test content",
		ContentType: ContentTypeExperience,
		Source:      SourceRetrieval,
		Salience:    0.8,
	}

	added := wm.Add(item)
	if added == nil {
		t.Fatal("Add returned nil")
	}

	if wm.Size() != 1 {
		t.Errorf("Expected size 1, got %d", wm.Size())
	}

	// Get should boost activation
	initialActivation := added.Activation
	retrieved, ok := wm.Get("test-1")
	if !ok {
		t.Fatal("Get returned false for existing item")
	}

	if retrieved.Activation <= initialActivation {
		t.Error("Get should boost activation (rehearsal effect)")
	}

	if retrieved.AccessCount != 2 { // Add + Get
		t.Errorf("Expected access count 2, got %d", retrieved.AccessCount)
	}
}

func TestWorkingMemory_CapacityEviction(t *testing.T) {
	config := DefaultWorkingMemoryConfig()
	config.Capacity = 5 // Use minimum for easier testing
	wm := NewCognitiveWorkingMemory(config)

	// Track evictions
	evictedIDs := make([]string, 0)
	wm.OnEviction(func(item *WorkingMemoryItem) {
		evictedIDs = append(evictedIDs, item.ID)
	})

	// Add items up to capacity
	for i := 0; i < 5; i++ {
		wm.Add(&WorkingMemoryItem{
			ID:         string(rune('a' + i)),
			Content:    i,
			Activation: float64(i) * 0.1, // Different activations
		})
	}

	if wm.Size() != 5 {
		t.Errorf("Expected size 5, got %d", wm.Size())
	}

	// Add one more - should evict lowest activation
	wm.Add(&WorkingMemoryItem{
		ID:         "new",
		Content:    "new item",
		Activation: 1.0,
	})

	if wm.Size() != 5 {
		t.Errorf("Expected size to remain 5 after eviction, got %d", wm.Size())
	}

	if len(evictedIDs) != 1 {
		t.Errorf("Expected 1 eviction, got %d", len(evictedIDs))
	}

	t.Logf("Evicted: %v", evictedIDs)
}

func TestWorkingMemory_Peek(t *testing.T) {
	wm := NewCognitiveWorkingMemory(DefaultWorkingMemoryConfig())

	item := &WorkingMemoryItem{ID: "peek-test", Content: "data"}
	wm.Add(item)

	initialActivation := item.Activation
	initialAccessCount := item.AccessCount

	// Peek should NOT boost activation
	peeked, ok := wm.Peek("peek-test")
	if !ok {
		t.Fatal("Peek returned false for existing item")
	}

	if peeked.Activation != initialActivation {
		t.Error("Peek should not change activation")
	}

	if peeked.AccessCount != initialAccessCount {
		t.Error("Peek should not increment access count")
	}
}

func TestWorkingMemory_Remove(t *testing.T) {
	wm := NewCognitiveWorkingMemory(DefaultWorkingMemoryConfig())

	wm.Add(&WorkingMemoryItem{ID: "remove-test", Content: "data"})

	if !wm.Contains("remove-test") {
		t.Fatal("Item should be in working memory")
	}

	removed := wm.Remove("remove-test")
	if !removed {
		t.Fatal("Remove should return true for existing item")
	}

	if wm.Contains("remove-test") {
		t.Fatal("Item should not be in working memory after removal")
	}

	// Removing non-existent item
	removed = wm.Remove("non-existent")
	if removed {
		t.Error("Remove should return false for non-existent item")
	}
}

func TestWorkingMemory_Clear(t *testing.T) {
	wm := NewCognitiveWorkingMemory(DefaultWorkingMemoryConfig())

	for i := 0; i < 5; i++ {
		wm.Add(&WorkingMemoryItem{ID: string(rune('a' + i)), Content: i})
	}

	if wm.Size() != 5 {
		t.Fatalf("Expected size 5, got %d", wm.Size())
	}

	wm.Clear()

	if wm.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", wm.Size())
	}
}

func TestWorkingMemory_ActivationDecay(t *testing.T) {
	config := DefaultWorkingMemoryConfig()
	config.DecayRate = 1.0 // Fast decay for testing
	config.ActivationThreshold = 0.5
	wm := NewCognitiveWorkingMemory(config)

	item := &WorkingMemoryItem{
		ID:             "decay-test",
		Content:        "data",
		Activation:     1.0,
		BaseActivation: 0.3,
	}
	wm.Add(item)

	initialActivation := item.Activation

	// Wait a bit and trigger decay
	time.Sleep(100 * time.Millisecond)
	wm.TriggerDecay()

	if item.Activation >= initialActivation {
		t.Error("Activation should decay over time")
	}

	t.Logf("Activation: %.4f -> %.4f", initialActivation, item.Activation)
}

func TestWorkingMemory_SpreadingActivation(t *testing.T) {
	config := DefaultWorkingMemoryConfig()
	config.SpreadingFactor = 0.5 // Strong spreading for testing
	wm := NewCognitiveWorkingMemory(config)

	// Add two associated items
	item1 := &WorkingMemoryItem{
		ID:           "item-1",
		Content:      "source",
		Activation:   1.0,
		Associations: []string{"item-2"},
	}
	item2 := &WorkingMemoryItem{
		ID:         "item-2",
		Content:    "target",
		Activation: 0.5,
	}

	wm.Add(item2) // Add target first
	initialActivation := item2.Activation

	wm.Add(item1) // Adding source should spread to target

	// item2's activation should have increased
	updated, _ := wm.Peek("item-2")
	if updated.Activation <= initialActivation {
		t.Error("Spreading activation should increase target's activation")
	}

	t.Logf("Target activation: %.4f -> %.4f", initialActivation, updated.Activation)
}

func TestWorkingMemory_Chunking(t *testing.T) {
	wm := NewCognitiveWorkingMemory(DefaultWorkingMemoryConfig())

	// Add items
	wm.Add(&WorkingMemoryItem{ID: "a", Content: "apple"})
	wm.Add(&WorkingMemoryItem{ID: "b", Content: "banana"})
	wm.Add(&WorkingMemoryItem{ID: "c", Content: "cherry"})

	// Create chunk
	chunk, err := wm.CreateChunk("fruits", "Fruit Group", []string{"a", "b", "c"}, "category:fruits")
	if err != nil {
		t.Fatalf("CreateChunk failed: %v", err)
	}

	if chunk.ID != "fruits" {
		t.Errorf("Expected chunk ID 'fruits', got '%s'", chunk.ID)
	}

	// Verify items are bound to chunk
	item, _ := wm.Peek("a")
	if item.ChunkID != "fruits" {
		t.Errorf("Item should be bound to chunk 'fruits', got '%s'", item.ChunkID)
	}

	// Get chunk items
	chunkItems := wm.GetChunkItems("fruits")
	if len(chunkItems) != 3 {
		t.Errorf("Expected 3 chunk items, got %d", len(chunkItems))
	}

	// Disband chunk
	disbanded := wm.DisbandChunk("fruits")
	if !disbanded {
		t.Error("DisbandChunk should return true")
	}

	// Verify items are unbound
	item, _ = wm.Peek("a")
	if item.ChunkID != "" {
		t.Error("Item should be unbound after chunk disbanded")
	}
}

func TestWorkingMemory_Focus(t *testing.T) {
	wm := NewCognitiveWorkingMemory(DefaultWorkingMemoryConfig())

	wm.Add(&WorkingMemoryItem{ID: "a", Content: "low", Activation: 0.5})
	wm.Add(&WorkingMemoryItem{ID: "b", Content: "high", Activation: 1.0})

	// Initially, highest activation should be focused
	focused, ok := wm.GetFocused()
	if !ok {
		t.Fatal("GetFocused should return true")
	}
	if focused.ID != "b" {
		t.Errorf("Expected focus on 'b', got '%s'", focused.ID)
	}

	// Explicitly focus on 'a'
	wm.Focus("a")

	focused, _ = wm.GetFocused()
	if focused.ID != "a" {
		t.Errorf("Expected focus on 'a' after Focus(), got '%s'", focused.ID)
	}
}

func TestWorkingMemory_GetByType(t *testing.T) {
	wm := NewCognitiveWorkingMemory(DefaultWorkingMemoryConfig())

	wm.Add(&WorkingMemoryItem{ID: "exp1", ContentType: ContentTypeExperience})
	wm.Add(&WorkingMemoryItem{ID: "exp2", ContentType: ContentTypeExperience})
	wm.Add(&WorkingMemoryItem{ID: "goal1", ContentType: ContentTypeGoal})
	wm.Add(&WorkingMemoryItem{ID: "ctx1", ContentType: ContentTypeContext})

	experiences := wm.GetByType(ContentTypeExperience)
	if len(experiences) != 2 {
		t.Errorf("Expected 2 experiences, got %d", len(experiences))
	}

	goals := wm.GetByType(ContentTypeGoal)
	if len(goals) != 1 {
		t.Errorf("Expected 1 goal, got %d", len(goals))
	}
}

func TestWorkingMemory_GetTopN(t *testing.T) {
	wm := NewCognitiveWorkingMemory(DefaultWorkingMemoryConfig())

	wm.Add(&WorkingMemoryItem{ID: "low", Activation: 0.3})
	wm.Add(&WorkingMemoryItem{ID: "medium", Activation: 0.6})
	wm.Add(&WorkingMemoryItem{ID: "high", Activation: 0.9})

	top2 := wm.GetTopN(2)
	if len(top2) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(top2))
	}

	// Verify ordering
	if top2[0].ID != "high" {
		t.Errorf("Expected 'high' first, got '%s'", top2[0].ID)
	}
	if top2[1].ID != "medium" {
		t.Errorf("Expected 'medium' second, got '%s'", top2[1].ID)
	}
}

func TestWorkingMemory_Associations(t *testing.T) {
	wm := NewCognitiveWorkingMemory(DefaultWorkingMemoryConfig())

	wm.Add(&WorkingMemoryItem{ID: "a", Content: "item a"})
	wm.Add(&WorkingMemoryItem{ID: "b", Content: "item b"})

	// Add association
	added := wm.AddAssociation("a", "b")
	if !added {
		t.Fatal("AddAssociation should return true")
	}

	// Verify bidirectional
	itemA, _ := wm.Peek("a")
	itemB, _ := wm.Peek("b")

	if len(itemA.Associations) != 1 || itemA.Associations[0] != "b" {
		t.Error("Item a should have association to b")
	}
	if len(itemB.Associations) != 1 || itemB.Associations[0] != "a" {
		t.Error("Item b should have association to a")
	}

	// Remove association
	removed := wm.RemoveAssociation("a", "b")
	if !removed {
		t.Fatal("RemoveAssociation should return true")
	}

	itemA, _ = wm.Peek("a")
	itemB, _ = wm.Peek("b")

	if len(itemA.Associations) != 0 {
		t.Error("Item a should have no associations after removal")
	}
	if len(itemB.Associations) != 0 {
		t.Error("Item b should have no associations after removal")
	}
}

func TestWorkingMemory_Rehearsal(t *testing.T) {
	wm := NewCognitiveWorkingMemory(DefaultWorkingMemoryConfig())

	item := &WorkingMemoryItem{ID: "rehearse-test", Content: "data"}
	wm.Add(item)

	// Access the same item multiple times
	for i := 0; i < 5; i++ {
		wm.Get("rehearse-test")
	}

	retrieved, _ := wm.Peek("rehearse-test")

	// Should have been accessed 6 times (1 add + 5 gets)
	if retrieved.AccessCount != 6 {
		t.Errorf("Expected access count 6, got %d", retrieved.AccessCount)
	}

	// Activation should be higher than base
	if retrieved.Activation <= DefaultBaseActivation {
		t.Error("Repeated access should increase activation above base")
	}

	t.Logf("Access count: %d, Activation: %.4f", retrieved.AccessCount, retrieved.Activation)
}

func TestWorkingMemory_Statistics(t *testing.T) {
	wm := NewCognitiveWorkingMemory(DefaultWorkingMemoryConfig())

	// Add some items
	for i := 0; i < 5; i++ {
		wm.Add(&WorkingMemoryItem{ID: string(rune('a' + i)), Content: i})
	}

	// Access some
	wm.Get("a")
	wm.Get("b")

	stats := wm.GetStats()

	if stats.TotalItemsAdded != 5 {
		t.Errorf("Expected 5 items added, got %d", stats.TotalItemsAdded)
	}

	if stats.TotalAccesses != 2 {
		t.Errorf("Expected 2 accesses, got %d", stats.TotalAccesses)
	}

	if stats.CapacityUtilization == 0 {
		t.Error("Capacity utilization should be > 0")
	}

	t.Logf("Stats: Added=%d, Evicted=%d, Accesses=%d, AvgActivation=%.4f, Utilization=%.2f",
		stats.TotalItemsAdded, stats.TotalItemsEvicted, stats.TotalAccesses,
		stats.AverageActivation, stats.CapacityUtilization)
}

func TestWorkingMemory_Snapshot(t *testing.T) {
	wm := NewCognitiveWorkingMemory(DefaultWorkingMemoryConfig())

	wm.Add(&WorkingMemoryItem{ID: "a", Content: "data a"})
	wm.Add(&WorkingMemoryItem{ID: "b", Content: "data b"})
	wm.CreateChunk("test-chunk", "Test", []string{"a", "b"}, "test")

	snapshot := wm.Snapshot()

	if snapshot.ItemCount != 2 {
		t.Errorf("Expected 2 items in snapshot, got %d", snapshot.ItemCount)
	}

	if snapshot.ChunkCount != 1 {
		t.Errorf("Expected 1 chunk in snapshot, got %d", snapshot.ChunkCount)
	}

	if snapshot.Capacity != DefaultWorkingMemoryCapacity {
		t.Errorf("Expected capacity %d, got %d", DefaultWorkingMemoryCapacity, snapshot.Capacity)
	}

	t.Logf("Snapshot: Items=%d, Chunks=%d, AvgActivation=%.4f",
		snapshot.ItemCount, snapshot.ChunkCount, snapshot.AverageActivation)
}

func TestWorkingMemory_IsFull(t *testing.T) {
	config := DefaultWorkingMemoryConfig()
	config.Capacity = 5
	wm := NewCognitiveWorkingMemory(config)

	if wm.IsFull() {
		t.Error("Empty working memory should not be full")
	}

	for i := 0; i < 5; i++ {
		wm.Add(&WorkingMemoryItem{ID: string(rune('a' + i)), Content: i})
	}

	if !wm.IsFull() {
		t.Error("Working memory at capacity should be full")
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkWorkingMemory_Add(b *testing.B) {
	wm := NewCognitiveWorkingMemory(DefaultWorkingMemoryConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wm.Add(&WorkingMemoryItem{
			ID:      string(rune(i % 100)),
			Content: i,
		})
	}
}

func BenchmarkWorkingMemory_Get(b *testing.B) {
	wm := NewCognitiveWorkingMemory(DefaultWorkingMemoryConfig())

	// Pre-populate
	for i := 0; i < 7; i++ {
		wm.Add(&WorkingMemoryItem{ID: string(rune('a' + i)), Content: i})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wm.Get(string(rune('a' + (i % 7))))
	}
}

func BenchmarkWorkingMemory_GetTopN(b *testing.B) {
	wm := NewCognitiveWorkingMemory(DefaultWorkingMemoryConfig())

	for i := 0; i < 7; i++ {
		wm.Add(&WorkingMemoryItem{
			ID:         string(rune('a' + i)),
			Content:    i,
			Activation: float64(i) * 0.1,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wm.GetTopN(3)
	}
}

func BenchmarkWorkingMemory_SpreadingActivation(b *testing.B) {
	config := DefaultWorkingMemoryConfig()
	config.SpreadingFactor = 0.3
	wm := NewCognitiveWorkingMemory(config)

	// Create items with associations
	for i := 0; i < 7; i++ {
		assocs := make([]string, 0)
		for j := 0; j < 7; j++ {
			if i != j {
				assocs = append(assocs, string(rune('a'+j)))
			}
		}
		wm.Add(&WorkingMemoryItem{
			ID:           string(rune('a' + i)),
			Content:      i,
			Associations: assocs,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wm.Get(string(rune('a' + (i % 7))))
	}
}
