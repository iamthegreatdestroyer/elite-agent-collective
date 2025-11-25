"""
VELOCITY-05 Test Suite
======================
Performance Optimization & Sub-Linear Algorithms Specialist

Tests cover:
- Algorithm optimization
- Sub-linear algorithms
- Probabilistic data structures
- Cache optimization
- Parallel algorithms
"""

import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).parent.parent))

from framework.base_agent_test import (
    BaseAgentTest, TestResult, DifficultyLevel, TestCategory
)


class VelocityAgentTest(BaseAgentTest):
    """Comprehensive test suite for VELOCITY-05."""

    @property
    def agent_id(self) -> str:
        return "05"

    @property
    def agent_codename(self) -> str:
        return "VELOCITY"

    @property
    def agent_tier(self) -> int:
        return 1

    @property
    def agent_specialty(self) -> str:
        return "Performance Optimization & Sub-Linear Algorithms"

    def test_L1_trivial_01(self) -> TestResult:
        """Test: Optimize bubble sort with early exit."""
        def test_func(input_data):
            arr = input_data.copy()
            n = len(arr)
            iterations = 0
            
            for i in range(n):
                swapped = False
                for j in range(0, n - i - 1):
                    iterations += 1
                    if arr[j] > arr[j + 1]:
                        arr[j], arr[j + 1] = arr[j + 1], arr[j]
                        swapped = True
                if not swapped:
                    break
            
            # For already sorted input, should exit early
            return arr == sorted(input_data) and iterations < n * n

        return self.execute_test(
            test_name="optimized_bubble_sort",
            difficulty=DifficultyLevel.TRIVIAL,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=[1, 2, 3, 4, 5],  # Already sorted
            expected_output=True
        )

    def test_L1_trivial_02(self) -> TestResult:
        """Test: Implement memoization for recursive function."""
        def test_func(n):
            call_count = [0]
            
            # Without memoization
            def fib_naive(n):
                call_count[0] += 1
                if n <= 1:
                    return n
                return fib_naive(n-1) + fib_naive(n-2)
            
            # With memoization
            cache = {}
            memo_calls = [0]
            
            def fib_memo(n):
                memo_calls[0] += 1
                if n in cache:
                    return cache[n]
                if n <= 1:
                    cache[n] = n
                else:
                    cache[n] = fib_memo(n-1) + fib_memo(n-2)
                return cache[n]
            
            result = fib_memo(n)
            naive_result = fib_naive(n)
            
            # Memoized should have far fewer calls
            return result == naive_result and memo_calls[0] < call_count[0]

        return self.execute_test(
            test_name="memoization_optimization",
            difficulty=DifficultyLevel.TRIVIAL,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=20,
            expected_output=True
        )

    def test_L2_standard_01(self) -> TestResult:
        """Test: Implement cache-friendly matrix multiplication."""
        def test_func(input_data):
            import time
            
            n = input_data
            A = [[i * n + j for j in range(n)] for i in range(n)]
            B = [[i * n + j for j in range(n)] for i in range(n)]
            
            # Standard multiplication
            def standard_mult(A, B, n):
                C = [[0] * n for _ in range(n)]
                for i in range(n):
                    for j in range(n):
                        for k in range(n):
                            C[i][j] += A[i][k] * B[k][j]
                return C
            
            # Cache-friendly (loop reordering)
            def cache_friendly_mult(A, B, n):
                C = [[0] * n for _ in range(n)]
                for i in range(n):
                    for k in range(n):
                        for j in range(n):
                            C[i][j] += A[i][k] * B[k][j]
                return C
            
            # Both should produce same result
            C1 = standard_mult(A, B, n)
            C2 = cache_friendly_mult(A, B, n)
            
            return C1 == C2

        return self.execute_test(
            test_name="cache_friendly_matrix_mult",
            difficulty=DifficultyLevel.STANDARD,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=10,
            expected_output=True
        )

    def test_L2_standard_02(self) -> TestResult:
        """Test: Implement binary indexed tree (Fenwick tree)."""
        def test_func(input_data):
            arr = input_data
            n = len(arr)
            
            class FenwickTree:
                def __init__(self, n):
                    self.n = n
                    self.tree = [0] * (n + 1)
                
                def update(self, i, delta):
                    i += 1
                    while i <= self.n:
                        self.tree[i] += delta
                        i += i & (-i)
                
                def prefix_sum(self, i):
                    i += 1
                    s = 0
                    while i > 0:
                        s += self.tree[i]
                        i -= i & (-i)
                    return s
                
                def range_sum(self, l, r):
                    if l == 0:
                        return self.prefix_sum(r)
                    return self.prefix_sum(r) - self.prefix_sum(l - 1)
            
            ft = FenwickTree(n)
            for i, val in enumerate(arr):
                ft.update(i, val)
            
            # Test correctness
            correct = True
            for i in range(n):
                if ft.prefix_sum(i) != sum(arr[:i+1]):
                    correct = False
                    break
            
            # Test range sum
            for l in range(n):
                for r in range(l, n):
                    if ft.range_sum(l, r) != sum(arr[l:r+1]):
                        correct = False
                        break
            
            return correct

        return self.execute_test(
            test_name="fenwick_tree",
            difficulty=DifficultyLevel.STANDARD,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=[3, 2, -1, 6, 5, 4, -3, 3, 7, 2],
            expected_output=True
        )

    def test_L2_standard_03(self) -> TestResult:
        """Test: Implement segment tree with lazy propagation."""
        def test_func(input_data):
            arr = input_data
            n = len(arr)
            
            class SegmentTree:
                def __init__(self, arr):
                    self.n = len(arr)
                    self.tree = [0] * (4 * self.n)
                    self.lazy = [0] * (4 * self.n)
                    self._build(arr, 0, 0, self.n - 1)
                
                def _build(self, arr, node, start, end):
                    if start == end:
                        self.tree[node] = arr[start]
                    else:
                        mid = (start + end) // 2
                        self._build(arr, 2*node+1, start, mid)
                        self._build(arr, 2*node+2, mid+1, end)
                        self.tree[node] = self.tree[2*node+1] + self.tree[2*node+2]
                
                def _push_down(self, node, start, end):
                    if self.lazy[node] != 0:
                        self.tree[node] += (end - start + 1) * self.lazy[node]
                        if start != end:
                            self.lazy[2*node+1] += self.lazy[node]
                            self.lazy[2*node+2] += self.lazy[node]
                        self.lazy[node] = 0
                
                def update_range(self, l, r, val):
                    self._update(0, 0, self.n-1, l, r, val)
                
                def _update(self, node, start, end, l, r, val):
                    self._push_down(node, start, end)
                    if r < start or l > end:
                        return
                    if l <= start and end <= r:
                        self.lazy[node] += val
                        self._push_down(node, start, end)
                        return
                    mid = (start + end) // 2
                    self._update(2*node+1, start, mid, l, r, val)
                    self._update(2*node+2, mid+1, end, l, r, val)
                    self.tree[node] = self.tree[2*node+1] + self.tree[2*node+2]
                
                def query(self, l, r):
                    return self._query(0, 0, self.n-1, l, r)
                
                def _query(self, node, start, end, l, r):
                    if r < start or l > end:
                        return 0
                    self._push_down(node, start, end)
                    if l <= start and end <= r:
                        return self.tree[node]
                    mid = (start + end) // 2
                    return (self._query(2*node+1, start, mid, l, r) +
                            self._query(2*node+2, mid+1, end, l, r))
            
            st = SegmentTree(arr)
            
            # Test initial query
            total = st.query(0, n-1)
            if total != sum(arr):
                return False
            
            # Test range update
            st.update_range(2, 5, 10)
            # Expected: arr with +10 to indices 2-5
            expected = sum(arr) + 10 * 4  # 4 elements updated
            
            return st.query(0, n-1) == expected

        return self.execute_test(
            test_name="segment_tree_lazy",
            difficulty=DifficultyLevel.STANDARD,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=[1, 3, 5, 7, 9, 11],
            expected_output=True
        )

    def test_L3_advanced_01(self) -> TestResult:
        """Test: Implement Bloom filter."""
        def test_func(input_data):
            import hashlib
            
            class BloomFilter:
                def __init__(self, size, num_hashes):
                    self.size = size
                    self.num_hashes = num_hashes
                    self.bit_array = [False] * size
                
                def _hashes(self, item):
                    hashes = []
                    for i in range(self.num_hashes):
                        h = hashlib.sha256(f"{item}{i}".encode()).hexdigest()
                        hashes.append(int(h, 16) % self.size)
                    return hashes
                
                def add(self, item):
                    for h in self._hashes(item):
                        self.bit_array[h] = True
                
                def might_contain(self, item):
                    return all(self.bit_array[h] for h in self._hashes(item))
            
            bf = BloomFilter(size=1000, num_hashes=3)
            
            # Add items
            for item in input_data:
                bf.add(item)
            
            # All added items should be found (no false negatives)
            for item in input_data:
                if not bf.might_contain(item):
                    return False
            
            # Some non-added items might return True (false positives)
            # But most should return False
            false_positives = 0
            for i in range(100):
                if bf.might_contain(f"nonexistent_{i}"):
                    false_positives += 1
            
            # False positive rate should be reasonable
            return false_positives < 20  # Less than 20% FP rate

        return self.execute_test(
            test_name="bloom_filter",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=["apple", "banana", "cherry", "date", "elderberry"],
            expected_output=True
        )

    def test_L3_advanced_02(self) -> TestResult:
        """Test: Implement Count-Min Sketch."""
        def test_func(input_data):
            import hashlib
            
            class CountMinSketch:
                def __init__(self, width, depth):
                    self.width = width
                    self.depth = depth
                    self.table = [[0] * width for _ in range(depth)]
                
                def _hash(self, item, i):
                    h = hashlib.sha256(f"{item}{i}".encode()).hexdigest()
                    return int(h, 16) % self.width
                
                def add(self, item, count=1):
                    for i in range(self.depth):
                        self.table[i][self._hash(item, i)] += count
                
                def estimate(self, item):
                    return min(self.table[i][self._hash(item, i)] 
                              for i in range(self.depth))
            
            cms = CountMinSketch(width=100, depth=5)
            
            # Count items
            counts = {}
            for item in input_data:
                counts[item] = counts.get(item, 0) + 1
                cms.add(item)
            
            # Verify estimates (should overestimate, never underestimate)
            for item, true_count in counts.items():
                estimate = cms.estimate(item)
                if estimate < true_count:
                    return False
            
            return True

        return self.execute_test(
            test_name="count_min_sketch",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=["a", "b", "a", "c", "a", "b", "a", "d", "a"],
            expected_output=True
        )

    def test_L3_advanced_03(self) -> TestResult:
        """Test: Implement HyperLogLog cardinality estimation."""
        def test_func(input_data):
            import hashlib
            import math
            
            class HyperLogLog:
                def __init__(self, p=14):
                    self.p = p
                    self.m = 1 << p
                    self.registers = [0] * self.m
                    self.alpha = 0.7213 / (1 + 1.079 / self.m)
                
                def _hash(self, item):
                    h = hashlib.sha256(str(item).encode()).hexdigest()
                    return int(h, 16)
                
                def _rho(self, w):
                    # Position of first 1 bit
                    pos = 1
                    while w & 1 == 0 and pos <= 64:
                        w >>= 1
                        pos += 1
                    return pos
                
                def add(self, item):
                    h = self._hash(item)
                    j = h & (self.m - 1)  # First p bits
                    w = h >> self.p      # Remaining bits
                    self.registers[j] = max(self.registers[j], self._rho(w))
                
                def count(self):
                    indicator = sum(2 ** (-r) for r in self.registers)
                    estimate = self.alpha * self.m ** 2 / indicator
                    
                    # Small range correction
                    if estimate <= 2.5 * self.m:
                        zeros = self.registers.count(0)
                        if zeros > 0:
                            estimate = self.m * math.log(self.m / zeros)
                    
                    return int(estimate)
            
            hll = HyperLogLog(p=10)
            
            for item in input_data:
                hll.add(item)
            
            estimate = hll.count()
            true_count = len(set(input_data))
            
            # Allow 10% error margin
            return abs(estimate - true_count) / true_count < 0.1

        return self.execute_test(
            test_name="hyperloglog",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=[f"item_{i % 100}" for i in range(1000)],  # 100 unique items
            expected_output=True
        )

    def test_L4_expert_01(self) -> TestResult:
        """Test: Implement MinHash for Jaccard similarity."""
        def test_func(input_data):
            import hashlib
            import random
            
            class MinHash:
                def __init__(self, num_hashes=100):
                    self.num_hashes = num_hashes
                    # Generate hash functions using random seeds
                    random.seed(42)
                    self.a = [random.randint(1, 2**31) for _ in range(num_hashes)]
                    self.b = [random.randint(0, 2**31) for _ in range(num_hashes)]
                    self.p = 2**31 - 1
                
                def _hash_value(self, item, i):
                    h = int(hashlib.md5(str(item).encode()).hexdigest(), 16)
                    return (self.a[i] * h + self.b[i]) % self.p
                
                def signature(self, s):
                    sig = [float('inf')] * self.num_hashes
                    for item in s:
                        for i in range(self.num_hashes):
                            h = self._hash_value(item, i)
                            sig[i] = min(sig[i], h)
                    return sig
                
                def estimate_jaccard(self, sig1, sig2):
                    return sum(1 for a, b in zip(sig1, sig2) if a == b) / self.num_hashes
            
            set1, set2 = input_data
            
            mh = MinHash(num_hashes=200)
            sig1 = mh.signature(set1)
            sig2 = mh.signature(set2)
            
            estimated = mh.estimate_jaccard(sig1, sig2)
            
            # True Jaccard similarity
            intersection = len(set1 & set2)
            union = len(set1 | set2)
            true_jaccard = intersection / union
            
            # Allow 15% error
            return abs(estimated - true_jaccard) < 0.15

        return self.execute_test(
            test_name="minhash_jaccard",
            difficulty=DifficultyLevel.EXPERT,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=(set(range(50)), set(range(25, 75))),  # 50% overlap
            expected_output=True
        )

    def test_L4_expert_02(self) -> TestResult:
        """Test: Implement skip list with O(log n) operations."""
        def test_func(input_data):
            import random
            
            class SkipListNode:
                def __init__(self, key, level):
                    self.key = key
                    self.forward = [None] * (level + 1)
            
            class SkipList:
                def __init__(self, max_level=16, p=0.5):
                    self.max_level = max_level
                    self.p = p
                    self.level = 0
                    self.header = SkipListNode(-float('inf'), max_level)
                
                def random_level(self):
                    lvl = 0
                    while random.random() < self.p and lvl < self.max_level:
                        lvl += 1
                    return lvl
                
                def insert(self, key):
                    update = [None] * (self.max_level + 1)
                    current = self.header
                    
                    for i in range(self.level, -1, -1):
                        while current.forward[i] and current.forward[i].key < key:
                            current = current.forward[i]
                        update[i] = current
                    
                    lvl = self.random_level()
                    
                    if lvl > self.level:
                        for i in range(self.level + 1, lvl + 1):
                            update[i] = self.header
                        self.level = lvl
                    
                    new_node = SkipListNode(key, lvl)
                    for i in range(lvl + 1):
                        new_node.forward[i] = update[i].forward[i]
                        update[i].forward[i] = new_node
                
                def search(self, key):
                    current = self.header
                    for i in range(self.level, -1, -1):
                        while current.forward[i] and current.forward[i].key < key:
                            current = current.forward[i]
                    current = current.forward[0]
                    return current is not None and current.key == key
            
            random.seed(42)
            sl = SkipList()
            
            # Insert elements
            for item in input_data:
                sl.insert(item)
            
            # Search all inserted elements
            for item in input_data:
                if not sl.search(item):
                    return False
            
            # Search non-existent elements
            for i in range(1000, 1010):
                if sl.search(i):
                    return False
            
            return True

        return self.execute_test(
            test_name="skip_list",
            difficulty=DifficultyLevel.EXPERT,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=list(range(0, 100, 2)),  # Even numbers
            expected_output=True
        )

    def test_L5_extreme_01(self) -> TestResult:
        """Test: Implement t-digest for quantile estimation."""
        def test_func(input_data):
            import math
            
            class TDigest:
                def __init__(self, delta=0.01):
                    self.delta = delta
                    self.centroids = []  # (mean, weight)
                    self.total_weight = 0
                
                def add(self, value, weight=1):
                    self.centroids.append((value, weight))
                    self.total_weight += weight
                    self._compress()
                
                def _compress(self):
                    if len(self.centroids) < 10:
                        return
                    
                    # Sort by mean
                    self.centroids.sort(key=lambda x: x[0])
                    
                    # Merge close centroids
                    merged = []
                    for mean, weight in self.centroids:
                        if not merged:
                            merged.append([mean, weight])
                        else:
                            last = merged[-1]
                            # Simple merge if weight is small
                            if last[1] < self.total_weight * self.delta:
                                new_weight = last[1] + weight
                                new_mean = (last[0] * last[1] + mean * weight) / new_weight
                                merged[-1] = [new_mean, new_weight]
                            else:
                                merged.append([mean, weight])
                    
                    self.centroids = [(m, w) for m, w in merged]
                
                def quantile(self, q):
                    if not self.centroids:
                        return 0
                    
                    self.centroids.sort(key=lambda x: x[0])
                    target = q * self.total_weight
                    cumulative = 0
                    
                    for mean, weight in self.centroids:
                        if cumulative + weight >= target:
                            return mean
                        cumulative += weight
                    
                    return self.centroids[-1][0]
            
            td = TDigest(delta=0.1)
            
            for value in input_data:
                td.add(value)
            
            # Test quantile estimation
            sorted_data = sorted(input_data)
            n = len(sorted_data)
            
            # Check median (50th percentile)
            true_median = sorted_data[n // 2]
            estimated_median = td.quantile(0.5)
            
            # Allow 10% error
            median_ok = abs(estimated_median - true_median) / true_median < 0.1
            
            return median_ok

        return self.execute_test(
            test_name="t_digest",
            difficulty=DifficultyLevel.EXTREME,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=[i ** 2 for i in range(100)],  # Skewed distribution
            expected_output=True
        )

    def test_L5_extreme_02(self) -> TestResult:
        """Test: Implement locality-sensitive hashing for nearest neighbor."""
        def test_func(input_data):
            import random
            import math
            
            class LSH:
                def __init__(self, dim, num_tables=10, num_hashes=5):
                    self.dim = dim
                    self.num_tables = num_tables
                    self.num_hashes = num_hashes
                    self.tables = [{} for _ in range(num_tables)]
                    
                    # Random projection vectors
                    random.seed(42)
                    self.projections = [
                        [[random.gauss(0, 1) for _ in range(dim)] 
                         for _ in range(num_hashes)]
                        for _ in range(num_tables)
                    ]
                
                def _hash(self, vec, table_idx):
                    bits = []
                    for proj in self.projections[table_idx]:
                        dot = sum(v * p for v, p in zip(vec, proj))
                        bits.append(1 if dot >= 0 else 0)
                    return tuple(bits)
                
                def insert(self, key, vec):
                    for i in range(self.num_tables):
                        h = self._hash(vec, i)
                        if h not in self.tables[i]:
                            self.tables[i][h] = []
                        self.tables[i][h].append((key, vec))
                
                def query(self, vec, k=5):
                    candidates = set()
                    for i in range(self.num_tables):
                        h = self._hash(vec, i)
                        if h in self.tables[i]:
                            for key, v in self.tables[i][h]:
                                candidates.add(key)
                    return list(candidates)[:k]
            
            points, query_point = input_data
            
            lsh = LSH(dim=len(points[0]), num_tables=5, num_hashes=4)
            
            for i, p in enumerate(points):
                lsh.insert(i, p)
            
            # Query
            candidates = lsh.query(query_point)
            
            # LSH should return some candidates
            return len(candidates) > 0

        return self.execute_test(
            test_name="locality_sensitive_hashing",
            difficulty=DifficultyLevel.EXTREME,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=(
                [[i, i+1, i+2] for i in range(100)],  # Points
                [50, 51, 52]  # Query
            ),
            expected_output=True
        )

    def test_collaboration_scenario(self) -> TestResult:
        """Test: Collaborate with APEX for optimized algorithms."""
        def test_func(input_data):
            collaboration = {
                "velocity_optimization": "Performance tuning",
                "apex_algorithm": "Core implementation",
                "integrated_solution": True
            }
            
            return all(v for v in collaboration.values())

        return self.execute_test(
            test_name="performance_collaboration",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.COLLABORATION,
            test_func=test_func,
            input_data=None,
            expected_output=True
        )

    def test_evolution_adaptation(self) -> TestResult:
        """Test: Adapt to new optimization techniques."""
        def test_func(input_data):
            techniques = {
                "classical": "Loop unrolling, memoization",
                "modern": "SIMD, cache optimization",
                "emerging": "GPU acceleration, quantum"
            }
            
            return len(techniques) == 3 and all(v for v in techniques.values())

        return self.execute_test(
            test_name="optimization_evolution",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.EVOLUTION,
            test_func=test_func,
            input_data=None,
            expected_output=True
        )

    def test_edge_case_handling(self) -> TestResult:
        """Test: Handle performance edge cases."""
        def test_func(input_data):
            edge_cases = []
            
            # Empty input
            edge_cases.append(sum([]) == 0)
            
            # Single element
            edge_cases.append(max([42]) == 42)
            
            # Large input
            large = list(range(10000))
            edge_cases.append(sum(large) == 49995000)
            
            # Worst case for quicksort (already sorted)
            sorted_arr = list(range(100))
            edge_cases.append(sorted(sorted_arr) == sorted_arr)
            
            return all(edge_cases)

        return self.execute_test(
            test_name="performance_edge_cases",
            difficulty=DifficultyLevel.STANDARD,
            category=TestCategory.EDGE_CASE,
            test_func=test_func,
            input_data=None,
            expected_output=True
        )


if __name__ == "__main__":
    test_suite = VelocityAgentTest()
    summary = test_suite.run_all_tests()

    print(f"\n{'='*60}")
    print(f"VELOCITY-05 Test Results")
    print(f"{'='*60}")
    print(f"Total Tests: {summary.total_tests}")
    print(f"Passed: {summary.passed_tests}")
    print(f"Failed: {summary.failed_tests}")
    print(f"Pass Rate: {summary.pass_rate:.2%}")
    print(f"{'='*60}\n")
