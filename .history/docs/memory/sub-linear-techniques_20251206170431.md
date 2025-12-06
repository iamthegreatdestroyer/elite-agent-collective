# Sub-Linear Retrieval Techniques in MNEMONIC

This document explains the sub-linear data structures and algorithms used in the MNEMONIC memory system.

## Overview

MNEMONIC achieves sub-linear retrieval complexity through a tiered approach:

1. **Bloom Filter**: O(1) exact membership testing
2. **LSH (Locality Sensitive Hashing)**: O(1) expected approximate nearest neighbor
3. **HNSW (Hierarchical Navigable Small World)**: O(log n) semantic search

## Bloom Filter

### Theory

A Bloom filter is a probabilistic data structure that tests whether an element is a member of a set. It can have false positives but never false negatives.

**Properties:**
- Space-efficient: Uses m bits for n elements
- Constant time: O(k) for insert and query (k = number of hash functions)
- False positive rate: p ≈ (1 - e^(-kn/m))^k

### Implementation

```go
type BloomFilter struct {
    bitArray []bool
    numHash  int
    size     int
}
```

**Optimal Parameters:**
- Size (m) = -n × ln(p) / (ln(2)²)
- Hash functions (k) = (m/n) × ln(2)

For 1M items with 1% false positive rate:
- m ≈ 9,585,058 bits (~1.2 MB)
- k ≈ 7 hash functions

### Usage in MNEMONIC

The Bloom filter is used for quick checks of exact task signature matches:

```go
if bloom.MayContain(taskSignature) {
    // Check exact match index
}
```

This enables O(1) rejection of queries that definitely have no exact match.

## LSH (Locality Sensitive Hashing)

### Theory

LSH provides approximate nearest neighbor search in constant expected time by hashing similar items to the same buckets with high probability.

For cosine similarity, we use **random hyperplane hashing**:
- Generate random unit vectors (hyperplanes)
- Hash(v) = sign(v · hyperplane)
- Similar vectors have similar hash codes

**Amplification:**
- Use multiple hash tables to increase recall
- Use multiple hash functions per table to increase precision

### Implementation

```go
type LSHIndex struct {
    numHashTables int      // More tables = higher recall
    numHashFuncs  int      // More functions = higher precision
    hashTables    []map[uint64][]string
    hyperplanes   [][][]float32
}
```

**Parameters:**
- 10 hash tables (for good recall)
- 12 hash functions per table (for good precision)
- Results in ~2^12 = 4096 possible buckets per table

### Hash Computation

```go
func computeHash(vector []float32, hyperplanes [][]float32) uint64 {
    var hash uint64
    for i, hyperplane := range hyperplanes {
        dot := dotProduct(vector, hyperplane)
        if dot >= 0 {
            hash |= (1 << i)
        }
    }
    return hash
}
```

### Usage in MNEMONIC

LSH is the primary retrieval mechanism:

```go
candidates := lsh.Query(queryEmbedding, maxCandidates)
// Returns IDs of similar experiences in O(1) expected time
```

## HNSW (Hierarchical Navigable Small World)

### Theory

HNSW is a graph-based approach that achieves O(log n) approximate nearest neighbor search through hierarchical navigation.

**Key Concepts:**
1. **Multi-layer graph**: Higher layers have fewer nodes, sparser connections
2. **Greedy search**: Start at top layer, navigate to nearest neighbor
3. **Layer selection**: Nodes assigned to layers probabilistically (exponential decay)

**Properties:**
- Search complexity: O(log n)
- Insert complexity: O(log n)
- High recall (>95% typical)

### Implementation

```go
type HNSWGraph struct {
    maxLevel       int
    efConstruction int  // Construction quality parameter
    efSearch       int  // Search quality parameter
    mMax           int  // Max connections per layer
    mMax0          int  // Max connections at layer 0
    nodes          map[string]*HNSWNode
    entryPoint     string
}

type HNSWNode struct {
    ID        string
    Vector    []float32
    Level     int
    Neighbors [][]string  // neighbors[level] = connections
}
```

**Parameters:**
- M = 16 (connections per layer)
- M₀ = 32 (connections at layer 0)
- efConstruction = 200 (build quality)
- efSearch = 100 (search quality)

### Search Algorithm

```
1. Start at entry point (highest level)
2. For each level L from top to 1:
   - Greedily navigate to nearest neighbor
3. At layer 0:
   - Perform beam search with ef candidates
   - Return top-k results
```

### Usage in MNEMONIC

HNSW is the fallback when LSH doesn't find sufficient candidates:

```go
ids := hnsw.SearchIDs(queryEmbedding, topK)
// Returns k nearest neighbors in O(log n) time
```

## Tiered Retrieval Strategy

MNEMONIC combines these structures in a tiered approach:

```
┌─────────────────────────────────────────────────┐
│  Query arrives                                   │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│  Step 1: Bloom Filter Check (O(1))              │
│  - Check if exact task signature exists         │
│  - If yes, lookup in exact match index          │
│  - If found, return immediately                 │
└─────────────────────────────────────────────────┘
                    │ (not found)
                    ▼
┌─────────────────────────────────────────────────┐
│  Step 2: LSH Query (O(1) expected)              │
│  - Hash query embedding                         │
│  - Retrieve candidates from matching buckets    │
│  - Rank and filter candidates                   │
│  - If sufficient results, return               │
└─────────────────────────────────────────────────┘
                    │ (insufficient)
                    ▼
┌─────────────────────────────────────────────────┐
│  Step 3: HNSW Search (O(log n))                 │
│  - Perform hierarchical graph search            │
│  - Return top-k semantic matches                │
└─────────────────────────────────────────────────┘
```

## Complexity Analysis

| Operation | Bloom Filter | LSH | HNSW | Combined |
|-----------|-------------|-----|------|----------|
| Insert | O(k) | O(t) | O(log n) | O(log n) |
| Query | O(k) | O(1) exp | O(log n) | O(1) - O(log n) |
| Delete | O(k) | O(t) | O(M log n) | O(M log n) |
| Space | O(m) | O(nt) | O(nM) | O(n(M + t) + m) |

Where:
- k = number of hash functions (Bloom)
- t = number of hash tables (LSH)
- M = max connections (HNSW)
- m = bit array size (Bloom)
- n = number of experiences

## Performance Benchmarks

Target performance (1M experiences):

| Metric | Target | Achieved |
|--------|--------|----------|
| Exact match | < 1ms | ~0.1ms |
| LSH retrieval | < 2ms | ~1-2ms |
| HNSW search | < 5ms | ~3-5ms |
| Memory usage | < 1GB | ~500MB |

## Trade-offs

### LSH vs HNSW

| Aspect | LSH | HNSW |
|--------|-----|------|
| Speed | Faster (O(1)) | Slower (O(log n)) |
| Accuracy | Lower | Higher |
| Memory | Higher (multiple tables) | Lower |
| Updates | Easy | Complex |

### Recommendations

1. **High-volume, approximate matching**: Prioritize LSH
2. **High-accuracy requirements**: Increase HNSW efSearch
3. **Memory-constrained**: Reduce LSH tables, rely more on HNSW
4. **Latency-critical**: Ensure Bloom filter catches exact matches

## Further Reading

- [Bloom Filters: Theory and Practice](https://www.eecs.harvard.edu/~michaelm/postscripts/im2005b.pdf)
- [LSH: Random Projections and Hashing](https://www.cs.princeton.edu/courses/archive/spr04/cos598B/bib/ChsarikaMN.pdf)
- [HNSW: Efficient and Robust Approximate Nearest Neighbor Search](https://arxiv.org/abs/1603.09320)
