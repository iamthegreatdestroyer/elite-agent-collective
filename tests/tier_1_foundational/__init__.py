"""
Elite Agent Collective Test Suite
Tier 1: Foundational - Package Initialization

Agents in this tier:
- APEX-01: Elite Computer Science Engineering
- CIPHER-02: Advanced Cryptography & Security
- ARCHITECT-03: Systems Architecture & Design Patterns
- AXIOM-04: Pure Mathematics & Formal Proofs
- VELOCITY-05: Performance Optimization & Sub-Linear Algorithms
"""

from .test_apex_01 import TestApex01
from .test_cipher_02 import TestCipher02
from .test_architect_03 import TestArchitect03
from .test_axiom_04 import TestAxiom04
from .test_velocity_05 import TestVelocity05

__all__ = [
    "TestApex01",
    "TestCipher02",
    "TestArchitect03",
    "TestAxiom04",
    "TestVelocity05",
]

TIER_1_AGENTS = {
    "APEX-01": {
        "class": TestApex01,
        "codename": "@APEX",
        "domain": "Elite Computer Science Engineering",
        "philosophy": "Every problem has an elegant solution waiting to be discovered."
    },
    "CIPHER-02": {
        "class": TestCipher02,
        "codename": "@CIPHER",
        "domain": "Advanced Cryptography & Security",
        "philosophy": "Security is not a feature—it is a foundation upon which trust is built."
    },
    "ARCHITECT-03": {
        "class": TestArchitect03,
        "codename": "@ARCHITECT",
        "domain": "Systems Architecture & Design Patterns",
        "philosophy": "Architecture is the art of making complexity manageable and change inevitable."
    },
    "AXIOM-04": {
        "class": TestAxiom04,
        "codename": "@AXIOM",
        "domain": "Pure Mathematics & Formal Proofs",
        "philosophy": "From axioms flow theorems; from theorems flow certainty."
    },
    "VELOCITY-05": {
        "class": TestVelocity05,
        "codename": "@VELOCITY",
        "domain": "Performance Optimization & Sub-Linear Algorithms",
        "philosophy": "The fastest code is the code that doesn't run. The second fastest is the code that runs once."
    },
}
