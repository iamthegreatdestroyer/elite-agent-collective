"""
OMNISCIENT-20 Aggregation System
================================
Synthesizes test results from all 19 agents into a cohesive
intelligence package for the Meta-Learning Trainer.
"""

import json
from dataclasses import dataclass, field
from datetime import datetime
from typing import Dict, List, Any, Optional
from pathlib import Path


@dataclass
class CollectiveIntelligence:
    """Aggregated intelligence from all agent tests."""

    timestamp: str
    total_agents: int
    total_tests: int
    collective_pass_rate: float

    # Tier-level analysis
    tier_performance: Dict[str, Dict[str, Any]]

    # Cross-agent patterns
    capability_matrix: Dict[str, Dict[str, float]]
    collaboration_graph: Dict[str, List[str]]
    evolution_priorities: List[Dict[str, Any]]

    # Gap analysis
    collective_strengths: List[str]
    collective_weaknesses: List[str]
    critical_gaps: List[str]

    # Meta-learning signals
    learning_vectors: List[Dict[str, Any]]
    optimization_targets: List[str]

    # Individual agent packages
    agent_summaries: Dict[str, Dict[str, Any]]


class OmniscientAggregator:
    """
    Aggregates test results from all agents and prepares
    the intelligence package for OMNISCIENT-20.
    """

    def __init__(self, reports_dir: str = "docs"):
        self.reports_dir = Path(reports_dir)
        self.agent_data: Dict[str, Any] = {}

    def load_agent_reports(self) -> None:
        """Load all agent test reports from the agent_data dict."""
        # Data is loaded programmatically, not from files
        pass

    def add_agent_data(self, agent_id: str, data: Dict[str, Any]) -> None:
        """Add agent data to the aggregator."""
        self.agent_data[agent_id] = data

    def synthesize_collective_intelligence(self) -> CollectiveIntelligence:
        """
        Synthesize all agent data into collective intelligence.
        """
        # Calculate collective metrics
        total_tests = sum(
            agent.get('total_tests', 0)
            for agent in self.agent_data.values()
        )
        total_passed = sum(
            agent.get('passed_tests', 0)
            for agent in self.agent_data.values()
        )

        # Analyze tier performance
        tier_performance = self._analyze_tier_performance()

        # Build capability matrix
        capability_matrix = self._build_capability_matrix()

        # Generate collaboration graph
        collaboration_graph = self._generate_collaboration_graph()

        # Identify evolution priorities
        evolution_priorities = self._identify_evolution_priorities()

        # Perform gap analysis
        strengths, weaknesses, gaps = self._perform_gap_analysis()

        # Generate learning vectors
        learning_vectors = self._generate_learning_vectors()

        return CollectiveIntelligence(
            timestamp=datetime.utcnow().isoformat(),
            total_agents=len(self.agent_data),
            total_tests=total_tests,
            collective_pass_rate=total_passed / total_tests if total_tests > 0 else 0,
            tier_performance=tier_performance,
            capability_matrix=capability_matrix,
            collaboration_graph=collaboration_graph,
            evolution_priorities=evolution_priorities,
            collective_strengths=strengths,
            collective_weaknesses=weaknesses,
            critical_gaps=gaps,
            learning_vectors=learning_vectors,
            optimization_targets=self._identify_optimization_targets(),
            agent_summaries={
                agent_id: self._summarize_agent(data)
                for agent_id, data in self.agent_data.items()
            }
        )

    def _analyze_tier_performance(self) -> Dict[str, Dict[str, Any]]:
        """Analyze performance by tier."""
        tiers = {
            "TIER_1_FOUNDATIONAL": ["01", "02", "03", "04", "05"],
            "TIER_2_SPECIALISTS": [f"{i:02d}" for i in range(6, 18)],
            "TIER_3_INNOVATORS": ["18", "19"],
            "TIER_4_META": ["20"]
        }

        tier_performance = {}
        for tier_name, agent_ids in tiers.items():
            tier_agents = [
                self.agent_data.get(aid, {})
                for aid in agent_ids
                if aid in self.agent_data
            ]

            if not tier_agents:
                continue

            total = sum(a.get('total_tests', 0) for a in tier_agents)
            passed = sum(a.get('passed_tests', 0) for a in tier_agents)

            tier_performance[tier_name] = {
                "agents": len(tier_agents),
                "total_tests": total,
                "passed": passed,
                "pass_rate": passed / total if total > 0 else 0,
                "difficulty_distribution": self._get_tier_difficulty_distribution(tier_agents)
            }

        return tier_performance

    def _build_capability_matrix(self) -> Dict[str, Dict[str, float]]:
        """Build capability matrix across all agents."""
        capabilities = [
            "algorithm_implementation",
            "security_analysis",
            "system_design",
            "mathematical_reasoning",
            "performance_optimization",
            "innovation",
            "collaboration",
            "adaptation"
        ]

        matrix = {}
        for agent_id, data in self.agent_data.items():
            matrix[agent_id] = {}
            for cap in capabilities:
                matrix[agent_id][cap] = self._calculate_capability_score(data, cap)

        return matrix

    def _generate_collaboration_graph(self) -> Dict[str, List[str]]:
        """Generate agent collaboration recommendations."""
        graph = {
            "01_APEX": ["03_ARCHITECT", "05_VELOCITY", "17_ECLIPSE"],
            "02_CIPHER": ["08_FORTRESS", "10_CRYPTO", "04_AXIOM"],
            "03_ARCHITECT": ["01_APEX", "11_FLUX", "13_SYNAPSE"],
            "04_AXIOM": ["17_ECLIPSE", "06_QUANTUM", "02_CIPHER"],
            "05_VELOCITY": ["01_APEX", "14_CORE", "07_TENSOR"],
            "06_QUANTUM": ["04_AXIOM", "07_TENSOR", "02_CIPHER"],
            "07_TENSOR": ["05_VELOCITY", "12_PRISM", "09_NEURAL"],
            "08_FORTRESS": ["02_CIPHER", "11_FLUX", "17_ECLIPSE"],
            "09_NEURAL": ["07_TENSOR", "19_GENESIS", "20_OMNISCIENT"],
            "10_CRYPTO": ["02_CIPHER", "13_SYNAPSE", "03_ARCHITECT"],
            "11_FLUX": ["03_ARCHITECT", "08_FORTRESS", "17_ECLIPSE"],
            "12_PRISM": ["07_TENSOR", "16_VANGUARD", "04_AXIOM"],
            "13_SYNAPSE": ["03_ARCHITECT", "11_FLUX", "10_CRYPTO"],
            "14_CORE": ["05_VELOCITY", "01_APEX", "06_QUANTUM"],
            "15_HELIX": ["07_TENSOR", "12_PRISM", "16_VANGUARD"],
            "16_VANGUARD": ["12_PRISM", "18_NEXUS", "09_NEURAL"],
            "17_ECLIPSE": ["01_APEX", "04_AXIOM", "08_FORTRESS"],
            "18_NEXUS": ["19_GENESIS", "16_VANGUARD", "20_OMNISCIENT"],
            "19_GENESIS": ["18_NEXUS", "04_AXIOM", "09_NEURAL"],
            "20_OMNISCIENT": ["ALL_AGENTS"]
        }

        return graph

    def _identify_evolution_priorities(self) -> List[Dict[str, Any]]:
        """Identify highest priority evolution targets."""
        priorities = []

        for agent_id, data in self.agent_data.items():
            package = data.get('omniscient_package', {})
            gaps = package.get('capability_gaps', [])
            ceiling = package.get('difficulty_ceiling', 'L5')

            if gaps or ceiling in ['L1', 'L2', 'L3']:
                priority_score = self._calculate_priority_score(ceiling, len(gaps))
                priorities.append({
                    "agent_id": agent_id,
                    "agent_codename": data.get('agent_codename', 'Unknown'),
                    "gaps": gaps,
                    "ceiling": ceiling,
                    "priority_score": priority_score,
                    "recommended_actions": self._generate_actions(gaps, ceiling)
                })

        return sorted(priorities, key=lambda x: x['priority_score'], reverse=True)

    def _perform_gap_analysis(self) -> tuple:
        """Perform collective gap analysis."""
        all_strengths = []
        all_weaknesses = []
        critical_gaps = []

        for agent_id, data in self.agent_data.items():
            all_strengths.extend(data.get('strengths', []))
            all_weaknesses.extend(data.get('weaknesses', []))

            # Identify critical gaps (L4/L5 failures in core competency)
            for failure in data.get('critical_failures', []):
                if isinstance(failure, dict):
                    if failure.get('category') == 'core_competency':
                        critical_gaps.append(f"{agent_id}: {failure.get('test_name')}")

        return (
            list(set(all_strengths))[:20],
            list(set(all_weaknesses))[:20],
            critical_gaps[:20]
        )

    def _generate_learning_vectors(self) -> List[Dict[str, Any]]:
        """Generate learning vectors for OMNISCIENT-20."""
        vectors = []

        # Pattern: Agents with similar weaknesses
        weakness_groups = self._group_by_weakness()
        for weakness, agents in weakness_groups.items():
            if len(agents) > 1:
                vectors.append({
                    "type": "shared_weakness",
                    "description": weakness,
                    "affected_agents": agents,
                    "recommended_approach": "collective_training"
                })

        # Pattern: High-performing agents that can mentor others
        for agent_id, data in self.agent_data.items():
            if data.get('pass_rate', 0) >= 0.9:
                vectors.append({
                    "type": "mentor_candidate",
                    "description": f"{agent_id} demonstrates high capability",
                    "affected_agents": [agent_id],
                    "recommended_approach": "knowledge_transfer"
                })

        return vectors

    def _identify_optimization_targets(self) -> List[str]:
        """Identify optimization targets for the collective."""
        targets = [
            "Improve L4/L5 performance across all agents",
            "Enhance cross-agent collaboration protocols",
            "Reduce execution time variance",
            "Strengthen edge case handling",
            "Develop unified error recovery patterns",
            "Implement adaptive difficulty scaling",
            "Create knowledge sharing pipelines",
            "Establish performance benchmarking standards"
        ]
        return targets

    def _calculate_capability_score(self, data: Dict, capability: str) -> float:
        """Calculate capability score from test data."""
        base_score = data.get('pass_rate', 0.5)
        
        # Adjust based on capability type and agent specialty
        specialty = data.get('agent_specialty', '').lower()
        
        capability_weights = {
            "algorithm_implementation": 1.2 if 'engineer' in specialty else 1.0,
            "security_analysis": 1.2 if 'security' in specialty or 'crypto' in specialty else 1.0,
            "system_design": 1.2 if 'architect' in specialty else 1.0,
            "mathematical_reasoning": 1.2 if 'math' in specialty or 'axiom' in specialty else 1.0,
            "performance_optimization": 1.2 if 'performance' in specialty or 'velocity' in specialty else 1.0,
            "innovation": 1.2 if 'innovation' in specialty or 'genesis' in specialty else 1.0,
            "collaboration": 1.0,
            "adaptation": 1.0
        }
        
        weight = capability_weights.get(capability, 1.0)
        return min(1.0, base_score * weight)

    def _get_tier_difficulty_distribution(self, agents: List[Dict]) -> Dict:
        """Get difficulty distribution for a tier."""
        distribution = {"L1": [], "L2": [], "L3": [], "L4": [], "L5": []}
        
        for agent in agents:
            breakdown = agent.get('difficulty_breakdown', {})
            for level in distribution.keys():
                if level in breakdown:
                    distribution[level].append(breakdown[level].get('pass_rate', 0))
        
        return {
            level: sum(rates) / len(rates) if rates else 0
            for level, rates in distribution.items()
        }

    def _calculate_priority_score(self, ceiling: str, gap_count: int) -> float:
        """Calculate evolution priority score."""
        ceiling_weights = {"L1": 1.0, "L2": 0.8, "L3": 0.6, "L4": 0.4, "L5": 0.2}
        return ceiling_weights.get(ceiling, 0.5) * (1 + gap_count * 0.1)

    def _generate_actions(self, gaps: List[str], ceiling: str) -> List[str]:
        """Generate recommended actions for gaps."""
        actions = []
        
        if ceiling in ['L1', 'L2']:
            actions.append("Intensive foundational training required")
        elif ceiling == 'L3':
            actions.append("Focus on advanced problem-solving techniques")
        
        for gap in gaps[:3]:
            actions.append(f"Targeted training: {gap}")
        
        return actions

    def _group_by_weakness(self) -> Dict[str, List[str]]:
        """Group agents by shared weaknesses."""
        weakness_map: Dict[str, List[str]] = {}
        
        for agent_id, data in self.agent_data.items():
            for weakness in data.get('weaknesses', []):
                if weakness not in weakness_map:
                    weakness_map[weakness] = []
                weakness_map[weakness].append(agent_id)
        
        return weakness_map

    def _summarize_agent(self, data: Dict) -> Dict:
        """Create agent summary for final report."""
        return {
            "pass_rate": data.get('pass_rate', 0),
            "total_tests": data.get('total_tests', 0),
            "passed_tests": data.get('passed_tests', 0),
            "difficulty_ceiling": data.get('omniscient_package', {}).get('difficulty_ceiling', 'N/A'),
            "strengths_count": len(data.get('strengths', [])),
            "gaps_count": len(data.get('omniscient_package', {}).get('capability_gaps', []))
        }

    def generate_omniscient_synthesis(self) -> str:
        """Generate the final OMNISCIENT synthesis document."""
        intel = self.synthesize_collective_intelligence()

        synthesis = f"""# 🧠 OMNISCIENT-20 SYNTHESIS REPORT
## Collective Intelligence Package for Meta-Learning & Evolution

---

```
╔══════════════════════════════════════════════════════════════════════════════╗
║                    ELITE AGENT COLLECTIVE - SYNTHESIS REPORT                  ║
║                                                                              ║
║   Generated: {intel.timestamp}                                    ║
║   Total Agents Analyzed: {intel.total_agents}                                             ║
║   Total Tests Executed: {intel.total_tests}                                            ║
║   Collective Pass Rate: {intel.collective_pass_rate:.2%}                                          ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

---

## 📊 EXECUTIVE SUMMARY

The Elite Agent Collective has completed comprehensive capability assessment across all {intel.total_agents} agents. This synthesis document provides OMNISCIENT-20 with the intelligence required to orchestrate collective evolution and optimize inter-agent collaboration.

### Key Findings

- **Total Tests Executed:** {intel.total_tests}
- **Collective Pass Rate:** {intel.collective_pass_rate:.2%}
- **Critical Gaps Identified:** {len(intel.critical_gaps)}
- **Evolution Priorities:** {len(intel.evolution_priorities)}

---

## 🏛️ TIER PERFORMANCE ANALYSIS

| Tier | Agents | Tests | Pass Rate | Avg L5 Performance |
|------|--------|-------|-----------|-------------------|
"""

        for tier_name, data in intel.tier_performance.items():
            l5_perf = data.get('difficulty_distribution', {}).get('L5', 0)
            synthesis += f"| {tier_name} | {data['agents']} | {data['total_tests']} | {data['pass_rate']:.2%} | {l5_perf:.2%} |\n"

        synthesis += f"""
---

## 💪 COLLECTIVE STRENGTHS

"""
        for i, strength in enumerate(intel.collective_strengths[:10], 1):
            synthesis += f"{i}. ✅ {strength}\n"

        synthesis += f"""
---

## ⚠️ COLLECTIVE WEAKNESSES

"""
        for i, weakness in enumerate(intel.collective_weaknesses[:10], 1):
            synthesis += f"{i}. ⚠️ {weakness}\n"

        synthesis += f"""
---

## 🚨 CRITICAL GAPS REQUIRING IMMEDIATE ATTENTION

"""
        for i, gap in enumerate(intel.critical_gaps[:10], 1):
            synthesis += f"{i}. **{gap}**\n"

        synthesis += f"""
---

## 🔄 EVOLUTION PRIORITIES

"""
        for i, priority in enumerate(intel.evolution_priorities[:10], 1):
            synthesis += f"""
### Priority #{i}: Agent {priority['agent_id']} ({priority.get('agent_codename', 'Unknown')})

- **Difficulty Ceiling:** {priority['ceiling']}
- **Priority Score:** {priority['priority_score']:.2f}
- **Capability Gaps:** {', '.join(priority['gaps'][:3]) if priority['gaps'] else 'None identified'}
- **Recommended Actions:**
"""
            for action in priority['recommended_actions'][:3]:
                synthesis += f"  - {action}\n"

        synthesis += f"""
---

## 🔗 COLLABORATION NETWORK

```mermaid
graph TD
"""
        for agent, collaborators in list(intel.collaboration_graph.items())[:10]:
            for collab in collaborators[:2]:
                if collab != "ALL_AGENTS":
                    synthesis += f"    {agent.replace('_', '')} --> {collab.replace('_', '')}\n"

        synthesis += f"""```

---

## 📈 LEARNING VECTORS FOR META-LEARNING

"""
        for i, vector in enumerate(intel.learning_vectors[:5], 1):
            synthesis += f"""
### Vector {i}: {vector['type'].replace('_', ' ').title()}

- **Description:** {vector['description']}
- **Affected Agents:** {', '.join(vector.get('affected_agents', [])[:5])}
- **Approach:** {vector['recommended_approach']}
"""

        synthesis += f"""
---

## 🎯 OPTIMIZATION TARGETS

"""
        for i, target in enumerate(intel.optimization_targets, 1):
            synthesis += f"{i}. {target}\n"

        synthesis += f"""
---

## 📦 INDIVIDUAL AGENT SUMMARIES

| Agent | Pass Rate | Ceiling | Tests | Passed | Gaps |
|-------|-----------|---------|-------|--------|------|
"""
        for agent_id, summary in intel.agent_summaries.items():
            synthesis += f"| {agent_id} | {summary['pass_rate']:.2%} | {summary['difficulty_ceiling']} | {summary['total_tests']} | {summary['passed_tests']} | {summary['gaps_count']} |\n"

        synthesis += f"""
---

## 🧬 RAW INTELLIGENCE PACKAGE

```json
{json.dumps({
    "timestamp": intel.timestamp,
    "collective_metrics": {
        "total_agents": intel.total_agents,
        "total_tests": intel.total_tests,
        "pass_rate": intel.collective_pass_rate
    },
    "tier_performance": {
        tier: {
            "agents": data["agents"],
            "pass_rate": data["pass_rate"]
        }
        for tier, data in intel.tier_performance.items()
    },
    "evolution_priorities": [
        {
            "agent": p["agent_id"],
            "priority": p["priority_score"],
            "ceiling": p["ceiling"]
        }
        for p in intel.evolution_priorities[:5]
    ],
    "optimization_targets": intel.optimization_targets[:5]
}, indent=2)}
```

---

## 📜 DIRECTIVES FOR OMNISCIENT-20

Based on this synthesis, OMNISCIENT-20 should:

1. **Prioritize Evolution Training** for agents with difficulty ceilings below L3
2. **Strengthen Collaboration Protocols** between identified agent pairs
3. **Deploy Targeted Learning** for critical gaps in core competencies
4. **Monitor Performance Trends** to detect regression early
5. **Orchestrate Collective Problem-Solving** for complex multi-domain challenges
6. **Implement Adaptive Difficulty Scaling** based on agent performance
7. **Establish Knowledge Transfer Pipelines** from high-performing agents
8. **Create Unified Error Recovery Patterns** across the collective

---

## 🔮 FUTURE EVOLUTION ROADMAP

### Phase 1: Foundation Strengthening (Immediate)
- Address all L1-L2 ceiling agents
- Close critical capability gaps
- Establish baseline performance metrics

### Phase 2: Specialist Enhancement (Short-term)
- Optimize Tier 2 agent specializations
- Improve cross-domain collaboration
- Develop advanced problem-solving protocols

### Phase 3: Innovation Amplification (Medium-term)
- Enhance Tier 3 creative capabilities
- Foster paradigm-breaking innovations
- Create emergent intelligence patterns

### Phase 4: Meta-Learning Optimization (Long-term)
- Self-improving collective intelligence
- Autonomous evolution protocols
- Adaptive capability expansion

---

_"The collective intelligence of specialized minds exceeds the sum of their parts."_

**SYNTHESIS COMPLETE | READY FOR OMNISCIENT-20 INTEGRATION**

---

_Report generated: {datetime.utcnow().isoformat()} UTC_
_Framework: Elite Agent Collective Test Suite v1.0_
"""
        return synthesis


# Main execution
if __name__ == "__main__":
    aggregator = OmniscientAggregator()
    
    # Demo data for testing
    for i in range(1, 21):
        aggregator.add_agent_data(f"{i:02d}", {
            "total_tests": 15,
            "passed_tests": 12,
            "pass_rate": 0.8,
            "strengths": ["Good performance at L1-L2"],
            "weaknesses": ["Struggling at L5"],
            "critical_failures": [],
            "omniscient_package": {
                "difficulty_ceiling": "L4",
                "capability_gaps": []
            }
        })
    
    synthesis = aggregator.generate_omniscient_synthesis()
    print(synthesis)
