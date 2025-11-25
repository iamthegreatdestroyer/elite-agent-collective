"""
Elite Agent Collective Test Suite
Tier 3: Innovators - Package Initialization

Agents in this tier:
- NEXUS-18: Paradigm Synthesis & Cross-Domain Innovation
- GENESIS-19: Zero-to-One Innovation & Novel Discovery
"""

from .test_nexus_18 import TestNexus18
from .test_genesis_19 import TestGenesis19

__all__ = [
    "TestNexus18",
    "TestGenesis19",
]

TIER_3_AGENTS = {
    "NEXUS-18": {
        "class": TestNexus18,
        "codename": "@NEXUS",
        "domain": "Paradigm Synthesis & Cross-Domain Innovation",
        "philosophy": "The most powerful ideas live at the intersection of domains that have never met."
    },
    "GENESIS-19": {
        "class": TestGenesis19,
        "codename": "@GENESIS",
        "domain": "Zero-to-One Innovation & Novel Discovery",
        "philosophy": "The greatest discoveries are not improvements—they are revelations."
    },
}
