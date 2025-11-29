"""
Evolution Tracker
=================
Capability growth monitoring for the Supreme Test Suite.
Tracks performance over time, identifies breakthrough opportunities,
and generates evolution reports for OMNISCIENT-20.
"""

from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple
import json


@dataclass
class CapabilitySnapshot:
    """Point-in-time capability measurement."""
    agent_id: str
    capability: str
    mastery_level: float
    timestamp: str
    context: str = ""


@dataclass
class GrowthTrajectory:
    """Growth trajectory for an agent-capability pair."""
    agent_id: str
    capability: str
    snapshots: List[CapabilitySnapshot]
    growth_rate: float  # Change per observation
    trend: str  # "improving", "stable", "declining"
    projection: float  # Projected future mastery
    breakthrough_potential: float  # Likelihood of breakthrough


@dataclass
class TierPerformanceRecord:
    """Performance record for a tier over time."""
    tier: int
    tier_name: str
    timestamps: List[str]
    pass_rates: List[float]
    agent_counts: List[int]
    average_growth: float


@dataclass
class EvolutionReport:
    """Complete evolution report for OMNISCIENT-20."""
    report_id: str
    generated_at: str
    reporting_period: Tuple[str, str]
    collective_growth: float
    tier_performances: Dict[int, TierPerformanceRecord]
    top_improvers: List[Dict[str, Any]]
    declining_agents: List[Dict[str, Any]]
    breakthrough_opportunities: List[Dict[str, Any]]
    recommendations: List[str]
    markdown_report: str


class EvolutionTracker:
    """
    Tracks capability growth and evolution across all 40 agents.
    Generates insights and reports for OMNISCIENT-20.
    """

    # Tier names
    TIER_NAMES = {
        1: "Foundational",
        2: "Specialists",
        3: "Innovators",
        4: "Meta",
        5: "Domain Specialists",
        6: "Emerging Tech",
        7: "Human-Centric",
        8: "Enterprise",
    }

    def __init__(self, storage_path: Optional[str] = None):
        """
        Initialize the Evolution Tracker.

        Args:
            storage_path: Optional path for persistent storage
        """
        self.storage_path = Path(storage_path) if storage_path else None
        self.snapshots: List[CapabilitySnapshot] = []
        self.tier_history: Dict[int, TierPerformanceRecord] = {}
        self.reports: List[EvolutionReport] = []
        self.report_counter = 0

        if self.storage_path and self.storage_path.exists():
            self._load_state()

    def _load_state(self) -> None:
        """Load state from storage."""
        try:
            state_file = self.storage_path / "evolution_state.json"
            if state_file.exists():
                with open(state_file, "r") as f:
                    state = json.load(f)
                    # Load snapshots
                    for s in state.get("snapshots", []):
                        self.snapshots.append(CapabilitySnapshot(**s))
        except Exception:
            pass  # Start fresh if load fails

    def _save_state(self) -> None:
        """Save state to storage."""
        if not self.storage_path:
            return

        try:
            self.storage_path.mkdir(parents=True, exist_ok=True)
            state_file = self.storage_path / "evolution_state.json"
            state = {
                "snapshots": [
                    {
                        "agent_id": s.agent_id,
                        "capability": s.capability,
                        "mastery_level": s.mastery_level,
                        "timestamp": s.timestamp,
                        "context": s.context,
                    }
                    for s in self.snapshots[-1000:]  # Keep last 1000
                ],
            }
            with open(state_file, "w") as f:
                json.dump(state, f, indent=2)
        except Exception:
            pass  # Continue even if save fails

    def capture_snapshot(
        self,
        test_result,
    ) -> List[CapabilitySnapshot]:
        """
        Capture periodic snapshot from test results.

        Args:
            test_result: CollectiveTestResult from MasterOrchestrator

        Returns:
            List of captured CapabilitySnapshot objects
        """
        timestamp = test_result.timestamp if hasattr(test_result, 'timestamp') else datetime.utcnow().isoformat()
        new_snapshots = []

        # Capture agent-level snapshots
        agent_results = test_result.agent_results if hasattr(test_result, 'agent_results') else {}
        for agent_id, data in agent_results.items():
            # Create snapshot for overall performance
            snapshot = CapabilitySnapshot(
                agent_id=agent_id,
                capability="overall_performance",
                mastery_level=data.get("pass_rate", 0.0),
                timestamp=timestamp,
                context=f"tier_{data.get('tier', 0)}",
            )
            new_snapshots.append(snapshot)
            self.snapshots.append(snapshot)

            # Create snapshots for specific capabilities
            for cap in data.get("capabilities_tested", []):
                cap_snapshot = CapabilitySnapshot(
                    agent_id=agent_id,
                    capability=cap,
                    mastery_level=data.get("pass_rate", 0.0),
                    timestamp=timestamp,
                    context=f"tier_{data.get('tier', 0)}",
                )
                new_snapshots.append(cap_snapshot)
                self.snapshots.append(cap_snapshot)

        # Update tier history
        tier_results = test_result.tier_results if hasattr(test_result, 'tier_results') else {}
        for tier, data in tier_results.items():
            tier_int = int(tier) if isinstance(tier, str) else tier
            if tier_int not in self.tier_history:
                self.tier_history[tier_int] = TierPerformanceRecord(
                    tier=tier_int,
                    tier_name=self.TIER_NAMES.get(tier_int, f"Tier {tier_int}"),
                    timestamps=[],
                    pass_rates=[],
                    agent_counts=[],
                    average_growth=0.0,
                )

            record = self.tier_history[tier_int]
            record.timestamps.append(timestamp)
            record.pass_rates.append(data.get("pass_rate", 0.0))
            record.agent_counts.append(data.get("agents_tested", 0))

            # Calculate average growth
            if len(record.pass_rates) >= 2:
                growths = [
                    record.pass_rates[i] - record.pass_rates[i - 1]
                    for i in range(1, len(record.pass_rates))
                ]
                record.average_growth = sum(growths) / len(growths)

        self._save_state()
        return new_snapshots

    def compute_growth_trajectory(
        self,
        agent_id: str,
        capability: str = "overall_performance",
    ) -> Optional[GrowthTrajectory]:
        """
        Compute growth trajectory for an agent-capability pair.

        Args:
            agent_id: Agent identifier
            capability: Capability to track

        Returns:
            GrowthTrajectory or None if insufficient data
        """
        # Filter relevant snapshots
        relevant = [
            s for s in self.snapshots
            if s.agent_id == agent_id and s.capability == capability
        ]

        if len(relevant) < 2:
            return None

        # Sort by timestamp
        relevant.sort(key=lambda x: x.timestamp)

        # Calculate growth rate
        growths = []
        for i in range(1, len(relevant)):
            growths.append(relevant[i].mastery_level - relevant[i - 1].mastery_level)

        growth_rate = sum(growths) / len(growths) if growths else 0.0

        # Determine trend
        if growth_rate > 0.02:
            trend = "improving"
        elif growth_rate < -0.02:
            trend = "declining"
        else:
            trend = "stable"

        # Project future mastery (simple linear projection)
        current_mastery = relevant[-1].mastery_level
        projection = min(1.0, max(0.0, current_mastery + growth_rate * 5))

        # Calculate breakthrough potential
        # High if improving rapidly and not yet at max
        if trend == "improving" and current_mastery < 0.95:
            breakthrough_potential = min(1.0, growth_rate * 10 + (0.95 - current_mastery))
        else:
            breakthrough_potential = 0.1

        return GrowthTrajectory(
            agent_id=agent_id,
            capability=capability,
            snapshots=relevant,
            growth_rate=growth_rate,
            trend=trend,
            projection=projection,
            breakthrough_potential=breakthrough_potential,
        )

    def identify_breakthrough_opportunities(
        self,
        min_potential: float = 0.5,
    ) -> List[Dict[str, Any]]:
        """
        Identify agents with high breakthrough potential.

        Args:
            min_potential: Minimum breakthrough potential threshold

        Returns:
            List of breakthrough opportunities
        """
        opportunities = []

        # Get unique agent-capability pairs
        pairs = set()
        for s in self.snapshots:
            pairs.add((s.agent_id, s.capability))

        for agent_id, capability in pairs:
            trajectory = self.compute_growth_trajectory(agent_id, capability)
            if trajectory and trajectory.breakthrough_potential >= min_potential:
                opportunities.append({
                    "agent_id": agent_id,
                    "capability": capability,
                    "current_mastery": trajectory.snapshots[-1].mastery_level,
                    "growth_rate": trajectory.growth_rate,
                    "breakthrough_potential": trajectory.breakthrough_potential,
                    "projection": trajectory.projection,
                    "recommendation": self._generate_breakthrough_recommendation(trajectory),
                })

        # Sort by breakthrough potential
        opportunities.sort(key=lambda x: x["breakthrough_potential"], reverse=True)
        return opportunities[:20]

    def _generate_breakthrough_recommendation(
        self,
        trajectory: GrowthTrajectory,
    ) -> str:
        """Generate recommendation for breakthrough opportunity."""
        if trajectory.growth_rate > 0.05:
            return f"Accelerate training - agent is rapidly improving"
        elif trajectory.growth_rate > 0.02:
            return f"Maintain current training intensity"
        elif trajectory.projection > 0.95:
            return f"Close to mastery - focus on edge cases"
        else:
            return f"Apply targeted exercises for {trajectory.capability}"

    def get_tier_performance(self, tier: int) -> Optional[TierPerformanceRecord]:
        """Get performance record for a tier."""
        return self.tier_history.get(tier)

    def get_all_tier_performances(self) -> Dict[int, Dict[str, Any]]:
        """Get all tier performance summaries."""
        summaries = {}
        for tier, record in self.tier_history.items():
            if record.pass_rates:
                summaries[tier] = {
                    "tier_name": record.tier_name,
                    "current_pass_rate": record.pass_rates[-1],
                    "average_pass_rate": sum(record.pass_rates) / len(record.pass_rates),
                    "growth": record.average_growth,
                    "observations": len(record.pass_rates),
                    "trend": "improving" if record.average_growth > 0.01 else (
                        "declining" if record.average_growth < -0.01 else "stable"
                    ),
                }
        return summaries

    def generate_evolution_report(
        self,
        period_start: Optional[str] = None,
        period_end: Optional[str] = None,
    ) -> EvolutionReport:
        """
        Generate comprehensive evolution report.

        Args:
            period_start: Start of reporting period (ISO format)
            period_end: End of reporting period (ISO format)

        Returns:
            EvolutionReport with full analysis
        """
        self.report_counter += 1
        report_id = f"EVOL-{datetime.utcnow().strftime('%Y%m%d')}-{self.report_counter:04d}"

        now = datetime.utcnow().isoformat()
        period_end = period_end or now
        period_start = period_start or (
            self.snapshots[0].timestamp if self.snapshots else now
        )

        # Filter snapshots for period
        period_snapshots = [
            s for s in self.snapshots
            if period_start <= s.timestamp <= period_end
        ]

        # Calculate collective growth
        if period_snapshots:
            first_half = period_snapshots[:len(period_snapshots) // 2]
            second_half = period_snapshots[len(period_snapshots) // 2:]

            first_avg = sum(s.mastery_level for s in first_half) / len(first_half) if first_half else 0
            second_avg = sum(s.mastery_level for s in second_half) / len(second_half) if second_half else 0
            collective_growth = second_avg - first_avg
        else:
            collective_growth = 0.0

        # Get tier performances
        tier_performances = {}
        for tier, record in self.tier_history.items():
            tier_performances[tier] = record

        # Identify top improvers
        agent_growth: Dict[str, List[float]] = {}
        for s in period_snapshots:
            if s.capability == "overall_performance":
                if s.agent_id not in agent_growth:
                    agent_growth[s.agent_id] = []
                agent_growth[s.agent_id].append(s.mastery_level)

        top_improvers = []
        declining_agents = []
        for agent_id, levels in agent_growth.items():
            if len(levels) >= 2:
                growth = levels[-1] - levels[0]
                entry = {
                    "agent_id": agent_id,
                    "initial": levels[0],
                    "current": levels[-1],
                    "growth": growth,
                }
                if growth > 0.02:
                    top_improvers.append(entry)
                elif growth < -0.02:
                    declining_agents.append(entry)

        top_improvers.sort(key=lambda x: x["growth"], reverse=True)
        declining_agents.sort(key=lambda x: x["growth"])

        # Get breakthrough opportunities
        breakthrough_opportunities = self.identify_breakthrough_opportunities()

        # Generate recommendations
        recommendations = self._generate_report_recommendations(
            collective_growth,
            top_improvers,
            declining_agents,
            breakthrough_opportunities,
        )

        # Generate markdown report
        markdown = self._generate_markdown_report(
            report_id=report_id,
            period=(period_start, period_end),
            collective_growth=collective_growth,
            tier_performances=self.get_all_tier_performances(),
            top_improvers=top_improvers[:10],
            declining_agents=declining_agents[:10],
            breakthrough_opportunities=breakthrough_opportunities[:10],
            recommendations=recommendations,
        )

        report = EvolutionReport(
            report_id=report_id,
            generated_at=now,
            reporting_period=(period_start, period_end),
            collective_growth=collective_growth,
            tier_performances=tier_performances,
            top_improvers=top_improvers[:10],
            declining_agents=declining_agents[:10],
            breakthrough_opportunities=breakthrough_opportunities[:10],
            recommendations=recommendations,
            markdown_report=markdown,
        )

        self.reports.append(report)
        return report

    def _generate_report_recommendations(
        self,
        collective_growth: float,
        top_improvers: List[Dict[str, Any]],
        declining_agents: List[Dict[str, Any]],
        breakthroughs: List[Dict[str, Any]],
    ) -> List[str]:
        """Generate report recommendations."""
        recommendations = []

        # Collective health recommendations
        if collective_growth > 0.05:
            recommendations.append("Collective growth is excellent - maintain current protocols")
        elif collective_growth > 0.02:
            recommendations.append("Collective growth is positive - continue monitoring")
        elif collective_growth < -0.02:
            recommendations.append("ALERT: Collective decline detected - immediate intervention needed")
        else:
            recommendations.append("Collective performance is stable - consider optimization")

        # Declining agent recommendations
        if declining_agents:
            worst = declining_agents[0]
            recommendations.append(
                f"Priority intervention for {worst['agent_id']} "
                f"(declined {abs(worst['growth']):.1%})"
            )

        # Breakthrough recommendations
        if breakthroughs:
            best = breakthroughs[0]
            recommendations.append(
                f"Invest in {best['agent_id']} for {best['capability']} "
                f"(breakthrough potential: {best['breakthrough_potential']:.1%})"
            )

        # Tier-level recommendations
        tier_perfs = self.get_all_tier_performances()
        for tier, perf in tier_perfs.items():
            if perf["trend"] == "declining":
                recommendations.append(
                    f"Tier {tier} ({perf['tier_name']}) is declining - investigate root cause"
                )

        return recommendations[:10]

    def _generate_markdown_report(
        self,
        report_id: str,
        period: Tuple[str, str],
        collective_growth: float,
        tier_performances: Dict[int, Dict[str, Any]],
        top_improvers: List[Dict[str, Any]],
        declining_agents: List[Dict[str, Any]],
        breakthrough_opportunities: List[Dict[str, Any]],
        recommendations: List[str],
    ) -> str:
        """Generate markdown evolution report."""
        report = f"""# 🧬 OMNISCIENT-20 Evolution Report

**Report ID:** {report_id}
**Generated:** {datetime.utcnow().strftime('%Y-%m-%d %H:%M:%S UTC')}
**Reporting Period:** {period[0][:10]} to {period[1][:10]}

---

## 📊 Executive Summary

| Metric | Value |
|--------|-------|
| **Collective Growth** | {collective_growth:+.2%} |
| **Tiers Tracked** | {len(tier_performances)} |
| **Top Improvers** | {len(top_improvers)} |
| **Declining Agents** | {len(declining_agents)} |
| **Breakthrough Opportunities** | {len(breakthrough_opportunities)} |

---

## 🏛️ Tier Performance

| Tier | Name | Current | Average | Trend | Growth |
|------|------|---------|---------|-------|--------|
"""
        for tier in sorted(tier_performances.keys()):
            perf = tier_performances[tier]
            trend_emoji = "📈" if perf["trend"] == "improving" else (
                "📉" if perf["trend"] == "declining" else "➡️"
            )
            report += f"| {tier} | {perf['tier_name']} | {perf['current_pass_rate']:.1%} | {perf['average_pass_rate']:.1%} | {trend_emoji} {perf['trend']} | {perf['growth']:+.2%} |\n"

        report += """
---

## 🚀 Top Improvers

"""
        if top_improvers:
            for i, agent in enumerate(top_improvers[:5], 1):
                report += f"{i}. **{agent['agent_id']}**: {agent['initial']:.1%} → {agent['current']:.1%} ({agent['growth']:+.2%})\n"
        else:
            report += "_No significant improvements detected in this period._\n"

        report += """
---

## ⚠️ Declining Agents

"""
        if declining_agents:
            for i, agent in enumerate(declining_agents[:5], 1):
                report += f"{i}. **{agent['agent_id']}**: {agent['initial']:.1%} → {agent['current']:.1%} ({agent['growth']:+.2%})\n"
        else:
            report += "_No significant declines detected in this period._\n"

        report += """
---

## 💡 Breakthrough Opportunities

"""
        if breakthrough_opportunities:
            for i, opp in enumerate(breakthrough_opportunities[:5], 1):
                report += f"""
### {i}. {opp['agent_id']} - {opp['capability']}

- **Current Mastery:** {opp['current_mastery']:.1%}
- **Growth Rate:** {opp['growth_rate']:+.3f}
- **Breakthrough Potential:** {opp['breakthrough_potential']:.1%}
- **Projection:** {opp['projection']:.1%}
- **Recommendation:** {opp['recommendation']}

"""
        else:
            report += "_No high-potential breakthrough opportunities identified._\n"

        report += """
---

## 📋 Recommendations

"""
        for i, rec in enumerate(recommendations, 1):
            priority = "🔴" if "ALERT" in rec or "intervention" in rec else (
                "🟡" if "Priority" in rec or "declining" in rec else "🟢"
            )
            report += f"{i}. {priority} {rec}\n"

        report += f"""
---

## 🔮 Next Steps for OMNISCIENT-20

1. Review declining agents and implement targeted training
2. Capitalize on breakthrough opportunities
3. Strengthen tier-level collaboration protocols
4. Continue monitoring collective health metrics
5. Schedule next evolution assessment

---

_"Evolution is not a destination, but a continuous journey of improvement."_

**Report Complete | Ready for OMNISCIENT-20 Integration**

---

_Generated by Evolution Tracker | Supreme Master Test Suite v2.0_
"""
        return report

    def save_report(self, report: EvolutionReport, output_path: str) -> None:
        """Save evolution report to file."""
        path = Path(output_path)
        path.parent.mkdir(parents=True, exist_ok=True)

        with open(path, "w") as f:
            f.write(report.markdown_report)

    def export_data(self) -> Dict[str, Any]:
        """Export all evolution tracking data."""
        return {
            "export_timestamp": datetime.utcnow().isoformat(),
            "snapshot_count": len(self.snapshots),
            "tier_history": {
                tier: {
                    "tier_name": record.tier_name,
                    "observations": len(record.pass_rates),
                    "current_rate": record.pass_rates[-1] if record.pass_rates else 0,
                    "average_growth": record.average_growth,
                }
                for tier, record in self.tier_history.items()
            },
            "reports_generated": len(self.reports),
        }


if __name__ == "__main__":
    # Demo usage
    from dataclasses import dataclass

    @dataclass
    class MockResult:
        timestamp: str = datetime.utcnow().isoformat()
        agent_results: Dict[str, Any] = None
        tier_results: Dict[int, Dict[str, Any]] = None

        def __post_init__(self):
            if self.agent_results is None:
                self.agent_results = {
                    "APEX-01": {"pass_rate": 0.94, "tier": 1, "capabilities_tested": ["algorithm"]},
                    "CIPHER-02": {"pass_rate": 0.92, "tier": 1, "capabilities_tested": ["security"]},
                    "TENSOR-07": {"pass_rate": 0.91, "tier": 2, "capabilities_tested": ["ml"]},
                }
            if self.tier_results is None:
                self.tier_results = {
                    1: {"pass_rate": 0.93, "agents_tested": 5},
                    2: {"pass_rate": 0.91, "agents_tested": 12},
                }

    tracker = EvolutionTracker()

    print("Evolution Tracker Demo")
    print("=" * 50)

    # Capture some snapshots
    for i in range(5):
        result = MockResult()
        # Simulate improvement
        for agent in result.agent_results.values():
            agent["pass_rate"] = min(1.0, agent["pass_rate"] + 0.01 * i)
        for tier in result.tier_results.values():
            tier["pass_rate"] = min(1.0, tier["pass_rate"] + 0.01 * i)

        snapshots = tracker.capture_snapshot(result)
        print(f"Captured {len(snapshots)} snapshots (iteration {i + 1})")

    # Compute trajectory
    trajectory = tracker.compute_growth_trajectory("APEX-01", "overall_performance")
    if trajectory:
        print(f"\nAPEX-01 Trajectory:")
        print(f"  Growth Rate: {trajectory.growth_rate:+.3f}")
        print(f"  Trend: {trajectory.trend}")
        print(f"  Projection: {trajectory.projection:.1%}")

    # Get breakthrough opportunities
    opportunities = tracker.identify_breakthrough_opportunities(min_potential=0.0)
    print(f"\nBreakthrough Opportunities: {len(opportunities)}")

    # Generate report
    report = tracker.generate_evolution_report()
    print(f"\nGenerated Report: {report.report_id}")
    print(f"Collective Growth: {report.collective_growth:+.2%}")
    print(f"Recommendations: {len(report.recommendations)}")
