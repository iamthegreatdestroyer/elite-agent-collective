"""
Elite Agent Collective - Framework Package
==========================================
"""

from .base_agent_test import (
    BaseAgentTest,
    TestResult,
    AgentTestSummary,
    DifficultyLevel,
    TestCategory
)
from .difficulty_engine import DifficultyEngine
from .documentation_generator import DocumentationGenerator
from .omniscient_aggregator import OmniscientAggregator, CollectiveIntelligence

__all__ = [
    'BaseAgentTest',
    'TestResult',
    'AgentTestSummary',
    'DifficultyLevel',
    'TestCategory',
    'DifficultyEngine',
    'DocumentationGenerator',
    'OmniscientAggregator',
    'CollectiveIntelligence'
]
