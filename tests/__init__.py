"""
═══════════════════════════════════════════════════════════════════════════════
                    ELITE AGENT COLLECTIVE TEST SUITE
                         Comprehensive Testing Framework
═══════════════════════════════════════════════════════════════════════════════

The Elite Agent Collective Test Suite provides comprehensive testing for all
20 specialized AI agents organized across 4 tiers:

Tier 1 - Foundational (5 agents):
    APEX-01, CIPHER-02, ARCHITECT-03, AXIOM-04, VELOCITY-05

Tier 2 - Specialists (12 agents):
    QUANTUM-06, TENSOR-07, FORTRESS-08, NEURAL-09, CRYPTO-10,
    FLUX-11, PRISM-12, SYNAPSE-13, CORE-14, HELIX-15,
    VANGUARD-16, ECLIPSE-17

Tier 3 - Innovators (2 agents):
    NEXUS-18, GENESIS-19

Tier 4 - Meta (1 agent):
    OMNISCIENT-20

Usage:
    from tests import run_all_tests
    from tests.framework import BaseAgentTest, TestResult, TestDifficulty
    from tests.tier_1_foundational import TestApex01

Author: Elite Agent Collective
Version: 1.0.0
"""

__version__ = "1.0.0"
__author__ = "Elite Agent Collective"

# Framework imports
from .framework.base_agent_test import BaseAgentTest, TestResult, DifficultyLevel as TestDifficulty
from .framework.difficulty_engine import DifficultyEngine
from .framework.test_runner import MasterTestRunner as AgentTestRunner
from .framework.documentation_generator import DocumentationGenerator
from .framework.omniscient_aggregator import OmniscientAggregator

# Tier imports
from . import tier_1_foundational
from . import tier_2_specialists
from . import tier_3_innovators
from . import tier_4_meta
from . import integration

__all__ = [
    # Version info
    "__version__",
    "__author__",
    # Framework
    "BaseAgentTest",
    "TestResult",
    "TestDifficulty",
    "DifficultyEngine",
    "AgentTestRunner",
    "DocumentationGenerator",
    "OmniscientAggregator",
    # Tiers
    "tier_1_foundational",
    "tier_2_specialists",
    "tier_3_innovators",
    "tier_4_meta",
    "integration",
]

# Agent registry for quick lookup
AGENT_REGISTRY = {
    # Tier 1: Foundational
    "APEX-01": {"tier": 1, "codename": "@APEX", "domain": "Software Engineering"},
    "CIPHER-02": {"tier": 1, "codename": "@CIPHER", "domain": "Cryptography & Security"},
    "ARCHITECT-03": {"tier": 1, "codename": "@ARCHITECT", "domain": "Systems Architecture"},
    "AXIOM-04": {"tier": 1, "codename": "@AXIOM", "domain": "Pure Mathematics"},
    "VELOCITY-05": {"tier": 1, "codename": "@VELOCITY", "domain": "Performance Optimization"},
    # Tier 2: Specialists
    "QUANTUM-06": {"tier": 2, "codename": "@QUANTUM", "domain": "Quantum Computing"},
    "TENSOR-07": {"tier": 2, "codename": "@TENSOR", "domain": "Machine Learning"},
    "FORTRESS-08": {"tier": 2, "codename": "@FORTRESS", "domain": "Defensive Security"},
    "NEURAL-09": {"tier": 2, "codename": "@NEURAL", "domain": "AGI Research"},
    "CRYPTO-10": {"tier": 2, "codename": "@CRYPTO", "domain": "Blockchain"},
    "FLUX-11": {"tier": 2, "codename": "@FLUX", "domain": "DevOps & Infrastructure"},
    "PRISM-12": {"tier": 2, "codename": "@PRISM", "domain": "Data Science"},
    "SYNAPSE-13": {"tier": 2, "codename": "@SYNAPSE", "domain": "Integration Engineering"},
    "CORE-14": {"tier": 2, "codename": "@CORE", "domain": "Low-Level Systems"},
    "HELIX-15": {"tier": 2, "codename": "@HELIX", "domain": "Bioinformatics"},
    "VANGUARD-16": {"tier": 2, "codename": "@VANGUARD", "domain": "Research Analysis"},
    "ECLIPSE-17": {"tier": 2, "codename": "@ECLIPSE", "domain": "Testing & Verification"},
    # Tier 3: Innovators
    "NEXUS-18": {"tier": 3, "codename": "@NEXUS", "domain": "Paradigm Synthesis"},
    "GENESIS-19": {"tier": 3, "codename": "@GENESIS", "domain": "Novel Discovery"},
    # Tier 4: Meta
    "OMNISCIENT-20": {"tier": 4, "codename": "@OMNISCIENT", "domain": "Collective Orchestration"},
}

def get_agent_info(agent_id: str) -> dict:
    """Get information about a specific agent."""
    return AGENT_REGISTRY.get(agent_id, {})

def list_agents_by_tier(tier: int) -> list:
    """List all agents in a specific tier."""
    return [
        agent_id for agent_id, info in AGENT_REGISTRY.items()
        if info["tier"] == tier
    ]

def get_total_test_count() -> int:
    """Get total number of tests in the suite."""
    return 328  # 20 agents * 15 tests + 28 integration tests
