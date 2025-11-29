"""
Supreme Master Suite - Scenario Modules
========================================
Predefined collaboration scenarios for testing.
"""

from .cross_tier_scenarios import (
    foundational_meets_enterprise,
    specialists_meets_innovation,
    all_tiers_grand_challenge,
)
from .domain_fusion_scenarios import (
    security_fusion_scenario,
    ml_integration_scenario,
    cloud_native_scenario,
)
from .stress_cascade_scenarios import (
    tier_cascade_stress,
    agent_failure_cascade,
    resource_exhaustion_scenario,
)
from .innovation_scenarios import (
    paradigm_breakthrough_scenario,
    novel_synthesis_scenario,
    emergent_capability_scenario,
)

__all__ = [
    # Cross-tier scenarios
    "foundational_meets_enterprise",
    "specialists_meets_innovation",
    "all_tiers_grand_challenge",
    # Domain fusion scenarios
    "security_fusion_scenario",
    "ml_integration_scenario",
    "cloud_native_scenario",
    # Stress cascade scenarios
    "tier_cascade_stress",
    "agent_failure_cascade",
    "resource_exhaustion_scenario",
    # Innovation scenarios
    "paradigm_breakthrough_scenario",
    "novel_synthesis_scenario",
    "emergent_capability_scenario",
]
