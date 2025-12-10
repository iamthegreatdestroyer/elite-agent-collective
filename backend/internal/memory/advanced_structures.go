// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements advanced sub-linear data structures for enhanced memory capabilities.
//
// @TENSOR @NEXUS @GENESIS - Elite Agent Collective Innovations v2.0
// - Count-Min Sketch: O(1) frequency tracking with bounded error
// - Cuckoo Filter: O(1) membership with deletion support
// - Product Quantization: 32x memory compression for embeddings
// - MinHash: O(1) semantic similarity estimation

package memory

import (
	"encoding/binary"
	"hash/fnv"
	"math"
	"math/rand"
	"sort"
	"sync"
)

// ============================================================================
// Count-Min Sketch - O(1) frequency estimation with bounded error
// ============================================================================

// CountMinSketch provides space-efficient frequency counting.
// Used for tracking strategy popularity and experience usage patterns.
// Complexity: O(1) update, O(1) query, O(d*w) space where d=depth, w=width.
type CountMinSketch struct {
	matrix [][]uint32
	width  int
	depth  int
	mu     sync.RWMutex
}

// NewCountMinSketch creates a Count-Min Sketch with specified error bounds.
// epsilon: error factor (ε), delta: probability of exceeding error
// width = ceil(e/ε), depth = ceil(ln(1/δ))
func NewCountMinSketch(epsilon, delta float64) *CountMinSketch {
	width := int(math.Ceil(math.E / epsilon))
	depth := int(math.Ceil(math.Log(1.0 / delta)))
	
	if width < 100 {
		width = 100
	}
	if depth < 3 {
		depth = 3
	}

	matrix := make([][]uint32, depth)
	for i := range matrix {
		matrix[i] = make([]uint32, width)
	}

	return &CountMinSketch{
		matrix: matrix,
		width:  width,
		depth:  depth,
	}
}

// NewCountMinSketchDefault creates a sketch with reasonable defaults.
// Approximately 0.1% error rate with 99.9% probability.
func NewCountMinSketchDefault() *CountMinSketch {
	return NewCountMinSketch(0.001, 0.001)
}

// getHashes generates d hash values for the given key.
func (c *CountMinSketch) getHashes(key string) []int {
	h1 := fnv.New64a()
	h1.Write([]byte(key))
	hash1 := h1.Sum64()

	h2 := fnv.New64()
	h2.Write([]byte(key))
	hash2 := h2.Sum64()

	hashes := make([]int, c.depth)
	for i := 0; i < c.depth; i++ {
		// Double hashing technique
		combined := hash1 + uint64(i)*hash2
		hashes[i] = int(combined % uint64(c.width))
	}
	return hashes
}

// Increment adds count to the frequency of key.
func (c *CountMinSketch) Increment(key string, count uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, hash := range c.getHashes(key) {
		c.matrix[i][hash] += count
	}
}

// Add increments the frequency of key by 1.
func (c *CountMinSketch) Add(key string) {
	c.Increment(key, 1)
}

// Estimate returns the estimated frequency of key (may overestimate).
func (c *CountMinSketch) Estimate(key string) uint32 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	hashes := c.getHashes(key)
	minCount := c.matrix[0][hashes[0]]

	for i := 1; i < c.depth; i++ {
		if count := c.matrix[i][hashes[i]]; count < minCount {
			minCount = count
		}
	}

	return minCount
}

// HeavyHitters returns items with estimated frequency >= threshold.
// Note: This is approximate and may include false positives.
func (c *CountMinSketch) HeavyHitters(items []string, threshold uint32) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, 0)
	for _, item := range items {
		if c.Estimate(item) >= threshold {
			result = append(result, item)
		}
	}
	return result
}

// Merge combines another Count-Min Sketch into this one.
func (c *CountMinSketch) Merge(other *CountMinSketch) error {
	if c.width != other.width || c.depth != other.depth {
		return ErrInvalidQuery // Mismatched dimensions
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 0; i < c.depth; i++ {
		for j := 0; j < c.width; j++ {
			c.matrix[i][j] += other.matrix[i][j]
		}
	}
	return nil
}

// ============================================================================
// Cuckoo Filter - O(1) membership with deletion support
// ============================================================================

// fingerprint is a compact hash stored in cuckoo buckets.
type fingerprint uint16

// CuckooFilter provides space-efficient set membership testing with deletion.
// Unlike Bloom filters, Cuckoo filters support O(1) deletion operations.
type CuckooFilter struct {
	buckets     [][]fingerprint
	numBuckets  int
	bucketSize  int
	maxKicks    int
	count       int
	mu          sync.RWMutex
}

// NewCuckooFilter creates a new Cuckoo filter.
// capacity: expected number of items
// bucketSize: slots per bucket (4 is recommended)
func NewCuckooFilter(capacity int, bucketSize int) *CuckooFilter {
	// Number of buckets = capacity / bucketSize / load_factor
	// Target 95% load factor for good performance
	numBuckets := int(math.Ceil(float64(capacity) / float64(bucketSize) / 0.95))
	if numBuckets < 2 {
		numBuckets = 2
	}

	buckets := make([][]fingerprint, numBuckets)
	for i := range buckets {
		buckets[i] = make([]fingerprint, 0, bucketSize)
	}

	return &CuckooFilter{
		buckets:    buckets,
		numBuckets: numBuckets,
		bucketSize: bucketSize,
		maxKicks:   500, // Maximum relocations before failure
	}
}

// NewCuckooFilterDefault creates a cuckoo filter for 100k items.
func NewCuckooFilterDefault() *CuckooFilter {
	return NewCuckooFilter(100000, 4)
}

// getFingerprint computes a fingerprint from a key.
func (c *CuckooFilter) getFingerprint(key string) fingerprint {
	h := fnv.New32a()
	h.Write([]byte(key))
	fp := fingerprint(h.Sum32() % 65535)
	if fp == 0 {
		fp = 1 // Reserve 0 for empty
	}
	return fp
}

// getBucketIndices returns the two candidate bucket indices for a key.
func (c *CuckooFilter) getBucketIndices(key string, fp fingerprint) (int, int) {
	h1 := fnv.New64a()
	h1.Write([]byte(key))
	i1 := int(h1.Sum64() % uint64(c.numBuckets))

	// i2 = i1 XOR hash(fingerprint), with bounds check
	h2 := fnv.New64a()
	binary.Write(h2, binary.LittleEndian, uint16(fp))
	fpHash := int(h2.Sum64() % uint64(c.numBuckets))
	i2 := (i1 ^ fpHash) % c.numBuckets
	if i2 < 0 {
		i2 = -i2
	}

	return i1, i2
}

// Add inserts a key into the filter. Returns false if filter is full.
func (c *CuckooFilter) Add(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	fp := c.getFingerprint(key)
	i1, i2 := c.getBucketIndices(key, fp)

	// Try to insert in bucket 1
	if len(c.buckets[i1]) < c.bucketSize {
		c.buckets[i1] = append(c.buckets[i1], fp)
		c.count++
		return true
	}

	// Try to insert in bucket 2
	if len(c.buckets[i2]) < c.bucketSize {
		c.buckets[i2] = append(c.buckets[i2], fp)
		c.count++
		return true
	}

	// Need to kick an existing entry
	rng := rand.New(rand.NewSource(rand.Int63()))
	currentIdx := i1
	if rng.Float32() < 0.5 {
		currentIdx = i2
	}
	currentFp := fp

	for kicks := 0; kicks < c.maxKicks; kicks++ {
		// Randomly select entry to kick
		kickIdx := rng.Intn(len(c.buckets[currentIdx]))
		kickedFp := c.buckets[currentIdx][kickIdx]
		c.buckets[currentIdx][kickIdx] = currentFp

		// Find alternate bucket for kicked entry
		h := fnv.New64a()
		binary.Write(h, binary.LittleEndian, uint16(kickedFp))
		fpHash := int(h.Sum64() % uint64(c.numBuckets))
		altIdx := (currentIdx ^ fpHash) % c.numBuckets // Ensure bounds

		// Try to insert kicked entry
		if len(c.buckets[altIdx]) < c.bucketSize {
			c.buckets[altIdx] = append(c.buckets[altIdx], kickedFp)
			c.count++
			return true
		}

		currentIdx = altIdx
		currentFp = kickedFp
	}

	return false // Filter is too full
}

// Contains checks if a key might be in the filter (may have false positives).
func (c *CuckooFilter) Contains(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fp := c.getFingerprint(key)
	i1, i2 := c.getBucketIndices(key, fp)

	// Check bucket 1
	for _, entry := range c.buckets[i1] {
		if entry == fp {
			return true
		}
	}

	// Check bucket 2
	for _, entry := range c.buckets[i2] {
		if entry == fp {
			return true
		}
	}

	return false
}

// Delete removes a key from the filter. Returns false if not found.
func (c *CuckooFilter) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	fp := c.getFingerprint(key)
	i1, i2 := c.getBucketIndices(key, fp)

	// Try to delete from bucket 1
	for i, entry := range c.buckets[i1] {
		if entry == fp {
			c.buckets[i1] = append(c.buckets[i1][:i], c.buckets[i1][i+1:]...)
			c.count--
			return true
		}
	}

	// Try to delete from bucket 2
	for i, entry := range c.buckets[i2] {
		if entry == fp {
			c.buckets[i2] = append(c.buckets[i2][:i], c.buckets[i2][i+1:]...)
			c.count--
			return true
		}
	}

	return false
}

// Count returns the number of items in the filter.
func (c *CuckooFilter) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.count
}

// LoadFactor returns the current load factor.
func (c *CuckooFilter) LoadFactor() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	totalSlots := c.numBuckets * c.bucketSize
	return float64(c.count) / float64(totalSlots)
}

// ============================================================================
// Product Quantization - O(1) compressed similarity search
// ============================================================================

// ProductQuantizer compresses high-dimensional vectors using codebook-based quantization.
// Achieves ~32x compression with minimal accuracy loss.
type ProductQuantizer struct {
	numSubvectors int         // M: number of subvectors
	codeSize      int         // K: number of centroids per subvector (usually 256)
	dimension     int         // D: original vector dimension
	subDim        int         // D/M: dimension per subvector
	codebooks     [][][]float32 // M codebooks, each with K centroids of subDim dimensions
	trained       bool
	mu            sync.RWMutex
}

// NewProductQuantizer creates a new Product Quantizer.
// dimension: original embedding dimension (must be divisible by numSubvectors)
// numSubvectors: number of segments to split vector into (typically 8-16)
// codeSize: number of centroids per segment (typically 256 for 8-bit codes)
func NewProductQuantizer(dimension, numSubvectors, codeSize int) *ProductQuantizer {
	if dimension%numSubvectors != 0 {
		// Adjust dimension to be divisible
		numSubvectors = gcd(dimension, numSubvectors)
	}

	subDim := dimension / numSubvectors

	// Initialize empty codebooks (will be trained later)
	codebooks := make([][][]float32, numSubvectors)
	for m := 0; m < numSubvectors; m++ {
		codebooks[m] = make([][]float32, codeSize)
		for k := 0; k < codeSize; k++ {
			codebooks[m][k] = make([]float32, subDim)
		}
	}

	return &ProductQuantizer{
		numSubvectors: numSubvectors,
		codeSize:      codeSize,
		dimension:     dimension,
		subDim:        subDim,
		codebooks:     codebooks,
		trained:       false,
	}
}

// NewProductQuantizerDefault creates a PQ for 384-dim embeddings with 8-bit codes.
func NewProductQuantizerDefault(dimension int) *ProductQuantizer {
	// Split into 8 subvectors, 256 centroids each = 8 bytes per vector
	numSubvectors := 8
	if dimension < 8 {
		numSubvectors = dimension
	}
	return NewProductQuantizer(dimension, numSubvectors, 256)
}

// gcd computes greatest common divisor.
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// Train builds codebooks from training vectors using k-means clustering.
// This is a simplified version - production would use more sophisticated clustering.
func (pq *ProductQuantizer) Train(vectors [][]float32, iterations int) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(vectors) == 0 {
		return
	}

	rng := rand.New(rand.NewSource(42))

	// Train each subvector codebook independently
	for m := 0; m < pq.numSubvectors; m++ {
		// Extract subvectors for this segment
		subvectors := make([][]float32, len(vectors))
		startIdx := m * pq.subDim
		endIdx := startIdx + pq.subDim

		for i, vec := range vectors {
			if len(vec) >= endIdx {
				subvectors[i] = vec[startIdx:endIdx]
			} else {
				subvectors[i] = make([]float32, pq.subDim)
			}
		}

		// Initialize centroids randomly
		perm := rng.Perm(len(subvectors))
		for k := 0; k < pq.codeSize && k < len(subvectors); k++ {
			copy(pq.codebooks[m][k], subvectors[perm[k]])
		}

		// K-means iterations
		for iter := 0; iter < iterations; iter++ {
			// Assign vectors to nearest centroid
			assignments := make([]int, len(subvectors))
			for i, sv := range subvectors {
				minDist := float32(math.MaxFloat32)
				for k := 0; k < pq.codeSize; k++ {
					dist := pq.l2Distance(sv, pq.codebooks[m][k])
					if dist < minDist {
						minDist = dist
						assignments[i] = k
					}
				}
			}

			// Update centroids
			counts := make([]int, pq.codeSize)
			newCentroids := make([][]float32, pq.codeSize)
			for k := 0; k < pq.codeSize; k++ {
				newCentroids[k] = make([]float32, pq.subDim)
			}

			for i, sv := range subvectors {
				k := assignments[i]
				counts[k]++
				for d := 0; d < pq.subDim; d++ {
					newCentroids[k][d] += sv[d]
				}
			}

			for k := 0; k < pq.codeSize; k++ {
				if counts[k] > 0 {
					for d := 0; d < pq.subDim; d++ {
						pq.codebooks[m][k][d] = newCentroids[k][d] / float32(counts[k])
					}
				}
			}
		}
	}

	pq.trained = true
}

// l2Distance computes squared L2 distance.
func (pq *ProductQuantizer) l2Distance(a, b []float32) float32 {
	var sum float32
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return sum
}

// Encode compresses a vector to M bytes (one byte per subvector when codeSize=256).
func (pq *ProductQuantizer) Encode(vector []float32) []uint8 {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	codes := make([]uint8, pq.numSubvectors)

	for m := 0; m < pq.numSubvectors; m++ {
		startIdx := m * pq.subDim
		endIdx := startIdx + pq.subDim
		
		subvector := vector[startIdx:endIdx]

		// Find nearest centroid
		minDist := float32(math.MaxFloat32)
		nearestCode := 0

		for k := 0; k < pq.codeSize; k++ {
			dist := pq.l2Distance(subvector, pq.codebooks[m][k])
			if dist < minDist {
				minDist = dist
				nearestCode = k
			}
		}

		codes[m] = uint8(nearestCode)
	}

	return codes
}

// Decode reconstructs an approximate vector from codes.
func (pq *ProductQuantizer) Decode(codes []uint8) []float32 {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	vector := make([]float32, pq.dimension)

	for m := 0; m < pq.numSubvectors && m < len(codes); m++ {
		startIdx := m * pq.subDim
		code := int(codes[m])
		if code < pq.codeSize {
			copy(vector[startIdx:startIdx+pq.subDim], pq.codebooks[m][code])
		}
	}

	return vector
}

// AsymmetricDistance computes distance between query vector and compressed code.
// This is faster than decode + distance for single queries.
func (pq *ProductQuantizer) AsymmetricDistance(query []float32, codes []uint8) float32 {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	var totalDist float32

	for m := 0; m < pq.numSubvectors && m < len(codes); m++ {
		startIdx := m * pq.subDim
		subQuery := query[startIdx : startIdx+pq.subDim]
		code := int(codes[m])
		
		if code < pq.codeSize {
			totalDist += pq.l2Distance(subQuery, pq.codebooks[m][code])
		}
	}

	return totalDist
}

// PrecomputeDistanceTable creates a lookup table for batch queries.
// table[m][k] = distance from query subvector m to centroid k
func (pq *ProductQuantizer) PrecomputeDistanceTable(query []float32) [][]float32 {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	table := make([][]float32, pq.numSubvectors)

	for m := 0; m < pq.numSubvectors; m++ {
		table[m] = make([]float32, pq.codeSize)
		startIdx := m * pq.subDim
		subQuery := query[startIdx : startIdx+pq.subDim]

		for k := 0; k < pq.codeSize; k++ {
			table[m][k] = pq.l2Distance(subQuery, pq.codebooks[m][k])
		}
	}

	return table
}

// DistanceWithTable computes distance using precomputed table (O(M) lookups).
func (pq *ProductQuantizer) DistanceWithTable(table [][]float32, codes []uint8) float32 {
	var totalDist float32
	for m := 0; m < len(codes) && m < len(table); m++ {
		code := int(codes[m])
		if code < len(table[m]) {
			totalDist += table[m][code]
		}
	}
	return totalDist
}

// CompressionRatio returns the compression ratio achieved.
func (pq *ProductQuantizer) CompressionRatio() float64 {
	originalBytes := pq.dimension * 4 // float32 = 4 bytes
	compressedBytes := pq.numSubvectors // 1 byte per subvector code
	return float64(originalBytes) / float64(compressedBytes)
}

// ============================================================================
// MinHash - O(1) Jaccard similarity estimation
// ============================================================================

// MinHash provides space-efficient similarity estimation using min-wise hashing.
// Used for finding semantically similar strategies without full embedding comparison.
type MinHash struct {
	numHashes   int
	hashSeeds   []uint64
	mu          sync.RWMutex
}

// MinHashSignature is a compact representation for similarity comparison.
type MinHashSignature []uint64

// NewMinHash creates a MinHash hasher with the specified number of hash functions.
// More hashes = better accuracy but larger signatures.
// Recommended: 100-200 hashes for good balance.
func NewMinHash(numHashes int) *MinHash {
	rng := rand.New(rand.NewSource(42))
	seeds := make([]uint64, numHashes)
	for i := range seeds {
		seeds[i] = rng.Uint64()
	}

	return &MinHash{
		numHashes: numHashes,
		hashSeeds: seeds,
	}
}

// NewMinHashDefault creates a MinHash with 128 hash functions.
func NewMinHashDefault() *MinHash {
	return NewMinHash(128)
}

// computeHash computes a hash for a token with a given seed.
func (m *MinHash) computeHash(token string, seed uint64) uint64 {
	h := fnv.New64a()
	h.Write([]byte(token))
	hash := h.Sum64()
	// Combine with seed using XOR and multiply
	return hash ^ (seed * 0x517cc1b727220a95)
}

// ComputeSignature creates a MinHash signature from a set of tokens.
func (m *MinHash) ComputeSignature(tokens []string) MinHashSignature {
	m.mu.RLock()
	defer m.mu.RUnlock()

	signature := make(MinHashSignature, m.numHashes)

	// Initialize to max values
	for i := range signature {
		signature[i] = math.MaxUint64
	}

	// For each token, update each hash function's minimum
	for _, token := range tokens {
		for i, seed := range m.hashSeeds {
			hash := m.computeHash(token, seed)
			if hash < signature[i] {
				signature[i] = hash
			}
		}
	}

	return signature
}

// ComputeSignatureFromText tokenizes text and computes signature.
func (m *MinHash) ComputeSignatureFromText(text string) MinHashSignature {
	// Simple tokenization: split by whitespace and lowercase
	tokens := tokenize(text)
	return m.ComputeSignature(tokens)
}

// tokenize splits text into lowercase tokens.
func tokenize(text string) []string {
	words := make([]string, 0)
	current := make([]byte, 0, 32)
	
	for i := 0; i < len(text); i++ {
		c := text[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			if c >= 'A' && c <= 'Z' {
				c += 32 // lowercase
			}
			current = append(current, c)
		} else if len(current) > 0 {
			words = append(words, string(current))
			current = current[:0]
		}
	}
	
	if len(current) > 0 {
		words = append(words, string(current))
	}
	
	return words
}

// EstimateSimilarity estimates Jaccard similarity between two signatures.
// Returns a value between 0 (no overlap) and 1 (identical).
func (m *MinHash) EstimateSimilarity(sig1, sig2 MinHashSignature) float64 {
	if len(sig1) != len(sig2) || len(sig1) == 0 {
		return 0
	}

	matches := 0
	for i := range sig1 {
		if sig1[i] == sig2[i] {
			matches++
		}
	}

	return float64(matches) / float64(len(sig1))
}

// FindSimilar finds items with similarity >= threshold.
func (m *MinHash) FindSimilar(query MinHashSignature, candidates map[string]MinHashSignature, threshold float64) []string {
	var results []string

	for id, sig := range candidates {
		if m.EstimateSimilarity(query, sig) >= threshold {
			results = append(results, id)
		}
	}

	return results
}

// ============================================================================
// LSH for MinHash (band-based) - O(1) expected similar item discovery
// ============================================================================

// MinHashLSH provides locality-sensitive hashing for MinHash signatures.
// Splits signature into bands and hashes each band for fast candidate retrieval.
type MinHashLSH struct {
	numBands  int
	rowsPerBand int
	buckets   []map[uint64][]string // One hash table per band
	mu        sync.RWMutex
}

// NewMinHashLSH creates an LSH index for MinHash signatures.
// threshold: target similarity threshold (0.5-0.9 typical)
// numHashes: number of hashes in MinHash signatures
func NewMinHashLSH(threshold float64, numHashes int) *MinHashLSH {
	// Optimal bands b and rows r: b*r = numHashes, threshold ≈ (1/b)^(1/r)
	// We solve for b given threshold and numHashes
	bestBands := 1
	minError := 1.0

	for b := 1; b <= numHashes; b++ {
		if numHashes%b != 0 {
			continue
		}
		r := numHashes / b
		predictedThreshold := math.Pow(1.0/float64(b), 1.0/float64(r))
		error := math.Abs(predictedThreshold - threshold)
		if error < minError {
			minError = error
			bestBands = b
		}
	}

	rowsPerBand := numHashes / bestBands
	if rowsPerBand == 0 {
		rowsPerBand = 1
		bestBands = numHashes
	}

	buckets := make([]map[uint64][]string, bestBands)
	for i := range buckets {
		buckets[i] = make(map[uint64][]string)
	}

	return &MinHashLSH{
		numBands:    bestBands,
		rowsPerBand: rowsPerBand,
		buckets:     buckets,
	}
}

// computeBandHash computes a hash for a band of the signature.
func (l *MinHashLSH) computeBandHash(sig MinHashSignature, bandIdx int) uint64 {
	start := bandIdx * l.rowsPerBand
	end := start + l.rowsPerBand
	if end > len(sig) {
		end = len(sig)
	}

	h := fnv.New64a()
	for i := start; i < end; i++ {
		binary.Write(h, binary.LittleEndian, sig[i])
	}
	return h.Sum64()
}

// Add inserts a signature into the LSH index.
func (l *MinHashLSH) Add(id string, sig MinHashSignature) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for b := 0; b < l.numBands; b++ {
		hash := l.computeBandHash(sig, b)
		l.buckets[b][hash] = append(l.buckets[b][hash], id)
	}
}

// Query finds candidate similar items (may have false positives).
func (l *MinHashLSH) Query(sig MinHashSignature) []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	candidateSet := make(map[string]int)

	for b := 0; b < l.numBands; b++ {
		hash := l.computeBandHash(sig, b)
		for _, id := range l.buckets[b][hash] {
			candidateSet[id]++
		}
	}

	// Sort by frequency (more band matches = more likely similar)
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

	results := make([]string, len(candidates))
	for i, c := range candidates {
		results[i] = c.id
	}
	return results
}

// Remove deletes a signature from the index.
func (l *MinHashLSH) Remove(id string, sig MinHashSignature) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for b := 0; b < l.numBands; b++ {
		hash := l.computeBandHash(sig, b)
		bucket := l.buckets[b][hash]
		for i, storedID := range bucket {
			if storedID == id {
				l.buckets[b][hash] = append(bucket[:i], bucket[i+1:]...)
				break
			}
		}
	}
}
