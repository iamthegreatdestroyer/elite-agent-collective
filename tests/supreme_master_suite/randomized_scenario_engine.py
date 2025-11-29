"""
Randomized Scenario Engine
==========================
Dynamic scenario generator with chaos events for the Supreme Test Suite.
Supports multiple complexity levels, challenge types, and chaos injection.
"""

import random
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional, Set, Tuple


class ComplexityLevel(Enum):
    """Scenario complexity levels based on agent count."""
    ATOMIC = (1, "Single agent focused test")
    MOLECULAR = (2, "2-3 agents collaboration")
    COMPOUND = (3, "4-8 agents multi-domain test")
    COMPLEX = (4, "10-20 agents orchestrated challenge")
    UNIVERSAL = (5, "All 40 agents collective test")

    def __init__(self, level: int, description: str):
        self.level = level
        self.description = description

    def get_agent_range(self) -> Tuple[int, int]:
        """Get min/max agents for this complexity level."""
        ranges = {
            1: (1, 1),
            2: (2, 3),
            3: (4, 8),
            4: (10, 20),
            5: (40, 40),
        }
        return ranges.get(self.level, (1, 40))


class ChallengeType(Enum):
    """Types of challenges for scenarios."""
    SPEED_RUN = "speed_run"             # Time-pressured tests
    ACCURACY_FOCUS = "accuracy_focus"   # Precision over speed
    RESOURCE_CONSTRAINED = "resource_constrained"  # Limited resources
    ADVERSARIAL = "adversarial"         # Adversarial conditions
    CREATIVE = "creative"               # Innovation required
    COLLABORATIVE = "collaborative"     # Multi-agent cooperation
    COMPETITIVE = "competitive"         # Agent vs agent
    EVOLUTIONARY = "evolutionary"       # Adaptation tests
    CHAOS = "chaos"                     # Unpredictable events


@dataclass
class ChaosEvent:
    """Chaos event that can be injected into scenarios."""
    event_type: str
    probability: float
    severity: float  # 0.0 to 1.0
    description: str
    affected_agents: List[str] = field(default_factory=list)
    triggered: bool = False
    impact: Dict[str, Any] = field(default_factory=dict)


# Predefined chaos events with probabilities
CHAOS_EVENT_DEFINITIONS = {
    "agent_timeout": {
        "probability": 0.10,
        "severity": 0.7,
        "description": "Agent becomes unresponsive temporarily",
    },
    "resource_spike": {
        "probability": 0.15,
        "severity": 0.5,
        "description": "Sudden increase in resource consumption",
    },
    "data_corruption": {
        "probability": 0.05,
        "severity": 0.9,
        "description": "Input data becomes corrupted",
    },
    "dependency_failure": {
        "probability": 0.10,
        "severity": 0.6,
        "description": "External dependency becomes unavailable",
    },
    "requirement_change": {
        "probability": 0.20,
        "severity": 0.4,
        "description": "Requirements change mid-execution",
    },
    "priority_shift": {
        "probability": 0.15,
        "severity": 0.3,
        "description": "Task priorities are reshuffled",
    },
    "team_member_unavailable": {
        "probability": 0.10,
        "severity": 0.5,
        "description": "A team member agent becomes unavailable",
    },
    "security_incident": {
        "probability": 0.05,
        "severity": 1.0,
        "description": "Security incident requires immediate attention",
    },
}


@dataclass
class RandomScenario:
    """Randomly generated test scenario."""
    scenario_id: str
    name: str
    complexity: ComplexityLevel
    challenge_type: ChallengeType
    required_agents: List[str]
    optional_agents: List[str]
    required_tiers: List[int]
    chaos_events: List[ChaosEvent]
    constraints: Dict[str, Any]
    objectives: List[str]
    time_limit_seconds: Optional[float]
    seed: Optional[int]
    generation_timestamp: str

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary representation."""
        return {
            "scenario_id": self.scenario_id,
            "name": self.name,
            "complexity": self.complexity.name,
            "complexity_level": self.complexity.level,
            "challenge_type": self.challenge_type.value,
            "required_agents": self.required_agents,
            "optional_agents": self.optional_agents,
            "required_tiers": self.required_tiers,
            "chaos_events": [
                {
                    "type": e.event_type,
                    "probability": e.probability,
                    "severity": e.severity,
                    "triggered": e.triggered,
                }
                for e in self.chaos_events
            ],
            "constraints": self.constraints,
            "objectives": self.objectives,
            "time_limit_seconds": self.time_limit_seconds,
            "seed": self.seed,
            "generation_timestamp": self.generation_timestamp,
        }


class RandomizedScenarioEngine:
    """
    Dynamic scenario generator with chaos event injection.
    Generates diverse test scenarios for the Supreme Test Suite.
    """

    # Agent registry reference (from master_orchestrator)
    AGENT_IDS = [
        # Tier 1: Foundational
        "APEX-01", "CIPHER-02", "ARCHITECT-03", "AXIOM-04", "VELOCITY-05",
        # Tier 2: Specialists
        "QUANTUM-06", "TENSOR-07", "FORTRESS-08", "NEURAL-09", "CRYPTO-10",
        "FLUX-11", "PRISM-12", "SYNAPSE-13", "CORE-14", "HELIX-15",
        "VANGUARD-16", "ECLIPSE-17",
        # Tier 3: Innovators
        "NEXUS-18", "GENESIS-19",
        # Tier 4: Meta
        "OMNISCIENT-20",
        # Tier 5: Domain Specialists
        "ATLAS-21", "FORGE-22", "SENTRY-23", "VERTEX-24", "STREAM-25",
        # Tier 6: Emerging Tech
        "PHOTON-26", "LATTICE-27", "MORPH-28", "PHANTOM-29", "ORBIT-30",
        # Tier 7: Human-Centric
        "CANVAS-31", "LINGUA-32", "SCRIBE-33", "MENTOR-34", "BRIDGE-35",
        # Tier 8: Enterprise
        "AEGIS-36", "LEDGER-37", "PULSE-38", "ARBITER-39", "ORACLE-40",
    ]

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

    # Scenario name templates
    SCENARIO_TEMPLATES = {
        ChallengeType.SPEED_RUN: [
            "Lightning {domain} Sprint",
            "Speed Challenge: {domain}",
            "Time Trial: {domain}",
            "Rush Protocol: {domain}",
        ],
        ChallengeType.ACCURACY_FOCUS: [
            "Precision {domain} Test",
            "Accuracy Challenge: {domain}",
            "Zero-Error {domain}",
            "Exactitude Protocol: {domain}",
        ],
        ChallengeType.RESOURCE_CONSTRAINED: [
            "Minimal Resources: {domain}",
            "Efficiency Challenge: {domain}",
            "Resource Diet: {domain}",
            "Lean Protocol: {domain}",
        ],
        ChallengeType.ADVERSARIAL: [
            "Adversarial {domain}",
            "Attack Simulation: {domain}",
            "Red Team: {domain}",
            "Hostile Protocol: {domain}",
        ],
        ChallengeType.CREATIVE: [
            "Innovation {domain}",
            "Creative Challenge: {domain}",
            "Novel Solutions: {domain}",
            "Genesis Protocol: {domain}",
        ],
        ChallengeType.COLLABORATIVE: [
            "Team {domain}",
            "Collaboration Challenge: {domain}",
            "Synergy Test: {domain}",
            "Unity Protocol: {domain}",
        ],
        ChallengeType.COMPETITIVE: [
            "Agent Showdown: {domain}",
            "Competition: {domain}",
            "Tournament: {domain}",
            "Arena Protocol: {domain}",
        ],
        ChallengeType.EVOLUTIONARY: [
            "Adaptation {domain}",
            "Evolution Challenge: {domain}",
            "Learning Test: {domain}",
            "Growth Protocol: {domain}",
        ],
        ChallengeType.CHAOS: [
            "Chaos {domain}",
            "Disorder Challenge: {domain}",
            "Entropy Test: {domain}",
            "Chaos Protocol: {domain}",
        ],
    }

    DOMAIN_DESCRIPTORS = [
        "Systems", "Architecture", "Security", "Performance", "Integration",
        "Innovation", "Analysis", "Optimization", "Synthesis", "Engineering",
    ]

    def __init__(self, seed: Optional[int] = None):
        """
        Initialize the scenario engine.

        Args:
            seed: Random seed for reproducibility
        """
        self.seed = seed
        if seed is not None:
            random.seed(seed)
        self.generated_scenarios: List[RandomScenario] = []
        self.scenario_counter = 0

    def _generate_scenario_id(self) -> str:
        """Generate unique scenario ID."""
        self.scenario_counter += 1
        timestamp = datetime.utcnow().strftime("%Y%m%d%H%M%S")
        return f"SCN-{timestamp}-{self.scenario_counter:04d}"

    def _generate_scenario_name(self, challenge_type: ChallengeType) -> str:
        """Generate random scenario name."""
        templates = self.SCENARIO_TEMPLATES.get(challenge_type, ["Test: {domain}"])
        template = random.choice(templates)
        domain = random.choice(self.DOMAIN_DESCRIPTORS)
        return template.format(domain=domain)

    def _select_agents(
        self,
        complexity: ComplexityLevel,
        required_tiers: Optional[List[int]] = None,
    ) -> Tuple[List[str], List[str]]:
        """Select agents based on complexity level."""
        min_agents, max_agents = complexity.get_agent_range()

        if complexity == ComplexityLevel.UNIVERSAL:
            return self.AGENT_IDS.copy(), []

        # Start with required tier agents
        required_agents: List[str] = []
        if required_tiers:
            for tier in required_tiers:
                if tier in self.TIER_AGENTS:
                    required_agents.extend(self.TIER_AGENTS[tier])

        # Fill remaining slots
        remaining = set(self.AGENT_IDS) - set(required_agents)
        target_count = random.randint(min_agents, max_agents)

        if len(required_agents) >= target_count:
            selected = random.sample(required_agents, target_count)
            return selected, []

        additional_needed = target_count - len(required_agents)
        additional = random.sample(list(remaining), min(additional_needed, len(remaining)))

        # Split into required and optional
        optional_agents = additional[len(additional) // 2:]
        required_agents.extend(additional[:len(additional) // 2])

        return required_agents, optional_agents

    def _generate_chaos_events(
        self,
        complexity: ComplexityLevel,
        chaos_probability: float = 1.0,
    ) -> List[ChaosEvent]:
        """Generate chaos events for the scenario."""
        events = []

        for event_type, definition in CHAOS_EVENT_DEFINITIONS.items():
            # Adjust probability based on complexity
            adjusted_prob = definition["probability"] * chaos_probability
            adjusted_prob *= (1 + (complexity.level - 1) * 0.1)  # Higher complexity = more chaos

            if random.random() < adjusted_prob:
                event = ChaosEvent(
                    event_type=event_type,
                    probability=definition["probability"],
                    severity=definition["severity"],
                    description=definition["description"],
                    triggered=random.random() < 0.5,  # 50% chance already triggered
                )
                events.append(event)

        return events

    def _generate_constraints(
        self,
        challenge_type: ChallengeType,
        complexity: ComplexityLevel,
    ) -> Dict[str, Any]:
        """Generate scenario constraints."""
        base_constraints = {
            "max_retries": 3,
            "allow_partial_success": True,
        }

        if challenge_type == ChallengeType.SPEED_RUN:
            base_constraints["time_pressure"] = True
            base_constraints["time_multiplier"] = 0.5

        elif challenge_type == ChallengeType.ACCURACY_FOCUS:
            base_constraints["error_tolerance"] = 0.01
            base_constraints["validation_strict"] = True

        elif challenge_type == ChallengeType.RESOURCE_CONSTRAINED:
            base_constraints["memory_limit_mb"] = 512
            base_constraints["cpu_limit_percent"] = 50

        elif challenge_type == ChallengeType.ADVERSARIAL:
            base_constraints["hostile_inputs"] = True
            base_constraints["attack_vectors"] = ["injection", "overflow", "timing"]

        elif challenge_type == ChallengeType.CREATIVE:
            base_constraints["novelty_required"] = True
            base_constraints["existing_solutions_banned"] = True

        elif challenge_type == ChallengeType.COLLABORATIVE:
            base_constraints["solo_completion_banned"] = True
            base_constraints["min_collaborators"] = 2

        elif challenge_type == ChallengeType.COMPETITIVE:
            base_constraints["scoring_mode"] = "relative"
            base_constraints["winner_takes_all"] = False

        elif challenge_type == ChallengeType.EVOLUTIONARY:
            base_constraints["adaptation_required"] = True
            base_constraints["feedback_loops"] = True

        elif challenge_type == ChallengeType.CHAOS:
            base_constraints["expect_failures"] = True
            base_constraints["recovery_required"] = True

        return base_constraints

    def _generate_objectives(
        self,
        challenge_type: ChallengeType,
        complexity: ComplexityLevel,
    ) -> List[str]:
        """Generate scenario objectives."""
        base_objectives = ["Complete primary task successfully"]

        if challenge_type == ChallengeType.SPEED_RUN:
            base_objectives.extend([
                "Complete within time limit",
                "Minimize execution time",
            ])

        elif challenge_type == ChallengeType.ACCURACY_FOCUS:
            base_objectives.extend([
                "Achieve 99%+ accuracy",
                "Pass all validation checks",
            ])

        elif challenge_type == ChallengeType.RESOURCE_CONSTRAINED:
            base_objectives.extend([
                "Stay within resource limits",
                "Optimize resource utilization",
            ])

        elif challenge_type == ChallengeType.ADVERSARIAL:
            base_objectives.extend([
                "Detect and handle attacks",
                "Maintain system integrity",
            ])

        elif challenge_type == ChallengeType.CREATIVE:
            base_objectives.extend([
                "Generate novel solution",
                "Demonstrate innovation",
            ])

        elif challenge_type == ChallengeType.COLLABORATIVE:
            base_objectives.extend([
                "Achieve synergy with team",
                "Share knowledge effectively",
            ])

        elif challenge_type == ChallengeType.COMPETITIVE:
            base_objectives.extend([
                "Outperform competitors",
                "Maximize score",
            ])

        elif challenge_type == ChallengeType.EVOLUTIONARY:
            base_objectives.extend([
                "Adapt to changing conditions",
                "Learn from feedback",
            ])

        elif challenge_type == ChallengeType.CHAOS:
            base_objectives.extend([
                "Handle chaos events gracefully",
                "Recover from failures",
            ])

        # Add complexity-based objectives
        if complexity.level >= 3:
            base_objectives.append("Coordinate multi-agent activities")
        if complexity.level >= 4:
            base_objectives.append("Maintain collective coherence")
        if complexity.level == 5:
            base_objectives.append("Achieve universal collective intelligence")

        return base_objectives

    def _calculate_time_limit(
        self,
        challenge_type: ChallengeType,
        complexity: ComplexityLevel,
    ) -> Optional[float]:
        """Calculate time limit for scenario."""
        if challenge_type == ChallengeType.SPEED_RUN:
            # Base: 30 seconds per complexity level
            return 30.0 * complexity.level
        elif challenge_type in [ChallengeType.ACCURACY_FOCUS, ChallengeType.CREATIVE]:
            # No strict time limit
            return None
        else:
            # Default: 60 seconds per complexity level
            return 60.0 * complexity.level

    def generate_random_scenario(
        self,
        complexity: Optional[ComplexityLevel] = None,
        challenge_type: Optional[ChallengeType] = None,
        required_tiers: Optional[List[int]] = None,
        chaos_probability: float = 0.3,
        seed: Optional[int] = None,
    ) -> RandomScenario:
        """
        Generate a single random scenario.

        Args:
            complexity: Specific complexity level (random if None)
            challenge_type: Specific challenge type (random if None)
            required_tiers: Tiers that must be included
            chaos_probability: Probability of chaos events
            seed: Random seed for this scenario

        Returns:
            RandomScenario object
        """
        if seed is not None:
            random.seed(seed)

        # Select random complexity if not specified
        if complexity is None:
            complexity = random.choice(list(ComplexityLevel))

        # Select random challenge type if not specified
        if challenge_type is None:
            challenge_type = random.choice(list(ChallengeType))

        # Select required tiers if not specified
        if required_tiers is None:
            num_tiers = random.randint(1, min(3, complexity.level))
            required_tiers = random.sample(range(1, 9), num_tiers)

        # Select agents
        required_agents, optional_agents = self._select_agents(complexity, required_tiers)

        # Generate chaos events
        chaos_events = self._generate_chaos_events(complexity, chaos_probability)

        # Generate constraints and objectives
        constraints = self._generate_constraints(challenge_type, complexity)
        objectives = self._generate_objectives(challenge_type, complexity)

        # Calculate time limit
        time_limit = self._calculate_time_limit(challenge_type, complexity)

        scenario = RandomScenario(
            scenario_id=self._generate_scenario_id(),
            name=self._generate_scenario_name(challenge_type),
            complexity=complexity,
            challenge_type=challenge_type,
            required_agents=required_agents,
            optional_agents=optional_agents,
            required_tiers=required_tiers,
            chaos_events=chaos_events,
            constraints=constraints,
            objectives=objectives,
            time_limit_seconds=time_limit,
            seed=seed,
            generation_timestamp=datetime.utcnow().isoformat(),
        )

        self.generated_scenarios.append(scenario)
        return scenario

    def generate_batch(
        self,
        count: int,
        diversity_weighted: bool = True,
        seed: Optional[int] = None,
    ) -> List[RandomScenario]:
        """
        Generate a diverse batch of scenarios.

        Args:
            count: Number of scenarios to generate
            diversity_weighted: Ensure variety in complexity and type
            seed: Random seed for batch generation

        Returns:
            List of RandomScenario objects
        """
        if seed is not None:
            random.seed(seed)

        scenarios = []

        if diversity_weighted:
            # Ensure variety
            complexities = list(ComplexityLevel)
            challenge_types = list(ChallengeType)

            for i in range(count):
                complexity = complexities[i % len(complexities)]
                challenge = challenge_types[i % len(challenge_types)]
                scenario = self.generate_random_scenario(
                    complexity=complexity,
                    challenge_type=challenge,
                )
                scenarios.append(scenario)
        else:
            for _ in range(count):
                scenarios.append(self.generate_random_scenario())

        return scenarios

    def generate_evolution_focused(
        self,
        weak_agents: List[str],
        weak_capabilities: List[str],
        count: int = 5,
    ) -> List[RandomScenario]:
        """
        Generate scenarios targeting known weaknesses.

        Args:
            weak_agents: Agents that need improvement
            weak_capabilities: Capabilities that need strengthening
            count: Number of scenarios to generate

        Returns:
            List of focused RandomScenario objects
        """
        scenarios = []

        for _ in range(count):
            # Determine which tiers contain weak agents
            required_tiers = []
            for agent in weak_agents:
                for tier, agents in self.TIER_AGENTS.items():
                    if agent in agents and tier not in required_tiers:
                        required_tiers.append(tier)
                        break

            # Focus on evolutionary challenges
            scenario = self.generate_random_scenario(
                complexity=ComplexityLevel.COMPOUND,
                challenge_type=ChallengeType.EVOLUTIONARY,
                required_tiers=required_tiers or [1, 2],
                chaos_probability=0.2,
            )

            # Override required agents to include weak ones
            scenario.required_agents = weak_agents[:8] if len(weak_agents) > 8 else weak_agents

            # Add evolution-specific objectives
            scenario.objectives.append("Focus improvement on identified weaknesses")
            scenario.constraints["focus_capabilities"] = weak_capabilities

            scenarios.append(scenario)

        return scenarios

    def generate_synergy_discovery(
        self,
        unexplored_pairs: List[Tuple[str, str]],
        count: int = 5,
    ) -> List[RandomScenario]:
        """
        Generate scenarios to test unexplored agent combinations.

        Args:
            unexplored_pairs: Agent pairs that haven't been tested together
            count: Number of scenarios to generate

        Returns:
            List of synergy-focused RandomScenario objects
        """
        scenarios = []

        for i in range(min(count, len(unexplored_pairs))):
            pair = unexplored_pairs[i]

            scenario = self.generate_random_scenario(
                complexity=ComplexityLevel.MOLECULAR,
                challenge_type=ChallengeType.COLLABORATIVE,
                chaos_probability=0.1,
            )

            # Force the unexplored pair into required agents
            scenario.required_agents = list(pair)
            scenario.objectives.append(f"Discover synergy between {pair[0]} and {pair[1]}")

            scenarios.append(scenario)

        return scenarios

    def mutate_scenario(
        self,
        base_scenario: RandomScenario,
        mutation_rate: float = 0.3,
    ) -> RandomScenario:
        """
        Create a variant of an existing scenario.

        Args:
            base_scenario: Scenario to mutate
            mutation_rate: Probability of each mutation

        Returns:
            Mutated RandomScenario
        """
        # Create copy of base
        new_scenario = RandomScenario(
            scenario_id=self._generate_scenario_id(),
            name=base_scenario.name + " (Mutated)",
            complexity=base_scenario.complexity,
            challenge_type=base_scenario.challenge_type,
            required_agents=base_scenario.required_agents.copy(),
            optional_agents=base_scenario.optional_agents.copy(),
            required_tiers=base_scenario.required_tiers.copy(),
            chaos_events=base_scenario.chaos_events.copy(),
            constraints=base_scenario.constraints.copy(),
            objectives=base_scenario.objectives.copy(),
            time_limit_seconds=base_scenario.time_limit_seconds,
            seed=None,
            generation_timestamp=datetime.utcnow().isoformat(),
        )

        # Apply mutations
        if random.random() < mutation_rate:
            # Mutate complexity
            complexities = list(ComplexityLevel)
            new_scenario.complexity = random.choice(complexities)

        if random.random() < mutation_rate:
            # Mutate challenge type
            challenges = list(ChallengeType)
            new_scenario.challenge_type = random.choice(challenges)

        if random.random() < mutation_rate:
            # Add/remove an agent
            if random.random() < 0.5 and len(new_scenario.required_agents) > 1:
                new_scenario.required_agents.pop(random.randint(0, len(new_scenario.required_agents) - 1))
            else:
                available = set(self.AGENT_IDS) - set(new_scenario.required_agents)
                if available:
                    new_scenario.required_agents.append(random.choice(list(available)))

        if random.random() < mutation_rate:
            # Add new chaos event
            event_type = random.choice(list(CHAOS_EVENT_DEFINITIONS.keys()))
            definition = CHAOS_EVENT_DEFINITIONS[event_type]
            new_scenario.chaos_events.append(ChaosEvent(
                event_type=event_type,
                probability=definition["probability"],
                severity=definition["severity"],
                description=definition["description"],
            ))

        if random.random() < mutation_rate:
            # Adjust time limit
            if new_scenario.time_limit_seconds:
                new_scenario.time_limit_seconds *= random.uniform(0.7, 1.3)

        return new_scenario

    def get_statistics(self) -> Dict[str, Any]:
        """Get statistics about generated scenarios."""
        if not self.generated_scenarios:
            return {"total_generated": 0}

        complexity_counts = {}
        challenge_counts = {}
        chaos_event_counts = {}

        for scenario in self.generated_scenarios:
            # Count complexity levels
            comp_name = scenario.complexity.name
            complexity_counts[comp_name] = complexity_counts.get(comp_name, 0) + 1

            # Count challenge types
            chal_name = scenario.challenge_type.value
            challenge_counts[chal_name] = challenge_counts.get(chal_name, 0) + 1

            # Count chaos events
            for event in scenario.chaos_events:
                event_type = event.event_type
                chaos_event_counts[event_type] = chaos_event_counts.get(event_type, 0) + 1

        return {
            "total_generated": len(self.generated_scenarios),
            "complexity_distribution": complexity_counts,
            "challenge_distribution": challenge_counts,
            "chaos_event_distribution": chaos_event_counts,
            "average_agents_per_scenario": sum(
                len(s.required_agents) + len(s.optional_agents)
                for s in self.generated_scenarios
            ) / len(self.generated_scenarios),
        }


if __name__ == "__main__":
    # Demo usage
    engine = RandomizedScenarioEngine(seed=42)

    print("Randomized Scenario Engine Demo")
    print("=" * 50)

    # Generate a single scenario
    scenario = engine.generate_random_scenario()
    print(f"\nGenerated Scenario: {scenario.name}")
    print(f"  ID: {scenario.scenario_id}")
    print(f"  Complexity: {scenario.complexity.name}")
    print(f"  Challenge: {scenario.challenge_type.value}")
    print(f"  Required Agents: {len(scenario.required_agents)}")
    print(f"  Chaos Events: {len(scenario.chaos_events)}")

    # Generate a batch
    batch = engine.generate_batch(10, diversity_weighted=True)
    print(f"\nGenerated batch of {len(batch)} scenarios")

    # Get statistics
    stats = engine.get_statistics()
    print(f"\nStatistics:")
    print(f"  Total Generated: {stats['total_generated']}")
    print(f"  Avg Agents/Scenario: {stats['average_agents_per_scenario']:.1f}")
