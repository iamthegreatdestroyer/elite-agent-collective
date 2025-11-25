"""
═══════════════════════════════════════════════════════════════════════════════
                    TIER 2 SPECIALISTS - TEST PACKAGE INITIALIZATION
                         Elite Agent Collective Test Suite
═══════════════════════════════════════════════════════════════════════════════

This package contains comprehensive test suites for all 12 Tier 2 Specialist Agents:

QUANTUM-06  - Quantum Mechanics & Quantum Computing
TENSOR-07   - Machine Learning & Deep Neural Networks
FORTRESS-08 - Defensive Security & Penetration Testing
NEURAL-09   - Cognitive Computing & AGI Research
CRYPTO-10   - Blockchain & Distributed Ledger Systems
FLUX-11     - DevOps & Infrastructure Automation
PRISM-12    - Data Science & Statistical Analysis
SYNAPSE-13  - Integration Engineering & API Design
CORE-14     - Low-Level Systems & Compiler Design
HELIX-15    - Bioinformatics & Computational Biology
VANGUARD-16 - Research Analysis & Literature Synthesis
ECLIPSE-17  - Testing, Verification & Formal Methods

Each agent test suite includes:
- Core competency tests (L1-L5 difficulty progression)
- Edge case handling tests
- Inter-agent collaboration tests
- Stress and performance tests
- Novelty generation and evolution tests

═══════════════════════════════════════════════════════════════════════════════
"""

from .test_quantum_06 import TestQuantum06
from .test_tensor_07 import TestTensor07
from .test_fortress_08 import TestFortress08
from .test_neural_09 import TestNeural09
from .test_crypto_10 import TestCrypto10
from .test_flux_11 import TestFlux11
from .test_prism_12 import TestPrism12
from .test_synapse_13 import TestSynapse13
from .test_core_14 import TestCore14
from .test_helix_15 import TestHelix15
from .test_vanguard_16 import TestVanguard16
from .test_eclipse_17 import TestEclipse17

# Registry of all Tier 2 Specialist agent test classes
TIER_2_AGENTS = {
    "QUANTUM-06": TestQuantum06,
    "TENSOR-07": TestTensor07,
    "FORTRESS-08": TestFortress08,
    "NEURAL-09": TestNeural09,
    "CRYPTO-10": TestCrypto10,
    "FLUX-11": TestFlux11,
    "PRISM-12": TestPrism12,
    "SYNAPSE-13": TestSynapse13,
    "CORE-14": TestCore14,
    "HELIX-15": TestHelix15,
    "VANGUARD-16": TestVanguard16,
    "ECLIPSE-17": TestEclipse17,
}

# Agent metadata for documentation
TIER_2_METADATA = {
    "tier_name": "Specialists",
    "tier_number": 2,
    "agent_count": 12,
    "total_tests": sum(
        len(agent_class().get_all_tests()) 
        for agent_class in TIER_2_AGENTS.values()
    ) if False else "~156",  # Lazy evaluation marker
    "domains": [
        "Quantum Computing",
        "Machine Learning",
        "Security",
        "AGI Research",
        "Blockchain",
        "DevOps",
        "Data Science",
        "API Design",
        "Systems Programming",
        "Bioinformatics",
        "Research Analysis",
        "Testing & Verification"
    ]
}

__all__ = [
    # Test Classes
    "TestQuantum06",
    "TestTensor07",
    "TestFortress08",
    "TestNeural09",
    "TestCrypto10",
    "TestFlux11",
    "TestPrism12",
    "TestSynapse13",
    "TestCore14",
    "TestHelix15",
    "TestVanguard16",
    "TestEclipse17",
    # Registries
    "TIER_2_AGENTS",
    "TIER_2_METADATA",
]
