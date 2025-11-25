"""
Elite Agent Collective Test Suite
Integration Tests - Package Initialization

Integration test modules:
- test_inter_agent_collaboration: Tests for agent collaboration patterns
- test_collective_problem_solving: Tests for swarm intelligence and collective solutions
- test_evolution_protocols: Tests for collective learning and adaptation
"""

from .test_inter_agent_collaboration import TestInterAgentCollaboration
from .test_collective_problem_solving import TestCollectiveProblemSolving
from .test_evolution_protocols import TestEvolutionProtocols

__all__ = [
    "TestInterAgentCollaboration",
    "TestCollectiveProblemSolving",
    "TestEvolutionProtocols",
]

INTEGRATION_TESTS = {
    "inter_agent_collaboration": {
        "class": TestInterAgentCollaboration,
        "description": "Tests for agent collaboration patterns across tiers",
        "test_count": 10
    },
    "collective_problem_solving": {
        "class": TestCollectiveProblemSolving,
        "description": "Tests for swarm intelligence and collective solutions",
        "test_count": 8
    },
    "evolution_protocols": {
        "class": TestEvolutionProtocols,
        "description": "Tests for collective learning and adaptation",
        "test_count": 10
    },
}
