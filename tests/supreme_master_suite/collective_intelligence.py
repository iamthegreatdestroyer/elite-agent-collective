"""
Collective Intelligence Module
==============================
Pattern synthesis across all 40 agents for the Supreme Test Suite.
Computes synergy matrices, detects emergent capabilities, and
generates meta-insights about collective behavior.
"""

from dataclasses import dataclass, field
from datetime import datetime
from typing import Any, Dict, List, Optional, Set, Tuple
import math


@dataclass
class EmergentCapability:
    """Capability that emerges from agent combinations."""
    capability_id: str
    name: str
    description: str
    contributing_agents: List[str]
    emergence_strength: float  # 0.0 to 1.0
    discovery_context: str
    prerequisites: List[str] = field(default_factory=list)
    synergy_requirements: Dict[str, float] = field(default_factory=dict)
    first_observed: str = ""
    observation_count: int = 1

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary representation."""
        return {
            "capability_id": self.capability_id,
            "name": self.name,
            "description": self.description,
            "contributing_agents": self.contributing_agents,
            "emergence_strength": self.emergence_strength,
            "discovery_context": self.discovery_context,
            "prerequisites": self.prerequisites,
            "synergy_requirements": self.synergy_requirements,
            "first_observed": self.first_observed,
            "observation_count": self.observation_count,
        }


@dataclass
class SynergyMatrix:
    """Matrix of synergy scores between agents."""
    matrix: Dict[str, Dict[str, float]]  # agent_id -> agent_id -> score
    tier_matrix: Dict[int, Dict[int, float]]  # tier -> tier -> score
    computed_at: str
    data_points: int

    def get_synergy(self, agent1: str, agent2: str) -> float:
        """Get synergy score between two agents."""
        if agent1 in self.matrix and agent2 in self.matrix[agent1]:
            return self.matrix[agent1][agent2]
        elif agent2 in self.matrix and agent1 in self.matrix[agent2]:
            return self.matrix[agent2][agent1]
        return 0.5  # Default neutral synergy

    def get_tier_synergy(self, tier1: int, tier2: int) -> float:
        """Get synergy score between two tiers."""
        if tier1 in self.tier_matrix and tier2 in self.tier_matrix[tier1]:
            return self.tier_matrix[tier1][tier2]
        elif tier2 in self.tier_matrix and tier1 in self.tier_matrix[tier2]:
            return self.tier_matrix[tier2][tier1]
        return 0.5

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary representation."""
        return {
            "agent_matrix": self.matrix,
            "tier_matrix": {str(k): {str(k2): v2 for k2, v2 in v.items()} for k, v in self.tier_matrix.items()},
            "computed_at": self.computed_at,
            "data_points": self.data_points,
        }


@dataclass
class FailureCascadePattern:
    """Pattern of cascading failures between agents."""
    pattern_id: str
    trigger_agent: str
    affected_agents: List[str]
    cascade_depth: int
    severity: float
    mitigation_strategies: List[str]
    prevention_agents: List[str]  # Agents that can prevent this cascade


@dataclass
class TeamPrediction:
    """Optimal team prediction for a problem type."""
    problem_type: str
    recommended_agents: List[str]
    confidence_score: float
    synergy_multiplier: float
    potential_risks: List[str]
    fallback_agents: List[str]


class CollectiveIntelligence:
    """
    Pattern synthesis and meta-insights for all 40 agents.
    Analyzes collective behavior and emergent properties.
    """

    # Tier mappings
    TIER_AGENTS = {
        1: ["APEX-01", "CIPHER-02", "ARCHITECT-03", "AXIOM-04", "VELOCITY-05"],
        2: ["QUANTUM-06", "TENSOR-07", "FORTRESS-08", "NEURAL-09", "CRYPTO-10",
            "FLUX-11", "PRISM-12", "SYNAPSE-13", "CORE-14", "HELIX-15",
            "VANGUARD-16", "ECLIPSE-17"],
        3: ["NEXUS-18", "GENESIS-19"],
        4: ["OMNISCIENT-20"],
        5: ["ATLAS-21", "FORGE-22", "SENTRY-23", "VERTEX-24", "STREAM-25"],
        6: ["PHOTON-26", "LATTICE-27", "MORPH-28", "PHANTOM-29", "ORBIT-30"],
        7: ["CANVAS-31", "LINGUA-32", "SCRIBE-33", "MENTOR-34", "BRIDGE-35"],
        8: ["AEGIS-36", "LEDGER-37", "PULSE-38", "ARBITER-39", "ORACLE-40"],
    }

    # Domain affinities for synergy calculation
    DOMAIN_AFFINITIES = {
        "security": ["CIPHER-02", "FORTRESS-08", "CRYPTO-10", "PHANTOM-29", "AEGIS-36"],
        "performance": ["VELOCITY-05", "CORE-14", "FORGE-22", "STREAM-25"],
        "architecture": ["ARCHITECT-03", "ATLAS-21", "SYNAPSE-13", "LATTICE-27"],
        "ml_ai": ["TENSOR-07", "NEURAL-09", "LINGUA-32", "ORACLE-40", "PRISM-12"],
        "innovation": ["NEXUS-18", "GENESIS-19", "VANGUARD-16"],
        "quality": ["ECLIPSE-17", "MENTOR-34", "SCRIBE-33"],
        "compliance": ["AEGIS-36", "LEDGER-37", "PULSE-38"],
    }

    def __init__(self):
        """Initialize the Collective Intelligence module."""
        self.synergy_matrix: Optional[SynergyMatrix] = None
        self.emergent_capabilities: List[EmergentCapability] = []
        self.failure_patterns: List[FailureCascadePattern] = []
        self.observation_history: List[Dict[str, Any]] = []

    def compute_synergy_matrix(
        self,
        test_results: List[Dict[str, Any]],
    ) -> SynergyMatrix:
        """
        Compute pairwise and tier-level synergy matrices.

        Args:
            test_results: List of test result dictionaries with agent performance

        Returns:
            SynergyMatrix with computed synergies
        """
        agent_matrix: Dict[str, Dict[str, float]] = {}
        tier_matrix: Dict[int, Dict[int, float]] = {}
        pair_observations: Dict[Tuple[str, str], List[float]] = {}

        # Process test results
        for result in test_results:
            agent_results = result.get("agent_results", {})
            agents = list(agent_results.keys())

            # Calculate pairwise synergies
            for i in range(len(agents)):
                for j in range(i + 1, len(agents)):
                    agent1, agent2 = agents[i], agents[j]
                    pair = (min(agent1, agent2), max(agent1, agent2))

                    # Synergy = combined performance vs individual
                    perf1 = agent_results[agent1].get("pass_rate", 0.5)
                    perf2 = agent_results[agent2].get("pass_rate", 0.5)
                    combined = (perf1 + perf2) / 2

                    # Synergy bonus if both perform above average
                    if perf1 > 0.8 and perf2 > 0.8:
                        synergy = combined * 1.1
                    elif perf1 < 0.6 or perf2 < 0.6:
                        synergy = combined * 0.9
                    else:
                        synergy = combined

                    synergy = max(0.0, min(1.0, synergy))

                    if pair not in pair_observations:
                        pair_observations[pair] = []
                    pair_observations[pair].append(synergy)

        # Aggregate observations into matrix
        for (agent1, agent2), scores in pair_observations.items():
            avg_synergy = sum(scores) / len(scores)

            if agent1 not in agent_matrix:
                agent_matrix[agent1] = {}
            agent_matrix[agent1][agent2] = avg_synergy

        # Calculate tier-level synergies
        tier_pair_observations: Dict[Tuple[int, int], List[float]] = {}

        for result in test_results:
            agent_results = result.get("agent_results", {})

            tier_perfs: Dict[int, List[float]] = {}
            for agent_id, data in agent_results.items():
                tier = data.get("tier", 0)
                if tier not in tier_perfs:
                    tier_perfs[tier] = []
                tier_perfs[tier].append(data.get("pass_rate", 0.5))

            tiers = list(tier_perfs.keys())
            for i in range(len(tiers)):
                for j in range(i, len(tiers)):
                    tier1, tier2 = tiers[i], tiers[j]
                    pair = (min(tier1, tier2), max(tier1, tier2))

                    avg1 = sum(tier_perfs[tier1]) / len(tier_perfs[tier1])
                    avg2 = sum(tier_perfs[tier2]) / len(tier_perfs[tier2])
                    tier_synergy = (avg1 + avg2) / 2

                    if pair not in tier_pair_observations:
                        tier_pair_observations[pair] = []
                    tier_pair_observations[pair].append(tier_synergy)

        for (tier1, tier2), scores in tier_pair_observations.items():
            avg_synergy = sum(scores) / len(scores)
            if tier1 not in tier_matrix:
                tier_matrix[tier1] = {}
            tier_matrix[tier1][tier2] = avg_synergy

        self.synergy_matrix = SynergyMatrix(
            matrix=agent_matrix,
            tier_matrix=tier_matrix,
            computed_at=datetime.utcnow().isoformat(),
            data_points=len(test_results),
        )

        return self.synergy_matrix

    def detect_emergent_capabilities(
        self,
        test_results: List[Dict[str, Any]],
    ) -> List[EmergentCapability]:
        """
        Detect emergent capabilities from agent combinations.

        Args:
            test_results: List of test result dictionaries

        Returns:
            List of detected EmergentCapability objects
        """
        capability_observations: Dict[str, Dict[str, Any]] = {}
        capability_counter = 0

        for result in test_results:
            agent_results = result.get("agent_results", {})
            synergies = result.get("cross_tier_synergies", [])
            patterns = result.get("emergent_patterns", [])

            # Check for high-performing combinations
            high_performers = [
                agent_id for agent_id, data in agent_results.items()
                if data.get("pass_rate", 0) > 0.9
            ]

            if len(high_performers) >= 3:
                # Potential emergent capability
                cap_key = "-".join(sorted(high_performers[:3]))

                if cap_key not in capability_observations:
                    capability_counter += 1
                    capability_observations[cap_key] = {
                        "agents": high_performers[:3],
                        "observations": [],
                        "id": f"EMC-{capability_counter:04d}",
                    }

                avg_perf = sum(
                    agent_results[a].get("pass_rate", 0)
                    for a in high_performers[:3]
                ) / 3
                capability_observations[cap_key]["observations"].append(avg_perf)

            # Process explicit emergent patterns
            for pattern in patterns:
                if pattern.get("type") == "capability_cluster":
                    cap_name = pattern.get("capability", "unknown")
                    cap_key = f"cluster_{cap_name}"

                    if cap_key not in capability_observations:
                        capability_counter += 1
                        capability_observations[cap_key] = {
                            "agents": [],
                            "observations": [],
                            "id": f"EMC-{capability_counter:04d}",
                            "capability": cap_name,
                        }

                    capability_observations[cap_key]["observations"].append(
                        pattern.get("average_performance", 0.5)
                    )

        # Convert observations to EmergentCapability objects
        emergent_capabilities = []
        for cap_key, data in capability_observations.items():
            if len(data["observations"]) >= 2:
                avg_strength = sum(data["observations"]) / len(data["observations"])

                if avg_strength > 0.85:
                    capability = EmergentCapability(
                        capability_id=data["id"],
                        name=self._generate_capability_name(cap_key, data),
                        description=f"Emergent capability from {len(data['agents'])} agents",
                        contributing_agents=data.get("agents", []),
                        emergence_strength=avg_strength,
                        discovery_context=cap_key,
                        first_observed=datetime.utcnow().isoformat(),
                        observation_count=len(data["observations"]),
                    )
                    emergent_capabilities.append(capability)

        self.emergent_capabilities.extend(emergent_capabilities)
        return emergent_capabilities

    def _generate_capability_name(
        self,
        cap_key: str,
        data: Dict[str, Any],
    ) -> str:
        """Generate human-readable capability name."""
        if "capability" in data:
            return f"Collective {data['capability'].replace('_', ' ').title()}"

        agents = data.get("agents", [])
        if len(agents) >= 2:
            return f"{agents[0]}-{agents[1]} Synergy"

        return f"Emergent Capability {cap_key[:10]}"

    def analyze_failure_cascades(
        self,
        test_results: List[Dict[str, Any]],
    ) -> List[FailureCascadePattern]:
        """
        Analyze failure cascade patterns.

        Args:
            test_results: List of test result dictionaries

        Returns:
            List of FailureCascadePattern objects
        """
        cascade_observations: Dict[str, Dict[str, Any]] = {}
        pattern_counter = 0

        for result in test_results:
            agent_results = result.get("agent_results", {})

            # Find agents with failures
            failed_agents = [
                (agent_id, data)
                for agent_id, data in agent_results.items()
                if data.get("pass_rate", 1.0) < 0.7
            ]

            if len(failed_agents) >= 2:
                # Sort by performance (worst first)
                failed_agents.sort(key=lambda x: x[1].get("pass_rate", 0))

                trigger = failed_agents[0][0]
                affected = [a[0] for a in failed_agents[1:]]

                pattern_key = f"{trigger}->{'_'.join(sorted(affected[:3]))}"

                if pattern_key not in cascade_observations:
                    pattern_counter += 1
                    cascade_observations[pattern_key] = {
                        "trigger": trigger,
                        "affected": affected,
                        "id": f"FCP-{pattern_counter:04d}",
                        "severities": [],
                    }

                # Calculate severity based on how low the performance dropped
                avg_failure_rate = 1.0 - sum(
                    a[1].get("pass_rate", 0) for a in failed_agents
                ) / len(failed_agents)
                cascade_observations[pattern_key]["severities"].append(avg_failure_rate)

        # Convert to FailureCascadePattern objects
        failure_patterns = []
        for pattern_key, data in cascade_observations.items():
            if len(data["severities"]) >= 1:
                avg_severity = sum(data["severities"]) / len(data["severities"])

                pattern = FailureCascadePattern(
                    pattern_id=data["id"],
                    trigger_agent=data["trigger"],
                    affected_agents=data["affected"][:5],
                    cascade_depth=len(data["affected"]),
                    severity=avg_severity,
                    mitigation_strategies=self._generate_mitigation_strategies(data["trigger"]),
                    prevention_agents=self._identify_prevention_agents(data["trigger"]),
                )
                failure_patterns.append(pattern)

        self.failure_patterns.extend(failure_patterns)
        return failure_patterns

    def _generate_mitigation_strategies(self, trigger_agent: str) -> List[str]:
        """Generate mitigation strategies for a failure cascade."""
        strategies = [
            f"Enhance isolation between {trigger_agent} and dependent agents",
            "Implement circuit breaker pattern",
            "Add redundancy in critical paths",
        ]

        # Add agent-specific strategies
        if "CIPHER" in trigger_agent or "FORTRESS" in trigger_agent:
            strategies.append("Deploy security fallback protocols")
        if "VELOCITY" in trigger_agent or "CORE" in trigger_agent:
            strategies.append("Implement graceful performance degradation")
        if "ARCHITECT" in trigger_agent or "SYNAPSE" in trigger_agent:
            strategies.append("Enable modular architecture failover")

        return strategies

    def _identify_prevention_agents(self, trigger_agent: str) -> List[str]:
        """Identify agents that can help prevent cascade from trigger."""
        prevention_agents = []

        # ECLIPSE can help with testing/verification to prevent failures
        prevention_agents.append("ECLIPSE-17")

        # OMNISCIENT can coordinate prevention
        prevention_agents.append("OMNISCIENT-20")

        # Add domain-specific prevention agents
        for domain, agents in self.DOMAIN_AFFINITIES.items():
            if trigger_agent in agents:
                for agent in agents:
                    if agent != trigger_agent and agent not in prevention_agents:
                        prevention_agents.append(agent)
                        if len(prevention_agents) >= 5:
                            break
                break

        return prevention_agents[:5]

    def predict_optimal_team(
        self,
        problem_type: str,
        team_size: int = 5,
        excluded_agents: Optional[List[str]] = None,
    ) -> TeamPrediction:
        """
        Predict optimal team for a problem type.

        Args:
            problem_type: Type of problem to solve
            team_size: Number of agents to recommend
            excluded_agents: Agents that should not be included

        Returns:
            TeamPrediction with recommended team
        """
        excluded = set(excluded_agents or [])

        # Get domain-relevant agents
        domain_agents = self.DOMAIN_AFFINITIES.get(problem_type, [])
        relevant_agents = [a for a in domain_agents if a not in excluded]

        # Add complementary agents based on synergy
        all_agents = []
        for tier_agents in self.TIER_AGENTS.values():
            all_agents.extend(tier_agents)

        remaining = [a for a in all_agents if a not in relevant_agents and a not in excluded]

        # Sort by synergy with relevant agents (if matrix available)
        if self.synergy_matrix and relevant_agents:
            def synergy_score(agent: str) -> float:
                scores = []
                for relevant in relevant_agents:
                    scores.append(self.synergy_matrix.get_synergy(agent, relevant))
                return sum(scores) / len(scores) if scores else 0.5

            remaining.sort(key=synergy_score, reverse=True)

        # Build team
        recommended = relevant_agents[:team_size]
        if len(recommended) < team_size:
            recommended.extend(remaining[:team_size - len(recommended)])

        # Calculate confidence and synergy multiplier
        confidence = 0.7  # Base confidence
        if problem_type in self.DOMAIN_AFFINITIES:
            confidence += 0.2

        synergy_multiplier = 1.0
        if self.synergy_matrix and len(recommended) >= 2:
            synergies = []
            for i in range(len(recommended)):
                for j in range(i + 1, len(recommended)):
                    synergies.append(
                        self.synergy_matrix.get_synergy(recommended[i], recommended[j])
                    )
            if synergies:
                avg_synergy = sum(synergies) / len(synergies)
                synergy_multiplier = 0.8 + avg_synergy * 0.4

        # Identify risks
        risks = []
        if len(relevant_agents) < 2:
            risks.append("Limited domain expertise in team")
        if any(a in excluded for a in domain_agents):
            risks.append("Key domain agents excluded")

        # Fallback agents
        fallback = remaining[team_size:team_size + 3] if len(remaining) > team_size else []

        return TeamPrediction(
            problem_type=problem_type,
            recommended_agents=recommended,
            confidence_score=min(1.0, confidence),
            synergy_multiplier=synergy_multiplier,
            potential_risks=risks,
            fallback_agents=fallback,
        )

    def generate_meta_insights(
        self,
        test_results: List[Dict[str, Any]],
    ) -> Dict[str, Any]:
        """
        Generate meta-insights about collective behavior.

        Args:
            test_results: List of test result dictionaries

        Returns:
            Dictionary of meta-insights
        """
        insights = {
            "timestamp": datetime.utcnow().isoformat(),
            "data_points": len(test_results),
        }

        # Tier performance analysis
        tier_performance: Dict[int, List[float]] = {}
        for result in test_results:
            for agent_id, data in result.get("agent_results", {}).items():
                tier = data.get("tier", 0)
                if tier not in tier_performance:
                    tier_performance[tier] = []
                tier_performance[tier].append(data.get("pass_rate", 0.5))

        insights["tier_performance"] = {
            tier: {
                "average": sum(rates) / len(rates) if rates else 0,
                "min": min(rates) if rates else 0,
                "max": max(rates) if rates else 0,
                "variance": self._calculate_variance(rates),
            }
            for tier, rates in tier_performance.items()
        }

        # Collective health metrics
        all_rates = []
        for rates in tier_performance.values():
            all_rates.extend(rates)

        if all_rates:
            insights["collective_health"] = {
                "average_performance": sum(all_rates) / len(all_rates),
                "uniformity": 1.0 - self._calculate_variance(all_rates),
                "agents_above_90": sum(1 for r in all_rates if r > 0.9) / len(all_rates),
                "agents_below_70": sum(1 for r in all_rates if r < 0.7) / len(all_rates),
            }
        else:
            insights["collective_health"] = {
                "average_performance": 0,
                "uniformity": 0,
                "agents_above_90": 0,
                "agents_below_70": 0,
            }

        # Emergent capability summary
        insights["emergent_capabilities"] = {
            "total_detected": len(self.emergent_capabilities),
            "strongest": [
                {"name": c.name, "strength": c.emergence_strength}
                for c in sorted(
                    self.emergent_capabilities,
                    key=lambda x: x.emergence_strength,
                    reverse=True,
                )[:5]
            ],
        }

        # Failure pattern summary
        insights["failure_patterns"] = {
            "total_detected": len(self.failure_patterns),
            "most_severe": [
                {
                    "trigger": p.trigger_agent,
                    "severity": p.severity,
                    "cascade_depth": p.cascade_depth,
                }
                for p in sorted(
                    self.failure_patterns,
                    key=lambda x: x.severity,
                    reverse=True,
                )[:3]
            ],
        }

        # Evolution recommendations
        insights["evolution_recommendations"] = self._generate_evolution_recommendations(
            tier_performance
        )

        return insights

    def _calculate_variance(self, values: List[float]) -> float:
        """Calculate variance of a list of values."""
        if len(values) < 2:
            return 0.0
        mean = sum(values) / len(values)
        return sum((x - mean) ** 2 for x in values) / len(values)

    def _generate_evolution_recommendations(
        self,
        tier_performance: Dict[int, List[float]],
    ) -> List[Dict[str, Any]]:
        """Generate evolution recommendations based on performance."""
        recommendations = []

        for tier, rates in tier_performance.items():
            if not rates:
                continue

            avg = sum(rates) / len(rates)

            if avg < 0.8:
                recommendations.append({
                    "type": "tier_improvement",
                    "tier": tier,
                    "current_performance": avg,
                    "target": 0.9,
                    "priority": "high" if avg < 0.7 else "medium",
                    "action": f"Intensive training for Tier {tier} agents",
                })

            variance = self._calculate_variance(rates)
            if variance > 0.05:
                recommendations.append({
                    "type": "tier_uniformity",
                    "tier": tier,
                    "variance": variance,
                    "priority": "medium",
                    "action": f"Reduce performance variance in Tier {tier}",
                })

        # Sort by priority
        priority_order = {"high": 0, "medium": 1, "low": 2}
        recommendations.sort(key=lambda x: priority_order.get(x.get("priority", "low"), 2))

        return recommendations[:10]

    def export_intelligence(self) -> Dict[str, Any]:
        """Export all collective intelligence data."""
        return {
            "export_timestamp": datetime.utcnow().isoformat(),
            "synergy_matrix": self.synergy_matrix.to_dict() if self.synergy_matrix else None,
            "emergent_capabilities": [c.to_dict() for c in self.emergent_capabilities],
            "failure_patterns": [
                {
                    "pattern_id": p.pattern_id,
                    "trigger_agent": p.trigger_agent,
                    "affected_agents": p.affected_agents,
                    "cascade_depth": p.cascade_depth,
                    "severity": p.severity,
                    "mitigation_strategies": p.mitigation_strategies,
                }
                for p in self.failure_patterns
            ],
            "observation_count": len(self.observation_history),
        }


if __name__ == "__main__":
    # Demo usage
    ci = CollectiveIntelligence()

    # Create mock test results
    mock_results = [
        {
            "agent_results": {
                "APEX-01": {"pass_rate": 0.95, "tier": 1},
                "CIPHER-02": {"pass_rate": 0.93, "tier": 1},
                "ARCHITECT-03": {"pass_rate": 0.94, "tier": 1},
                "TENSOR-07": {"pass_rate": 0.92, "tier": 2},
                "NEXUS-18": {"pass_rate": 0.91, "tier": 3},
            },
            "cross_tier_synergies": [],
            "emergent_patterns": [
                {"type": "capability_cluster", "capability": "system_design", "average_performance": 0.93},
            ],
        },
        {
            "agent_results": {
                "APEX-01": {"pass_rate": 0.94, "tier": 1},
                "VELOCITY-05": {"pass_rate": 0.96, "tier": 1},
                "CORE-14": {"pass_rate": 0.91, "tier": 2},
                "FORGE-22": {"pass_rate": 0.88, "tier": 5},
            },
            "cross_tier_synergies": [],
            "emergent_patterns": [],
        },
    ]

    print("Collective Intelligence Demo")
    print("=" * 50)

    # Compute synergy matrix
    matrix = ci.compute_synergy_matrix(mock_results)
    print(f"\nSynergy Matrix computed with {matrix.data_points} data points")

    # Detect emergent capabilities
    emergent = ci.detect_emergent_capabilities(mock_results)
    print(f"Detected {len(emergent)} emergent capabilities")

    # Analyze failures
    failures = ci.analyze_failure_cascades(mock_results)
    print(f"Detected {len(failures)} failure cascade patterns")

    # Predict optimal team
    team = ci.predict_optimal_team("security", team_size=4)
    print(f"\nOptimal team for security:")
    print(f"  Agents: {team.recommended_agents}")
    print(f"  Confidence: {team.confidence_score:.2f}")
    print(f"  Synergy Multiplier: {team.synergy_multiplier:.2f}")

    # Generate meta-insights
    insights = ci.generate_meta_insights(mock_results)
    print(f"\nCollective Health: {insights['collective_health']['average_performance']:.2%}")
