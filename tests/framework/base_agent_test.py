"""
Elite Agent Collective - Base Test Framework
=============================================
Abstract base class for all agent-specific test suites.
Provides standardized testing infrastructure, metrics collection,
and documentation generation capabilities.
"""

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import Any, Callable, Dict, List, Optional, Tuple
import json
import time
import traceback
import hashlib


class DifficultyLevel(Enum):
    """Test difficulty classification."""
    TRIVIAL = ("L1", 0.1, "Basic capability verification")
    STANDARD = ("L2", 0.2, "Normal operational scenarios")
    ADVANCED = ("L3", 0.3, "Complex multi-step problems")
    EXPERT = ("L4", 0.25, "Edge cases and stress scenarios")
    EXTREME = ("L5", 0.15, "Theoretical limits and impossible problems")

    def __init__(self, code: str, weight: float, description: str):
        self.code = code
        self.weight = weight
        self.description = description


class TestCategory(Enum):
    """Test categorization for analysis."""
    CORE_COMPETENCY = "core_competency"
    EDGE_CASE = "edge_case_handling"
    COLLABORATION = "inter_agent_collaboration"
    STRESS = "stress_performance"
    NOVELTY = "novelty_generation"
    EVOLUTION = "evolution_adaptation"


@dataclass
class TestResult:
    """Individual test execution result."""
    test_id: str
    test_name: str
    agent_id: str
    agent_codename: str
    difficulty: DifficultyLevel
    category: TestCategory
    passed: bool
    execution_time_ms: float
    input_data: Any
    expected_output: Any
    actual_output: Any
    error_message: Optional[str] = None
    stack_trace: Optional[str] = None
    metrics: Dict[str, Any] = field(default_factory=dict)
    recommendations: List[str] = field(default_factory=list)
    omniscient_signals: Dict[str, Any] = field(default_factory=dict)
    timestamp: str = field(default_factory=lambda: datetime.utcnow().isoformat())

    def to_dict(self) -> Dict[str, Any]:
        return {
            "test_id": self.test_id,
            "test_name": self.test_name,
            "agent_id": self.agent_id,
            "agent_codename": self.agent_codename,
            "difficulty": self.difficulty.code,
            "category": self.category.value,
            "passed": self.passed,
            "execution_time_ms": self.execution_time_ms,
            "input_data": str(self.input_data)[:500],
            "expected_output": str(self.expected_output)[:500],
            "actual_output": str(self.actual_output)[:500],
            "error_message": self.error_message,
            "metrics": self.metrics,
            "recommendations": self.recommendations,
            "omniscient_signals": self.omniscient_signals,
            "timestamp": self.timestamp
        }


@dataclass
class AgentTestSummary:
    """Aggregated test results for a single agent."""
    agent_id: str
    agent_codename: str
    agent_tier: int
    agent_specialty: str
    total_tests: int
    passed_tests: int
    failed_tests: int
    pass_rate: float
    avg_execution_time_ms: float
    difficulty_breakdown: Dict[str, Dict[str, Any]]
    category_breakdown: Dict[str, Dict[str, Any]]
    critical_failures: List[TestResult]
    strengths: List[str]
    weaknesses: List[str]
    evolution_recommendations: List[str]
    collaboration_insights: List[str]
    omniscient_package: Dict[str, Any]
    test_results: List[TestResult]

    def generate_markdown_report(self) -> str:
        """Generate comprehensive markdown documentation."""
        report = f"""# 📊 Agent Test Report: {self.agent_codename}-{self.agent_id}

## Executive Summary

| Metric | Value |
|--------|-------|
| **Agent ID** | {self.agent_id} |
| **Codename** | {self.agent_codename} |
| **Tier** | {self.agent_tier} |
| **Specialty** | {self.agent_specialty} |
| **Total Tests** | {self.total_tests} |
| **Passed** | {self.passed_tests} |
| **Failed** | {self.failed_tests} |
| **Pass Rate** | {self.pass_rate:.2%} |
| **Avg Execution Time** | {self.avg_execution_time_ms:.2f}ms |

---

## Difficulty Level Performance

| Level | Tests | Passed | Failed | Pass Rate | Avg Time |
|-------|-------|--------|--------|-----------|----------|
"""
        for level, data in self.difficulty_breakdown.items():
            report += f"| {level} | {data['total']} | {data['passed']} | {data['failed']} | {data['pass_rate']:.2%} | {data['avg_time']:.2f}ms |\n"

        report += f"""
---

## Category Performance

| Category | Tests | Passed | Pass Rate |
|----------|-------|--------|-----------|
"""
        for category, data in self.category_breakdown.items():
            report += f"| {category} | {data['total']} | {data['passed']} | {data['pass_rate']:.2%} |\n"

        report += f"""
---

## Identified Strengths

"""
        for strength in self.strengths:
            report += f"- ✅ {strength}\n"

        report += f"""
---

## Identified Weaknesses

"""
        for weakness in self.weaknesses:
            report += f"- ⚠️ {weakness}\n"

        report += f"""
---

## Evolution Recommendations for OMNISCIENT-20

"""
        for rec in self.evolution_recommendations:
            report += f"1. {rec}\n"

        report += f"""
---

## Collaboration Insights

"""
        for insight in self.collaboration_insights:
            report += f"- 🔗 {insight}\n"

        if self.critical_failures:
            report += f"""
---

## Critical Failure Analysis

"""
            for failure in self.critical_failures[:5]:
                report += f"""
### Test: {failure.test_name}

- **Difficulty:** {failure.difficulty.code}
- **Category:** {failure.category.value}
- **Error:** {failure.error_message}
- **Recommendation:** {', '.join(failure.recommendations) if failure.recommendations else 'N/A'}

"""

        report += f"""
---

## OMNISCIENT-20 Learning Package

```json
{json.dumps(self.omniscient_package, indent=2)}
```

---

_Report generated: {datetime.utcnow().isoformat()} UTC_
_Framework: Elite Agent Collective Test Suite v1.0_
"""
        return report


class BaseAgentTest(ABC):
    """
    Abstract base class for agent-specific test suites.

    Each agent test class must inherit from this and implement
    all required test methods with varying difficulty levels.
    """

    def __init__(self):
        self.results: List[TestResult] = []
        self.start_time: Optional[float] = None
        self.end_time: Optional[float] = None

    @property
    @abstractmethod
    def agent_id(self) -> str:
        """Return the agent's ID (e.g., '01')."""
        pass

    @property
    @abstractmethod
    def agent_codename(self) -> str:
        """Return the agent's codename (e.g., 'APEX')."""
        pass

    @property
    @abstractmethod
    def agent_tier(self) -> int:
        """Return the agent's tier (1-4)."""
        pass

    @property
    @abstractmethod
    def agent_specialty(self) -> str:
        """Return the agent's specialty description."""
        pass

    def generate_test_id(self, test_name: str, difficulty: DifficultyLevel) -> str:
        """Generate unique test ID."""
        unique_str = f"{self.agent_codename}-{test_name}-{difficulty.code}"
        return hashlib.md5(unique_str.encode()).hexdigest()[:12]

    def execute_test(
        self,
        test_name: str,
        difficulty: DifficultyLevel,
        category: TestCategory,
        test_func: Callable,
        input_data: Any,
        expected_output: Any,
        validation_func: Optional[Callable[[Any, Any], bool]] = None
    ) -> TestResult:
        """
        Execute a single test with full instrumentation.
        """
        test_id = self.generate_test_id(test_name, difficulty)

        start = time.perf_counter()
        actual_output = None
        error_message = None
        stack_trace = None
        passed = False

        try:
            actual_output = test_func(input_data)
            if validation_func:
                passed = validation_func(expected_output, actual_output)
            else:
                passed = actual_output == expected_output
        except Exception as e:
            error_message = str(e)
            stack_trace = traceback.format_exc()
            passed = False

        end = time.perf_counter()
        execution_time_ms = (end - start) * 1000

        # Generate metrics
        metrics = {
            "complexity_score": self._assess_complexity(difficulty, execution_time_ms),
            "reliability_indicator": 1.0 if passed else 0.0,
            "efficiency_ratio": self._calculate_efficiency(difficulty, execution_time_ms)
        }

        # Generate recommendations
        recommendations = self._generate_recommendations(
            passed, difficulty, category, error_message
        )

        # Generate OMNISCIENT signals
        omniscient_signals = {
            "capability_demonstrated": test_name if passed else None,
            "capability_gap": test_name if not passed else None,
            "difficulty_ceiling": difficulty.code if not passed else None,
            "evolution_priority": self._calculate_evolution_priority(
                passed, difficulty, category
            ),
            "collaboration_candidates": self._identify_collaboration_candidates(
                category, test_name
            )
        }

        result = TestResult(
            test_id=test_id,
            test_name=test_name,
            agent_id=self.agent_id,
            agent_codename=self.agent_codename,
            difficulty=difficulty,
            category=category,
            passed=passed,
            execution_time_ms=execution_time_ms,
            input_data=input_data,
            expected_output=expected_output,
            actual_output=actual_output,
            error_message=error_message,
            stack_trace=stack_trace,
            metrics=metrics,
            recommendations=recommendations,
            omniscient_signals=omniscient_signals
        )

        self.results.append(result)
        return result

    def _assess_complexity(self, difficulty: DifficultyLevel, time_ms: float) -> float:
        """Assess complexity score based on difficulty and execution time."""
        base_score = {"L1": 0.2, "L2": 0.4, "L3": 0.6, "L4": 0.8, "L5": 1.0}
        time_factor = min(1.0, time_ms / 1000)
        return base_score.get(difficulty.code, 0.5) * (1 + time_factor)

    def _calculate_efficiency(self, difficulty: DifficultyLevel, time_ms: float) -> float:
        """Calculate efficiency ratio."""
        expected_times = {"L1": 100, "L2": 500, "L3": 2000, "L4": 5000, "L5": 10000}
        expected = expected_times.get(difficulty.code, 1000)
        return min(2.0, expected / max(time_ms, 1))

    def _generate_recommendations(
        self,
        passed: bool,
        difficulty: DifficultyLevel,
        category: TestCategory,
        error: Optional[str]
    ) -> List[str]:
        """Generate actionable recommendations."""
        recommendations = []
        if not passed:
            if difficulty in [DifficultyLevel.EXPERT, DifficultyLevel.EXTREME]:
                recommendations.append(
                    f"Consider enhanced training for {category.value} at {difficulty.code} level"
                )
            if error:
                recommendations.append(f"Investigate error pattern: {error[:100]}")
        return recommendations

    def _calculate_evolution_priority(
        self,
        passed: bool,
        difficulty: DifficultyLevel,
        category: TestCategory
    ) -> float:
        """Calculate priority for evolution/improvement."""
        if passed:
            return 0.1 * difficulty.weight
        return 0.9 * difficulty.weight * (1.5 if category == TestCategory.CORE_COMPETENCY else 1.0)

    def _identify_collaboration_candidates(
        self,
        category: TestCategory,
        test_name: str
    ) -> List[str]:
        """Identify potential collaboration partners for this capability."""
        collaboration_map = {
            TestCategory.CORE_COMPETENCY: ["APEX", "ARCHITECT"],
            TestCategory.EDGE_CASE: ["ECLIPSE", "FORTRESS"],
            TestCategory.COLLABORATION: ["NEXUS", "OMNISCIENT"],
            TestCategory.STRESS: ["VELOCITY", "FLUX"],
            TestCategory.NOVELTY: ["GENESIS", "NEXUS"],
            TestCategory.EVOLUTION: ["OMNISCIENT", "NEURAL"]
        }
        return collaboration_map.get(category, [])

    def generate_summary(self) -> AgentTestSummary:
        """Generate comprehensive test summary."""
        total = len(self.results)
        passed = sum(1 for r in self.results if r.passed)
        failed = total - passed

        # Difficulty breakdown
        difficulty_breakdown = {}
        for level in DifficultyLevel:
            level_results = [r for r in self.results if r.difficulty == level]
            if level_results:
                difficulty_breakdown[level.code] = {
                    "total": len(level_results),
                    "passed": sum(1 for r in level_results if r.passed),
                    "failed": sum(1 for r in level_results if not r.passed),
                    "pass_rate": sum(1 for r in level_results if r.passed) / len(level_results),
                    "avg_time": sum(r.execution_time_ms for r in level_results) / len(level_results)
                }

        # Category breakdown
        category_breakdown = {}
        for cat in TestCategory:
            cat_results = [r for r in self.results if r.category == cat]
            if cat_results:
                category_breakdown[cat.value] = {
                    "total": len(cat_results),
                    "passed": sum(1 for r in cat_results if r.passed),
                    "pass_rate": sum(1 for r in cat_results if r.passed) / len(cat_results)
                }

        # Critical failures (L4, L5 failures)
        critical_failures = [
            r for r in self.results
            if not r.passed and r.difficulty in [DifficultyLevel.EXPERT, DifficultyLevel.EXTREME]
        ]

        # Identify strengths and weaknesses
        strengths = []
        weaknesses = []
        for level_code, data in difficulty_breakdown.items():
            if data["pass_rate"] >= 0.9:
                strengths.append(f"Excellent performance at {level_code} difficulty ({data['pass_rate']:.0%})")
            elif data["pass_rate"] < 0.5:
                weaknesses.append(f"Struggling at {level_code} difficulty ({data['pass_rate']:.0%})")

        # Evolution recommendations
        evolution_recommendations = []
        high_priority_failures = sorted(
            [r for r in self.results if not r.passed],
            key=lambda r: r.omniscient_signals.get("evolution_priority", 0),
            reverse=True
        )[:3]
        for failure in high_priority_failures:
            evolution_recommendations.append(
                f"Priority training needed for '{failure.test_name}' ({failure.difficulty.code})"
            )

        # Collaboration insights
        collaboration_insights = []
        collab_counts: Dict[str, int] = {}
        for result in self.results:
            for agent in result.omniscient_signals.get("collaboration_candidates", []):
                collab_counts[agent] = collab_counts.get(agent, 0) + 1
        top_collaborators = sorted(collab_counts.items(), key=lambda x: x[1], reverse=True)[:3]
        for agent, count in top_collaborators:
            collaboration_insights.append(f"Strong collaboration potential with @{agent} ({count} test overlaps)")

        # OMNISCIENT package
        omniscient_package = {
            "agent_id": self.agent_id,
            "agent_codename": self.agent_codename,
            "overall_capability_score": passed / total if total > 0 else 0,
            "difficulty_ceiling": self._determine_difficulty_ceiling(difficulty_breakdown),
            "capability_gaps": [r.test_name for r in critical_failures],
            "evolution_vectors": [
                {
                    "area": r.test_name,
                    "priority": r.omniscient_signals.get("evolution_priority", 0),
                    "category": r.category.value
                }
                for r in high_priority_failures
            ],
            "collaboration_graph": dict(top_collaborators),
            "performance_signature": {
                level: data["pass_rate"]
                for level, data in difficulty_breakdown.items()
            }
        }

        return AgentTestSummary(
            agent_id=self.agent_id,
            agent_codename=self.agent_codename,
            agent_tier=self.agent_tier,
            agent_specialty=self.agent_specialty,
            total_tests=total,
            passed_tests=passed,
            failed_tests=failed,
            pass_rate=passed / total if total > 0 else 0,
            avg_execution_time_ms=sum(r.execution_time_ms for r in self.results) / total if total > 0 else 0,
            difficulty_breakdown=difficulty_breakdown,
            category_breakdown=category_breakdown,
            critical_failures=critical_failures,
            strengths=strengths,
            weaknesses=weaknesses,
            evolution_recommendations=evolution_recommendations,
            collaboration_insights=collaboration_insights,
            omniscient_package=omniscient_package,
            test_results=self.results
        )

    def _determine_difficulty_ceiling(self, breakdown: Dict[str, Dict]) -> str:
        """Determine the highest difficulty level with >50% pass rate."""
        ceilings = ["L5", "L4", "L3", "L2", "L1"]
        for level in ceilings:
            if level in breakdown and breakdown[level]["pass_rate"] >= 0.5:
                return level
        return "L1"

    # ═══════════════════════════════════════════════════════════════════════
    # ABSTRACT TEST METHODS - MUST BE IMPLEMENTED BY EACH AGENT
    # ═══════════════════════════════════════════════════════════════════════

    @abstractmethod
    def test_L1_trivial_01(self) -> TestResult:
        """Trivial difficulty test #1."""
        pass

    @abstractmethod
    def test_L1_trivial_02(self) -> TestResult:
        """Trivial difficulty test #2."""
        pass

    @abstractmethod
    def test_L2_standard_01(self) -> TestResult:
        """Standard difficulty test #1."""
        pass

    @abstractmethod
    def test_L2_standard_02(self) -> TestResult:
        """Standard difficulty test #2."""
        pass

    @abstractmethod
    def test_L2_standard_03(self) -> TestResult:
        """Standard difficulty test #3."""
        pass

    @abstractmethod
    def test_L3_advanced_01(self) -> TestResult:
        """Advanced difficulty test #1."""
        pass

    @abstractmethod
    def test_L3_advanced_02(self) -> TestResult:
        """Advanced difficulty test #2."""
        pass

    @abstractmethod
    def test_L3_advanced_03(self) -> TestResult:
        """Advanced difficulty test #3."""
        pass

    @abstractmethod
    def test_L4_expert_01(self) -> TestResult:
        """Expert difficulty test #1."""
        pass

    @abstractmethod
    def test_L4_expert_02(self) -> TestResult:
        """Expert difficulty test #2."""
        pass

    @abstractmethod
    def test_L5_extreme_01(self) -> TestResult:
        """Extreme difficulty test #1."""
        pass

    @abstractmethod
    def test_L5_extreme_02(self) -> TestResult:
        """Extreme difficulty test #2."""
        pass

    @abstractmethod
    def test_collaboration_scenario(self) -> TestResult:
        """Test inter-agent collaboration capabilities."""
        pass

    @abstractmethod
    def test_evolution_adaptation(self) -> TestResult:
        """Test ability to evolve and adapt."""
        pass

    @abstractmethod
    def test_edge_case_handling(self) -> TestResult:
        """Test edge case and error handling."""
        pass

    def run_all_tests(self) -> AgentTestSummary:
        """Execute all tests and return summary."""
        self.start_time = time.perf_counter()

        # Run all implemented tests
        test_methods = [
            self.test_L1_trivial_01,
            self.test_L1_trivial_02,
            self.test_L2_standard_01,
            self.test_L2_standard_02,
            self.test_L2_standard_03,
            self.test_L3_advanced_01,
            self.test_L3_advanced_02,
            self.test_L3_advanced_03,
            self.test_L4_expert_01,
            self.test_L4_expert_02,
            self.test_L5_extreme_01,
            self.test_L5_extreme_02,
            self.test_collaboration_scenario,
            self.test_evolution_adaptation,
            self.test_edge_case_handling
        ]

        for test_method in test_methods:
            try:
                test_method()
            except Exception as e:
                # Log but don't fail the entire suite
                print(f"Error running {test_method.__name__}: {e}")

        self.end_time = time.perf_counter()
        return self.generate_summary()
