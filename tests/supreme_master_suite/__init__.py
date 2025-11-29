"""
═══════════════════════════════════════════════════════════════════════════════
                    SUPREME MASTER TEST SUITE
         OMNISCIENT-20 Learning Database & Collective Intelligence
═══════════════════════════════════════════════════════════════════════════════

The Supreme Master Test Suite provides comprehensive testing infrastructure
for all 40 agents across 8 tiers, with integrated learning capabilities
for OMNISCIENT-20 (Agent #20).

Components:
- MasterOrchestrator: Coordinates all 40 agents in unified testing scenarios
- OmniscientLearningDB: SQLite-based learning repository for Agent #20
- RandomizedScenarioEngine: Dynamic scenario generator with chaos events
- CollectiveIntelligence: Cross-agent pattern synthesis
- EvolutionTracker: Capability growth monitoring over time

Author: Elite Agent Collective
Version: 2.0.0
"""

__version__ = "2.0.0"
__author__ = "Elite Agent Collective"

from .master_orchestrator import (
    MasterOrchestrator,
    TestMode,
    AgentProfile,
    CollectiveTestResult,
    AGENT_REGISTRY,
)
from .omniscient_learning_db import OmniscientLearningDB
from .randomized_scenario_engine import (
    RandomizedScenarioEngine,
    ComplexityLevel,
    ChallengeType,
    ChaosEvent,
    RandomScenario,
)
from .collective_intelligence import (
    CollectiveIntelligence,
    EmergentCapability,
    SynergyMatrix,
)
from .evolution_tracker import EvolutionTracker

__all__ = [
    # Version info
    "__version__",
    "__author__",
    # Master Orchestrator
    "MasterOrchestrator",
    "TestMode",
    "AgentProfile",
    "CollectiveTestResult",
    "AGENT_REGISTRY",
    # Learning Database
    "OmniscientLearningDB",
    # Scenario Engine
    "RandomizedScenarioEngine",
    "ComplexityLevel",
    "ChallengeType",
    "ChaosEvent",
    "RandomScenario",
    # Collective Intelligence
    "CollectiveIntelligence",
    "EmergentCapability",
    "SynergyMatrix",
    # Evolution Tracker
    "EvolutionTracker",
]
