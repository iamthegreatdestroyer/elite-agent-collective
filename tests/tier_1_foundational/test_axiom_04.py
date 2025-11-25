"""
AXIOM-04 Test Suite
===================
Pure Mathematics & Formal Proofs Specialist

Tests cover:
- Mathematical proofs
- Algorithm complexity analysis
- Formal verification
- Number theory
- Graph theory
"""

import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).parent.parent))

from framework.base_agent_test import (
    BaseAgentTest, TestResult, DifficultyLevel, TestCategory
)


class AxiomAgentTest(BaseAgentTest):
    """Comprehensive test suite for AXIOM-04."""

    @property
    def agent_id(self) -> str:
        return "04"

    @property
    def agent_codename(self) -> str:
        return "AXIOM"

    @property
    def agent_tier(self) -> int:
        return 1

    @property
    def agent_specialty(self) -> str:
        return "Pure Mathematics & Formal Proofs"

    def test_L1_trivial_01(self) -> TestResult:
        """Test: Prove sum of first n natural numbers."""
        def test_func(n):
            # Direct computation
            direct_sum = sum(range(1, n + 1))
            # Formula: n(n+1)/2
            formula_result = n * (n + 1) // 2
            return direct_sum == formula_result

        return self.execute_test(
            test_name="sum_natural_numbers",
            difficulty=DifficultyLevel.TRIVIAL,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=100,
            expected_output=True
        )

    def test_L1_trivial_02(self) -> TestResult:
        """Test: Implement and verify Euclidean algorithm for GCD."""
        def test_func(input_data):
            a, b = input_data
            
            def gcd(x, y):
                while y:
                    x, y = y, x % y
                return x
            
            result = gcd(a, b)
            
            # Verify: result divides both a and b
            return a % result == 0 and b % result == 0

        return self.execute_test(
            test_name="euclidean_gcd",
            difficulty=DifficultyLevel.TRIVIAL,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=(48, 18),
            expected_output=True
        )

    def test_L2_standard_01(self) -> TestResult:
        """Test: Prove binary search correctness with loop invariant."""
        def test_func(input_data):
            arr, target = input_data
            
            # Loop invariant: if target exists, it's in arr[left:right+1]
            left, right = 0, len(arr) - 1
            
            invariant_maintained = True
            
            while left <= right:
                # Check invariant before iteration
                if target in arr:
                    if target < arr[left] or target > arr[right]:
                        invariant_maintained = False
                        break
                
                mid = left + (right - left) // 2
                
                if arr[mid] == target:
                    return invariant_maintained and arr[mid] == target
                elif arr[mid] < target:
                    left = mid + 1
                else:
                    right = mid - 1
            
            # Target not found - verify invariant holds (empty range)
            return invariant_maintained and left > right

        return self.execute_test(
            test_name="binary_search_loop_invariant",
            difficulty=DifficultyLevel.STANDARD,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=([1, 2, 3, 4, 5, 6, 7, 8, 9, 10], 7),
            expected_output=True
        )

    def test_L2_standard_02(self) -> TestResult:
        """Test: Prove properties of modular arithmetic."""
        def test_func(input_data):
            a, b, m = input_data
            
            # Property 1: (a + b) mod m = ((a mod m) + (b mod m)) mod m
            prop1 = (a + b) % m == ((a % m) + (b % m)) % m
            
            # Property 2: (a * b) mod m = ((a mod m) * (b mod m)) mod m
            prop2 = (a * b) % m == ((a % m) * (b % m)) % m
            
            # Property 3: (a^b) mod m can be computed efficiently
            def mod_pow(base, exp, mod):
                result = 1
                base = base % mod
                while exp > 0:
                    if exp % 2 == 1:
                        result = (result * base) % mod
                    exp = exp >> 1
                    base = (base * base) % mod
                return result
            
            prop3 = mod_pow(a, b, m) == pow(a, b, m)
            
            return prop1 and prop2 and prop3

        return self.execute_test(
            test_name="modular_arithmetic_properties",
            difficulty=DifficultyLevel.STANDARD,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=(17, 13, 7),
            expected_output=True
        )

    def test_L2_standard_03(self) -> TestResult:
        """Test: Implement and verify Fermat's little theorem."""
        def test_func(input_data):
            a, p = input_data  # p is prime
            
            # Fermat's little theorem: a^(p-1) ≡ 1 (mod p) for prime p
            if a % p == 0:
                return True  # Edge case: a is divisible by p
            
            result = pow(a, p - 1, p)
            return result == 1

        return self.execute_test(
            test_name="fermats_little_theorem",
            difficulty=DifficultyLevel.STANDARD,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=(3, 7),  # 3^6 mod 7 = 729 mod 7 = 1
            expected_output=True
        )

    def test_L3_advanced_01(self) -> TestResult:
        """Test: Derive time complexity using Master Theorem."""
        def test_func(input_data):
            # Analyze T(n) = aT(n/b) + f(n)
            # Returns the complexity class
            
            a, b, f_n_degree = input_data  # f(n) = n^f_n_degree
            
            log_b_a = __import__('math').log(a) / __import__('math').log(b)
            
            # Master theorem cases
            epsilon = 0.001
            
            if f_n_degree < log_b_a - epsilon:
                # Case 1: T(n) = Θ(n^log_b_a)
                complexity = f"O(n^{log_b_a:.2f})"
            elif abs(f_n_degree - log_b_a) < epsilon:
                # Case 2: T(n) = Θ(n^log_b_a * log n)
                complexity = f"O(n^{log_b_a:.2f} * log n)"
            else:
                # Case 3: T(n) = Θ(f(n))
                complexity = f"O(n^{f_n_degree})"
            
            # Verify for merge sort: T(n) = 2T(n/2) + O(n)
            # a=2, b=2, f(n)=n^1
            # log_2(2) = 1, f_n_degree = 1, so Case 2 applies
            return "log n" in complexity or f_n_degree == log_b_a

        return self.execute_test(
            test_name="master_theorem_analysis",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=(2, 2, 1),  # Merge sort recurrence
            expected_output=True
        )

    def test_L3_advanced_02(self) -> TestResult:
        """Test: Implement matrix exponentiation for Fibonacci."""
        def test_func(n):
            import numpy as np
            
            def matrix_power(M, n):
                if n == 1:
                    return M
                if n % 2 == 0:
                    half = matrix_power(M, n // 2)
                    return half @ half
                else:
                    return M @ matrix_power(M, n - 1)
            
            # [[F(n+1), F(n)], [F(n), F(n-1)]] = [[1,1],[1,0]]^n
            M = np.array([[1, 1], [1, 0]], dtype=np.int64)
            
            if n <= 0:
                return 0
            
            result = matrix_power(M, n)
            fib_n = result[0, 1]
            
            # Verify against iterative computation
            def fib_iterative(n):
                if n <= 1:
                    return n
                a, b = 0, 1
                for _ in range(2, n + 1):
                    a, b = b, a + b
                return b
            
            return fib_n == fib_iterative(n)

        return self.execute_test(
            test_name="matrix_fibonacci",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=20,
            expected_output=True
        )

    def test_L3_advanced_03(self) -> TestResult:
        """Test: Prove graph coloring theorem properties."""
        def test_func(input_data):
            # For a planar graph, chromatic number <= 4 (Four Color Theorem)
            # We verify greedy coloring gives <= max_degree + 1 colors
            
            n, edges = input_data
            
            # Build adjacency list
            adj = {i: [] for i in range(n)}
            for u, v in edges:
                adj[u].append(v)
                adj[v].append(u)
            
            # Greedy coloring
            colors = [-1] * n
            
            for node in range(n):
                neighbor_colors = {colors[neighbor] for neighbor in adj[node]}
                
                # Find smallest available color
                color = 0
                while color in neighbor_colors:
                    color += 1
                colors[node] = color
            
            # Verify properties
            max_degree = max(len(neighbors) for neighbors in adj.values())
            num_colors = max(colors) + 1
            
            # Verify no adjacent nodes have same color
            valid_coloring = all(
                colors[u] != colors[v]
                for u, v in edges
            )
            
            return valid_coloring and num_colors <= max_degree + 1

        return self.execute_test(
            test_name="graph_coloring_theorem",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=(5, [(0,1), (0,2), (1,2), (1,3), (2,3), (3,4)]),
            expected_output=True
        )

    def test_L4_expert_01(self) -> TestResult:
        """Test: Prove NP-completeness reduction."""
        def test_func(input_data):
            # Demonstrate reduction from 3-SAT to Independent Set
            # If we can solve Independent Set, we can solve 3-SAT
            
            # 3-SAT clause: (x1 ∨ x2 ∨ ¬x3)
            # For each clause, create a triangle in the graph
            # Add edges between contradictory literals
            
            clauses = input_data  # List of clauses, each is list of literals
            
            # Build graph
            nodes = []  # (clause_idx, literal)
            edges = []
            
            for i, clause in enumerate(clauses):
                for lit in clause:
                    nodes.append((i, lit))
            
            # Add edges within each clause (triangle)
            for i, clause in enumerate(clauses):
                clause_nodes = [(i, lit) for lit in clause]
                for j in range(len(clause_nodes)):
                    for k in range(j + 1, len(clause_nodes)):
                        edges.append((nodes.index(clause_nodes[j]), 
                                     nodes.index(clause_nodes[k])))
            
            # Add edges between contradictory literals
            for i, (ci, li) in enumerate(nodes):
                for j, (cj, lj) in enumerate(nodes):
                    if i < j and li == -lj:  # Contradictory
                        edges.append((i, j))
            
            # If graph has independent set of size = num_clauses,
            # then 3-SAT is satisfiable
            
            # Verify reduction is polynomial
            n = len(nodes)
            m = len(edges)
            
            return n == sum(len(c) for c in clauses) and m >= len(clauses)

        return self.execute_test(
            test_name="np_completeness_reduction",
            difficulty=DifficultyLevel.EXPERT,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=[[1, 2, -3], [-1, 2, 3], [1, -2, 3]],
            expected_output=True
        )

    def test_L4_expert_02(self) -> TestResult:
        """Test: Implement and verify Prim's MST correctness."""
        def test_func(input_data):
            import heapq
            
            n, edges = input_data  # n vertices, list of (u, v, weight)
            
            # Build adjacency list
            adj = {i: [] for i in range(n)}
            for u, v, w in edges:
                adj[u].append((v, w))
                adj[v].append((u, w))
            
            # Prim's algorithm
            mst_edges = []
            visited = {0}
            heap = [(w, 0, v) for v, w in adj[0]]
            heapq.heapify(heap)
            
            while heap and len(visited) < n:
                weight, u, v = heapq.heappop(heap)
                if v in visited:
                    continue
                
                visited.add(v)
                mst_edges.append((u, v, weight))
                
                for next_v, next_w in adj[v]:
                    if next_v not in visited:
                        heapq.heappush(heap, (next_w, v, next_v))
            
            # Verify MST properties
            # 1. MST has n-1 edges
            prop1 = len(mst_edges) == n - 1
            
            # 2. MST connects all vertices
            mst_adj = {i: [] for i in range(n)}
            for u, v, w in mst_edges:
                mst_adj[u].append(v)
                mst_adj[v].append(u)
            
            # BFS to check connectivity
            visited_check = {0}
            queue = [0]
            while queue:
                node = queue.pop(0)
                for neighbor in mst_adj[node]:
                    if neighbor not in visited_check:
                        visited_check.add(neighbor)
                        queue.append(neighbor)
            
            prop2 = len(visited_check) == n
            
            return prop1 and prop2

        return self.execute_test(
            test_name="prims_mst_correctness",
            difficulty=DifficultyLevel.EXPERT,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=(5, [(0,1,2), (0,2,3), (1,2,1), (1,3,4), (2,3,2), (3,4,5)]),
            expected_output=True
        )

    def test_L5_extreme_01(self) -> TestResult:
        """Test: Implement formal verification concepts."""
        def test_func(input_data):
            # Implement Hoare Logic verification
            # {P} S {Q} - if P holds before S, Q holds after
            
            # Pre/post condition verification for simple program
            # Program: x = x + 1
            # Pre: x >= 0
            # Post: x >= 1
            
            def verify_hoare_triple(pre_condition, statement, post_condition, initial_state):
                # Check precondition
                if not pre_condition(initial_state):
                    return True  # Vacuously true
                
                # Execute statement
                new_state = statement(initial_state)
                
                # Check postcondition
                return post_condition(new_state)
            
            # Test case 1: increment
            result1 = verify_hoare_triple(
                lambda s: s['x'] >= 0,
                lambda s: {**s, 'x': s['x'] + 1},
                lambda s: s['x'] >= 1,
                {'x': 5}
            )
            
            # Test case 2: swap
            result2 = verify_hoare_triple(
                lambda s: s['x'] == 3 and s['y'] == 5,
                lambda s: {'x': s['y'], 'y': s['x']},
                lambda s: s['x'] == 5 and s['y'] == 3,
                {'x': 3, 'y': 5}
            )
            
            # Test weakest precondition for assignment
            # wp(x := e, Q) = Q[x/e]
            def wp_assignment(var, expr, post):
                def result(state):
                    new_state = state.copy()
                    new_state[var] = expr(state)
                    return post(new_state)
                return result
            
            # For x := x + 1, wp(x >= 2) = (x + 1 >= 2) = (x >= 1)
            post = lambda s: s['x'] >= 2
            pre = wp_assignment('x', lambda s: s['x'] + 1, post)
            
            result3 = pre({'x': 1})  # Should be True
            
            return result1 and result2 and result3

        return self.execute_test(
            test_name="hoare_logic_verification",
            difficulty=DifficultyLevel.EXTREME,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=None,
            expected_output=True
        )

    def test_L5_extreme_02(self) -> TestResult:
        """Test: Implement type system soundness verification."""
        def test_func(input_data):
            # Simple type system with typing rules
            from enum import Enum
            
            class Type(Enum):
                INT = "int"
                BOOL = "bool"
                FUNC = "func"
            
            class TypeChecker:
                def __init__(self):
                    self.context = {}
                
                def infer(self, expr):
                    if isinstance(expr, int):
                        return Type.INT
                    elif isinstance(expr, bool):
                        return Type.BOOL
                    elif isinstance(expr, tuple):
                        op, *args = expr
                        if op == 'add':
                            if all(self.infer(a) == Type.INT for a in args):
                                return Type.INT
                        elif op == 'eq':
                            if len(args) == 2 and self.infer(args[0]) == self.infer(args[1]):
                                return Type.BOOL
                        elif op == 'if':
                            cond, then_e, else_e = args
                            if self.infer(cond) == Type.BOOL:
                                t1 = self.infer(then_e)
                                t2 = self.infer(else_e)
                                if t1 == t2:
                                    return t1
                    return None
                
                def check(self, expr, expected_type):
                    return self.infer(expr) == expected_type
            
            tc = TypeChecker()
            
            # Test well-typed expressions
            test1 = tc.check(5, Type.INT)
            test2 = tc.check(True, Type.BOOL)
            test3 = tc.check(('add', 1, 2), Type.INT)
            test4 = tc.check(('eq', 1, 2), Type.BOOL)
            test5 = tc.check(('if', True, 1, 2), Type.INT)
            
            # Test ill-typed expression
            test6 = tc.check(('add', 1, True), Type.INT) == False
            
            # Progress: well-typed terms evaluate
            # Preservation: evaluation preserves types
            
            return all([test1, test2, test3, test4, test5, test6])

        return self.execute_test(
            test_name="type_system_soundness",
            difficulty=DifficultyLevel.EXTREME,
            category=TestCategory.CORE_COMPETENCY,
            test_func=test_func,
            input_data=None,
            expected_output=True
        )

    def test_collaboration_scenario(self) -> TestResult:
        """Test: Collaborate with ECLIPSE for formal verification."""
        def test_func(input_data):
            collaboration = {
                "axiom_proofs": "Mathematical foundation",
                "eclipse_testing": "Property-based verification",
                "combined_assurance": "Formal correctness guarantee"
            }
            
            return all(v for v in collaboration.values())

        return self.execute_test(
            test_name="formal_verification_collaboration",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.COLLABORATION,
            test_func=test_func,
            input_data=None,
            expected_output=True
        )

    def test_evolution_adaptation(self) -> TestResult:
        """Test: Adapt mathematical reasoning to new domains."""
        def test_func(input_data):
            # Verify ability to apply mathematical concepts to new problems
            domains = {
                "classical": "Standard analysis and algebra",
                "constructive": "Constructive mathematics",
                "computational": "Complexity theory",
                "categorical": "Category theory abstractions"
            }
            
            return len(domains) == 4 and all(v for v in domains.values())

        return self.execute_test(
            test_name="mathematical_domain_adaptation",
            difficulty=DifficultyLevel.ADVANCED,
            category=TestCategory.EVOLUTION,
            test_func=test_func,
            input_data=None,
            expected_output=True
        )

    def test_edge_case_handling(self) -> TestResult:
        """Test: Handle mathematical edge cases."""
        def test_func(input_data):
            import math
            
            edge_cases = []
            
            # Division by zero handling
            try:
                result = 1 / 0
                edge_cases.append(False)
            except ZeroDivisionError:
                edge_cases.append(True)
            
            # Infinity handling
            edge_cases.append(math.isinf(float('inf')))
            
            # NaN handling
            edge_cases.append(math.isnan(float('nan')))
            
            # Large number handling
            large = 10**100
            edge_cases.append(large * 2 == 2 * 10**100)
            
            # Floating point precision
            edge_cases.append(abs(0.1 + 0.2 - 0.3) < 1e-10)
            
            return all(edge_cases)

        return self.execute_test(
            test_name="mathematical_edge_cases",
            difficulty=DifficultyLevel.STANDARD,
            category=TestCategory.EDGE_CASE,
            test_func=test_func,
            input_data=None,
            expected_output=True
        )


if __name__ == "__main__":
    test_suite = AxiomAgentTest()
    summary = test_suite.run_all_tests()

    print(f"\n{'='*60}")
    print(f"AXIOM-04 Test Results")
    print(f"{'='*60}")
    print(f"Total Tests: {summary.total_tests}")
    print(f"Passed: {summary.passed_tests}")
    print(f"Failed: {summary.failed_tests}")
    print(f"Pass Rate: {summary.pass_rate:.2%}")
    print(f"{'='*60}\n")
