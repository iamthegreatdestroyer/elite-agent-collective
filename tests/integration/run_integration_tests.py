"""
═══════════════════════════════════════════════════════════════════════════════
                      COMPREHENSIVE INTEGRATION TEST RUNNER
                      Elite Agent Collective Test Suite
═══════════════════════════════════════════════════════════════════════════════
Purpose: Master test runner that orchestrates all integration tests
Coverage: Executes all test suites with progress tracking and reporting

Runs:
1. Agent Invocation Tests - Verify all 40 agents work correctly
2. Multi-Agent Collaboration Tests - Test cross-agent interactions  
3. Evolution Protocol Tests - Validate learning and adaptation
4. MNEMONIC Memory Tests - Memory system validation
5. Performance Benchmarks - Load and stress testing
6. GitHub Actions Workflow Tests - CI/CD validation
═══════════════════════════════════════════════════════════════════════════════
"""

import sys
import time
import json
from pathlib import Path
from datetime import datetime
from typing import List, Dict, Any, Tuple
from dataclasses import dataclass, asdict
from enum import Enum

# Add test framework to path
sys.path.insert(0, str(Path(__file__).parent / "framework"))

# Import test modules
from test_agent_invocation import run_agent_invocation_tests
from test_collective_problem_solving import run_collective_tests
from test_evolution_protocols import run_evolution_tests
from test_mnemonic_memory import run_mnemonic_tests
from test_performance_benchmarks import run_performance_benchmarks
from test_github_actions_workflow import run_github_actions_tests


class TestCategory(Enum):
    """Categories of integration tests."""
    AGENT_INVOCATION = "Agent Invocation"
    COLLECTIVE = "Multi-Agent Collaboration"
    EVOLUTION = "Evolution Protocols"
    MEMORY = "MNEMONIC Memory"
    PERFORMANCE = "Performance Benchmarks"
    GITHUB_ACTIONS = "GitHub Actions Workflow"


@dataclass
class TestSuiteResult:
    """Results from a test suite."""
    category: str
    suite_name: str
    passed: int
    failed: int
    skipped: int
    duration_seconds: float
    success: bool
    details: Dict[str, Any]


class ComprehensiveTestRunner:
    """Master test runner orchestrating all integration tests."""
    
    def __init__(self):
        self.results: List[TestSuiteResult] = []
        self.start_time = None
        self.end_time = None
        self.total_passed = 0
        self.total_failed = 0
        self.total_skipped = 0
    
    def print_header(self):
        """Print test suite header."""
        print("\n" + "█"*80)
        print("█" + " "*78 + "█")
        print("█" + "ELITE AGENT COLLECTIVE - COMPREHENSIVE INTEGRATION TEST SUITE".center(78) + "█")
        print("█" + " "*78 + "█")
        print("█"*80)
        print(f"\n  Started: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"  Workspace: Elite Agent Collective v2.0")
        print(f"  Test Scope: Full integration validation\n")
    
    def run_all_tests(self):
        """Execute all integration test suites."""
        self.start_time = time.perf_counter()
        self.print_header()
        
        # Test suite definitions with their runners
        test_suites = [
            {
                "category": TestCategory.AGENT_INVOCATION.value,
                "suite_name": "Agent Invocation Test Suite",
                "runner": run_agent_invocation_tests,
                "description": "Verify all 40 agents can be invoked and respond correctly"
            },
            {
                "category": TestCategory.COLLECTIVE.value,
                "suite_name": "Multi-Agent Collaboration Test Suite",
                "runner": run_collective_tests,
                "description": "Validate cross-agent collaboration and communication"
            },
            {
                "category": TestCategory.EVOLUTION.value,
                "suite_name": "Evolution Protocol Test Suite",
                "runner": run_evolution_tests,
                "description": "Test learning protocols and adaptation mechanisms"
            },
            {
                "category": TestCategory.MEMORY.value,
                "suite_name": "MNEMONIC Memory Test Suite",
                "runner": run_mnemonic_tests,
                "description": "Validate memory system functionality and performance"
            },
            {
                "category": TestCategory.PERFORMANCE.value,
                "suite_name": "Performance Benchmark Suite",
                "runner": run_performance_benchmarks,
                "description": "Load testing and performance measurement"
            },
            {
                "category": TestCategory.GITHUB_ACTIONS.value,
                "suite_name": "GitHub Actions Workflow Test Suite",
                "runner": run_github_actions_tests,
                "description": "CI/CD workflow validation and verification"
            },
        ]
        
        # Run each test suite
        for suite_idx, suite_config in enumerate(test_suites, 1):
            print(f"\n{'='*80}")
            print(f"[{suite_idx}/{len(test_suites)}] {suite_config['suite_name']}")
            print(f"{'='*80}")
            print(f"Description: {suite_config['description']}")
            print(f"{'─'*80}\n")
            
            suite_start = time.perf_counter()
            
            try:
                # Run the test suite
                result = suite_config["runner"]()
                
                # Parse results based on runner return type
                if isinstance(result, tuple):
                    passed, failed = result
                    skipped = 0
                else:
                    passed = result if isinstance(result, int) else 0
                    failed = 0
                    skipped = 0
                
                suite_duration = time.perf_counter() - suite_start
                success = failed == 0
                
                # Record results
                suite_result = TestSuiteResult(
                    category=suite_config['category'],
                    suite_name=suite_config['suite_name'],
                    passed=passed,
                    failed=failed,
                    skipped=skipped,
                    duration_seconds=suite_duration,
                    success=success,
                    details={
                        "description": suite_config['description'],
                        "completion_time": datetime.now().isoformat()
                    }
                )
                
                self.results.append(suite_result)
                self.total_passed += passed
                self.total_failed += failed
                self.total_skipped += skipped
                
                # Print suite result
                status = "✓ PASSED" if success else "✗ FAILED"
                print(f"\n{status}")
                print(f"  Passed: {passed}, Failed: {failed}, Skipped: {skipped}")
                print(f"  Duration: {suite_duration:.2f}s")
                
            except Exception as e:
                print(f"\n✗ SUITE ERROR")
                print(f"  Error: {str(e)}")
                
                suite_duration = time.perf_counter() - suite_start
                suite_result = TestSuiteResult(
                    category=suite_config['category'],
                    suite_name=suite_config['suite_name'],
                    passed=0,
                    failed=1,
                    skipped=0,
                    duration_seconds=suite_duration,
                    success=False,
                    details={"error": str(e)}
                )
                
                self.results.append(suite_result)
                self.total_failed += 1
        
        self.end_time = time.perf_counter()
        self.print_summary()
    
    def print_summary(self):
        """Print comprehensive test summary."""
        total_duration = self.end_time - self.start_time
        total_tests = self.total_passed + self.total_failed
        success_rate = (self.total_passed / total_tests * 100) if total_tests > 0 else 0
        
        print("\n" + "█"*80)
        print("█" + " "*78 + "█")
        print("█" + "TEST EXECUTION SUMMARY".center(78) + "█")
        print("█" + " "*78 + "█")
        print("█"*80)
        
        # Test suites summary table
        print("\nTest Suite Results:")
        print("─"*80)
        print(f"{'Suite':<40} {'Pass':<8} {'Fail':<8} {'Status':<10}")
        print("─"*80)
        
        for result in self.results:
            status = "✓ PASS" if result.success else "✗ FAIL"
            suite_name = result.suite_name[:37] + "..." if len(result.suite_name) > 40 else result.suite_name
            print(f"{suite_name:<40} {result.passed:<8} {result.failed:<8} {status:<10}")
        
        print("─"*80)
        
        # Overall statistics
        print(f"\nOverall Statistics:")
        print(f"  Total Test Suites: {len(self.results)}")
        print(f"  Total Tests Run: {total_tests}")
        print(f"  Passed: {self.total_passed}")
        print(f"  Failed: {self.total_failed}")
        print(f"  Skipped: {self.total_skipped}")
        print(f"  Success Rate: {success_rate:.1f}%")
        print(f"  Total Duration: {total_duration:.2f}s")
        print(f"  Completion Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        
        # Result categorization
        print(f"\nResult Categories:")
        all_passed = all(r.success for r in self.results)
        no_failures = self.total_failed == 0
        all_completed = all(r.duration_seconds > 0 for r in self.results)
        
        status_items = [
            ("All test suites passed", all_passed, "✓"),
            ("No test failures", no_failures, "✓"),
            ("All tests completed", all_completed, "✓"),
        ]
        
        for item, condition, symbol in status_items:
            print(f"  {symbol} {item:<50} {'YES' if condition else 'NO'}")
        
        # Performance summary by category
        print(f"\nPerformance by Category:")
        print("─"*80)
        
        for result in sorted(self.results, key=lambda r: r.duration_seconds, reverse=True):
            print(f"  {result.category:<35} {result.duration_seconds:>8.2f}s")
        
        print("─"*80)
        
        # Final status
        print(f"\n{'█'*80}")
        if self.total_failed == 0:
            print(f"█ ✓ ALL INTEGRATION TESTS PASSED".ljust(79) + "█")
        else:
            print(f"█ ✗ SOME TESTS FAILED - REVIEW RESULTS".ljust(79) + "█")
        print(f"█"*80 + "\n")
    
    def generate_report(self) -> Dict[str, Any]:
        """Generate detailed test report."""
        report = {
            "title": "Elite Agent Collective - Integration Test Report",
            "timestamp": datetime.now().isoformat(),
            "duration_seconds": self.end_time - self.start_time if self.end_time else 0,
            "summary": {
                "total_suites": len(self.results),
                "total_tests": self.total_passed + self.total_failed,
                "passed": self.total_passed,
                "failed": self.total_failed,
                "skipped": self.total_skipped,
                "success_rate": (self.total_passed / (self.total_passed + self.total_failed) * 100) 
                               if (self.total_passed + self.total_failed) > 0 else 0
            },
            "test_suites": [asdict(r) for r in self.results]
        }
        
        return report
    
    def save_report(self, output_path: str = "test_report.json"):
        """Save test report to JSON file."""
        report = self.generate_report()
        
        output_file = Path(output_path)
        output_file.parent.mkdir(parents=True, exist_ok=True)
        
        with open(output_file, 'w') as f:
            json.dump(report, f, indent=2, default=str)
        
        print(f"Test report saved to: {output_file}")
        
        return output_file


def main():
    """Main entry point."""
    runner = ComprehensiveTestRunner()
    runner.run_all_tests()
    
    # Save detailed report
    report_path = Path(__file__).parent.parent / "test_results" / f"integration_tests_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
    runner.save_report(str(report_path))
    
    # Return exit code based on results
    return 0 if runner.total_failed == 0 else 1


if __name__ == "__main__":
    exit_code = main()
    sys.exit(exit_code)
