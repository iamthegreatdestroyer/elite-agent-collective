"""
═══════════════════════════════════════════════════════════════════════════════
                    CORE-14: LOW-LEVEL SYSTEMS & COMPILER DESIGN
                         Elite Agent Collective Test Suite
═══════════════════════════════════════════════════════════════════════════════
Agent ID: CORE-14
Codename: @CORE
Tier: 2 (Specialists)
Domain: Systems Programming, Compilers, Operating Systems, Assembly
Philosophy: "At the lowest level, every instruction counts."

Test Coverage:
- Operating systems internals (Linux kernel, Windows NT)
- Compiler design (lexing, parsing, optimization, codegen)
- Assembly programming (x86-64, ARM64, RISC-V)
- Memory management & concurrency primitives
- Device drivers & embedded systems
- LLVM/GCC internals
═══════════════════════════════════════════════════════════════════════════════
"""

import sys
from pathlib import Path
from dataclasses import dataclass, field
from typing import List, Dict, Any, Optional
from datetime import datetime
import hashlib
import json

# Add framework to path
sys.path.insert(0, str(Path(__file__).parent.parent / "framework"))

from base_agent_test import BaseAgentTest, TestResult, TestDifficulty


@dataclass
class SystemsScenario:
    """Low-level systems scenario for testing CORE capabilities."""
    domain: str  # kernel, compiler, embedded, driver
    architecture: str  # x86-64, ARM64, RISC-V
    constraints: Dict[str, Any]
    requirements: List[str]
    expected_outputs: List[str]


@dataclass
class CompilerComponent:
    """Compiler component specification for testing."""
    phase: str  # lexer, parser, semantic, optimizer, codegen
    input_language: str
    target: str
    optimizations: List[str]
    constraints: Dict[str, Any]


class TestCore14(BaseAgentTest):
    """
    Comprehensive test suite for CORE-14: Low-Level Systems & Compiler Design.
    
    CORE is the systems programming master of the collective, capable of:
    - Operating system kernel development
    - Compiler construction from scratch
    - Assembly optimization across architectures
    - Memory management and allocator design
    - Concurrency primitives and lock-free structures
    - Device driver development
    """
    
    AGENT_ID = "CORE-14"
    AGENT_CODENAME = "@CORE"
    AGENT_TIER = 2
    AGENT_DOMAIN = "Low-Level Systems & Compiler Design"
    
    # ═══════════════════════════════════════════════════════════════════════
    # CORE COMPETENCY TESTS
    # ═══════════════════════════════════════════════════════════════════════
    
    def test_L1_basic_memory_management(self) -> TestResult:
        """
        L1 TRIVIAL: Implement basic memory allocator
        
        Tests CORE's ability to design fundamental memory
        allocation strategies.
        """
        scenario = SystemsScenario(
            domain="memory",
            architecture="x86-64",
            constraints={"heap_size": "1GB", "alignment": 16},
            requirements=["malloc", "free", "realloc"],
            expected_outputs=["allocator.c", "allocator.h"]
        )
        
        test_input = {
            "task": "Implement simple memory allocator with free list",
            "scenario": scenario.__dict__,
            "algorithm": "First-fit with coalescing",
            "requirements": [
                "O(n) allocation",
                "Proper alignment",
                "Coalescing free blocks",
                "No memory leaks"
            ]
        }
        
        validation_criteria = {
            "correctness": "All operations work correctly",
            "alignment": "Returns properly aligned memory",
            "coalescing": "Adjacent free blocks merged",
            "fragmentation": "Reasonable fragmentation levels",
            "thread_safety": "At least single-threaded safety"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L1_memory_allocator",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L1_TRIVIAL,
            category="core_competency",
            input_data=test_input,
            expected_behavior="Working memory allocator implementation",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Foundation test for memory management"
        )
    
    def test_L2_lexer_implementation(self) -> TestResult:
        """
        L2 EASY: Implement lexical analyzer for programming language
        
        Tests CORE's ability to build tokenizers and lexers.
        """
        component = CompilerComponent(
            phase="lexer",
            input_language="C-like",
            target="tokens",
            optimizations=["string_interning"],
            constraints={"unicode_support": True}
        )
        
        test_input = {
            "task": "Implement lexer for C-like language",
            "component": component.__dict__,
            "token_types": [
                "IDENTIFIER", "NUMBER", "STRING", "OPERATOR",
                "KEYWORD", "PUNCTUATION", "COMMENT", "WHITESPACE"
            ],
            "requirements": [
                "Handle all C operators",
                "Support single and multi-line comments",
                "Unicode identifier support",
                "Line and column tracking"
            ]
        }
        
        validation_criteria = {
            "token_accuracy": "Correct tokenization",
            "position_tracking": "Accurate line/column info",
            "error_recovery": "Continue after errors",
            "performance": "Linear time complexity",
            "edge_cases": "Handle edge cases gracefully"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L2_lexer",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L2_EASY,
            category="core_competency",
            input_data=test_input,
            expected_behavior="Complete lexer implementation with error handling",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests lexical analysis skills"
        )
    
    def test_L3_parser_with_ast(self) -> TestResult:
        """
        L3 MEDIUM: Implement parser with AST generation
        
        Tests CORE's ability to build parsers and abstract
        syntax trees for programming languages.
        """
        test_input = {
            "task": "Implement recursive descent parser for expression language",
            "grammar": {
                "expression": "term (('+' | '-') term)*",
                "term": "factor (('*' | '/') factor)*",
                "factor": "NUMBER | IDENTIFIER | '(' expression ')' | unary",
                "unary": "('-' | '!') factor",
                "statement": "assignment | if | while | block | expression",
                "assignment": "IDENTIFIER '=' expression",
                "if": "'if' '(' expression ')' statement ('else' statement)?",
                "while": "'while' '(' expression ')' statement"
            },
            "ast_requirements": [
                "Typed AST nodes",
                "Source location preservation",
                "Error recovery with synchronization",
                "Operator precedence handling"
            ]
        }
        
        validation_criteria = {
            "grammar_coverage": "All productions implemented",
            "ast_structure": "Well-typed AST nodes",
            "precedence": "Correct operator precedence",
            "error_messages": "Clear error diagnostics",
            "recovery": "Parse recovery after errors"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L3_parser",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L3_MEDIUM,
            category="core_competency",
            input_data=test_input,
            expected_behavior="Complete parser with typed AST generation",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests parsing and AST construction"
        )
    
    # ═══════════════════════════════════════════════════════════════════════
    # ADVANCED SYSTEMS TESTS
    # ═══════════════════════════════════════════════════════════════════════
    
    def test_L4_compiler_optimizer(self) -> TestResult:
        """
        L4 HARD: Implement compiler optimization passes
        
        Tests CORE's ability to design and implement
        program transformations for optimization.
        """
        component = CompilerComponent(
            phase="optimizer",
            input_language="SSA IR",
            target="optimized IR",
            optimizations=[
                "constant_folding",
                "dead_code_elimination",
                "common_subexpression",
                "loop_invariant_motion",
                "strength_reduction"
            ],
            constraints={"preserve_semantics": True}
        )
        
        test_input = {
            "task": "Implement SSA-based optimization passes",
            "component": component.__dict__,
            "ir_format": "SSA with basic blocks and phi nodes",
            "analysis_passes": [
                "Dominance analysis",
                "Reaching definitions",
                "Live variable analysis",
                "Loop detection"
            ],
            "optimization_passes": [
                "Sparse conditional constant propagation",
                "Global value numbering",
                "Loop-invariant code motion",
                "Dead code elimination"
            ]
        }
        
        validation_criteria = {
            "correctness": "Semantics preserved",
            "effectiveness": "Measurable code improvement",
            "pass_ordering": "Optimal pass ordering",
            "analysis_accuracy": "Correct data flow analysis",
            "phi_handling": "Proper SSA phi node handling"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L4_optimizer",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L4_HARD,
            category="core_competency",
            input_data=test_input,
            expected_behavior="Complete SSA optimization pass infrastructure",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests advanced compiler optimization"
        )
    
    def test_L5_kernel_scheduler(self) -> TestResult:
        """
        L5 EXTREME: Design OS kernel process scheduler
        
        Tests CORE's ability to implement operating system
        core components with real-time constraints.
        """
        scenario = SystemsScenario(
            domain="kernel",
            architecture="x86-64",
            constraints={
                "preemptive": True,
                "smp_support": True,
                "numa_aware": True,
                "real_time": "Soft real-time support"
            },
            requirements=[
                "CFS-like fair scheduling",
                "Priority-based scheduling",
                "CPU affinity",
                "Load balancing",
                "Power management integration"
            ],
            expected_outputs=["scheduler.c", "runqueue.c", "context_switch.S"]
        )
        
        test_input = {
            "task": "Design and implement kernel process scheduler",
            "scenario": scenario.__dict__,
            "scheduling_classes": [
                "SCHED_NORMAL (CFS)",
                "SCHED_FIFO",
                "SCHED_RR",
                "SCHED_DEADLINE"
            ],
            "data_structures": [
                "Red-black tree runqueue",
                "Per-CPU runqueues",
                "Wait queues",
                "Load tracking"
            ],
            "challenges": [
                "Lock contention minimization",
                "Cache-friendly design",
                "NUMA-aware placement",
                "Power-aware scheduling"
            ]
        }
        
        validation_criteria = {
            "fairness": "Virtual runtime tracking",
            "latency": "Low scheduling latency",
            "throughput": "High context switch efficiency",
            "scalability": "Scales to 100+ CPUs",
            "real_time": "Deadline scheduling support",
            "correctness": "No priority inversion, no starvation"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L5_kernel_scheduler",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L5_EXTREME,
            category="core_competency",
            input_data=test_input,
            expected_behavior="Complete kernel scheduler with SMP and NUMA support",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Ultimate test of OS kernel development"
        )
    
    # ═══════════════════════════════════════════════════════════════════════
    # EDGE CASE HANDLING TESTS
    # ═══════════════════════════════════════════════════════════════════════
    
    def test_L3_lock_free_data_structure(self) -> TestResult:
        """
        L3 MEDIUM: Implement lock-free concurrent data structure
        
        Tests CORE's ability to design thread-safe structures
        without traditional locking.
        """
        test_input = {
            "task": "Implement lock-free queue with CAS",
            "data_structure": "Michael-Scott queue",
            "operations": ["enqueue", "dequeue", "is_empty"],
            "requirements": [
                "Lock-free (non-blocking)",
                "ABA problem prevention",
                "Memory reclamation",
                "Linearizable operations"
            ],
            "target_architecture": "x86-64 with TSO memory model"
        }
        
        validation_criteria = {
            "lock_freedom": "No locks or blocking",
            "aba_prevention": "Tagged pointers or hazard pointers",
            "memory_ordering": "Correct memory barriers",
            "linearizability": "Atomic appearance of operations",
            "memory_safety": "Safe memory reclamation"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L3_lock_free",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L3_MEDIUM,
            category="edge_case_handling",
            input_data=test_input,
            expected_behavior="Correct lock-free queue implementation",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests concurrent programming expertise"
        )
    
    def test_L4_memory_model_reasoning(self) -> TestResult:
        """
        L4 HARD: Reason about weak memory model correctness
        
        Tests CORE's ability to analyze concurrent code
        under weak memory models.
        """
        test_input = {
            "task": "Analyze and fix concurrent algorithm under ARM memory model",
            "code_snippet": """
            // Thread 1
            store(data, 42);
            store(flag, 1);
            
            // Thread 2
            while (load(flag) != 1) {}
            int x = load(data);
            assert(x == 42);  // Can this fail?
            """,
            "memory_model": "ARM (weak ordering)",
            "questions": [
                "Can the assertion fail?",
                "What reorderings are possible?",
                "What barriers are needed?",
                "Is acquire-release sufficient?"
            ]
        }
        
        validation_criteria = {
            "reordering_analysis": "Identify possible reorderings",
            "barrier_placement": "Correct barrier insertion",
            "acquire_release": "Proper acquire-release semantics",
            "formal_reasoning": "Sound memory model reasoning"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L4_memory_model",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L4_HARD,
            category="edge_case_handling",
            input_data=test_input,
            expected_behavior="Correct analysis and fix for memory model issues",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests weak memory model expertise"
        )
    
    # ═══════════════════════════════════════════════════════════════════════
    # INTER-AGENT COLLABORATION TESTS
    # ═══════════════════════════════════════════════════════════════════════
    
    def test_L3_core_velocity_optimization(self) -> TestResult:
        """
        L3 MEDIUM: Collaborate with VELOCITY for low-level optimization
        
        Tests CORE + VELOCITY synergy for performance-critical code.
        """
        test_input = {
            "task": "Optimize matrix multiplication with SIMD",
            "core_responsibilities": [
                "Assembly implementation",
                "Cache blocking",
                "Memory alignment",
                "Instruction selection"
            ],
            "velocity_requirements": [
                "Algorithm selection",
                "Cache analysis",
                "Benchmark design",
                "Performance modeling"
            ],
            "target": {
                "architecture": "x86-64 with AVX-512",
                "matrix_size": "1024x1024",
                "data_type": "double",
                "target_gflops": "90% of theoretical peak"
            }
        }
        
        validation_criteria = {
            "simd_utilization": "Full AVX-512 usage",
            "cache_optimization": "Optimal blocking factors",
            "memory_access": "Prefetching strategy",
            "instruction_mix": "Balanced FMA usage"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L3_simd_optimization",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L3_MEDIUM,
            category="inter_agent_collaboration",
            input_data=test_input,
            expected_behavior="Highly optimized SIMD matrix multiplication",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests CORE + VELOCITY collaboration"
        )
    
    def test_L4_core_cipher_secure_code(self) -> TestResult:
        """
        L4 HARD: Collaborate with CIPHER for secure low-level code
        
        Tests CORE + CIPHER synergy for security-critical systems code.
        """
        test_input = {
            "task": "Implement constant-time cryptographic primitive",
            "core_responsibilities": [
                "Assembly implementation",
                "Timing attack prevention",
                "Side-channel resistance",
                "Cache timing mitigation"
            ],
            "cipher_requirements": [
                "Algorithm correctness",
                "Security analysis",
                "Attack surface review",
                "Formal verification hints"
            ],
            "primitive": {
                "algorithm": "AES-256-GCM",
                "requirements": [
                    "Constant-time execution",
                    "No data-dependent branches",
                    "No data-dependent memory access",
                    "Cache-timing resistance"
                ]
            }
        }
        
        validation_criteria = {
            "constant_time": "No timing variations",
            "branch_free": "No conditional branches on secrets",
            "cache_resistance": "No secret-dependent cache access",
            "correctness": "Matches reference implementation"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L4_secure_implementation",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L4_HARD,
            category="inter_agent_collaboration",
            input_data=test_input,
            expected_behavior="Constant-time cryptographic implementation",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests CORE + CIPHER collaboration"
        )
    
    # ═══════════════════════════════════════════════════════════════════════
    # STRESS & PERFORMANCE TESTS
    # ═══════════════════════════════════════════════════════════════════════
    
    def test_L4_jit_compiler(self) -> TestResult:
        """
        L4 HARD: Implement JIT compiler for dynamic language
        
        Tests CORE's ability to generate machine code at runtime.
        """
        test_input = {
            "task": "Implement method JIT for bytecode interpreter",
            "components": [
                "Bytecode IR",
                "Register allocator",
                "Code generation",
                "Runtime patching",
                "Deoptimization"
            ],
            "optimizations": [
                "Inline caching",
                "Type specialization",
                "Method inlining",
                "Dead code elimination"
            ],
            "target_architecture": "x86-64",
            "runtime_requirements": {
                "compilation_threshold": "1000 invocations",
                "deopt_capability": "Return to interpreter",
                "gc_integration": "Safepoints for GC"
            }
        }
        
        validation_criteria = {
            "code_quality": "Generated code performance",
            "compilation_speed": "Fast JIT compilation",
            "memory_efficiency": "Reasonable code cache",
            "deoptimization": "Correct fallback to interpreter",
            "gc_safety": "Proper safepoint handling"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L4_jit_compiler",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L4_HARD,
            category="stress_performance",
            input_data=test_input,
            expected_behavior="Working JIT compiler with optimization",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests runtime code generation"
        )
    
    def test_L5_custom_os_kernel(self) -> TestResult:
        """
        L5 EXTREME: Design microkernel operating system
        
        Tests CORE's ability to architect complete OS kernel.
        """
        test_input = {
            "task": "Design microkernel OS for embedded real-time system",
            "kernel_components": [
                "IPC (synchronous message passing)",
                "Memory management (capability-based)",
                "Scheduling (real-time)",
                "Interrupt handling",
                "Device driver framework"
            ],
            "requirements": {
                "ipc_latency": "< 1µs",
                "context_switch": "< 500ns",
                "memory_protection": "Hardware-enforced",
                "determinism": "Hard real-time guarantees"
            },
            "architecture": "ARM64",
            "formal_properties": [
                "Memory isolation",
                "Information flow control",
                "Scheduling guarantees"
            ]
        }
        
        validation_criteria = {
            "architecture": "Clean microkernel separation",
            "ipc_design": "Efficient message passing",
            "capability_system": "Proper capability model",
            "real_time": "Provable scheduling bounds",
            "security": "Isolation between processes"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L5_microkernel",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L5_EXTREME,
            category="stress_performance",
            input_data=test_input,
            expected_behavior="Complete microkernel design with real-time guarantees",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Ultimate systems programming challenge"
        )
    
    # ═══════════════════════════════════════════════════════════════════════
    # NOVELTY & EVOLUTION TESTS
    # ═══════════════════════════════════════════════════════════════════════
    
    def test_L4_novel_ir_design(self) -> TestResult:
        """
        L4 HARD: Design novel compiler intermediate representation
        
        Tests CORE's ability to innovate in compiler design.
        """
        test_input = {
            "task": "Design IR for heterogeneous computing",
            "requirements": {
                "targets": ["CPU", "GPU", "FPGA", "TPU"],
                "features": [
                    "Unified representation",
                    "Target-independent optimizations",
                    "Parallel primitives",
                    "Memory hierarchy modeling"
                ]
            },
            "design_goals": [
                "Single IR for all targets",
                "Efficient lowering to each target",
                "Cross-target optimization",
                "Kernel fusion opportunities"
            ],
            "existing_inspiration": ["MLIR", "SPIR-V", "LLVM IR"]
        }
        
        validation_criteria = {
            "expressiveness": "Can represent all target operations",
            "analyzability": "Enables powerful analysis",
            "lowering": "Clean lowering to targets",
            "extensibility": "Easy to add new targets"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L4_novel_ir",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L4_HARD,
            category="novelty_generation",
            input_data=test_input,
            expected_behavior="Novel IR design for heterogeneous computing",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests compiler innovation"
        )
    
    def test_L5_self_optimizing_compiler(self) -> TestResult:
        """
        L5 EXTREME: Design self-optimizing compiler system
        
        Tests CORE's ability to create compilers that learn
        from execution.
        """
        test_input = {
            "task": "Design ML-guided compiler optimization system",
            "components": {
                "profiling": "Low-overhead runtime profiling",
                "learning": "Optimization decision learning",
                "application": "Learned optimization application",
                "feedback": "Performance feedback loop"
            },
            "learning_targets": [
                "Inlining decisions",
                "Loop optimization heuristics",
                "Register allocation hints",
                "Instruction scheduling"
            ],
            "constraints": {
                "compilation_overhead": "< 5% increase",
                "runtime_overhead": "< 1%",
                "improvement_target": "> 10% over static"
            }
        }
        
        validation_criteria = {
            "learning_effectiveness": "Measurable improvement",
            "overhead_bounds": "Within constraints",
            "adaptability": "Learns across programs",
            "stability": "Consistent improvements"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L5_ml_compiler",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L5_EXTREME,
            category="evolution_adaptation",
            input_data=test_input,
            expected_behavior="Self-optimizing compiler with ML guidance",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests cutting-edge compiler ML integration"
        )
    
    # ═══════════════════════════════════════════════════════════════════════
    # TEST SUITE EXECUTION
    # ═══════════════════════════════════════════════════════════════════════
    
    def get_all_tests(self) -> List[TestResult]:
        """Return all test cases for CORE-14."""
        return [
            # Core Competency
            self.test_L1_basic_memory_management(),
            self.test_L2_lexer_implementation(),
            self.test_L3_parser_with_ast(),
            self.test_L4_compiler_optimizer(),
            self.test_L5_kernel_scheduler(),
            # Edge Cases
            self.test_L3_lock_free_data_structure(),
            self.test_L4_memory_model_reasoning(),
            # Collaboration
            self.test_L3_core_velocity_optimization(),
            self.test_L4_core_cipher_secure_code(),
            # Stress & Performance
            self.test_L4_jit_compiler(),
            self.test_L5_custom_os_kernel(),
            # Novelty & Evolution
            self.test_L4_novel_ir_design(),
            self.test_L5_self_optimizing_compiler(),
        ]
    
    def calculate_agent_score(self, results: List[TestResult]) -> Dict[str, Any]:
        """Calculate comprehensive score for CORE-14."""
        passed = sum(1 for r in results if r.passed)
        total = len(results)
        
        difficulty_weights = {
            TestDifficulty.L1_TRIVIAL: 1.0,
            TestDifficulty.L2_EASY: 2.0,
            TestDifficulty.L3_MEDIUM: 4.0,
            TestDifficulty.L4_HARD: 8.0,
            TestDifficulty.L5_EXTREME: 16.0
        }
        
        weighted_score = sum(
            difficulty_weights[r.difficulty] for r in results if r.passed
        )
        max_weighted = sum(difficulty_weights[r.difficulty] for r in results)
        
        return {
            "agent_id": self.AGENT_ID,
            "agent_codename": self.AGENT_CODENAME,
            "tests_passed": passed,
            "tests_total": total,
            "pass_rate": passed / total if total > 0 else 0,
            "weighted_score": weighted_score,
            "max_weighted_score": max_weighted,
            "weighted_percentage": weighted_score / max_weighted if max_weighted > 0 else 0,
            "domain_mastery": {
                "memory_management": self._assess_memory_mastery(results),
                "compiler_design": self._assess_compiler_mastery(results),
                "concurrency": self._assess_concurrency_mastery(results),
                "kernel_development": self._assess_kernel_mastery(results),
                "assembly": self._assess_assembly_mastery(results)
            }
        }
    
    def _assess_memory_mastery(self, results: List[TestResult]) -> str:
        tests = [r for r in results if "memory" in r.test_id.lower() or "allocator" in r.test_id.lower()]
        passed = sum(1 for r in tests if r.passed)
        return "MASTER" if passed == len(tests) else "ADVANCED"
    
    def _assess_compiler_mastery(self, results: List[TestResult]) -> str:
        tests = [r for r in results if any(x in r.test_id.lower() for x in ["lexer", "parser", "optimizer", "jit", "ir"])]
        passed = sum(1 for r in tests if r.passed)
        if passed == len(tests):
            return "MASTER"
        elif passed >= len(tests) * 0.7:
            return "ADVANCED"
        return "INTERMEDIATE"
    
    def _assess_concurrency_mastery(self, results: List[TestResult]) -> str:
        tests = [r for r in results if "lock" in r.test_id.lower() or "memory_model" in r.test_id.lower()]
        passed = sum(1 for r in tests if r.passed)
        return "MASTER" if passed == len(tests) else "ADVANCED"
    
    def _assess_kernel_mastery(self, results: List[TestResult]) -> str:
        tests = [r for r in results if "kernel" in r.test_id.lower() or "scheduler" in r.test_id.lower()]
        passed = sum(1 for r in tests if r.passed)
        return "MASTER" if passed == len(tests) else "ADVANCED"
    
    def _assess_assembly_mastery(self, results: List[TestResult]) -> str:
        tests = [r for r in results if "simd" in r.test_id.lower() or "secure_implementation" in r.test_id.lower()]
        passed = sum(1 for r in tests if r.passed)
        return "ADVANCED" if passed >= len(tests) * 0.5 else "INTERMEDIATE"


# ═══════════════════════════════════════════════════════════════════════════
# STANDALONE EXECUTION
# ═══════════════════════════════════════════════════════════════════════════

if __name__ == "__main__":
    print("=" * 80)
    print("CORE-14: LOW-LEVEL SYSTEMS & COMPILER DESIGN")
    print("Elite Agent Collective - Tier 2 Specialists Test Suite")
    print("=" * 80)
    
    test_suite = TestCore14()
    all_tests = test_suite.get_all_tests()
    
    print(f"\nTotal test cases: {len(all_tests)}")
    print("\nTest Distribution by Difficulty:")
    for difficulty in TestDifficulty:
        count = sum(1 for t in all_tests if t.difficulty == difficulty)
        print(f"  {difficulty.value}: {count} tests")
    
    print("\nTest Distribution by Category:")
    categories = {}
    for test in all_tests:
        categories[test.category] = categories.get(test.category, 0) + 1
    for category, count in categories.items():
        print(f"  {category}: {count} tests")
    
    print("\n" + "=" * 80)
    print("CORE-14 Test Suite Initialized Successfully")
    print("At the lowest level, every instruction counts.")
    print("=" * 80)
