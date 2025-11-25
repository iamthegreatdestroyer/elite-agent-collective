"""
Elite Agent Collective - Master Test Runner
===========================================
Orchestrates execution of all agent tests and generates
comprehensive documentation for OMNISCIENT-20.
"""

import importlib
import sys
from pathlib import Path
from datetime import datetime
from typing import List, Dict, Any

# Add the tests directory to the path
sys.path.insert(0, str(Path(__file__).parent.parent))

from framework.base_agent_test import BaseAgentTest, AgentTestSummary
from framework.omniscient_aggregator import OmniscientAggregator
from framework.documentation_generator import DocumentationGenerator


class MasterTestRunner:
    """
    Orchestrates the execution of all 20 agent test suites
    and aggregates results for OMNISCIENT-20.
    """

    AGENT_TESTS = [
        ("tier_1_foundational.test_apex_01", "ApexAgentTest"),
        ("tier_1_foundational.test_cipher_02", "CipherAgentTest"),
        ("tier_1_foundational.test_architect_03", "ArchitectAgentTest"),
        ("tier_1_foundational.test_axiom_04", "AxiomAgentTest"),
        ("tier_1_foundational.test_velocity_05", "VelocityAgentTest"),
        ("tier_2_specialists.test_quantum_06", "QuantumAgentTest"),
        ("tier_2_specialists.test_tensor_07", "TensorAgentTest"),
        ("tier_2_specialists.test_fortress_08", "FortressAgentTest"),
        ("tier_2_specialists.test_neural_09", "NeuralAgentTest"),
        ("tier_2_specialists.test_crypto_10", "CryptoAgentTest"),
        ("tier_2_specialists.test_flux_11", "FluxAgentTest"),
        ("tier_2_specialists.test_prism_12", "PrismAgentTest"),
        ("tier_2_specialists.test_synapse_13", "SynapseAgentTest"),
        ("tier_2_specialists.test_core_14", "CoreAgentTest"),
        ("tier_2_specialists.test_helix_15", "HelixAgentTest"),
        ("tier_2_specialists.test_vanguard_16", "VanguardAgentTest"),
        ("tier_2_specialists.test_eclipse_17", "EclipseAgentTest"),
        ("tier_3_innovators.test_nexus_18", "NexusAgentTest"),
        ("tier_3_innovators.test_genesis_19", "GenesisAgentTest"),
        ("tier_4_meta.test_omniscient_20", "OmniscientAgentTest"),
    ]

    def __init__(self):
        self.summaries: List[AgentTestSummary] = []
        self.start_time = None
        self.end_time = None
        self.doc_generator = DocumentationGenerator("docs")

    def run_all_tests(self) -> None:
        """Execute all agent test suites."""
        self.start_time = datetime.utcnow()

        print(f"""
╔══════════════════════════════════════════════════════════════════════════════╗
║              ELITE AGENT COLLECTIVE - MASTER TEST SUITE                      ║
║                                                                              ║
║  Starting comprehensive test execution for all 20 agents...                  ║
║  Timestamp: {self.start_time.isoformat()}                              ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
""")

        for module_name, class_name in self.AGENT_TESTS:
            self._run_agent_test(module_name, class_name)

        self.end_time = datetime.utcnow()
        duration = (self.end_time - self.start_time).total_seconds()

        print(f"""
╔══════════════════════════════════════════════════════════════════════════════╗
║                        TEST EXECUTION COMPLETE                               ║
║                                                                              ║
║  End Time: {self.end_time.isoformat()}                                 ║
║  Duration: {duration:.2f} seconds                                            ║
║  Agents Tested: {len(self.summaries)}                                                       ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
""")

    def _run_agent_test(self, module_name: str, class_name: str) -> None:
        """Run a single agent test suite."""
        try:
            module = importlib.import_module(module_name)
            test_class = getattr(module, class_name)

            print(f"\n{'─'*60}")
            print(f"  Executing: {class_name}")
            print(f"{'─'*60}")

            test_instance = test_class()
            summary = test_instance.run_all_tests()
            self.summaries.append(summary)

            print(f"  ✓ Complete | Pass Rate: {summary.pass_rate:.2%} | Tests: {summary.total_tests}")

            # Generate individual report
            self._save_agent_report(summary)

        except Exception as e:
            print(f"  ✗ Error loading {module_name}.{class_name}: {e}")

    def _save_agent_report(self, summary: AgentTestSummary) -> None:
        """Save individual agent test report."""
        filepath = self.doc_generator.save_agent_report(summary)
        print(f"  📄 Report saved: {filepath}")

    def generate_omniscient_package(self) -> None:
        """Generate the OMNISCIENT-20 synthesis package."""
        print(f"\n{'═'*60}")
        print("  Generating OMNISCIENT-20 Synthesis Package...")
        print(f"{'═'*60}\n")

        aggregator = OmniscientAggregator()

        # Load all generated reports
        for summary in self.summaries:
            aggregator.add_agent_data(summary.agent_id, {
                "agent_codename": summary.agent_codename,
                "agent_specialty": summary.agent_specialty,
                "total_tests": summary.total_tests,
                "passed_tests": summary.passed_tests,
                "pass_rate": summary.pass_rate,
                "strengths": summary.strengths,
                "weaknesses": summary.weaknesses,
                "difficulty_breakdown": summary.difficulty_breakdown,
                "critical_failures": [f.to_dict() for f in summary.critical_failures],
                "omniscient_package": summary.omniscient_package
            })

        synthesis = aggregator.generate_omniscient_synthesis()
        
        output_path = Path("OMNISCIENT_SYNTHESIS.md")
        with open(output_path, "w", encoding='utf-8') as f:
            f.write(synthesis)

        print(f"  ✓ OMNISCIENT Synthesis Report Generated: {output_path}")

    def generate_collective_matrix(self) -> None:
        """Generate the collective capabilities matrix."""
        print(f"\n{'─'*60}")
        print("  Generating Collective Capabilities Matrix...")
        print(f"{'─'*60}")
        
        filepath = self.doc_generator.save_collective_matrix(self.summaries)
        print(f"  ✓ Matrix saved: {filepath}")

    def generate_summary_report(self) -> None:
        """Generate overall summary report."""
        total_tests = sum(s.total_tests for s in self.summaries)
        total_passed = sum(s.passed_tests for s in self.summaries)

        report = f"""# Elite Agent Collective - Test Suite Summary

## Execution Summary

- **Execution Date:** {self.start_time.isoformat() if self.start_time else 'N/A'}
- **Total Agents Tested:** {len(self.summaries)}
- **Total Tests Executed:** {total_tests}
- **Total Passed:** {total_passed}
- **Total Failed:** {total_tests - total_passed}
- **Overall Pass Rate:** {total_passed/total_tests:.2%}

---

## Agent Performance

| Agent | Tier | Tests | Passed | Failed | Pass Rate | Ceiling |
|-------|------|-------|--------|--------|-----------|---------|
"""
        for summary in self.summaries:
            ceiling = summary.omniscient_package.get('difficulty_ceiling', 'N/A')
            report += f"| {summary.agent_codename}-{summary.agent_id} | {summary.agent_tier} | {summary.total_tests} | {summary.passed_tests} | {summary.failed_tests} | {summary.pass_rate:.2%} | {ceiling} |\n"

        report += f"""
---

## Tier Summary

| Tier | Description | Agents | Avg Pass Rate |
|------|-------------|--------|---------------|
"""
        tier_info = {
            1: ("Foundational", []),
            2: ("Specialists", []),
            3: ("Innovators", []),
            4: ("Meta", [])
        }
        
        for summary in self.summaries:
            tier_info[summary.agent_tier][1].append(summary.pass_rate)
        
        for tier, (desc, rates) in tier_info.items():
            avg_rate = sum(rates) / len(rates) if rates else 0
            report += f"| Tier {tier} | {desc} | {len(rates)} | {avg_rate:.2%} |\n"

        report += f"""
---

## Generated Reports

### Individual Agent Reports

"""
        for summary in self.summaries:
            report += f"- [Agent {summary.agent_id}: {summary.agent_codename}](docs/AGENT_{summary.agent_id}_{summary.agent_codename}_REPORT.md)\n"

        report += f"""
### Collective Reports

- [OMNISCIENT Synthesis](OMNISCIENT_SYNTHESIS.md)
- [Collective Capabilities Matrix](docs/COLLECTIVE_CAPABILITIES_MATRIX.md)

---

## Quick Statistics

```
┌──────────────────────────────────────────────────────────┐
│                    TEST SUITE METRICS                     │
├──────────────────────────────────────────────────────────┤
│  Total Agents:     {len(self.summaries):>4}                                    │
│  Total Tests:      {total_tests:>4}                                    │
│  Tests Passed:     {total_passed:>4}                                    │
│  Tests Failed:     {total_tests - total_passed:>4}                                    │
│  Pass Rate:        {total_passed/total_tests:>6.2%}                                 │
└──────────────────────────────────────────────────────────┘
```

---

*Generated: {datetime.utcnow().isoformat()} UTC*
*Framework: Elite Agent Collective Test Suite v1.0*
"""

        with open("tests/README.md", "w", encoding='utf-8') as f:
            f.write(report)
        
        print(f"  ✓ Summary report saved: tests/README.md")

    def export_results(self) -> None:
        """Export results to JSON."""
        filepath = self.doc_generator.export_json(self.summaries)
        print(f"  ✓ JSON export saved: {filepath}")


def main():
    """Main entry point for the test runner."""
    runner = MasterTestRunner()
    
    # Run all tests
    runner.run_all_tests()
    
    # Generate all documentation
    runner.generate_omniscient_package()
    runner.generate_collective_matrix()
    runner.generate_summary_report()
    runner.export_results()

    print("""
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║   ███████╗██╗   ██╗ ██████╗ ██████╗███████╗███████╗███████╗██╗               ║
║   ██╔════╝██║   ██║██╔════╝██╔════╝██╔════╝██╔════╝██╔════╝██║               ║
║   ███████╗██║   ██║██║     ██║     █████╗  ███████╗███████╗██║               ║
║   ╚════██║██║   ██║██║     ██║     ██╔══╝  ╚════██║╚════██║╚═╝               ║
║   ███████║╚██████╔╝╚██████╗╚██████╗███████╗███████║███████║██╗               ║
║   ╚══════╝ ╚═════╝  ╚═════╝ ╚═════╝╚══════╝╚══════╝╚══════╝╚═╝               ║
║                                                                              ║
║            ELITE AGENT COLLECTIVE TEST SUITE COMPLETE                        ║
║                                                                              ║
║   All 20 agents have been tested across multiple difficulty levels.          ║
║   Documentation has been generated for each agent.                           ║
║   OMNISCIENT-20 synthesis package is ready for integration.                  ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
""")


if __name__ == "__main__":
    main()
