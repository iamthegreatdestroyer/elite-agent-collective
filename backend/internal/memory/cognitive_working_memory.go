// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements Cognitive Working Memory from @NEURAL's Cognitive Architecture Analysis.
//
// Working Memory is based on cognitive science research, particularly:
// - Miller's Law (7±2 capacity limit)
// - ACT-R's activation-based retrieval
// - Spreading activation for associative priming
// - Chunking for binding related items

package memory

import (
	"container/heap"
	"math"
	"sync"
	"time"
)

// ============================================================================
// Constants
// ============================================================================

const (
	// DefaultWorkingMemoryCapacity is Miller's magic number (7±2)
	DefaultWorkingMemoryCapacity = 7

	// MinWorkingMemoryCapacity is the lower bound
	MinWorkingMemoryCapacity = 5

	// MaxWorkingMemoryCapacity is the upper bound
	MaxWorkingMemoryCapacity = 9

	// DefaultActivationDecayRate controls how fast activation fades
	// Higher = faster decay
	DefaultActivationDecayRate = 0.1

	// DefaultActivationThreshold below which items are forgotten
	DefaultActivationThreshold = 0.2

	// DefaultSpreadingFactor controls activation spread strength
	DefaultSpreadingFactor = 0.3

	// DefaultBaseActivation for new items
	DefaultBaseActivation = 1.0

	// DefaultRehearsalBoost when an item is accessed
	DefaultRehearsalBoost = 0.5
)

// ============================================================================
// Working Memory Item
// ============================================================================

// WorkingMemoryItem represents a single item in working memory.
type WorkingMemoryItem struct {
	// ID uniquely identifies this item
	ID string

	// Content is the actual data stored
	Content interface{}

	// ContentType describes what kind of content this is
	ContentType WorkingMemoryContentType

	// Activation is the current activation level (0.0 to 1.0+)
	Activation float64

	// BaseActivation is the minimum activation for this item
	BaseActivation float64

	// LastAccess is when this item was last retrieved/used
	LastAccess time.Time

	// CreatedAt is when this item entered working memory
	CreatedAt time.Time

	// AccessCount is how many times this item has been accessed
	AccessCount int

	// ChunkID links this item to a chunk (if bound)
	ChunkID string

	// Source indicates where this item came from
	Source WorkingMemorySource

	// Salience is the inherent importance (0.0 to 1.0)
	Salience float64

	// Associations are IDs of related items for spreading activation
	Associations []string

	// Metadata for additional context
	Metadata map[string]interface{}

	// index is used by the heap
	index int
}

// WorkingMemoryContentType categorizes content.
type WorkingMemoryContentType string

const (
	ContentTypeExperience   WorkingMemoryContentType = "experience"
	ContentTypeGoal         WorkingMemoryContentType = "goal"
	ContentTypeContext      WorkingMemoryContentType = "context"
	ContentTypeIntermediate WorkingMemoryContentType = "intermediate"
	ContentTypeChunk        WorkingMemoryContentType = "chunk"
	ContentTypeTask         WorkingMemoryContentType = "task"
	ContentTypeAgent        WorkingMemoryContentType = "agent"
	ContentTypeGeneral      WorkingMemoryContentType = "general"
)

// WorkingMemorySource indicates the origin of an item.
type WorkingMemorySource string

const (
	SourceRetrieval   WorkingMemorySource = "retrieval"   // From long-term memory
	SourcePerception  WorkingMemorySource = "perception"  // From input
	SourceGoal        WorkingMemorySource = "goal"        // From goal stack
	SourceComputation WorkingMemorySource = "computation" // Generated during processing
	SourceAttention   WorkingMemorySource = "attention"   // From attention controller
)

// ============================================================================
// Chunk - For binding related items
// ============================================================================

// Chunk binds multiple items together as a single unit.
// Chunking allows working memory to hold more information by treating
// related items as a single cognitive unit.
type Chunk struct {
	// ID uniquely identifies this chunk
	ID string

	// Name is a human-readable label
	Name string

	// ItemIDs are the items bound in this chunk
	ItemIDs []string

	// Pattern describes the relationship
	Pattern string

	// Strength indicates how well-formed the chunk is
	Strength float64

	// CreatedAt timestamp
	CreatedAt time.Time

	// UsageCount tracks how often this chunk is used
	UsageCount int
}

// ============================================================================
// Activation Priority Queue (for capacity management)
// ============================================================================

// activationHeap implements heap.Interface for activation-based eviction.
type activationHeap []*WorkingMemoryItem

func (h activationHeap) Len() int           { return len(h) }
func (h activationHeap) Less(i, j int) bool { return h[i].Activation < h[j].Activation }
func (h activationHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *activationHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*WorkingMemoryItem)
	item.index = n
	*h = append(*h, item)
}

func (h *activationHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[0 : n-1]
	return item
}

// ============================================================================
// Cognitive Working Memory
// ============================================================================

// CognitiveWorkingMemory implements a capacity-limited, activation-based
// working memory system inspired by cognitive architectures like ACT-R.
type CognitiveWorkingMemory struct {
	mu sync.RWMutex

	// capacity is the maximum number of items (Miller's 7±2)
	capacity int

	// items stores all current working memory items by ID
	items map[string]*WorkingMemoryItem

	// activationQueue for capacity management
	activationQueue activationHeap

	// chunks stores bound item groups
	chunks map[string]*Chunk

	// itemToChunk maps item IDs to their chunk
	itemToChunk map[string]string

	// decayRate controls activation decay speed
	decayRate float64

	// activationThreshold below which items are removed
	activationThreshold float64

	// spreadingFactor controls activation spread strength
	spreadingFactor float64

	// baseActivation for new items
	baseActivation float64

	// rehearsalBoost when items are accessed
	rehearsalBoost float64

	// lastDecayTime for decay calculations
	lastDecayTime time.Time

	// focusedItem is the currently attended item (highest activation)
	focusedItem string

	// evictionCallback called when items are evicted
	evictionCallback func(*WorkingMemoryItem)

	// stats tracks working memory statistics
	stats *WorkingMemoryStats
}

// WorkingMemoryStats tracks usage statistics.
type WorkingMemoryStats struct {
	TotalItemsAdded     int64
	TotalItemsEvicted   int64
	TotalAccesses       int64
	TotalChunksFormed   int64
	AverageActivation   float64
	CapacityUtilization float64
}

// WorkingMemoryConfig configures the working memory.
type WorkingMemoryConfig struct {
	Capacity            int
	DecayRate           float64
	ActivationThreshold float64
	SpreadingFactor     float64
	BaseActivation      float64
	RehearsalBoost      float64
}

// DefaultWorkingMemoryConfig returns sensible defaults.
func DefaultWorkingMemoryConfig() WorkingMemoryConfig {
	return WorkingMemoryConfig{
		Capacity:            DefaultWorkingMemoryCapacity,
		DecayRate:           DefaultActivationDecayRate,
		ActivationThreshold: DefaultActivationThreshold,
		SpreadingFactor:     DefaultSpreadingFactor,
		BaseActivation:      DefaultBaseActivation,
		RehearsalBoost:      DefaultRehearsalBoost,
	}
}

// NewCognitiveWorkingMemory creates a new working memory instance.
func NewCognitiveWorkingMemory(config WorkingMemoryConfig) *CognitiveWorkingMemory {
	// Validate capacity to Miller's 7±2
	capacity := config.Capacity
	if capacity < MinWorkingMemoryCapacity {
		capacity = MinWorkingMemoryCapacity
	}
	if capacity > MaxWorkingMemoryCapacity {
		capacity = MaxWorkingMemoryCapacity
	}

	wm := &CognitiveWorkingMemory{
		capacity:            capacity,
		items:               make(map[string]*WorkingMemoryItem),
		activationQueue:     make(activationHeap, 0, capacity),
		chunks:              make(map[string]*Chunk),
		itemToChunk:         make(map[string]string),
		decayRate:           config.DecayRate,
		activationThreshold: config.ActivationThreshold,
		spreadingFactor:     config.SpreadingFactor,
		baseActivation:      config.BaseActivation,
		rehearsalBoost:      config.RehearsalBoost,
		lastDecayTime:       time.Now(),
		stats:               &WorkingMemoryStats{},
	}

	heap.Init(&wm.activationQueue)
	return wm
}

// ============================================================================
// Core Operations
// ============================================================================

// Add inserts an item into working memory.
// If capacity is exceeded, the lowest-activation item is evicted.
func (wm *CognitiveWorkingMemory) Add(item *WorkingMemoryItem) *WorkingMemoryItem {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Apply decay to existing items first
	wm.applyDecayLocked()

	// Check if item already exists
	if existing, ok := wm.items[item.ID]; ok {
		// Boost activation (rehearsal effect)
		existing.Activation += wm.rehearsalBoost
		existing.LastAccess = time.Now()
		existing.AccessCount++
		heap.Fix(&wm.activationQueue, existing.index)
		wm.stats.TotalAccesses++
		return existing
	}

	// Initialize item properties
	now := time.Now()
	item.CreatedAt = now
	item.LastAccess = now
	item.AccessCount = 1
	if item.Activation == 0 {
		item.Activation = wm.baseActivation
	}
	if item.BaseActivation == 0 {
		item.BaseActivation = wm.baseActivation * 0.5
	}
	if item.Metadata == nil {
		item.Metadata = make(map[string]interface{})
	}

	// Check capacity and evict if necessary
	for len(wm.items) >= wm.capacity {
		wm.evictLowestActivationLocked()
	}

	// Add item
	wm.items[item.ID] = item
	heap.Push(&wm.activationQueue, item)
	wm.stats.TotalItemsAdded++

	// Update focus
	wm.updateFocusLocked()

	// Apply spreading activation from new item
	wm.spreadActivationLocked(item)

	return item
}

// Get retrieves an item by ID, boosting its activation.
func (wm *CognitiveWorkingMemory) Get(id string) (*WorkingMemoryItem, bool) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	item, ok := wm.items[id]
	if !ok {
		return nil, false
	}

	// Rehearsal: boost activation
	item.Activation += wm.rehearsalBoost
	item.LastAccess = time.Now()
	item.AccessCount++
	heap.Fix(&wm.activationQueue, item.index)
	wm.stats.TotalAccesses++

	// Update focus
	wm.updateFocusLocked()

	// Spread activation
	wm.spreadActivationLocked(item)

	return item, true
}

// Peek retrieves an item without boosting its activation.
func (wm *CognitiveWorkingMemory) Peek(id string) (*WorkingMemoryItem, bool) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	item, ok := wm.items[id]
	return item, ok
}

// Remove explicitly removes an item from working memory.
func (wm *CognitiveWorkingMemory) Remove(id string) bool {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	item, ok := wm.items[id]
	if !ok {
		return false
	}

	// Remove from chunk if bound
	if item.ChunkID != "" {
		wm.unbindFromChunkLocked(item.ID, item.ChunkID)
	}

	// Remove from heap
	heap.Remove(&wm.activationQueue, item.index)
	delete(wm.items, id)

	// Update focus
	wm.updateFocusLocked()

	return true
}

// Contains checks if an item is in working memory.
func (wm *CognitiveWorkingMemory) Contains(id string) bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	_, ok := wm.items[id]
	return ok
}

// Size returns the current number of items.
func (wm *CognitiveWorkingMemory) Size() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return len(wm.items)
}

// Capacity returns the maximum capacity.
func (wm *CognitiveWorkingMemory) Capacity() int {
	return wm.capacity
}

// IsFull returns true if at capacity.
func (wm *CognitiveWorkingMemory) IsFull() bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return len(wm.items) >= wm.capacity
}

// Clear removes all items from working memory.
func (wm *CognitiveWorkingMemory) Clear() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.items = make(map[string]*WorkingMemoryItem)
	wm.activationQueue = make(activationHeap, 0, wm.capacity)
	wm.chunks = make(map[string]*Chunk)
	wm.itemToChunk = make(map[string]string)
	wm.focusedItem = ""
	heap.Init(&wm.activationQueue)
}

// ============================================================================
// Activation Management
// ============================================================================

// applyDecayLocked applies time-based decay to all items.
// Must be called with lock held.
func (wm *CognitiveWorkingMemory) applyDecayLocked() {
	now := time.Now()
	elapsed := now.Sub(wm.lastDecayTime).Seconds()
	wm.lastDecayTime = now

	if elapsed <= 0 {
		return
	}

	// Decay factor based on elapsed time
	decayFactor := math.Exp(-wm.decayRate * elapsed)

	toRemove := make([]string, 0)

	for id, item := range wm.items {
		// Apply decay: A(t) = A(0) * e^(-d*t)
		item.Activation = item.BaseActivation + (item.Activation-item.BaseActivation)*decayFactor

		// Mark for removal if below threshold
		if item.Activation < wm.activationThreshold {
			toRemove = append(toRemove, id)
		}
	}

	// Remove decayed items
	for _, id := range toRemove {
		item := wm.items[id]
		if wm.evictionCallback != nil {
			wm.evictionCallback(item)
		}
		if item.ChunkID != "" {
			wm.unbindFromChunkLocked(item.ID, item.ChunkID)
		}
		heap.Remove(&wm.activationQueue, item.index)
		delete(wm.items, id)
		wm.stats.TotalItemsEvicted++
	}

	// Rebuild heap after modifications
	if len(toRemove) > 0 {
		heap.Init(&wm.activationQueue)
	}
}

// spreadActivationLocked spreads activation to associated items.
// Must be called with lock held.
func (wm *CognitiveWorkingMemory) spreadActivationLocked(source *WorkingMemoryItem) {
	if len(source.Associations) == 0 {
		return
	}

	spreadAmount := source.Activation * wm.spreadingFactor / float64(len(source.Associations))

	for _, assocID := range source.Associations {
		if assoc, ok := wm.items[assocID]; ok {
			assoc.Activation += spreadAmount
			heap.Fix(&wm.activationQueue, assoc.index)
		}
	}
}

// evictLowestActivationLocked removes the lowest-activation item.
// Must be called with lock held.
func (wm *CognitiveWorkingMemory) evictLowestActivationLocked() {
	if len(wm.activationQueue) == 0 {
		return
	}

	item := heap.Pop(&wm.activationQueue).(*WorkingMemoryItem)

	// Callback before removal
	if wm.evictionCallback != nil {
		wm.evictionCallback(item)
	}

	// Remove from chunk if bound
	if item.ChunkID != "" {
		wm.unbindFromChunkLocked(item.ID, item.ChunkID)
	}

	delete(wm.items, item.ID)
	wm.stats.TotalItemsEvicted++
}

// updateFocusLocked updates the focused item to highest activation.
// Must be called with lock held.
func (wm *CognitiveWorkingMemory) updateFocusLocked() {
	if len(wm.items) == 0 {
		wm.focusedItem = ""
		return
	}

	var maxActivation float64
	var maxID string

	for id, item := range wm.items {
		if item.Activation > maxActivation {
			maxActivation = item.Activation
			maxID = id
		}
	}

	wm.focusedItem = maxID
}

// BoostActivation manually boosts an item's activation.
func (wm *CognitiveWorkingMemory) BoostActivation(id string, amount float64) bool {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	item, ok := wm.items[id]
	if !ok {
		return false
	}

	item.Activation += amount
	item.LastAccess = time.Now()
	heap.Fix(&wm.activationQueue, item.index)
	wm.updateFocusLocked()

	return true
}

// SetActivation sets an item's activation to a specific value.
func (wm *CognitiveWorkingMemory) SetActivation(id string, activation float64) bool {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	item, ok := wm.items[id]
	if !ok {
		return false
	}

	item.Activation = activation
	heap.Fix(&wm.activationQueue, item.index)
	wm.updateFocusLocked()

	return true
}

// ============================================================================
// Chunking Operations
// ============================================================================

// CreateChunk binds multiple items into a single cognitive unit.
// This allows working memory to effectively hold more information.
func (wm *CognitiveWorkingMemory) CreateChunk(id, name string, itemIDs []string, pattern string) (*Chunk, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Verify all items exist
	for _, itemID := range itemIDs {
		if _, ok := wm.items[itemID]; !ok {
			return nil, ErrExperienceNotFound
		}
	}

	chunk := &Chunk{
		ID:         id,
		Name:       name,
		ItemIDs:    itemIDs,
		Pattern:    pattern,
		Strength:   1.0,
		CreatedAt:  time.Now(),
		UsageCount: 1,
	}

	wm.chunks[id] = chunk

	// Bind items to chunk
	for _, itemID := range itemIDs {
		wm.items[itemID].ChunkID = id
		wm.itemToChunk[itemID] = id
	}

	wm.stats.TotalChunksFormed++

	return chunk, nil
}

// GetChunk retrieves a chunk by ID.
func (wm *CognitiveWorkingMemory) GetChunk(id string) (*Chunk, bool) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	chunk, ok := wm.chunks[id]
	if ok {
		chunk.UsageCount++
	}
	return chunk, ok
}

// GetChunkItems retrieves all items in a chunk.
func (wm *CognitiveWorkingMemory) GetChunkItems(chunkID string) []*WorkingMemoryItem {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	chunk, ok := wm.chunks[chunkID]
	if !ok {
		return nil
	}

	items := make([]*WorkingMemoryItem, 0, len(chunk.ItemIDs))
	for _, itemID := range chunk.ItemIDs {
		if item, ok := wm.items[itemID]; ok {
			items = append(items, item)
		}
	}

	return items
}

// DisbandChunk removes a chunk, unbinding its items.
func (wm *CognitiveWorkingMemory) DisbandChunk(id string) bool {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	chunk, ok := wm.chunks[id]
	if !ok {
		return false
	}

	// Unbind all items
	for _, itemID := range chunk.ItemIDs {
		if item, ok := wm.items[itemID]; ok {
			item.ChunkID = ""
		}
		delete(wm.itemToChunk, itemID)
	}

	delete(wm.chunks, id)
	return true
}

// unbindFromChunkLocked removes an item from its chunk.
// Must be called with lock held.
func (wm *CognitiveWorkingMemory) unbindFromChunkLocked(itemID, chunkID string) {
	chunk, ok := wm.chunks[chunkID]
	if !ok {
		return
	}

	// Remove item from chunk's list
	newItemIDs := make([]string, 0, len(chunk.ItemIDs)-1)
	for _, id := range chunk.ItemIDs {
		if id != itemID {
			newItemIDs = append(newItemIDs, id)
		}
	}
	chunk.ItemIDs = newItemIDs

	// Remove mapping
	delete(wm.itemToChunk, itemID)

	// Disband chunk if empty
	if len(chunk.ItemIDs) == 0 {
		delete(wm.chunks, chunkID)
	}
}

// ============================================================================
// Focus & Attention
// ============================================================================

// GetFocused returns the currently focused (highest activation) item.
func (wm *CognitiveWorkingMemory) GetFocused() (*WorkingMemoryItem, bool) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if wm.focusedItem == "" {
		return nil, false
	}

	return wm.items[wm.focusedItem], true
}

// Focus explicitly sets focus to an item, boosting its activation.
func (wm *CognitiveWorkingMemory) Focus(id string) bool {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	item, ok := wm.items[id]
	if !ok {
		return false
	}

	// Boost activation significantly
	item.Activation += wm.rehearsalBoost * 2
	item.LastAccess = time.Now()
	heap.Fix(&wm.activationQueue, item.index)

	wm.focusedItem = id

	// Spread activation from focused item
	wm.spreadActivationLocked(item)

	return true
}

// ============================================================================
// Query Operations
// ============================================================================

// GetByType returns all items of a specific type.
func (wm *CognitiveWorkingMemory) GetByType(contentType WorkingMemoryContentType) []*WorkingMemoryItem {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	result := make([]*WorkingMemoryItem, 0)
	for _, item := range wm.items {
		if item.ContentType == contentType {
			result = append(result, item)
		}
	}
	return result
}

// GetBySource returns all items from a specific source.
func (wm *CognitiveWorkingMemory) GetBySource(source WorkingMemorySource) []*WorkingMemoryItem {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	result := make([]*WorkingMemoryItem, 0)
	for _, item := range wm.items {
		if item.Source == source {
			result = append(result, item)
		}
	}
	return result
}

// GetTopN returns the N highest-activation items.
func (wm *CognitiveWorkingMemory) GetTopN(n int) []*WorkingMemoryItem {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if n > len(wm.items) {
		n = len(wm.items)
	}

	// Collect all items
	items := make([]*WorkingMemoryItem, 0, len(wm.items))
	for _, item := range wm.items {
		items = append(items, item)
	}

	// Sort by activation (descending)
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].Activation > items[i].Activation {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	if n > len(items) {
		return items
	}
	return items[:n]
}

// GetAll returns all items in working memory.
func (wm *CognitiveWorkingMemory) GetAll() []*WorkingMemoryItem {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	items := make([]*WorkingMemoryItem, 0, len(wm.items))
	for _, item := range wm.items {
		items = append(items, item)
	}
	return items
}

// ============================================================================
// Association Management
// ============================================================================

// AddAssociation creates a bidirectional association between items.
func (wm *CognitiveWorkingMemory) AddAssociation(id1, id2 string) bool {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	item1, ok1 := wm.items[id1]
	item2, ok2 := wm.items[id2]

	if !ok1 || !ok2 {
		return false
	}

	// Add bidirectional association
	item1.Associations = appendUnique(item1.Associations, id2)
	item2.Associations = appendUnique(item2.Associations, id1)

	return true
}

// RemoveAssociation removes a bidirectional association.
func (wm *CognitiveWorkingMemory) RemoveAssociation(id1, id2 string) bool {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	item1, ok1 := wm.items[id1]
	item2, ok2 := wm.items[id2]

	if !ok1 || !ok2 {
		return false
	}

	item1.Associations = removeString(item1.Associations, id2)
	item2.Associations = removeString(item2.Associations, id1)

	return true
}

// ============================================================================
// Callbacks
// ============================================================================

// OnEviction sets a callback for when items are evicted.
func (wm *CognitiveWorkingMemory) OnEviction(callback func(*WorkingMemoryItem)) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.evictionCallback = callback
}

// ============================================================================
// Decay Trigger
// ============================================================================

// TriggerDecay manually triggers decay processing.
// Call this periodically if not using automatic decay.
func (wm *CognitiveWorkingMemory) TriggerDecay() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.applyDecayLocked()
	wm.updateFocusLocked()
}

// ============================================================================
// Statistics
// ============================================================================

// GetStats returns working memory statistics.
func (wm *CognitiveWorkingMemory) GetStats() *WorkingMemoryStats {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	stats := *wm.stats

	// Calculate current metrics
	if len(wm.items) > 0 {
		var totalActivation float64
		for _, item := range wm.items {
			totalActivation += item.Activation
		}
		stats.AverageActivation = totalActivation / float64(len(wm.items))
	}

	stats.CapacityUtilization = float64(len(wm.items)) / float64(wm.capacity)

	return &stats
}

// ============================================================================
// Snapshot
// ============================================================================

// Snapshot returns a snapshot of the current working memory state.
type WorkingMemorySnapshot struct {
	Timestamp           time.Time
	ItemCount           int
	Capacity            int
	FocusedItem         string
	ChunkCount          int
	Items               []*WorkingMemoryItem
	AverageActivation   float64
	CapacityUtilization float64
}

// Snapshot returns current state for debugging/monitoring.
func (wm *CognitiveWorkingMemory) Snapshot() *WorkingMemorySnapshot {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	snapshot := &WorkingMemorySnapshot{
		Timestamp:           time.Now(),
		ItemCount:           len(wm.items),
		Capacity:            wm.capacity,
		FocusedItem:         wm.focusedItem,
		ChunkCount:          len(wm.chunks),
		Items:               make([]*WorkingMemoryItem, 0, len(wm.items)),
		CapacityUtilization: float64(len(wm.items)) / float64(wm.capacity),
	}

	var totalActivation float64
	for _, item := range wm.items {
		snapshot.Items = append(snapshot.Items, item)
		totalActivation += item.Activation
	}

	if len(wm.items) > 0 {
		snapshot.AverageActivation = totalActivation / float64(len(wm.items))
	}

	return snapshot
}

// ============================================================================
// Utility Functions
// ============================================================================

// appendUnique appends a string if not already present.
func appendUnique(slice []string, s string) []string {
	for _, existing := range slice {
		if existing == s {
			return slice
		}
	}
	return append(slice, s)
}

// removeString removes a string from a slice.
func removeString(slice []string, s string) []string {
	result := make([]string, 0, len(slice))
	for _, existing := range slice {
		if existing != s {
			result = append(result, existing)
		}
	}
	return result
}
