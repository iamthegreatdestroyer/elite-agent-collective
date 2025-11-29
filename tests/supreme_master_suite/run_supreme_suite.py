#!/usr/bin/env python3
"""
Supreme Master Test Suite CLI Runner
=====================================
Command-line interface for running the Supreme Test Suite
with support for all test modes and configuration options.

Usage:
    python run_supreme_suite.py supreme [OPTIONS]
    python run_supreme_suite.py randomize [OPTIONS]
    python run_supreme_suite.py evolution-report [OPTIONS]
    python run_supreme_suite.py export-knowledge [OPTIONS]
"""

import argparse
import json
import sys
from datetime import datetime
from pathlib import Path
from typing import Optional

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent))

from supreme_master_suite.master_orchestrator import (
    MasterOrchestrator,
    TestMode,
    AGENT_REGISTRY,
    TIER_DEFINITIONS,
)
from supreme_master_suite.omniscient_learning_db import OmniscientLearningDB
from supreme_master_suite.randomized_scenario_engine import (
    RandomizedScenarioEngine,
    ComplexityLevel,
    ChallengeType,
)
from supreme_master_suite.collective_intelligence import CollectiveIntelligence
from supreme_master_suite.evolution_tracker import EvolutionTracker


def print_banner():
    """Print the Supreme Suite banner."""
    print("""
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║   ███████╗██╗   ██╗██████╗ ██████╗ ███████╗███╗   ███╗███████╗               ║
║   ██╔════╝██║   ██║██╔══██╗██╔══██╗██╔════╝████╗ ████║██╔════╝               ║
║   ███████╗██║   ██║██████╔╝██████╔╝█████╗  ██╔████╔██║█████╗                 ║
║   ╚════██║██║   ██║██╔═══╝ ██╔══██╗██╔══╝  ██║╚██╔╝██║██╔══╝                 ║
║   ███████║╚██████╔╝██║     ██║  ██║███████╗██║ ╚═╝ ██║███████╗               ║
║   ╚══════╝ ╚═════╝ ╚═╝     ╚═╝  ╚═╝╚══════╝╚═╝     ╚═╝╚══════╝               ║
║                                                                              ║
║                    MASTER TEST SUITE FOR 40 ELITE AGENTS                     ║
║                                                                              ║
║   OMNISCIENT-20 Learning Database │ Collective Intelligence Engine          ║
║   Randomized Scenario Generator │ Evolution Tracking System                  ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
    """)


def cmd_supreme(args):
    """Run the full supreme test suite."""
    print_banner()
    print("\n📋 SUPREME TEST SUITE EXECUTION")
    print("=" * 60)

    # Parse test mode
    mode_map = {
        "structured": TestMode.STRUCTURED,
        "randomized": TestMode.RANDOMIZED,
        "adaptive": TestMode.ADAPTIVE,
        "chaos": TestMode.CHAOS,
        "evolution": TestMode.EVOLUTION,
    }
    mode = mode_map.get(args.mode.lower(), TestMode.STRUCTURED)

    print(f"\n🔧 Configuration:")
    print(f"   Mode: {mode.value}")
    print(f"   Chaos Probability: {args.chaos_probability:.1%}")
    print(f"   Seed: {args.seed or 'Random'}")
    if args.tiers:
        print(f"   Target Tiers: {args.tiers}")
    if args.agents:
        print(f"   Target Agents: {args.agents}")

    # Initialize components
    db_path = args.db_path or str(Path(__file__).parent / "reports" / "learning.db")
    learning_db = OmniscientLearningDB(db_path)
    orchestrator = MasterOrchestrator(learning_db=learning_db)

    print(f"\n🚀 Starting test execution...")
    print(f"   Total Agents: {len(AGENT_REGISTRY)}")
    print(f"   Total Tiers: {len(TIER_DEFINITIONS)}")

    # Parse target tiers and agents
    target_tiers = None
    if args.tiers:
        target_tiers = [int(t) for t in args.tiers.split(",")]

    target_agents = None
    if args.agents:
        target_agents = [a.strip() for a in args.agents.split(",")]

    # Run tests
    start_time = datetime.now()
    result = orchestrator.run_supreme_test(
        mode=mode,
        target_agents=target_agents,
        target_tiers=target_tiers,
        chaos_probability=args.chaos_probability,
        seed=args.seed,
    )
    duration = (datetime.now() - start_time).total_seconds()

    # Print results
    print("\n" + "=" * 60)
    print("📊 TEST RESULTS")
    print("=" * 60)

    print(f"\n   Execution ID: {result.execution_id}")
    print(f"   Duration: {duration:.2f} seconds")
    print(f"   Agents Tested: {result.agents_tested}")
    print(f"   Total Tests: {result.total_tests}")
    print(f"   Passed: {result.passed_tests}")
    print(f"   Failed: {result.failed_tests}")
    print(f"   Pass Rate: {result.pass_rate:.1%}")

    print(f"\n📈 Scores:")
    print(f"   Collaboration Score: {result.collaboration_score:.2f}")
    print(f"   Innovation Score: {result.innovation_score:.2f}")
    print(f"   Efficiency Score: {result.efficiency_score:.2f}")

    if result.chaos_events_handled > 0:
        print(f"\n⚡ Chaos Events Handled: {result.chaos_events_handled}")

    print(f"\n🏛️ Tier Results:")
    for tier, data in sorted(result.tier_results.items()):
        tier_name = TIER_DEFINITIONS.get(tier, {}).get("name", f"Tier {tier}")
        pass_rate = data.get("pass_rate", 0)
        print(f"   Tier {tier} ({tier_name}): {pass_rate:.1%}")

    if result.cross_tier_synergies:
        print(f"\n🔗 Synergies Detected: {len(result.cross_tier_synergies)}")

    if result.emergent_patterns:
        print(f"\n✨ Emergent Patterns: {len(result.emergent_patterns)}")

    print(f"\n📋 Evolution Recommendations:")
    for rec in result.evolution_recommendations[:5]:
        print(f"   • {rec}")

    # Save results if output path specified
    if args.output:
        output_path = Path(args.output)
        output_path.parent.mkdir(parents=True, exist_ok=True)
        with open(output_path, "w") as f:
            json.dump(result.to_dict(), f, indent=2)
        print(f"\n💾 Results saved to: {output_path}")

    # Update evolution tracker
    tracker = EvolutionTracker(str(Path(__file__).parent / "reports"))
    tracker.capture_snapshot(result)

    print("\n" + "=" * 60)
    print("✅ SUPREME TEST SUITE COMPLETE")
    print("=" * 60)

    learning_db.close()

    # Return exit code based on pass rate
    return 0 if result.pass_rate >= 0.85 else 1


def cmd_randomize(args):
    """Generate and run randomized scenarios."""
    print_banner()
    print("\n🎲 RANDOMIZED SCENARIO GENERATION")
    print("=" * 60)

    engine = RandomizedScenarioEngine(seed=args.seed)

    # Parse complexity
    complexity = None
    if args.complexity:
        complexity_map = {
            "atomic": ComplexityLevel.ATOMIC,
            "molecular": ComplexityLevel.MOLECULAR,
            "compound": ComplexityLevel.COMPOUND,
            "complex": ComplexityLevel.COMPLEX,
            "universal": ComplexityLevel.UNIVERSAL,
        }
        complexity = complexity_map.get(args.complexity.lower())

    # Parse challenge type
    challenge = None
    if args.challenge:
        challenge_map = {
            "speed": ChallengeType.SPEED_RUN,
            "accuracy": ChallengeType.ACCURACY_FOCUS,
            "resource": ChallengeType.RESOURCE_CONSTRAINED,
            "adversarial": ChallengeType.ADVERSARIAL,
            "creative": ChallengeType.CREATIVE,
            "collaborative": ChallengeType.COLLABORATIVE,
            "competitive": ChallengeType.COMPETITIVE,
            "evolutionary": ChallengeType.EVOLUTIONARY,
            "chaos": ChallengeType.CHAOS,
        }
        challenge = challenge_map.get(args.challenge.lower())

    print(f"\n🔧 Configuration:")
    print(f"   Count: {args.count}")
    print(f"   Complexity: {args.complexity or 'Random'}")
    print(f"   Challenge: {args.challenge or 'Random'}")
    print(f"   Chaos Probability: {args.chaos_probability:.1%}")
    print(f"   Seed: {args.seed or 'Random'}")

    # Generate scenarios
    print(f"\n🎲 Generating {args.count} scenarios...")

    if args.diversity_weighted:
        scenarios = engine.generate_batch(
            count=args.count,
            diversity_weighted=True,
            seed=args.seed,
        )
    else:
        scenarios = []
        for _ in range(args.count):
            scenario = engine.generate_random_scenario(
                complexity=complexity,
                challenge_type=challenge,
                chaos_probability=args.chaos_probability,
            )
            scenarios.append(scenario)

    # Print scenarios
    print(f"\n📋 Generated Scenarios:")
    for i, scenario in enumerate(scenarios, 1):
        print(f"\n   {i}. {scenario.name}")
        print(f"      ID: {scenario.scenario_id}")
        print(f"      Complexity: {scenario.complexity.name}")
        print(f"      Challenge: {scenario.challenge_type.value}")
        print(f"      Required Agents: {len(scenario.required_agents)}")
        print(f"      Chaos Events: {len(scenario.chaos_events)}")

    # Get statistics
    stats = engine.get_statistics()
    print(f"\n📊 Generation Statistics:")
    print(f"   Total Generated: {stats['total_generated']}")
    print(f"   Avg Agents/Scenario: {stats['average_agents_per_scenario']:.1f}")

    # Run scenarios if requested
    if args.execute:
        print(f"\n🚀 Executing scenarios...")
        db_path = str(Path(__file__).parent / "reports" / "learning.db")
        learning_db = OmniscientLearningDB(db_path)
        orchestrator = MasterOrchestrator(learning_db=learning_db)

        for scenario in scenarios[:5]:  # Limit to 5 for demo
            result = orchestrator.run_supreme_test(
                mode=TestMode.RANDOMIZED,
                target_agents=scenario.required_agents,
                chaos_probability=len(scenario.chaos_events) / 10.0,
            )
            print(f"   {scenario.name}: {result.pass_rate:.1%}")

        learning_db.close()

    # Save scenarios if output path specified
    if args.output:
        output_path = Path(args.output)
        output_path.parent.mkdir(parents=True, exist_ok=True)
        with open(output_path, "w") as f:
            json.dump([s.to_dict() for s in scenarios], f, indent=2)
        print(f"\n💾 Scenarios saved to: {output_path}")

    print("\n" + "=" * 60)
    print("✅ SCENARIO GENERATION COMPLETE")
    print("=" * 60)

    return 0


def cmd_evolution_report(args):
    """Generate OMNISCIENT evolution report."""
    print_banner()
    print("\n🧬 EVOLUTION REPORT GENERATION")
    print("=" * 60)

    # Initialize components
    tracker = EvolutionTracker(str(Path(__file__).parent / "reports"))
    db_path = str(Path(__file__).parent / "reports" / "learning.db")

    # Load existing data
    learning_db = OmniscientLearningDB(db_path)

    print(f"\n📊 Analyzing evolution data...")

    # Generate report
    report = tracker.generate_evolution_report()

    print(f"\n📋 Report Generated:")
    print(f"   Report ID: {report.report_id}")
    print(f"   Collective Growth: {report.collective_growth:+.2%}")
    print(f"   Top Improvers: {len(report.top_improvers)}")
    print(f"   Declining Agents: {len(report.declining_agents)}")
    print(f"   Breakthrough Opportunities: {len(report.breakthrough_opportunities)}")

    print(f"\n📋 Recommendations:")
    for rec in report.recommendations[:5]:
        print(f"   • {rec}")

    # Save report
    output_path = Path(args.output) if args.output else Path(__file__).parent / "reports" / f"{report.report_id}.md"
    output_path.parent.mkdir(parents=True, exist_ok=True)
    tracker.save_report(report, str(output_path))

    print(f"\n💾 Report saved to: {output_path}")

    # Print full report if requested
    if args.verbose:
        print("\n" + "=" * 60)
        print("FULL REPORT")
        print("=" * 60)
        print(report.markdown_report)

    learning_db.close()

    print("\n" + "=" * 60)
    print("✅ EVOLUTION REPORT COMPLETE")
    print("=" * 60)

    return 0


def cmd_export_knowledge(args):
    """Export learning database for Agent #20."""
    print_banner()
    print("\n📦 KNOWLEDGE EXPORT FOR OMNISCIENT-20")
    print("=" * 60)

    # Initialize learning database
    db_path = args.db_path or str(Path(__file__).parent / "reports" / "learning.db")
    learning_db = OmniscientLearningDB(db_path)

    print(f"\n📊 Exporting knowledge from: {db_path}")

    # Export knowledge
    knowledge = learning_db.export_omniscient_knowledge()

    print(f"\n📋 Knowledge Package:")
    print(f"   Export Timestamp: {knowledge['export_timestamp']}")
    print(f"   Agents Tracked: {knowledge['metadata']['unique_agents']}")
    print(f"   Total Learning Records: {knowledge['metadata']['total_learning_records']}")
    print(f"   Collaboration Pairs: {knowledge['metadata']['collaboration_pairs']}")

    if knowledge['synthesized_patterns']:
        patterns = knowledge['synthesized_patterns']
        print(f"\n   Capability Patterns: {len(patterns.get('capability_patterns', []))}")
        print(f"   Synergy Patterns: {len(patterns.get('synergy_patterns', []))}")
        print(f"   Anti-Patterns: {len(patterns.get('anti_patterns', []))}")

    if knowledge['evolution_recommendations']:
        print(f"\n📋 Evolution Recommendations ({len(knowledge['evolution_recommendations'])}):")
        for rec in knowledge['evolution_recommendations'][:5]:
            print(f"   • [{rec['priority']}] {rec['action']}")

    # Save knowledge package
    output_path = Path(args.output) if args.output else Path(__file__).parent / "reports" / "omniscient_knowledge.json"
    output_path.parent.mkdir(parents=True, exist_ok=True)
    with open(output_path, "w") as f:
        json.dump(knowledge, f, indent=2)

    print(f"\n💾 Knowledge package saved to: {output_path}")

    learning_db.close()

    print("\n" + "=" * 60)
    print("✅ KNOWLEDGE EXPORT COMPLETE")
    print("=" * 60)

    return 0


def main():
    """Main entry point for the CLI."""
    parser = argparse.ArgumentParser(
        description="Supreme Master Test Suite for 40 Elite Agents",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s supreme --mode structured
  %(prog)s supreme --mode chaos --chaos-probability 0.3
  %(prog)s randomize --count 10 --diversity-weighted
  %(prog)s evolution-report --verbose
  %(prog)s export-knowledge --output knowledge.json
        """,
    )

    subparsers = parser.add_subparsers(dest="command", help="Commands")

    # Supreme command
    supreme_parser = subparsers.add_parser(
        "supreme",
        help="Run full supreme test suite",
    )
    supreme_parser.add_argument(
        "--mode",
        choices=["structured", "randomized", "adaptive", "chaos", "evolution"],
        default="structured",
        help="Test mode (default: structured)",
    )
    supreme_parser.add_argument(
        "--chaos-probability",
        type=float,
        default=0.0,
        help="Probability of chaos events (0.0-1.0)",
    )
    supreme_parser.add_argument(
        "--seed",
        type=int,
        help="Random seed for reproducibility",
    )
    supreme_parser.add_argument(
        "--tiers",
        type=str,
        help="Comma-separated tier numbers to test",
    )
    supreme_parser.add_argument(
        "--agents",
        type=str,
        help="Comma-separated agent IDs to test",
    )
    supreme_parser.add_argument(
        "--output",
        type=str,
        help="Output file for results (JSON)",
    )
    supreme_parser.add_argument(
        "--db-path",
        type=str,
        help="Path to learning database",
    )
    supreme_parser.set_defaults(func=cmd_supreme)

    # Randomize command
    randomize_parser = subparsers.add_parser(
        "randomize",
        help="Generate and run randomized scenarios",
    )
    randomize_parser.add_argument(
        "--count",
        type=int,
        default=5,
        help="Number of scenarios to generate",
    )
    randomize_parser.add_argument(
        "--complexity",
        choices=["atomic", "molecular", "compound", "complex", "universal"],
        help="Specific complexity level",
    )
    randomize_parser.add_argument(
        "--challenge",
        choices=["speed", "accuracy", "resource", "adversarial", "creative", "collaborative", "competitive", "evolutionary", "chaos"],
        help="Specific challenge type",
    )
    randomize_parser.add_argument(
        "--chaos-probability",
        type=float,
        default=0.3,
        help="Probability of chaos events",
    )
    randomize_parser.add_argument(
        "--diversity-weighted",
        action="store_true",
        help="Generate diverse scenarios",
    )
    randomize_parser.add_argument(
        "--execute",
        action="store_true",
        help="Execute generated scenarios",
    )
    randomize_parser.add_argument(
        "--seed",
        type=int,
        help="Random seed for reproducibility",
    )
    randomize_parser.add_argument(
        "--output",
        type=str,
        help="Output file for scenarios (JSON)",
    )
    randomize_parser.set_defaults(func=cmd_randomize)

    # Evolution report command
    evolution_parser = subparsers.add_parser(
        "evolution-report",
        help="Generate OMNISCIENT evolution report",
    )
    evolution_parser.add_argument(
        "--output",
        type=str,
        help="Output file for report (Markdown)",
    )
    evolution_parser.add_argument(
        "--verbose",
        action="store_true",
        help="Print full report to console",
    )
    evolution_parser.set_defaults(func=cmd_evolution_report)

    # Export knowledge command
    export_parser = subparsers.add_parser(
        "export-knowledge",
        help="Export learning database for Agent #20",
    )
    export_parser.add_argument(
        "--output",
        type=str,
        help="Output file for knowledge package (JSON)",
    )
    export_parser.add_argument(
        "--db-path",
        type=str,
        help="Path to learning database",
    )
    export_parser.set_defaults(func=cmd_export_knowledge)

    # Parse arguments
    args = parser.parse_args()

    if args.command is None:
        parser.print_help()
        return 0

    # Run command
    return args.func(args)


if __name__ == "__main__":
    sys.exit(main())
