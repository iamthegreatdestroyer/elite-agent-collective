// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements the Attention Controller from @NEURAL's Cognitive Architecture Analysis.
//
// Attention Mechanism (Critical Cognitive Function):
// - Models attention as a capacity-constrained resource
// - Forces prioritization through salience computation
// - Implements focus stack with priority-based management
// - Connects to working memory to modulate what enters/persists
// - Enables bounded rationality through cognitive load limits
//
// Human cognition is fundamentally constrained by attention. This constraint
// forces prioritization, which is essential for general intelligence.

package memory

import (
	"container/heap"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// Error Definitions
// ============================================================================

var (
	ErrAttentionCapacityExceeded = errors.New("attention capacity exceeded")
	ErrFocusItemNotFound         = errors.New("focus item not found")
	ErrInvalidSalience           = errors.New("invalid salience value")
	ErrAttentionBlocked          = errors.New("attention blocked by higher priority item")
)

// ============================================================================
// Focus Item Types
// ============================================================================

// FocusItemType classifies what is being attended to.
type FocusItemType int

const (
	// FocusGoal represents an active goal being pursued
	FocusGoal FocusItemType = iota

	// FocusTask represents a specific task or subtask
	FocusTask

	// FocusContext represents contextual information
	FocusContext

	// FocusExperience represents a retrieved memory/experience
	FocusExperience

	// FocusAgent represents attention to a specific agent's output
	FocusAgent

	// FocusInterrupt represents an urgent interrupt requiring attention
	FocusInterrupt

	// FocusReflection represents self-reflective attention
	FocusReflection
)

// String returns human-readable focus item type.
func (t FocusItemType) String() string {
	switch t {
	case FocusGoal:
		return "goal"
	case FocusTask:
		return "task"
	case FocusContext:
		return "context"
	case FocusExperience:
		return "experience"
	case FocusAgent:
		return "agent"
	case FocusInterrupt:
		return "interrupt"
	case FocusReflection:
		return "reflection"
	default:
		return "unknown"
	}
}

// BasePriority returns the base priority for this type (interrupts highest).
func (t FocusItemType) BasePriority() float64 {
	switch t {
	case FocusInterrupt:
		return 1.0
	case FocusGoal:
		return 0.8
	case FocusTask:
		return 0.7
	case FocusAgent:
		return 0.6
	case FocusExperience:
		return 0.5
	case FocusContext:
		return 0.4
	case FocusReflection:
		return 0.3
	default:
		return 0.1
	}
}

// ============================================================================
// Focus Item
// ============================================================================

// focusItemIDCounter provides unique IDs
var focusItemIDCounter uint64

// FocusItem represents something being attended to.
type FocusItem struct {
	// ID uniquely identifies this focus item
	ID string

	// Type classifies the focus item
	Type FocusItemType

	// Content is the actual content being attended to
	Content interface{}

	// Label is a human-readable description
	Label string

	// Salience measures how attention-grabbing this item is (0.0 to 1.0)
	Salience float64

	// Priority is the computed priority (salience * type weight * recency)
	Priority float64

	// CognitiveLoad is how much attention capacity this item requires
	CognitiveLoad float64

	// EntryTime is when this item entered the focus
	EntryTime time.Time

	// LastAccessTime is when this item was last accessed
	LastAccessTime time.Time

	// AccessCount tracks how many times this item has been accessed
	AccessCount int

	// DecayRate controls how quickly salience decays (per second)
	DecayRate float64

	// Sticky items resist being evicted
	Sticky bool

	// SourceID links to the original source (goal ID, experience ID, etc.)
	SourceID string

	// Metadata holds additional properties
	Metadata map[string]interface{}

	// index is used by the priority queue
	index int
}

// NewFocusItem creates a new focus item.
func NewFocusItem(itemType FocusItemType, content interface{}, label string, salience float64) *FocusItem {
	now := time.Now()
	item := &FocusItem{
		ID:             fmt.Sprintf("focus-%d", atomic.AddUint64(&focusItemIDCounter, 1)),
		Type:           itemType,
		Content:        content,
		Label:          label,
		Salience:       clampFloat(salience, 0.0, 1.0),
		CognitiveLoad:  computeDefaultLoad(itemType, salience),
		EntryTime:      now,
		LastAccessTime: now,
		DecayRate:      0.01, // Default: 1% per second
		Metadata:       make(map[string]interface{}),
	}
	item.Priority = item.computePriority()
	return item
}

// computePriority calculates the current priority.
func (f *FocusItem) computePriority() float64 {
	// Priority = salience * type_weight * recency_bonus
	typeWeight := f.Type.BasePriority()
	recency := f.recencyBonus()
	stickyBonus := 1.0
	if f.Sticky {
		stickyBonus = 1.5
	}
	return f.Salience * typeWeight * recency * stickyBonus
}

// recencyBonus returns a bonus based on how recently the item was accessed.
func (f *FocusItem) recencyBonus() float64 {
	elapsed := time.Since(f.LastAccessTime).Seconds()
	// Exponential decay of recency bonus
	return math.Exp(-elapsed / 60.0) // 1-minute half-life
}

// DecaySalience applies time-based decay to the salience.
func (f *FocusItem) DecaySalience(elapsed time.Duration) {
	decayFactor := math.Exp(-f.DecayRate * elapsed.Seconds())
	f.Salience = clampFloat(f.Salience*decayFactor, 0.0, 1.0)
	f.Priority = f.computePriority()
}

// Touch updates access time and count.
func (f *FocusItem) Touch() {
	f.LastAccessTime = time.Now()
	f.AccessCount++
	f.Priority = f.computePriority()
}

// Clone creates a deep copy of the focus item.
func (f *FocusItem) Clone() *FocusItem {
	clone := &FocusItem{
		ID:             f.ID,
		Type:           f.Type,
		Content:        f.Content,
		Label:          f.Label,
		Salience:       f.Salience,
		Priority:       f.Priority,
		CognitiveLoad:  f.CognitiveLoad,
		EntryTime:      f.EntryTime,
		LastAccessTime: f.LastAccessTime,
		AccessCount:    f.AccessCount,
		DecayRate:      f.DecayRate,
		Sticky:         f.Sticky,
		SourceID:       f.SourceID,
		Metadata:       make(map[string]interface{}),
	}
	for k, v := range f.Metadata {
		clone.Metadata[k] = v
	}
	return clone
}

// computeDefaultLoad estimates cognitive load based on type and salience.
func computeDefaultLoad(itemType FocusItemType, salience float64) float64 {
	// Base load by type
	baseLoad := 0.1
	switch itemType {
	case FocusGoal:
		baseLoad = 0.3
	case FocusTask:
		baseLoad = 0.2
	case FocusInterrupt:
		baseLoad = 0.4
	case FocusExperience:
		baseLoad = 0.15
	case FocusContext:
		baseLoad = 0.1
	case FocusAgent:
		baseLoad = 0.2
	case FocusReflection:
		baseLoad = 0.25
	}
	// Scale by salience
	return baseLoad * (0.5 + 0.5*salience)
}

// ============================================================================
// Focus Priority Queue (Max-Heap)
// ============================================================================

// FocusPriorityQueue implements heap.Interface for priority-based focus management.
type FocusPriorityQueue []*FocusItem

func (pq FocusPriorityQueue) Len() int { return len(pq) }

func (pq FocusPriorityQueue) Less(i, j int) bool {
	// Max-heap: higher priority first
	return pq[i].Priority > pq[j].Priority
}

func (pq FocusPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *FocusPriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*FocusItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *FocusPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// ============================================================================
// Salience Computer
// ============================================================================

// SalienceFactors holds factors that influence salience computation.
type SalienceFactors struct {
	// Novelty: how new/unexpected is this item?
	Novelty float64

	// Relevance: how relevant to current goals?
	Relevance float64

	// Urgency: how time-sensitive?
	Urgency float64

	// Emotional: emotional significance (for future expansion)
	Emotional float64

	// SourceTrust: trust in the source of this item
	SourceTrust float64
}

// SalienceComputer computes salience for various inputs.
type SalienceComputer struct {
	// weights for different factors
	noveltyWeight   float64
	relevanceWeight float64
	urgencyWeight   float64
	emotionalWeight float64
	trustWeight     float64

	// currentGoals for relevance computation
	currentGoals []string

	// noveltyBaseline for comparison
	noveltyBaseline map[string]float64
}

// NewSalienceComputer creates a new salience computer with default weights.
func NewSalienceComputer() *SalienceComputer {
	return &SalienceComputer{
		noveltyWeight:   0.25,
		relevanceWeight: 0.35,
		urgencyWeight:   0.25,
		emotionalWeight: 0.05,
		trustWeight:     0.10,
		currentGoals:    make([]string, 0),
		noveltyBaseline: make(map[string]float64),
	}
}

// SetWeights configures the salience factor weights.
func (sc *SalienceComputer) SetWeights(novelty, relevance, urgency, emotional, trust float64) {
	sc.noveltyWeight = novelty
	sc.relevanceWeight = relevance
	sc.urgencyWeight = urgency
	sc.emotionalWeight = emotional
	sc.trustWeight = trust
}

// SetCurrentGoals updates the goals used for relevance computation.
func (sc *SalienceComputer) SetCurrentGoals(goals []string) {
	sc.currentGoals = goals
}

// ComputeSalience calculates the overall salience from factors.
func (sc *SalienceComputer) ComputeSalience(factors *SalienceFactors) float64 {
	if factors == nil {
		return 0.5 // Default salience
	}

	salience := sc.noveltyWeight*factors.Novelty +
		sc.relevanceWeight*factors.Relevance +
		sc.urgencyWeight*factors.Urgency +
		sc.emotionalWeight*factors.Emotional +
		sc.trustWeight*factors.SourceTrust

	return clampFloat(salience, 0.0, 1.0)
}

// ComputeNovelty estimates novelty based on how different this is from baseline.
func (sc *SalienceComputer) ComputeNovelty(key string, value float64) float64 {
	baseline, exists := sc.noveltyBaseline[key]
	if !exists {
		sc.noveltyBaseline[key] = value
		return 1.0 // Completely novel
	}

	// Update baseline with exponential moving average
	alpha := 0.1
	sc.noveltyBaseline[key] = alpha*value + (1-alpha)*baseline

	// Novelty is deviation from baseline
	deviation := math.Abs(value - baseline)
	if baseline == 0 {
		return deviation
	}
	return clampFloat(deviation/baseline, 0.0, 1.0)
}

// ============================================================================
// Attention Controller
// ============================================================================

// AttentionController manages the attention focus with capacity constraints.
type AttentionController struct {
	mu sync.RWMutex

	// capacity is the maximum cognitive load
	capacity float64

	// currentLoad is the sum of cognitive loads of focused items
	currentLoad float64

	// focusHeap is the priority queue of focus items
	focusHeap FocusPriorityQueue

	// focusMap provides O(1) lookup by ID
	focusMap map[string]*FocusItem

	// salienceComputer computes salience values
	salienceComputer *SalienceComputer

	// config holds controller configuration
	config *AttentionConfig

	// stats tracks attention statistics
	stats *AttentionStats

	// callbacks
	onFocusGained func(*FocusItem)
	onFocusLost   func(*FocusItem, string) // item, reason
	onOverload    func(float64)            // current load

	// workingMemory reference (optional integration)
	workingMemory *CognitiveWorkingMemory
}

// AttentionConfig configures the attention controller.
type AttentionConfig struct {
	// Capacity is the maximum cognitive load (default: 7 ± 2, Miller's Law)
	Capacity float64

	// DecayInterval is how often to apply salience decay
	DecayInterval time.Duration

	// MinSalience is the threshold below which items are evicted
	MinSalience float64

	// MaxFocusItems limits the number of items regardless of load
	MaxFocusItems int

	// InterruptThreshold is the minimum salience for an interrupt to succeed
	InterruptThreshold float64

	// StickyDecayMultiplier slows decay for sticky items
	StickyDecayMultiplier float64
}

// DefaultAttentionConfig returns sensible defaults based on cognitive science.
func DefaultAttentionConfig() *AttentionConfig {
	return &AttentionConfig{
		Capacity:              7.0, // Miller's Law: 7 ± 2
		DecayInterval:         time.Second,
		MinSalience:           0.05,
		MaxFocusItems:         15,
		InterruptThreshold:    0.7,
		StickyDecayMultiplier: 0.5,
	}
}

// AttentionStats tracks attention-related statistics.
type AttentionStats struct {
	TotalItemsFocused  int64
	TotalItemsEvicted  int64
	TotalInterrupts    int64
	TotalOverloads     int64
	AverageLoadPercent float64
	PeakLoad           float64
	FocusGainedCount   int64
	FocusLostCount     int64
}

// NewAttentionController creates a new attention controller.
func NewAttentionController(config *AttentionConfig) *AttentionController {
	if config == nil {
		config = DefaultAttentionConfig()
	}

	ac := &AttentionController{
		capacity:         config.Capacity,
		focusHeap:        make(FocusPriorityQueue, 0),
		focusMap:         make(map[string]*FocusItem),
		salienceComputer: NewSalienceComputer(),
		config:           config,
		stats:            &AttentionStats{},
	}

	heap.Init(&ac.focusHeap)
	return ac
}

// ============================================================================
// Focus Management
// ============================================================================

// Focus attempts to add an item to the focus.
// Returns true if successful, false if capacity exceeded and item couldn't be added.
func (ac *AttentionController) Focus(item *FocusItem) (bool, error) {
	if item == nil {
		return false, errors.New("nil focus item")
	}

	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Check if already focused
	if _, exists := ac.focusMap[item.ID]; exists {
		// Update existing item
		ac.touchItem(item.ID)
		return true, nil
	}

	// Check item limit
	if len(ac.focusHeap) >= ac.config.MaxFocusItems {
		// Try to evict lowest priority item
		if !ac.evictLowest(item.Priority) {
			return false, ErrAttentionCapacityExceeded
		}
	}

	// Check cognitive load capacity
	if ac.currentLoad+item.CognitiveLoad > ac.capacity {
		// Try to make room by evicting low-priority items
		if !ac.makeRoom(item.CognitiveLoad, item.Priority) {
			ac.stats.TotalOverloads++
			if ac.onOverload != nil {
				ac.onOverload(ac.currentLoad)
			}
			return false, ErrAttentionCapacityExceeded
		}
	}

	// Add item to focus
	ac.addItem(item)
	return true, nil
}

// FocusInterrupt forces attention to an interrupt item, potentially evicting others.
func (ac *AttentionController) FocusInterrupt(item *FocusItem) error {
	if item == nil {
		return errors.New("nil interrupt item")
	}

	if item.Salience < ac.config.InterruptThreshold {
		return ErrAttentionBlocked
	}

	item.Type = FocusInterrupt
	item.Priority = item.computePriority()

	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Force room for interrupt by evicting if necessary
	for ac.currentLoad+item.CognitiveLoad > ac.capacity && len(ac.focusHeap) > 0 {
		ac.evictLowestUnlocked()
	}

	ac.addItem(item)
	ac.stats.TotalInterrupts++
	return nil
}

// Unfocus removes an item from focus.
func (ac *AttentionController) Unfocus(itemID string) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	item, exists := ac.focusMap[itemID]
	if !exists {
		return ErrFocusItemNotFound
	}

	ac.removeItem(item, "manual")
	return nil
}

// Touch refreshes an item's access time and priority.
func (ac *AttentionController) Touch(itemID string) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	return ac.touchItem(itemID)
}

// touchItem updates access time (must be called with lock held).
func (ac *AttentionController) touchItem(itemID string) error {
	item, exists := ac.focusMap[itemID]
	if !exists {
		return ErrFocusItemNotFound
	}

	item.Touch()
	heap.Fix(&ac.focusHeap, item.index)
	return nil
}

// addItem adds an item to focus (must be called with lock held).
func (ac *AttentionController) addItem(item *FocusItem) {
	heap.Push(&ac.focusHeap, item)
	ac.focusMap[item.ID] = item
	ac.currentLoad += item.CognitiveLoad
	ac.stats.TotalItemsFocused++
	ac.stats.FocusGainedCount++

	// Update peak load
	if ac.currentLoad > ac.stats.PeakLoad {
		ac.stats.PeakLoad = ac.currentLoad
	}

	if ac.onFocusGained != nil {
		ac.onFocusGained(item)
	}

	// Sync to working memory if integrated
	if ac.workingMemory != nil {
		ac.syncToWorkingMemory(item, true)
	}
}

// removeItem removes an item from focus (must be called with lock held).
func (ac *AttentionController) removeItem(item *FocusItem, reason string) {
	if item.index >= 0 && item.index < len(ac.focusHeap) {
		heap.Remove(&ac.focusHeap, item.index)
	}
	delete(ac.focusMap, item.ID)
	ac.currentLoad -= item.CognitiveLoad
	if ac.currentLoad < 0 {
		ac.currentLoad = 0
	}
	ac.stats.FocusLostCount++

	if ac.onFocusLost != nil {
		ac.onFocusLost(item, reason)
	}

	// Sync to working memory if integrated
	if ac.workingMemory != nil {
		ac.syncToWorkingMemory(item, false)
	}
}

// evictLowest removes the lowest priority item if its priority is below threshold.
func (ac *AttentionController) evictLowest(threshold float64) bool {
	if len(ac.focusHeap) == 0 {
		return false
	}

	// Find lowest priority item
	lowestIdx := -1
	lowestPriority := math.MaxFloat64
	for i, item := range ac.focusHeap {
		if !item.Sticky && item.Priority < lowestPriority {
			lowestPriority = item.Priority
			lowestIdx = i
		}
	}

	if lowestIdx == -1 || lowestPriority >= threshold {
		return false
	}

	item := ac.focusHeap[lowestIdx]
	ac.removeItem(item, "evicted_by_priority")
	ac.stats.TotalItemsEvicted++
	return true
}

// evictLowestUnlocked removes the lowest priority non-sticky item.
func (ac *AttentionController) evictLowestUnlocked() bool {
	if len(ac.focusHeap) == 0 {
		return false
	}

	// Find lowest priority non-sticky item
	lowestIdx := -1
	lowestPriority := math.MaxFloat64
	for i, item := range ac.focusHeap {
		if !item.Sticky && item.Priority < lowestPriority {
			lowestPriority = item.Priority
			lowestIdx = i
		}
	}

	if lowestIdx == -1 {
		return false
	}

	item := ac.focusHeap[lowestIdx]
	ac.removeItem(item, "evicted_for_interrupt")
	ac.stats.TotalItemsEvicted++
	return true
}

// makeRoom attempts to evict items to make room for newLoad.
func (ac *AttentionController) makeRoom(newLoad, newPriority float64) bool {
	needed := (ac.currentLoad + newLoad) - ac.capacity

	// Collect evictable items sorted by priority
	evictable := make([]*FocusItem, 0)
	for _, item := range ac.focusHeap {
		if !item.Sticky && item.Priority < newPriority {
			evictable = append(evictable, item)
		}
	}

	// Sort by priority ascending (lowest first)
	sort.Slice(evictable, func(i, j int) bool {
		return evictable[i].Priority < evictable[j].Priority
	})

	// Evict until we have enough room
	freed := 0.0
	for _, item := range evictable {
		if freed >= needed {
			break
		}
		ac.removeItem(item, "evicted_for_capacity")
		ac.stats.TotalItemsEvicted++
		freed += item.CognitiveLoad
	}

	return freed >= needed
}

// ============================================================================
// Decay and Maintenance
// ============================================================================

// DecayAll applies salience decay to all focus items.
func (ac *AttentionController) DecayAll(elapsed time.Duration) int {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	evicted := 0
	itemsToEvict := make([]*FocusItem, 0)

	for _, item := range ac.focusHeap {
		// Apply decay (slower for sticky items)
		decayDuration := elapsed
		if item.Sticky {
			decayDuration = time.Duration(float64(elapsed) * ac.config.StickyDecayMultiplier)
		}
		item.DecaySalience(decayDuration)

		// Check if below threshold
		if item.Salience < ac.config.MinSalience {
			itemsToEvict = append(itemsToEvict, item)
		}
	}

	// Evict decayed items
	for _, item := range itemsToEvict {
		ac.removeItem(item, "salience_decay")
		evicted++
		ac.stats.TotalItemsEvicted++
	}

	// Rebuild heap after priority changes
	heap.Init(&ac.focusHeap)

	return evicted
}

// Tick performs periodic maintenance (decay, stats update).
func (ac *AttentionController) Tick() {
	ac.DecayAll(ac.config.DecayInterval)
	ac.updateAverageLoad()
}

// updateAverageLoad updates the running average load.
func (ac *AttentionController) updateAverageLoad() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	loadPercent := (ac.currentLoad / ac.capacity) * 100
	alpha := 0.1
	ac.stats.AverageLoadPercent = alpha*loadPercent + (1-alpha)*ac.stats.AverageLoadPercent
}

// ============================================================================
// Query Methods
// ============================================================================

// GetFocused returns a copy of the current focus item by ID.
func (ac *AttentionController) GetFocused(itemID string) (*FocusItem, bool) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	item, exists := ac.focusMap[itemID]
	if !exists {
		return nil, false
	}
	return item.Clone(), true
}

// GetAllFocused returns copies of all focused items.
func (ac *AttentionController) GetAllFocused() []*FocusItem {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	result := make([]*FocusItem, len(ac.focusHeap))
	for i, item := range ac.focusHeap {
		result[i] = item.Clone()
	}
	return result
}

// GetTopFocused returns the top N items by priority.
func (ac *AttentionController) GetTopFocused(n int) []*FocusItem {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	// Copy and sort by priority
	items := make([]*FocusItem, len(ac.focusHeap))
	for i, item := range ac.focusHeap {
		items[i] = item.Clone()
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Priority > items[j].Priority
	})

	if n > len(items) {
		n = len(items)
	}
	return items[:n]
}

// GetFocusedByType returns items of a specific type.
func (ac *AttentionController) GetFocusedByType(itemType FocusItemType) []*FocusItem {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	result := make([]*FocusItem, 0)
	for _, item := range ac.focusHeap {
		if item.Type == itemType {
			result = append(result, item.Clone())
		}
	}
	return result
}

// FocusCount returns the number of focused items.
func (ac *AttentionController) FocusCount() int {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return len(ac.focusHeap)
}

// CurrentLoad returns the current cognitive load.
func (ac *AttentionController) CurrentLoad() float64 {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.currentLoad
}

// Capacity returns the maximum cognitive load.
func (ac *AttentionController) Capacity() float64 {
	return ac.capacity
}

// LoadPercent returns current load as percentage of capacity.
func (ac *AttentionController) LoadPercent() float64 {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return (ac.currentLoad / ac.capacity) * 100
}

// CanFocus checks if an item with the given load can be focused.
func (ac *AttentionController) CanFocus(load float64) bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	// Must have capacity AND item count below limit
	return ac.currentLoad+load <= ac.capacity && len(ac.focusHeap) < ac.config.MaxFocusItems
}

// ============================================================================
// Working Memory Integration
// ============================================================================

// SetWorkingMemory links the attention controller to working memory.
func (ac *AttentionController) SetWorkingMemory(wm *CognitiveWorkingMemory) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.workingMemory = wm
}

// syncToWorkingMemory syncs focus state to working memory.
func (ac *AttentionController) syncToWorkingMemory(item *FocusItem, gained bool) {
	if ac.workingMemory == nil {
		return
	}

	if gained {
		// Add or boost item in working memory
		wmItem := &WorkingMemoryItem{
			ID:          "attn-" + item.ID,
			ContentType: focusTypeToWMType(item.Type),
			Content:     item.Content,
			Source:      SourceAttention,
			Activation:  item.Salience,
			Metadata: map[string]interface{}{
				"focus_id":    item.ID,
				"focus_type":  item.Type.String(),
				"focus_label": item.Label,
			},
		}
		ac.workingMemory.Add(wmItem)
	} else {
		// Remove or decay item in working memory
		ac.workingMemory.Remove("attn-" + item.ID)
	}
}

// focusTypeToWMType converts focus type to working memory content type.
func focusTypeToWMType(ft FocusItemType) WorkingMemoryContentType {
	switch ft {
	case FocusGoal:
		return ContentTypeGoal
	case FocusTask:
		return ContentTypeTask
	case FocusExperience:
		return ContentTypeExperience
	case FocusContext:
		return ContentTypeContext
	case FocusAgent:
		return ContentTypeAgent
	default:
		return ContentTypeGeneral
	}
}

// ============================================================================
// Callbacks
// ============================================================================

// OnFocusGained sets callback for when an item gains focus.
func (ac *AttentionController) OnFocusGained(fn func(*FocusItem)) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.onFocusGained = fn
}

// OnFocusLost sets callback for when an item loses focus.
func (ac *AttentionController) OnFocusLost(fn func(*FocusItem, string)) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.onFocusLost = fn
}

// OnOverload sets callback for attention overload.
func (ac *AttentionController) OnOverload(fn func(float64)) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.onOverload = fn
}

// ============================================================================
// Statistics
// ============================================================================

// GetStats returns attention statistics.
func (ac *AttentionController) GetStats() AttentionStats {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return *ac.stats
}

// ============================================================================
// Snapshot/Restore
// ============================================================================

// AttentionSnapshot captures the attention state.
type AttentionSnapshot struct {
	Items       []*FocusItem
	CurrentLoad float64
	Stats       AttentionStats
	Timestamp   time.Time
}

// Snapshot captures current attention state.
func (ac *AttentionController) Snapshot() *AttentionSnapshot {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	items := make([]*FocusItem, len(ac.focusHeap))
	for i, item := range ac.focusHeap {
		items[i] = item.Clone()
	}

	return &AttentionSnapshot{
		Items:       items,
		CurrentLoad: ac.currentLoad,
		Stats:       *ac.stats,
		Timestamp:   time.Now(),
	}
}

// Restore restores attention state from snapshot.
func (ac *AttentionController) Restore(snapshot *AttentionSnapshot) error {
	if snapshot == nil {
		return errors.New("nil snapshot")
	}

	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.focusHeap = make(FocusPriorityQueue, len(snapshot.Items))
	ac.focusMap = make(map[string]*FocusItem)

	for i, item := range snapshot.Items {
		clone := item.Clone()
		clone.index = i
		ac.focusHeap[i] = clone
		ac.focusMap[clone.ID] = clone
	}

	heap.Init(&ac.focusHeap)
	ac.currentLoad = snapshot.CurrentLoad
	ac.stats = &AttentionStats{}
	*ac.stats = snapshot.Stats

	return nil
}

// Clear removes all focus items.
func (ac *AttentionController) Clear() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.focusHeap = make(FocusPriorityQueue, 0)
	ac.focusMap = make(map[string]*FocusItem)
	ac.currentLoad = 0
	heap.Init(&ac.focusHeap)
}

// ============================================================================
// Attention-Based Filtering
// ============================================================================

// FilterByAttention filters a slice based on attention capacity.
// Returns items that can fit in remaining attention, sorted by computed salience.
func (ac *AttentionController) FilterByAttention(items []interface{}, salienceFunc func(interface{}) float64) []interface{} {
	ac.mu.RLock()
	availableLoad := ac.capacity - ac.currentLoad
	ac.mu.RUnlock()

	if availableLoad <= 0 {
		return nil
	}

	// Compute salience for each item
	type scored struct {
		item     interface{}
		salience float64
		load     float64
	}

	scored_items := make([]scored, len(items))
	for i, item := range items {
		sal := salienceFunc(item)
		scored_items[i] = scored{
			item:     item,
			salience: sal,
			load:     0.1 * (0.5 + 0.5*sal), // Estimate load from salience
		}
	}

	// Sort by salience descending
	sort.Slice(scored_items, func(i, j int) bool {
		return scored_items[i].salience > scored_items[j].salience
	})

	// Take items until capacity exhausted
	result := make([]interface{}, 0)
	usedLoad := 0.0
	for _, s := range scored_items {
		if usedLoad+s.load > availableLoad {
			continue
		}
		result = append(result, s.item)
		usedLoad += s.load
	}

	return result
}

// ============================================================================
// Helper Functions
// ============================================================================

// clampFloat clamps a float64 value to the given range.
func clampFloat(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
