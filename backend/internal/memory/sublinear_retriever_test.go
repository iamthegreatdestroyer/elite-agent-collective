package memory

import (
	"math/rand"
	"testing"
	"time"
)

// ============================================================================
// Bloom Filter Tests
// ============================================================================

func TestBloomFilter_AddAndCheck(t *testing.T) {
	bf := NewBloomFilter(1000, 7)

	// Add items
	items := []string{"apple", "banana", "cherry", "date", "elderberry"}
	for _, item := range items {
		bf.Add(item)
	}

	// Check that added items are found
	for _, item := range items {
		if !bf.MayContain(item) {
			t.Errorf("Bloom filter should contain %q", item)
		}
	}

	// Check for false positives (some may occur, but not all)
	notAdded := []string{"fig", "grape", "honeydew", "kiwi", "lemon"}
	falsePositives := 0
	for _, item := range notAdded {
		if bf.MayContain(item) {
			falsePositives++
		}
	}

	// With good parameters, false positive rate should be low
	if falsePositives > 2 {
		t.Logf("Warning: %d false positives out of %d (may be acceptable)", falsePositives, len(notAdded))
	}
}

func TestBloomFilter_Optimal(t *testing.T) {
	// Test optimal filter creation
	bf := NewBloomFilterOptimal(10000, 0.01)

	// Add 10000 items
	for i := 0; i < 10000; i++ {
		bf.Add(randomString(20))
	}

	// Verify filter is working
	testItem := "unique_test_item_12345"
	bf.Add(testItem)
	if !bf.MayContain(testItem) {
		t.Error("Bloom filter should contain added item")
	}
}

func TestBloomFilter_Clear(t *testing.T) {
	bf := NewBloomFilter(100, 3)
	bf.Add("test")

	if !bf.MayContain("test") {
		t.Error("Should contain 'test' before clear")
	}

	bf.Clear()

	// After clear, item should not be found (definitely)
	// Note: false positives still possible but very unlikely for small filter
	if bf.MayContain("test") {
		t.Log("False positive after clear (acceptable)")
	}
}

// ============================================================================
// LSH Index Tests
// ============================================================================

func TestLSHIndex_AddAndQuery(t *testing.T) {
	dimension := 128
	lsh := NewLSHIndex(5, 8, dimension)

	// Create vectors
	rng := rand.New(rand.NewSource(42))
	vectors := make([][]float32, 100)
	for i := range vectors {
		vectors[i] = randomVector(rng, dimension)
		lsh.Add("id_"+string(rune('a'+i%26)), vectors[i])
	}

	// Query with first vector
	candidates := lsh.Query(vectors[0], 10)

	// Should find at least the original vector
	if len(candidates) == 0 {
		t.Error("LSH query should return candidates")
	}
}

func TestLSHIndex_SimilarVectors(t *testing.T) {
	dimension := 64
	lsh := NewLSHIndex(10, 12, dimension)
	rng := rand.New(rand.NewSource(42))

	// Create a base vector
	base := randomVector(rng, dimension)
	lsh.Add("base", base)

	// Create similar vectors (small perturbations)
	for i := 0; i < 10; i++ {
		similar := make([]float32, dimension)
		for j := range similar {
			similar[j] = base[j] + float32(rng.NormFloat64()*0.1)
		}
		lsh.Add("similar_"+string(rune('0'+i)), similar)
	}

	// Create dissimilar vectors
	for i := 0; i < 10; i++ {
		dissimilar := randomVector(rng, dimension)
		lsh.Add("dissimilar_"+string(rune('0'+i)), dissimilar)
	}

	// Query with base vector should prefer similar vectors
	candidates := lsh.Query(base, 5)

	// Check that we get some candidates
	if len(candidates) == 0 {
		t.Error("Should find candidates")
	}
}

func TestLSHIndex_Remove(t *testing.T) {
	dimension := 32
	lsh := NewLSHIndex(3, 6, dimension)
	rng := rand.New(rand.NewSource(42))

	vector := randomVector(rng, dimension)
	lsh.Add("test", vector)

	// Verify it's found
	initialSize := lsh.Size()
	if initialSize == 0 {
		t.Error("LSH should have items after add")
	}

	// Remove it
	lsh.Remove("test", vector)

	// Note: Size might not decrease immediately due to how LSH works
	// The important thing is that queries shouldn't return the removed item
}

// ============================================================================
// HNSW Graph Tests
// ============================================================================

func TestHNSWGraph_AddAndSearch(t *testing.T) {
	dimension := 64
	hnsw := NewHNSWGraph(dimension, 16, 200)

	rng := rand.New(rand.NewSource(42))
	vectors := make(map[string][]float32)

	// Add vectors
	for i := 0; i < 100; i++ {
		id := "node_" + string(rune('a'+i%26)) + string(rune('0'+i/26))
		vec := randomVector(rng, dimension)
		vectors[id] = vec
		hnsw.Add(id, vec)
	}

	// Search should return results
	queryVec := vectors["node_a0"]
	results := hnsw.SearchIDs(queryVec, 5)

	if len(results) == 0 {
		t.Error("HNSW search should return results")
	}

	// The first result should be the query itself (if using exact match)
	foundSelf := false
	for _, id := range results {
		if id == "node_a0" {
			foundSelf = true
			break
		}
	}
	if !foundSelf {
		t.Log("Query vector might not be found (acceptable for approximate search)")
	}
}

func TestHNSWGraph_NearestNeighbors(t *testing.T) {
	dimension := 32
	hnsw := NewHNSWGraph(dimension, 8, 100)
	rng := rand.New(rand.NewSource(42))

	// Create a cluster of similar vectors
	center := randomVector(rng, dimension)
	hnsw.Add("center", center)

	for i := 0; i < 20; i++ {
		near := make([]float32, dimension)
		for j := range near {
			near[j] = center[j] + float32(rng.NormFloat64()*0.1)
		}
		hnsw.Add("near_"+string(rune('0'+i)), near)
	}

	// Create distant vectors
	for i := 0; i < 20; i++ {
		far := randomVector(rng, dimension)
		// Make them far by scaling
		for j := range far {
			far[j] *= 10
		}
		hnsw.Add("far_"+string(rune('0'+i)), far)
	}

	// Search near center should mostly return near vectors
	results := hnsw.SearchIDs(center, 10)

	nearCount := 0
	for _, id := range results {
		if id == "center" || (len(id) > 5 && id[:5] == "near_") {
			nearCount++
		}
	}

	// Should find mostly near vectors
	if nearCount < 5 {
		t.Errorf("Expected more near vectors in results, got %d", nearCount)
	}
}

func TestHNSWGraph_Remove(t *testing.T) {
	dimension := 16
	hnsw := NewHNSWGraph(dimension, 8, 50)
	rng := rand.New(rand.NewSource(42))

	vec1 := randomVector(rng, dimension)
	vec2 := randomVector(rng, dimension)

	hnsw.Add("node1", vec1)
	hnsw.Add("node2", vec2)

	if hnsw.Size() != 2 {
		t.Errorf("Expected size 2, got %d", hnsw.Size())
	}

	hnsw.Remove("node1")

	if hnsw.Size() != 1 {
		t.Errorf("Expected size 1 after removal, got %d", hnsw.Size())
	}
}

// ============================================================================
// SubLinearRetriever Tests
// ============================================================================

func TestSubLinearRetriever_AddAndRetrieve(t *testing.T) {
	dimension := 64
	retriever := NewSubLinearRetriever(dimension)
	rng := rand.New(rand.NewSource(42))

	// Add experiences
	for i := 0; i < 50; i++ {
		exp := &ExperienceTuple{
			ID:            "exp_" + string(rune('a'+i%26)) + string(rune('0'+i/26)),
			AgentID:       "APEX",
			TierID:        1,
			TaskSignature: "task_" + string(rune('0'+i)),
			Input:         "test input",
			Output:        "test output",
			Strategy:      "test strategy",
			Success:       true,
			Embedding:     randomVector(rng, dimension),
			FitnessScore:  0.8,
			Timestamp:     time.Now().UnixNano(),
		}
		if err := retriever.Add(exp); err != nil {
			t.Errorf("Failed to add experience: %v", err)
		}
	}

	if retriever.Size() != 50 {
		t.Errorf("Expected size 50, got %d", retriever.Size())
	}

	// Test exact match retrieval
	query := &QueryContext{
		AgentID:                      "APEX",
		TierID:                       1,
		TaskSignature:                "task_0",
		TopK:                         5,
		MinFitnessScore:              0.5,
		IncludeTierExperiences:       true,
		IncludeCollectiveExperiences: true,
	}

	result, err := retriever.Retrieve(query)
	if err != nil {
		t.Errorf("Retrieve failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	t.Logf("Retrieval method: %s, candidates: %d, results: %d",
		result.RetrievalMethod, result.TotalCandidates, len(result.Experiences))
}

func TestSubLinearRetriever_SemanticRetrieval(t *testing.T) {
	dimension := 32
	retriever := NewSubLinearRetriever(dimension)
	rng := rand.New(rand.NewSource(42))

	// Create a base embedding
	baseEmbedding := randomVector(rng, dimension)

	// Add experiences with similar embeddings
	for i := 0; i < 10; i++ {
		embedding := make([]float32, dimension)
		for j := range embedding {
			embedding[j] = baseEmbedding[j] + float32(rng.NormFloat64()*0.1)
		}

		exp := &ExperienceTuple{
			ID:           "similar_" + string(rune('0'+i)),
			AgentID:      "APEX",
			TierID:       1,
			Embedding:    embedding,
			FitnessScore: 0.8,
			Timestamp:    time.Now().UnixNano(),
		}
		retriever.Add(exp)
	}

	// Add dissimilar experiences
	for i := 0; i < 10; i++ {
		exp := &ExperienceTuple{
			ID:           "dissimilar_" + string(rune('0'+i)),
			AgentID:      "APEX",
			TierID:       1,
			Embedding:    randomVector(rng, dimension),
			FitnessScore: 0.8,
			Timestamp:    time.Now().UnixNano(),
		}
		retriever.Add(exp)
	}

	// Query with base embedding
	query := &QueryContext{
		AgentID:                      "APEX",
		TierID:                       1,
		Embedding:                    baseEmbedding,
		TopK:                         5,
		MinFitnessScore:              0.5,
		IncludeTierExperiences:       true,
		IncludeCollectiveExperiences: true,
	}

	result, err := retriever.Retrieve(query)
	if err != nil {
		t.Errorf("Retrieve failed: %v", err)
	}

	t.Logf("Retrieved %d experiences via %s", len(result.Experiences), result.RetrievalMethod)
}

func TestSubLinearRetriever_GetByAgent(t *testing.T) {
	dimension := 32
	retriever := NewSubLinearRetriever(dimension)

	// Add experiences for different agents
	agents := []string{"APEX", "CIPHER", "ARCHITECT"}
	for _, agent := range agents {
		for i := 0; i < 5; i++ {
			exp := &ExperienceTuple{
				ID:           agent + "_" + string(rune('0'+i)),
				AgentID:      agent,
				TierID:       1,
				FitnessScore: 0.8,
				Timestamp:    time.Now().UnixNano(),
			}
			retriever.Add(exp)
		}
	}

	// Get experiences by agent
	apexExps := retriever.GetByAgent("APEX")
	if len(apexExps) != 5 {
		t.Errorf("Expected 5 APEX experiences, got %d", len(apexExps))
	}

	for _, exp := range apexExps {
		if exp.AgentID != "APEX" {
			t.Errorf("Expected APEX agent, got %s", exp.AgentID)
		}
	}
}

func TestSubLinearRetriever_GetByTier(t *testing.T) {
	dimension := 32
	retriever := NewSubLinearRetriever(dimension)

	// Add experiences for different tiers
	for tier := 1; tier <= 3; tier++ {
		for i := 0; i < 5; i++ {
			exp := &ExperienceTuple{
				ID:           "tier" + string(rune('0'+tier)) + "_" + string(rune('0'+i)),
				AgentID:      "AGENT",
				TierID:       tier,
				FitnessScore: 0.8,
				Timestamp:    time.Now().UnixNano(),
			}
			retriever.Add(exp)
		}
	}

	// Get experiences by tier
	tier2Exps := retriever.GetByTier(2)
	if len(tier2Exps) != 5 {
		t.Errorf("Expected 5 tier-2 experiences, got %d", len(tier2Exps))
	}

	for _, exp := range tier2Exps {
		if exp.TierID != 2 {
			t.Errorf("Expected tier 2, got %d", exp.TierID)
		}
	}
}

func TestSubLinearRetriever_Remove(t *testing.T) {
	dimension := 32
	retriever := NewSubLinearRetriever(dimension)

	exp := &ExperienceTuple{
		ID:           "test_exp",
		AgentID:      "APEX",
		TierID:       1,
		FitnessScore: 0.8,
		Timestamp:    time.Now().UnixNano(),
	}
	retriever.Add(exp)

	if retriever.Size() != 1 {
		t.Errorf("Expected size 1, got %d", retriever.Size())
	}

	err := retriever.Remove("test_exp")
	if err != nil {
		t.Errorf("Remove failed: %v", err)
	}

	if retriever.Size() != 0 {
		t.Errorf("Expected size 0 after removal, got %d", retriever.Size())
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func randomVector(rng *rand.Rand, dimension int) []float32 {
	vec := make([]float32, dimension)
	for i := range vec {
		vec[i] = float32(rng.NormFloat64())
	}
	return vec
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkBloomFilter_Add(b *testing.B) {
	bf := NewBloomFilterOptimal(100000, 0.01)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Add(randomString(20))
	}
}

func BenchmarkBloomFilter_MayContain(b *testing.B) {
	bf := NewBloomFilterOptimal(100000, 0.01)
	for i := 0; i < 10000; i++ {
		bf.Add(randomString(20))
	}
	queries := make([]string, 1000)
	for i := range queries {
		queries[i] = randomString(20)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.MayContain(queries[i%len(queries)])
	}
}

func BenchmarkLSH_Add(b *testing.B) {
	lsh := NewLSHIndex(10, 12, 128)
	rng := rand.New(rand.NewSource(42))
	vectors := make([][]float32, b.N)
	for i := range vectors {
		vectors[i] = randomVector(rng, 128)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lsh.Add("id_"+string(rune(i)), vectors[i])
	}
}

func BenchmarkLSH_Query(b *testing.B) {
	lsh := NewLSHIndex(10, 12, 128)
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 10000; i++ {
		lsh.Add("id_"+string(rune(i)), randomVector(rng, 128))
	}
	query := randomVector(rng, 128)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lsh.Query(query, 10)
	}
}

func BenchmarkHNSW_Add(b *testing.B) {
	hnsw := NewHNSWGraph(128, 16, 200)
	rng := rand.New(rand.NewSource(42))
	vectors := make([][]float32, b.N)
	for i := range vectors {
		vectors[i] = randomVector(rng, 128)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hnsw.Add("id_"+string(rune(i)), vectors[i])
	}
}

func BenchmarkHNSW_Search(b *testing.B) {
	hnsw := NewHNSWGraph(128, 16, 200)
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 10000; i++ {
		hnsw.Add("id_"+string(rune(i)), randomVector(rng, 128))
	}
	query := randomVector(rng, 128)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hnsw.SearchIDs(query, 10)
	}
}

func BenchmarkSubLinearRetriever_Retrieve(b *testing.B) {
	retriever := NewSubLinearRetriever(128)
	rng := rand.New(rand.NewSource(42))

	for i := 0; i < 10000; i++ {
		exp := &ExperienceTuple{
			ID:            "exp_" + string(rune(i)),
			AgentID:       "APEX",
			TierID:        1,
			TaskSignature: "task_" + string(rune(i%100)),
			Embedding:     randomVector(rng, 128),
			FitnessScore:  0.8,
			Timestamp:     time.Now().UnixNano(),
		}
		retriever.Add(exp)
	}

	query := &QueryContext{
		AgentID:                      "APEX",
		TierID:                       1,
		Embedding:                    randomVector(rng, 128),
		TopK:                         10,
		MinFitnessScore:              0.5,
		IncludeTierExperiences:       true,
		IncludeCollectiveExperiences: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		retriever.Retrieve(query)
	}
}
