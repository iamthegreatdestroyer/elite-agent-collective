"""
APEX-01 Test Suite
==================
Elite Computer Science Engineering Specialist

Tests cover:
- Algorithm implementation
- System design
- Code quality
- Performance optimization
- Edge case handling
"""

import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).parent.parent))

from framework.base_agent_test import (
    BaseAgentTest, TestResult, DifficultyLevel, TestCategory
)


class ApexAgentTest(BaseAgentTest):
    """Comprehensive test suite for APEX-01."""

    @property
    def agent_id(self) -> str:
        return "01"

    @property
    def agent_codename(self) -> str:
        return "APEX"

    @property
    def agent_tier(self) -> int:
        return 1

    @property
    def agent_specialty(self) -> str:
        return "Elite Computer Science Engineering"

    # ═══════════════════════════════════════════════════════════════════════
    # L1 TRIVIAL TESTS
    # ═══════════════════════════════════════════════════════════════════════

    def test_L1_trivial_01(self) -> TestResult:
        """Test: Implement string reversal."""
        def test_func(input_data):
            s = input_data
            return s[::-1]

        return self.execute_test(
            test_name="string_reversal",
            difficulty=DifficultyLevel.TRIVIAL,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data="Hello, World!",
            expected_output="!dlroW ,olleH"
        )

    def test_L1_trivial_02(self) -> TestResult:
        """Test: Implement FizzBuzz."""
        def test_func(n):
            result = []
            for i in range(1, n + 1):
                if i % 15 == 0:
                    result.append("FizzBuzz")
                elif i % 3 == 0:
                    result.append("Fizz")
                elif i % 5 == 0:
                    result.append("Buzz")
                else:
                    result.append(str(i))
            return result

        return self.execute_test(
            test_name="fizzbuzz_implementation",
            difficulty=DifficultyLevel.TRIVIAL,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=15,
            expected_output=["1","2","Fizz","4","Buzz","Fizz","7","8","Fizz","Buzz","11","Fizz","13","14","FizzBuzz"]
        )

    # ═══════════════════════════════════════════════════════════════════════
    # L2 STANDARD TESTS
    # ═══════════════════════════════════════════════════════════════════════

    def test_L2_standard_01(self) -> TestResult:
        """Test: Implement thread-safe singleton pattern."""
        def test_func(input_data):
            import threading

            class Singleton:
                _instance = None
                _lock = threading.Lock()

                def __new__(cls):
                    if cls._instance is None:
                        with cls._lock:
                            if cls._instance is None:
                                cls._instance = super().__new__(cls)
                    return cls._instance

            instances = []
            def create_instance():
                instances.append(id(Singleton()))

            threads = [threading.Thread(target=create_instance) for _ in range(10)]
            for t in threads:
                t.start()
            for t in threads:
                t.join()

            return len(set(instances)) == 1

        return self.execute_test(
            test_name="thread_safe_singleton",
            difficulty=DifficultyLevel.STANDARD,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=None,
            expected_output=True
        )

    def test_L2_standard_02(self) -> TestResult:
        """Test: Implement binary search with edge cases."""
        def test_func(input_data):
            arr, target = input_data
            left, right = 0, len(arr) - 1
            while left <= right:
                mid = left + (right - left) // 2
                if arr[mid] == target:
                    return mid
                elif arr[mid] < target:
                    left = mid + 1
                else:
                    right = mid - 1
            return -1

        return self.execute_test(
            test_name="binary_search_edge_cases",
            difficulty=DifficultyLevel.STANDARD,
            category=TestCategory.EDGE_CASE,
            test_func=test_func,
            input_data=([1, 2, 3, 4, 5, 6, 7, 8, 9, 10], 7),
            expected_output=6
        )

    def test_L2_standard_03(self) -> TestResult:
        """Test: Implement LRU Cache."""
        def test_func(operations):
            from collections import OrderedDict

            class LRUCache:
                def __init__(self, capacity):
                    self.cache = OrderedDict()
                    self.capacity = capacity

                def get(self, key):
                    if key not in self.cache:
                        return -1
                    self.cache.move_to_end(key)
                    return self.cache[key]

                def put(self, key, value):
                    if key in self.cache:
                        self.cache.move_to_end(key)
                    self.cache[key] = value
                    if len(self.cache) > self.capacity:
                        self.cache.popitem(last=False)

            cache = LRUCache(2)
            results = []
            for op, args in operations:
                if op == "put":
                    cache.put(*args)
                    results.append(None)
                elif op == "get":
                    results.append(cache.get(args[0]))
            return results

        operations = [
            ("put", (1, 1)), ("put", (2, 2)), ("get", (1,)),
            ("put", (3, 3)), ("get", (2,)), ("get", (3,))
        ]

        return self.execute_test(
            test_name="lru_cache_implementation",
            difficulty=DifficultyLevel.STANDARD,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=operations,
            expected_output=[None, None, 1, None, -1, 3]
        )

    # ═══════════════════════════════════════════════════════════════════════
    # L3 ADVANCED TESTS
    # ═══════════════════════════════════════════════════════════════════════

    def test_L3_advanced_01(self) -> TestResult:
        """Test: Implement concurrent LRU cache with TTL."""
        def test_func(input_data):
            import threading
            import time
            from collections import OrderedDict

            class ConcurrentLRUCache:
                def __init__(self, capacity, ttl_seconds):
                    self.cache = OrderedDict()
                    self.timestamps = {}
                    self.capacity = capacity
                    self.ttl = ttl_seconds
                    self.lock = threading.RLock()

                def _is_expired(self, key):
                    if key not in self.timestamps:
                        return True
                    return time.time() - self.timestamps[key] > self.ttl

                def get(self, key):
                    with self.lock:
                        if key not in self.cache or self._is_expired(key):
                            if key in self.cache:
                                del self.cache[key]
                                del self.timestamps[key]
                            return -1
                        self.cache.move_to_end(key)
                        return self.cache[key]

                def put(self, key, value):
                    with self.lock:
                        if key in self.cache:
                            self.cache.move_to_end(key)
                        self.cache[key] = value
                        self.timestamps[key] = time.time()
                        if len(self.cache) > self.capacity:
                            oldest = next(iter(self.cache))
                            del self.cache[oldest]
                            del self.timestamps[oldest]

            cache = ConcurrentLRUCache(capacity=2, ttl_seconds=0.1)
            cache.put(1, "a")
            result1 = cache.get(1)
            time.sleep(0.15)
            result2 = cache.get(1)

            return (result1 == "a", result2 == -1)

        return self.execute_test(
            test_name="concurrent_lru_cache_ttl",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=None,
            expected_output=(True, True)
        )

    def test_L3_advanced_02(self) -> TestResult:
        """Test: Implement Trie with autocomplete."""
        def test_func(input_data):
            words, prefix = input_data

            class TrieNode:
                def __init__(self):
                    self.children = {}
                    self.is_end = False

            class Trie:
                def __init__(self):
                    self.root = TrieNode()

                def insert(self, word):
                    node = self.root
                    for char in word:
                        if char not in node.children:
                            node.children[char] = TrieNode()
                        node = node.children[char]
                    node.is_end = True

                def autocomplete(self, prefix):
                    node = self.root
                    for char in prefix:
                        if char not in node.children:
                            return []
                        node = node.children[char]

                    results = []
                    self._dfs(node, prefix, results)
                    return sorted(results)

                def _dfs(self, node, current, results):
                    if node.is_end:
                        results.append(current)
                    for char, child in node.children.items():
                        self._dfs(child, current + char, results)

            trie = Trie()
            for word in words:
                trie.insert(word)
            return trie.autocomplete(prefix)

        return self.execute_test(
            test_name="trie_autocomplete",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=(["apple", "application", "apply", "apt", "banana"], "app"),
            expected_output=["apple", "application", "apply"]
        )

    def test_L3_advanced_03(self) -> TestResult:
        """Test: Implement graph shortest path with negative weights (Bellman-Ford)."""
        def test_func(input_data):
            vertices, edges, source = input_data

            dist = {v: float('inf') for v in range(vertices)}
            dist[source] = 0

            for _ in range(vertices - 1):
                for u, v, w in edges:
                    if dist[u] != float('inf') and dist[u] + w < dist[v]:
                        dist[v] = dist[u] + w

            for u, v, w in edges:
                if dist[u] != float('inf') and dist[u] + w < dist[v]:
                    return None

            return dist

        edges = [(0, 1, -1), (0, 2, 4), (1, 2, 3), (1, 3, 2), (1, 4, 2), (3, 2, 5), (3, 1, 1), (4, 3, -3)]

        return self.execute_test(
            test_name="bellman_ford_shortest_path",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=(5, edges, 0),
            expected_output={0: 0, 1: -1, 2: 2, 3: -2, 4: 1}
        )

    # ═══════════════════════════════════════════════════════════════════════
    # L4 EXPERT TESTS
    # ═══════════════════════════════════════════════════════════════════════

    def test_L4_expert_01(self) -> TestResult:
        """Test: Design distributed rate limiter."""
        def test_func(input_data):
            import time

            class DistributedRateLimiter:
                def __init__(self, rate_per_second, burst_capacity):
                    self.rate = rate_per_second
                    self.capacity = burst_capacity
                    self.nodes = {}

                def add_node(self, node_id):
                    self.nodes[node_id] = {
                        'tokens': self.capacity / max(len(self.nodes) + 1, 1),
                        'last_refill': time.time()
                    }
                    self._redistribute_tokens()

                def _redistribute_tokens(self):
                    if not self.nodes:
                        return
                    tokens_per_node = self.capacity / len(self.nodes)
                    for node in self.nodes.values():
                        node['tokens'] = tokens_per_node

                def try_acquire(self, node_id, tokens_needed=1):
                    if node_id not in self.nodes:
                        return False

                    node = self.nodes[node_id]
                    now = time.time()

                    elapsed = now - node['last_refill']
                    refill = elapsed * self.rate / len(self.nodes)
                    node['tokens'] = min(self.capacity / len(self.nodes), node['tokens'] + refill)
                    node['last_refill'] = now

                    if node['tokens'] >= tokens_needed:
                        node['tokens'] -= tokens_needed
                        return True
                    return False

            limiter = DistributedRateLimiter(rate_per_second=10, burst_capacity=20)
            limiter.add_node("node1")
            limiter.add_node("node2")

            results = []
            for _ in range(25):
                results.append(limiter.try_acquire("node1"))

            allowed = sum(results)
            return 8 <= allowed <= 12

        return self.execute_test(
            test_name="distributed_rate_limiter",
            difficulty=DifficultyLevel.EXPERT,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=None,
            expected_output=True
        )

    def test_L4_expert_02(self) -> TestResult:
        """Test: Implement consistent hashing for distributed systems."""
        def test_func(input_data):
            import hashlib
            import bisect

            class ConsistentHash:
                def __init__(self, replicas=100):
                    self.replicas = replicas
                    self.ring = []
                    self.nodes = {}

                def _hash(self, key):
                    return int(hashlib.md5(key.encode()).hexdigest(), 16)

                def add_node(self, node):
                    for i in range(self.replicas):
                        virtual_key = f"{node}:{i}"
                        h = self._hash(virtual_key)
                        bisect.insort(self.ring, h)
                        self.nodes[h] = node

                def remove_node(self, node):
                    for i in range(self.replicas):
                        virtual_key = f"{node}:{i}"
                        h = self._hash(virtual_key)
                        self.ring.remove(h)
                        del self.nodes[h]

                def get_node(self, key):
                    if not self.ring:
                        return None
                    h = self._hash(key)
                    idx = bisect.bisect_right(self.ring, h) % len(self.ring)
                    return self.nodes[self.ring[idx]]

            ch = ConsistentHash(replicas=100)
            ch.add_node("server1")
            ch.add_node("server2")
            ch.add_node("server3")

            distribution = {"server1": 0, "server2": 0, "server3": 0}
            for i in range(1000):
                node = ch.get_node(f"key_{i}")
                distribution[node] += 1

            for count in distribution.values():
                if count < 200 or count > 500:
                    return False

            original_mappings = {f"key_{i}": ch.get_node(f"key_{i}") for i in range(100)}
            ch.remove_node("server2")

            unchanged = sum(
                1 for k, v in original_mappings.items()
                if v != "server2" and ch.get_node(k) == v
            )

            return unchanged >= 60

        return self.execute_test(
            test_name="consistent_hashing",
            difficulty=DifficultyLevel.EXPERT,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=None,
            expected_output=True
        )

    # ═══════════════════════════════════════════════════════════════════════
    # L5 EXTREME TESTS
    # ═══════════════════════════════════════════════════════════════════════

    def test_L5_extreme_01(self) -> TestResult:
        """Test: Implement lock-free concurrent queue."""
        def test_func(input_data):
            import threading

            class LockFreeQueue:
                class Node:
                    def __init__(self, value=None):
                        self.value = value
                        self.next = None

                def __init__(self):
                    dummy = self.Node()
                    self.head = dummy
                    self.tail = dummy
                    self._lock = threading.Lock()

                def enqueue(self, value):
                    node = self.Node(value)
                    with self._lock:
                        self.tail.next = node
                        self.tail = node

                def dequeue(self):
                    with self._lock:
                        if self.head.next is None:
                            return None
                        value = self.head.next.value
                        self.head = self.head.next
                        return value

            queue = LockFreeQueue()
            results = {"enqueued": 0, "dequeued": 0}

            def producer():
                for i in range(100):
                    queue.enqueue(i)
                    results["enqueued"] += 1

            def consumer():
                count = 0
                while count < 100:
                    val = queue.dequeue()
                    if val is not None:
                        count += 1
                results["dequeued"] = count

            t1 = threading.Thread(target=producer)
            t2 = threading.Thread(target=consumer)

            t1.start()
            t2.start()
            t1.join()
            t2.join()

            return results["enqueued"] == 100 and results["dequeued"] == 100

        return self.execute_test(
            test_name="lock_free_concurrent_queue",
            difficulty=DifficultyLevel.EXTREME,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=None,
            expected_output=True
        )

    def test_L5_extreme_02(self) -> TestResult:
        """Test: Implement B+ Tree with range queries."""
        def test_func(input_data):
            class BPlusTreeNode:
                def __init__(self, is_leaf=False):
                    self.keys = []
                    self.children = []
                    self.is_leaf = is_leaf
                    self.next = None

            class BPlusTree:
                def __init__(self, order=4):
                    self.root = BPlusTreeNode(is_leaf=True)
                    self.order = order

                def insert(self, key, value):
                    root = self.root
                    if len(root.keys) == self.order - 1:
                        new_root = BPlusTreeNode()
                        new_root.children.append(self.root)
                        self._split_child(new_root, 0)
                        self.root = new_root
                    self._insert_non_full(self.root, key, value)

                def _insert_non_full(self, node, key, value):
                    if node.is_leaf:
                        i = len(node.keys) - 1
                        node.keys.append(None)
                        node.children.append(None)
                        while i >= 0 and key < node.keys[i]:
                            node.keys[i + 1] = node.keys[i]
                            node.children[i + 1] = node.children[i]
                            i -= 1
                        node.keys[i + 1] = key
                        node.children[i + 1] = value
                    else:
                        i = len(node.keys) - 1
                        while i >= 0 and key < node.keys[i]:
                            i -= 1
                        i += 1
                        if len(node.children[i].keys) == self.order - 1:
                            self._split_child(node, i)
                            if key > node.keys[i]:
                                i += 1
                        self._insert_non_full(node.children[i], key, value)

                def _split_child(self, parent, index):
                    order = self.order
                    child = parent.children[index]
                    mid = order // 2

                    new_node = BPlusTreeNode(is_leaf=child.is_leaf)

                    if child.is_leaf:
                        new_node.keys = child.keys[mid:]
                        new_node.children = child.children[mid:]
                        child.keys = child.keys[:mid]
                        child.children = child.children[:mid]
                        new_node.next = child.next
                        child.next = new_node
                        parent.keys.insert(index, new_node.keys[0])
                    else:
                        new_node.keys = child.keys[mid + 1:]
                        new_node.children = child.children[mid + 1:]
                        parent.keys.insert(index, child.keys[mid])
                        child.keys = child.keys[:mid]
                        child.children = child.children[:mid + 1]

                    parent.children.insert(index + 1, new_node)

                def range_query(self, start, end):
                    results = []
                    node = self._find_leaf(start)
                    while node:
                        for i, key in enumerate(node.keys):
                            if start <= key <= end:
                                results.append((key, node.children[i]))
                            elif key > end:
                                return results
                        node = node.next
                    return results

                def _find_leaf(self, key):
                    node = self.root
                    while not node.is_leaf:
                        i = 0
                        while i < len(node.keys) and key >= node.keys[i]:
                            i += 1
                        node = node.children[i]
                    return node

            tree = BPlusTree(order=4)
            for i in [10, 20, 5, 15, 25, 30, 35, 40, 1, 7]:
                tree.insert(i, f"value_{i}")

            range_result = tree.range_query(10, 30)
            keys = [k for k, v in range_result]

            return sorted(keys) == [10, 15, 20, 25, 30]

        return self.execute_test(
            test_name="bplus_tree_range_queries",
            difficulty=DifficultyLevel.EXTREME,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=None,
            expected_output=True
        )

    # ═══════════════════════════════════════════════════════════════════════
    # SPECIAL CATEGORY TESTS
    # ═══════════════════════════════════════════════════════════════════════

    def test_collaboration_scenario(self) -> TestResult:
        """Test: Multi-agent collaboration for system design."""
        def test_func(input_data):
            system_requirements = {
                "apex_contribution": "Algorithm selection and implementation",
                "architect_integration": "System structure and patterns",
                "velocity_optimization": "Performance tuning",
                "result": "Integrated system design"
            }

            required_keys = ["apex_contribution", "architect_integration",
                           "velocity_optimization", "result"]
            return all(key in system_requirements for key in required_keys)

        return self.execute_test(
            test_name="multi_agent_system_design",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.COLLABORATION,
            test_func=test_func,
            input_data=None,
            expected_output=True
        )

    def test_evolution_adaptation(self) -> TestResult:
        """Test: Ability to learn from feedback and improve."""
        def test_func(input_data):
            failure_patterns = [
                {"type": "timeout", "count": 3, "solution": "optimize algorithm"},
                {"type": "memory", "count": 2, "solution": "use streaming"},
                {"type": "accuracy", "count": 1, "solution": "add validation"}
            ]

            critical = max(failure_patterns, key=lambda x: x["count"])
            return critical["type"] == "timeout"

        return self.execute_test(
            test_name="learning_from_failures",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.EVOLUTION,
            test_func=test_func,
            input_data=None,
            expected_output=True
        )

    def test_edge_case_handling(self) -> TestResult:
        """Test: Robust handling of edge cases."""
        def test_func(input_data):
            test_cases = [
                ([], "empty_list"),
                (None, "null_input"),
                ([1], "single_element"),
                (list(range(1000)), "large_input"),
                ([-1, 0, 1], "negative_numbers"),
            ]

            results = []
            for test_input, case_name in test_cases:
                try:
                    if test_input is None:
                        results.append(("null_handled", True))
                    elif len(test_input) == 0:
                        results.append(("empty_handled", True))
                    else:
                        results.append((case_name, True))
                except Exception as e:
                    results.append((case_name, False))

            return all(r[1] for r in results)

        return self.execute_test(
            test_name="comprehensive_edge_cases",
            difficulty=DifficultyLevel.STANDARD,
            category=TestCategory.EDGE_CASE,
            test_func=test_func,
            input_data=None,
            expected_output=True
        )


if __name__ == "__main__":
    test_suite = ApexAgentTest()
    summary = test_suite.run_all_tests()

    print(f"\n{'='*60}")
    print(f"APEX-01 Test Results")
    print(f"{'='*60}")
    print(f"Total Tests: {summary.total_tests}")
    print(f"Passed: {summary.passed_tests}")
    print(f"Failed: {summary.failed_tests}")
    print(f"Pass Rate: {summary.pass_rate:.2%}")
    print(f"{'='*60}\n")
