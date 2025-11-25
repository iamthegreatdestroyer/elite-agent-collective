"""
Elite Agent Collective - Documentation Generator
================================================
Automated documentation generation for test results.
"""

from dataclasses import dataclass
from datetime import datetime
from typing import Dict, List, Any, Optional
from pathlib import Path
import json


class DocumentationGenerator:
    """
    Generates comprehensive documentation from test results.
    """

    def __init__(self, output_dir: str = "docs"):
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(parents=True, exist_ok=True)

    def generate_agent_report(self, summary: Any) -> str:
        """Generate a comprehensive report for a single agent."""
        return summary.generate_markdown_report()

    def save_agent_report(self, summary: Any) -> Path:
        """Save agent report to file."""
        report = self.generate_agent_report(summary)
        filename = f"AGENT_{summary.agent_id}_{summary.agent_codename}_REPORT.md"
        filepath = self.output_dir / filename
        
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(report)
        
        return filepath

    def generate_collective_matrix(self, all_summaries: List[Any]) -> str:
        """Generate a collective capabilities matrix."""
        report = """# 🎯 Elite Agent Collective - Capabilities Matrix

## Overview

This matrix provides a comprehensive view of all agent capabilities across the collective.

---

## Agent Performance Summary

| Agent | Tier | Specialty | Tests | Pass Rate | Ceiling |
|-------|------|-----------|-------|-----------|---------|
"""
        for summary in all_summaries:
            ceiling = summary.omniscient_package.get('difficulty_ceiling', 'N/A')
            report += f"| {summary.agent_codename}-{summary.agent_id} | {summary.agent_tier} | {summary.agent_specialty[:30]}... | {summary.total_tests} | {summary.pass_rate:.2%} | {ceiling} |\n"

        report += """
---

## Difficulty Level Distribution

| Level | Description | Collective Pass Rate |
|-------|-------------|---------------------|
"""
        # Calculate collective pass rates per level
        level_totals = {"L1": [0, 0], "L2": [0, 0], "L3": [0, 0], "L4": [0, 0], "L5": [0, 0]}
        for summary in all_summaries:
            for level, data in summary.difficulty_breakdown.items():
                level_totals[level][0] += data.get('passed', 0)
                level_totals[level][1] += data.get('total', 0)

        level_descriptions = {
            "L1": "Basic capability verification",
            "L2": "Normal operational scenarios",
            "L3": "Complex multi-step problems",
            "L4": "Edge cases and stress scenarios",
            "L5": "Theoretical limits and impossible problems"
        }

        for level in ["L1", "L2", "L3", "L4", "L5"]:
            passed, total = level_totals[level]
            rate = passed / total if total > 0 else 0
            report += f"| {level} | {level_descriptions[level]} | {rate:.2%} |\n"

        report += """
---

## Tier Analysis

"""
        for tier in range(1, 5):
            tier_agents = [s for s in all_summaries if s.agent_tier == tier]
            if tier_agents:
                tier_names = {1: "Foundational", 2: "Specialists", 3: "Innovators", 4: "Meta"}
                report += f"""
### Tier {tier}: {tier_names.get(tier, 'Unknown')}

| Agent | Pass Rate | Strengths | Gaps |
|-------|-----------|-----------|------|
"""
                for agent in tier_agents:
                    strengths_count = len(agent.strengths)
                    gaps_count = len(agent.omniscient_package.get('capability_gaps', []))
                    report += f"| {agent.agent_codename} | {agent.pass_rate:.2%} | {strengths_count} | {gaps_count} |\n"

        report += f"""
---

## Collaboration Network

Based on test overlaps and capability complementarity:

```
Tier 1 (Foundational)
├── APEX ─────────── ARCHITECT ─────────── VELOCITY
│      \\                  |                   /
│       \\                 |                  /
│        └───────── ECLIPSE ───────────────┘
│                         |
├── CIPHER ───────── FORTRESS ─────────── CRYPTO
│                         |
└── AXIOM ────────── QUANTUM ─────────── PRISM

Tier 2 (Specialists)
├── TENSOR ─────────── NEURAL ─────────── HELIX
├── FLUX ─────────── SYNAPSE ─────────── CORE
└── VANGUARD ─────────── Integration Hub

Tier 3 (Innovators)
└── NEXUS ←─────────→ GENESIS
           \\         /
            \\       /
             \\     /
              ↓   ↓
Tier 4 (Meta)
└── OMNISCIENT (Orchestrator)
```

---

## Critical Gaps Across Collective

"""
        all_gaps = []
        for summary in all_summaries:
            for gap in summary.omniscient_package.get('capability_gaps', []):
                all_gaps.append(f"{summary.agent_codename}: {gap}")

        for gap in all_gaps[:10]:
            report += f"- ⚠️ {gap}\n"

        report += f"""
---

## Evolution Priorities

"""
        all_priorities = []
        for summary in all_summaries:
            for rec in summary.evolution_recommendations:
                all_priorities.append(f"{summary.agent_codename}: {rec}")

        for i, priority in enumerate(all_priorities[:15], 1):
            report += f"{i}. {priority}\n"

        report += f"""
---

_Matrix generated: {datetime.utcnow().isoformat()} UTC_
_Framework: Elite Agent Collective Test Suite v1.0_
"""
        return report

    def save_collective_matrix(self, all_summaries: List[Any]) -> Path:
        """Save collective matrix to file."""
        report = self.generate_collective_matrix(all_summaries)
        filepath = self.output_dir / "COLLECTIVE_CAPABILITIES_MATRIX.md"
        
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(report)
        
        return filepath

    def generate_test_index(self, all_summaries: List[Any]) -> str:
        """Generate an index of all tests."""
        report = """# 📋 Elite Agent Collective - Test Index

## All Tests by Agent

"""
        for summary in all_summaries:
            report += f"""
### {summary.agent_codename}-{summary.agent_id}

| Test Name | Difficulty | Category | Status |
|-----------|------------|----------|--------|
"""
            for result in summary.test_results:
                status = "✅ PASS" if result.passed else "❌ FAIL"
                report += f"| {result.test_name} | {result.difficulty.code} | {result.category.value} | {status} |\n"

        return report

    def export_json(self, all_summaries: List[Any], filename: str = "test_results.json") -> Path:
        """Export all results to JSON format."""
        data = {
            "timestamp": datetime.utcnow().isoformat(),
            "total_agents": len(all_summaries),
            "agents": []
        }

        for summary in all_summaries:
            agent_data = {
                "agent_id": summary.agent_id,
                "agent_codename": summary.agent_codename,
                "agent_tier": summary.agent_tier,
                "total_tests": summary.total_tests,
                "passed_tests": summary.passed_tests,
                "pass_rate": summary.pass_rate,
                "omniscient_package": summary.omniscient_package,
                "test_results": [r.to_dict() for r in summary.test_results]
            }
            data["agents"].append(agent_data)

        filepath = self.output_dir / filename
        with open(filepath, 'w', encoding='utf-8') as f:
            json.dump(data, f, indent=2, default=str)

        return filepath
