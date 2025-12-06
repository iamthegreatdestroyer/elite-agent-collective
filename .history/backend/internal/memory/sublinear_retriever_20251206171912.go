// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements sub-linear retrieval using LSH, HNSW, and Bloom Filters.

package memory

import (
	"encoding/binary"
	"hash/fnv"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// ============================================================================
// Bloom Filter - O(1) membership testing
// ============================================================================

// BloomFilter provides probabilistic set membership testing with O(1) complexity.
// Used for quick checks before expensive operations.
type BloomFilter struct {
	bitArray []bool
	numHash  int
	size     int
	mu       sync.RWMutex
}

// NewBloomFilter creates a new Bloom filter with the specified size and hash count.
// Optimal hash count k ≈ (m/n) * ln(2), where m = size, n = expected elements.
func NewBloomFilter(size int, numHash int) *BloomFilter {
	return &BloomFilter{
		bitArray: make([]bool, size),
		numHash:  numHash,
		size:     size,
	}
}

// NewBloomFilterOptimal creates a Bloom filter optimized for expected elements and false positive rate.
func NewBloomFilterOptimal(expectedElements int, falsePositiveRate float64) *BloomFilter {
	// m = -n * ln(p) / (ln(2)^2)
	m := int(math.Ceil(-float64(expectedElements) * math.Log(falsePositiveRate) / (math.Ln2 * math.Ln2)))
	// k = (m/n) * ln(2)
	k := int(math.Ceil(float64(m) / float64(expectedElements) * math.Ln2))
	return NewBloomFilter(m, k)
}

// getHashes generates k hash values for the given key using double hashing.
func (b *BloomFilter) getHashes(key string) []int {
	h1 := fnv.New64a()
	h1.Write([]byte(key))
	hash1 := h1.Sum64()

	h2 := fnv.New64()
	h2.Write([]byte(key))
	hash2 := h2.Sum64()

	hashes := make([]int, b.numHash)
	for i := 0; i < b.numHash; i++ {
		// Double hashing: hash_i = hash1 + i * hash2
		combined := hash1 + uint64(i)*hash2
		hashes[i] = int(combined % uint64(b.size))
	}
	return hashes
}

// Add inserts a key into the Bloom filter.
func (b *BloomFilter) Add(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, h := range b.getHashes(key) {
		b.bitArray[h] = true
	}
}

// MayContain checks if a key might be in the set (may have false positives).
func (b *BloomFilter) MayContain(key string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, h := range b.getHashes(key) {
		if !b.bitArray[h] {
			return false
		}
	}
	return true
}

// Clear resets the Bloom filter.
func (b *BloomFilter) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.bitArray = make([]bool, b.size)
}

// ============================================================================
// LSH Index - O(1) expected approximate nearest neighbor lookup
// ============================================================================

// LSHIndex implements Locality Sensitive Hashing for O(1) approximate nearest neighbor queries.
// Uses random hyperplane hashing for cosine similarity.
type LSHIndex struct {
	numHashTables int
	numHashFuncs  int
	dimension     int
	hashTables    []map[uint64][]string // hash -> experience IDs
	hyperplanes   [][][]float32         // [table][hash_func][dimension]
	mu            sync.RWMutex
	rng           *rand.Rand
}

// NewLSHIndex creates a new LSH index with the specified parameters.
// numHashTables: more tables = higher recall, lower precision
// numHashFuncs: more functions = higher precision, lower recall
func NewLSHIndex(numHashTables, numHashFuncs, dimension int) *LSHIndex {
	lsh := &LSHIndex{
		numHashTables: numHashTables,
		numHashFuncs:  numHashFuncs,
		dimension:     dimension,
		hashTables:    make([]map[uint64][]string, numHashTables),
		hyperplanes:   make([][][]float32, numHashTables),
		rng:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// Initialize hash tables and random hyperplanes
	for t := 0; t < numHashTables; t++ {
		lsh.hashTables[t] = make(map[uint64][]string)
		lsh.hyperplanes[t] = make([][]float32, numHashFuncs)
		for h := 0; h < numHashFuncs; h++ {
			lsh.hyperplanes[t][h] = lsh.randomHyperplane()
		}
	}

	return lsh
}

// randomHyperplane generates a random unit vector for hyperplane hashing.
func (l *LSHIndex) randomHyperplane() []float32 {
	plane := make([]float32, l.dimension)
	var norm float64
	for i := 0; i < l.dimension; i++ {
		plane[i] = float32(l.rng.NormFloat64())
		norm += float64(plane[i] * plane[i])
	}
	norm = math.Sqrt(norm)
	for i := range plane {
		plane[i] /= float32(norm)
	}
	return plane
}

// computeHash computes the LSH hash for a vector in a specific table.
func (l *LSHIndex) computeHash(vector []float32, tableIdx int) uint64 {
	if len(vector) != l.dimension {
		return 0
	}

	var hash uint64
	for h, hyperplane := range l.hyperplanes[tableIdx] {
		// Compute dot product
		var dot float32
		for i, v := range vector {
			dot += v * hyperplane[i]
		}
		// Set bit based on sign of dot product
		if dot >= 0 {
			hash |= (1 << uint(h))
		}
	}
	return hash
}

// Add inserts a vector with its ID into the LSH index.
func (l *LSHIndex) Add(id string, vector []float32) {
	if len(vector) != l.dimension {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	for t := 0; t < l.numHashTables; t++ {
		hash := l.computeHash(vector, t)
		l.hashTables[t][hash] = append(l.hashTables[t][hash], id)
	}
}

// Remove removes a vector from the LSH index by its ID.
func (l *LSHIndex) Remove(id string, vector []float32) {
	if len(vector) != l.dimension {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	for t := 0; t < l.numHashTables; t++ {
		hash := l.computeHash(vector, t)
		bucket := l.hashTables[t][hash]
		for i, existingID := range bucket {
			if existingID == id {
				l.hashTables[t][hash] = append(bucket[:i], bucket[i+1:]...)
				break
			}
		}
	}
}

// Query finds candidate IDs that are similar to the query vector.
// Returns up to maxCandidates IDs. Complexity: O(1) expected.
func (l *LSHIndex) Query(vector []float32, maxCandidates int) []string {
	if len(vector) != l.dimension {
		return nil
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	// Use a map to deduplicate candidates across tables
	candidateSet := make(map[string]int) // ID -> count of appearances

	for t := 0; t < l.numHashTables; t++ {
		hash := l.computeHash(vector, t)
		for _, id := range l.hashTables[t][hash] {
			candidateSet[id]++
		}
	}

	// Convert to slice and sort by frequency (more appearances = more likely similar)
	type candidate struct {
		id    string
		count int
	}
	candidates := make([]candidate, 0, len(candidateSet))
	for id, count := range candidateSet {
		candidates = append(candidates, candidate{id, count})
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].count > candidates[j].count
	})

	// Return top candidates
	result := make([]string, 0, maxCandidates)
	for i := 0; i < len(candidates) && i < maxCandidates; i++ {
		result = append(result, candidates[i].id)
	}
	return result
}

// Size returns the total number of entries across all hash tables.
func (l *LSHIndex) Size() int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	total := 0
	seen := make(map[string]bool)
	for t := 0; t < l.numHashTables; t++ {
		for _, bucket := range l.hashTables[t] {
			for _, id := range bucket {
				if !seen[id] {
					seen[id] = true
					total++
				}
			}
		}
	}
	return total
}

// ============================================================================
// HNSW Graph - O(log n) hierarchical navigable small world search
// ============================================================================

// HNSWNode represents a node in the HNSW graph.
type HNSWNode struct {
	ID         string
	Vector     []float32
	Level      int
	Neighbors  [][]string // neighbors[level] = list of neighbor IDs
}

// HNSWGraph implements a Hierarchical Navigable Small World graph for
// O(log n) approximate nearest neighbor search.
type HNSWGraph struct {
	maxLevel       int
	efConstruction int // Size of dynamic candidate list during construction
	efSearch       int // Size of dynamic candidate list during search
	mMax           int // Maximum number of connections per element per layer
	mMax0          int // Maximum number of connections for layer 0
	ml             float64 // Level generation factor (1/ln(M))
	nodes          map[string]*HNSWNode
	entryPoint     string
	dimension      int
	mu             sync.RWMutex
	rng            *rand.Rand
}

// NewHNSWGraph creates a new HNSW graph with the specified parameters.
func NewHNSWGraph(dimension, mMax, efConstruction int) *HNSWGraph {
	return &HNSWGraph{
		maxLevel:       0,
		efConstruction: efConstruction,
		efSearch:       efConstruction, // Can be tuned separately
		mMax:           mMax,
		mMax0:          mMax * 2, // Convention: double connections at layer 0
		ml:             1.0 / math.Log(float64(mMax)),
		nodes:          make(map[string]*HNSWNode),
		dimension:      dimension,
		rng:            rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SetEfSearch sets the search ef parameter (controls recall/speed tradeoff).
func (h *HNSWGraph) SetEfSearch(ef int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.efSearch = ef
}

// randomLevel generates a random level for a new node.
func (h *HNSWGraph) randomLevel() int {
	r := h.rng.Float64()
	level := int(math.Floor(-math.Log(r) * h.ml))
	return level
}

// distance computes Euclidean distance between two vectors.
func (h *HNSWGraph) distance(a, b []float32) float32 {
	if len(a) != len(b) {
		return math.MaxFloat32
	}
	var sum float32
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return float32(math.Sqrt(float64(sum)))
}

// cosineSimilarity computes cosine similarity between two vectors.
func (h *HNSWGraph) cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return -1
	}
	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// priorityQueue implements a min-heap for nearest neighbor search.
type priorityQueue struct {
	items []pqItem
}

type pqItem struct {
	id       string
	distance float32
}

func (pq *priorityQueue) Len() int { return len(pq.items) }

func (pq *priorityQueue) Push(id string, dist float32) {
	pq.items = append(pq.items, pqItem{id, dist})
	// Bubble up
	i := len(pq.items) - 1
	for i > 0 {
		parent := (i - 1) / 2
		if pq.items[parent].distance <= pq.items[i].distance {
			break
		}
		pq.items[parent], pq.items[i] = pq.items[i], pq.items[parent]
		i = parent
	}
}

func (pq *priorityQueue) Pop() (string, float32) {
	if len(pq.items) == 0 {
		return "", 0
	}
	item := pq.items[0]
	last := len(pq.items) - 1
	pq.items[0] = pq.items[last]
	pq.items = pq.items[:last]
	// Bubble down
	i := 0
	for {
		left := 2*i + 1
		right := 2*i + 2
		smallest := i
		if left < len(pq.items) && pq.items[left].distance < pq.items[smallest].distance {
			smallest = left
		}
		if right < len(pq.items) && pq.items[right].distance < pq.items[smallest].distance {
			smallest = right
		}
		if smallest == i {
			break
		}
		pq.items[i], pq.items[smallest] = pq.items[smallest], pq.items[i]
		i = smallest
	}
	return item.id, item.distance
}

func (pq *priorityQueue) Peek() (string, float32) {
	if len(pq.items) == 0 {
		return "", 0
	}
	return pq.items[0].id, pq.items[0].distance
}

// maxHeap wraps priorityQueue to act as a max-heap (for keeping k nearest).
type maxHeap struct {
	pq priorityQueue
}

func (mh *maxHeap) Push(id string, dist float32) {
	// Store negative distance to simulate max-heap with min-heap
	mh.pq.Push(id, -dist)
}

func (mh *maxHeap) Pop() (string, float32) {
	id, negDist := mh.pq.Pop()
	return id, -negDist
}

func (mh *maxHeap) Peek() (string, float32) {
	id, negDist := mh.pq.Peek()
	return id, -negDist
}

func (mh *maxHeap) Len() int { return mh.pq.Len() }

// Add inserts a new node into the HNSW graph.
func (h *HNSWGraph) Add(id string, vector []float32) {
	if len(vector) != h.dimension {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	level := h.randomLevel()

	node := &HNSWNode{
		ID:        id,
		Vector:    vector,
		Level:     level,
		Neighbors: make([][]string, level+1),
	}
	for i := range node.Neighbors {
		node.Neighbors[i] = make([]string, 0)
	}

	// First node case
	if len(h.nodes) == 0 {
		h.nodes[id] = node
		h.entryPoint = id
		h.maxLevel = level
		return
	}

	h.nodes[id] = node

	// Find entry point for this level
	currentID := h.entryPoint
	currentNode := h.nodes[currentID]

	// Traverse from top level down to node's level + 1
	for l := h.maxLevel; l > level; l-- {
		currentID = h.searchLayer(vector, currentID, 1, l)[0]
		currentNode = h.nodes[currentID]
	}

	// For levels from min(level, maxLevel) down to 0, find and connect neighbors
	for l := min(level, h.maxLevel); l >= 0; l-- {
		neighbors := h.searchLayer(vector, currentID, h.efConstruction, l)

		// Select M best neighbors
		m := h.mMax
		if l == 0 {
			m = h.mMax0
		}
		selectedNeighbors := h.selectNeighbors(vector, neighbors, m, l)

		// Connect node to neighbors
		node.Neighbors[l] = selectedNeighbors

		// Connect neighbors back to node
		for _, neighborID := range selectedNeighbors {
			neighbor := h.nodes[neighborID]
			if neighbor != nil && l < len(neighbor.Neighbors) {
				neighbor.Neighbors[l] = append(neighbor.Neighbors[l], id)
				// Prune if too many connections
				maxConn := h.mMax
				if l == 0 {
					maxConn = h.mMax0
				}
				if len(neighbor.Neighbors[l]) > maxConn {
					neighbor.Neighbors[l] = h.selectNeighbors(neighbor.Vector, neighbor.Neighbors[l], maxConn, l)
				}
			}
		}

		if len(neighbors) > 0 {
			currentID = neighbors[0]
		}
	}

	// Update entry point if new node has higher level
	if level > h.maxLevel {
		h.entryPoint = id
		h.maxLevel = level
	}
}

// searchLayer performs greedy search within a single layer.
func (h *HNSWGraph) searchLayer(query []float32, entryID string, ef int, level int) []string {
	visited := make(map[string]bool)
	candidates := &priorityQueue{}
	result := &maxHeap{}

	entryNode := h.nodes[entryID]
	if entryNode == nil {
		return nil
	}

	dist := h.distance(query, entryNode.Vector)
	candidates.Push(entryID, dist)
	result.Push(entryID, dist)
	visited[entryID] = true

	for candidates.Len() > 0 {
		currentID, currentDist := candidates.Pop()
		_, furthestDist := result.Peek()

		if currentDist > furthestDist {
			break
		}

		currentNode := h.nodes[currentID]
		if currentNode == nil || level >= len(currentNode.Neighbors) {
			continue
		}

		for _, neighborID := range currentNode.Neighbors[level] {
			if visited[neighborID] {
				continue
			}
			visited[neighborID] = true

			neighborNode := h.nodes[neighborID]
			if neighborNode == nil {
				continue
			}

			dist := h.distance(query, neighborNode.Vector)
			_, furthestDist := result.Peek()

			if dist < furthestDist || result.Len() < ef {
				candidates.Push(neighborID, dist)
				result.Push(neighborID, dist)
				if result.Len() > ef {
					result.Pop()
				}
			}
		}
	}

	// Extract results
	results := make([]string, 0, result.Len())
	for result.Len() > 0 {
		id, _ := result.Pop()
		results = append(results, id)
	}
	// Reverse to get nearest first
	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}
	return results
}

// selectNeighbors selects the best neighbors using simple heuristic.
func (h *HNSWGraph) selectNeighbors(query []float32, candidates []string, m int, level int) []string {
	if len(candidates) <= m {
		return candidates
	}

	// Sort by distance
	type neighbor struct {
		id   string
		dist float32
	}
	neighbors := make([]neighbor, 0, len(candidates))
	for _, id := range candidates {
		node := h.nodes[id]
		if node != nil {
			neighbors = append(neighbors, neighbor{id, h.distance(query, node.Vector)})
		}
	}
	sort.Slice(neighbors, func(i, j int) bool {
		return neighbors[i].dist < neighbors[j].dist
	})

	result := make([]string, 0, m)
	for i := 0; i < len(neighbors) && len(result) < m; i++ {
		result = append(result, neighbors[i].id)
	}
	return result
}

// Search finds the k nearest neighbors to the query vector.
// Complexity: O(log n) expected.
func (h *HNSWGraph) Search(query []float32, k int) []*ExperienceTuple {
	if len(query) != h.dimension || len(h.nodes) == 0 {
		return nil
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	// Start from entry point
	currentID := h.entryPoint

	// Traverse from top level down to level 1
	for l := h.maxLevel; l > 0; l-- {
		results := h.searchLayer(query, currentID, 1, l)
		if len(results) > 0 {
			currentID = results[0]
		}
	}

	// Search in layer 0 with ef parameter
	ef := max(h.efSearch, k)
	results := h.searchLayer(query, currentID, ef, 0)

	// Return top k
	if len(results) > k {
		results = results[:k]
	}

	// Note: This returns IDs only. The caller should look up full ExperienceTuples.
	// Returning nil here as a placeholder - the actual retriever will handle this.
	return nil
}

// SearchIDs finds the k nearest neighbor IDs to the query vector.
func (h *HNSWGraph) SearchIDs(query []float32, k int) []string {
	if len(query) != h.dimension || len(h.nodes) == 0 {
		return nil
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	currentID := h.entryPoint

	for l := h.maxLevel; l > 0; l-- {
		results := h.searchLayer(query, currentID, 1, l)
		if len(results) > 0 {
			currentID = results[0]
		}
	}

	ef := max(h.efSearch, k)
	results := h.searchLayer(query, currentID, ef, 0)

	if len(results) > k {
		results = results[:k]
	}
	return results
}

// Remove removes a node from the HNSW graph.
func (h *HNSWGraph) Remove(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	node, exists := h.nodes[id]
	if !exists {
		return
	}

	// Remove connections from neighbors
	for l, neighbors := range node.Neighbors {
		for _, neighborID := range neighbors {
			neighbor := h.nodes[neighborID]
			if neighbor != nil && l < len(neighbor.Neighbors) {
				// Remove id from neighbor's connections
				for i, nid := range neighbor.Neighbors[l] {
					if nid == id {
						neighbor.Neighbors[l] = append(neighbor.Neighbors[l][:i], neighbor.Neighbors[l][i+1:]...)
						break
					}
				}
			}
		}
	}

	delete(h.nodes, id)

	// Update entry point if necessary
	if h.entryPoint == id {
		// Find new entry point with highest level
		h.entryPoint = ""
		h.maxLevel = 0
		for nid, n := range h.nodes {
			if n.Level >= h.maxLevel {
				h.maxLevel = n.Level
				h.entryPoint = nid
			}
		}
	}
}

// Size returns the number of nodes in the graph.
func (h *HNSWGraph) Size() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.nodes)
}

// ============================================================================
// SubLinearRetriever - Combined retrieval system
// ============================================================================

// SubLinearRetriever combines multiple sub-linear data structures for
// efficient experience retrieval.
type SubLinearRetriever struct {
	// Sub-linear indices
	lsh   *LSHIndex
	hnsw  *HNSWGraph
	bloom *BloomFilter

	// Experience storage
	experiences map[string]*ExperienceTuple
	expMu       sync.RWMutex

	// Secondary indices for efficient filtering
	agentIndex    map[string][]string // agent_id -> experience_ids
	tierIndex     map[int][]string    // tier -> experience_ids
	taskSigIndex  map[string]string   // task_signature -> experience_id (exact match)
	agentMu       sync.RWMutex
	tierMu        sync.RWMutex
	taskSigMu     sync.RWMutex

	// Configuration
	dimension int

	// Statistics
	stats *MemoryStats
}

// NewSubLinearRetriever creates a new sub-linear retriever with the specified embedding dimension.
func NewSubLinearRetriever(dimension int) *SubLinearRetriever {
	return &SubLinearRetriever{
		lsh:          NewLSHIndex(10, 12, dimension),      // 10 tables, 12 hash functions
		hnsw:         NewHNSWGraph(dimension, 16, 200),    // M=16, efConstruction=200
		bloom:        NewBloomFilterOptimal(1000000, 0.01), // 1M items, 1% FP rate
		experiences:  make(map[string]*ExperienceTuple),
		agentIndex:   make(map[string][]string),
		tierIndex:    make(map[int][]string),
		taskSigIndex: make(map[string]string),
		dimension:    dimension,
		stats:        NewMemoryStats(),
	}
}

// Add inserts an experience into all indices.
func (r *SubLinearRetriever) Add(exp *ExperienceTuple) error {
	if exp == nil || exp.ID == "" {
		return ErrInvalidExperience
	}

	// Store experience
	r.expMu.Lock()
	r.experiences[exp.ID] = exp
	r.expMu.Unlock()

	// Add to Bloom filter for exact match
	r.bloom.Add(exp.TaskSignature)

	// Add to task signature index
	r.taskSigMu.Lock()
	r.taskSigIndex[exp.TaskSignature] = exp.ID
	r.taskSigMu.Unlock()

	// Add to agent index
	r.agentMu.Lock()
	r.agentIndex[exp.AgentID] = append(r.agentIndex[exp.AgentID], exp.ID)
	r.agentMu.Unlock()

	// Add to tier index
	r.tierMu.Lock()
	r.tierIndex[exp.TierID] = append(r.tierIndex[exp.TierID], exp.ID)
	r.tierMu.Unlock()

	// Add to LSH and HNSW if embedding exists
	if len(exp.Embedding) == r.dimension {
		r.lsh.Add(exp.ID, exp.Embedding)
		r.hnsw.Add(exp.ID, exp.Embedding)
	}

	// Update statistics
	r.stats.IncrementExperiences(exp.AgentID, exp.TierID)

	return nil
}

// Remove removes an experience from all indices.
func (r *SubLinearRetriever) Remove(id string) error {
	r.expMu.Lock()
	exp, exists := r.experiences[id]
	if !exists {
		r.expMu.Unlock()
		return ErrExperienceNotFound
	}
	delete(r.experiences, id)
	r.expMu.Unlock()

	// Remove from task signature index
	r.taskSigMu.Lock()
	delete(r.taskSigIndex, exp.TaskSignature)
	r.taskSigMu.Unlock()

	// Remove from agent index
	r.agentMu.Lock()
	if ids, ok := r.agentIndex[exp.AgentID]; ok {
		for i, eid := range ids {
			if eid == id {
				r.agentIndex[exp.AgentID] = append(ids[:i], ids[i+1:]...)
				break
			}
		}
	}
	r.agentMu.Unlock()

	// Remove from tier index
	r.tierMu.Lock()
	if ids, ok := r.tierIndex[exp.TierID]; ok {
		for i, eid := range ids {
			if eid == id {
				r.tierIndex[exp.TierID] = append(ids[:i], ids[i+1:]...)
				break
			}
		}
	}
	r.tierMu.Unlock()

	// Remove from LSH and HNSW
	if len(exp.Embedding) == r.dimension {
		r.lsh.Remove(id, exp.Embedding)
		r.hnsw.Remove(id)
	}

	return nil
}

// Retrieve performs sub-linear experience retrieval using a tiered approach.
// 1. Bloom filter check for exact task signature (O(1))
// 2. LSH for approximate matching (O(1) expected)
// 3. HNSW for semantic search (O(log n))
func (r *SubLinearRetriever) Retrieve(query *QueryContext) (*RetrievalResult, error) {
	if query == nil {
		return nil, ErrInvalidQuery
	}

	startTime := time.Now()
	result := &RetrievalResult{
		Experiences: make([]*ExperienceTuple, 0, query.TopK),
	}

	// Step 1: Bloom filter check for exact task signature (O(1))
	if query.TaskSignature != "" && r.bloom.MayContain(query.TaskSignature) {
		r.taskSigMu.RLock()
		if expID, ok := r.taskSigIndex[query.TaskSignature]; ok {
			r.taskSigMu.RUnlock()
			r.expMu.RLock()
			if exp, exists := r.experiences[expID]; exists {
				result.Experiences = append(result.Experiences, exp)
				result.RetrievalMethod = "exact"
				result.TotalCandidates = 1
				// Update access statistics
				exp.UsageCount++
				exp.LastAccessTime = time.Now().UnixNano()
			}
			r.expMu.RUnlock()

			if len(result.Experiences) > 0 {
				result.RetrievalLatencyNs = time.Since(startTime).Nanoseconds()
				r.stats.UpdateRetrievalStats(result.RetrievalLatencyNs, true)
				return result, nil
			}
		} else {
			r.taskSigMu.RUnlock()
		}
	}

	// Step 2: LSH for approximate matching (O(1) expected)
	if len(query.Embedding) == r.dimension {
		candidates := r.lsh.Query(query.Embedding, query.TopK*3) // Get more candidates for filtering
		result.TotalCandidates = len(candidates)

		if len(candidates) > 0 {
			experiences := r.lookupAndFilter(candidates, query)
			if len(experiences) > 0 {
				result.Experiences = experiences
				result.RetrievalMethod = "lsh"
				result.RetrievalLatencyNs = time.Since(startTime).Nanoseconds()
				r.stats.UpdateRetrievalStats(result.RetrievalLatencyNs, false)
				return result, nil
			}
		}
	}

	// Step 3: HNSW for semantic search (O(log n))
	if len(query.Embedding) == r.dimension {
		ids := r.hnsw.SearchIDs(query.Embedding, query.TopK*3)
		result.TotalCandidates += len(ids)

		if len(ids) > 0 {
			experiences := r.lookupAndFilter(ids, query)
			if len(experiences) > 0 {
				result.Experiences = experiences
				result.RetrievalMethod = "hnsw"
				result.RetrievalLatencyNs = time.Since(startTime).Nanoseconds()
				r.stats.UpdateRetrievalStats(result.RetrievalLatencyNs, false)
				return result, nil
			}
		}
	}

	// Fallback: Return empty result
	result.RetrievalMethod = "fallback"
	result.RetrievalLatencyNs = time.Since(startTime).Nanoseconds()
	r.stats.UpdateRetrievalStats(result.RetrievalLatencyNs, false)
	return result, nil
}

// lookupAndFilter retrieves experiences by ID and applies query filters.
func (r *SubLinearRetriever) lookupAndFilter(ids []string, query *QueryContext) []*ExperienceTuple {
	r.expMu.RLock()
	defer r.expMu.RUnlock()

	now := time.Now().UnixNano()
	results := make([]*ExperienceTuple, 0, query.TopK)

	for _, id := range ids {
		if len(results) >= query.TopK {
			break
		}

		exp, exists := r.experiences[id]
		if !exists {
			continue
		}

		// Apply filters
		if exp.FitnessScore < query.MinFitnessScore {
			continue
		}

		if query.MaxAge > 0 && (now-exp.Timestamp) > query.MaxAge {
			continue
		}

		// Include experiences from the same agent, same tier (if enabled), or collective
		sameAgent := exp.AgentID == query.AgentID
		sameTier := query.IncludeTierExperiences && exp.TierID == query.TierID
		isCollective := query.IncludeCollectiveExperiences && exp.AgentID == "COLLECTIVE"

		if sameAgent || sameTier || isCollective {
			results = append(results, exp)
			// Update access statistics
			exp.UsageCount++
			exp.LastAccessTime = now
		}
	}

	return results
}

// GetByAgent returns all experiences for a specific agent.
func (r *SubLinearRetriever) GetByAgent(agentID string) []*ExperienceTuple {
	r.agentMu.RLock()
	ids := r.agentIndex[agentID]
	r.agentMu.RUnlock()

	r.expMu.RLock()
	defer r.expMu.RUnlock()

	results := make([]*ExperienceTuple, 0, len(ids))
	for _, id := range ids {
		if exp, exists := r.experiences[id]; exists {
			results = append(results, exp)
		}
	}
	return results
}

// GetByTier returns all experiences for a specific tier.
func (r *SubLinearRetriever) GetByTier(tierID int) []*ExperienceTuple {
	r.tierMu.RLock()
	ids := r.tierIndex[tierID]
	r.tierMu.RUnlock()

	r.expMu.RLock()
	defer r.expMu.RUnlock()

	results := make([]*ExperienceTuple, 0, len(ids))
	for _, id := range ids {
		if exp, exists := r.experiences[id]; exists {
			results = append(results, exp)
		}
	}
	return results
}

// GetStats returns the current memory statistics.
func (r *SubLinearRetriever) GetStats() MemoryStats {
	return r.stats.GetStats()
}

// Size returns the total number of stored experiences.
func (r *SubLinearRetriever) Size() int {
	r.expMu.RLock()
	defer r.expMu.RUnlock()
	return len(r.experiences)
}

// ============================================================================
// Utility functions
// ============================================================================

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// hashString computes a hash for a string.
func hashString(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

// hashBytes computes a hash for bytes.
func hashBytes(b []byte) uint64 {
	h := fnv.New64a()
	h.Write(b)
	return h.Sum64()
}

// float32ToBytes converts a float32 to bytes for hashing.
func float32ToBytes(f float32) []byte {
	bits := math.Float32bits(f)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)
	return bytes
}
