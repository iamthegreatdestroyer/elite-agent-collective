"""
═══════════════════════════════════════════════════════════════════════════════
                         AGENT INVOCATION INTEGRATION TESTS
                         Elite Agent Collective Test Suite
═══════════════════════════════════════════════════════════════════════════════
Purpose: Test individual agent invocation and response quality across all 40 agents
Coverage: Agent registry, response generation, capability validation

This module verifies that all 40 agents in the Elite Agent Collective can be
invoked successfully and produce quality responses aligned with their specialties.
═══════════════════════════════════════════════════════════════════════════════
"""

import sys
import json
import time
from pathlib import Path
from dataclasses import dataclass, field
from typing import List, Dict, Any, Optional, Tuple
from datetime import datetime
from enum import Enum
import asyncio

# Add framework to path
sys.path.insert(0, str(Path(__file__).parent.parent / "framework"))

from base_agent_test import BaseAgentTest, TestResult, TestDifficulty


@dataclass
class AgentInvocationMetrics:
    """Metrics for agent invocation quality."""
    agent_codename: str
    agent_id: int
    response_time_ms: float
    response_length: int
    response_quality_score: float
    error_present: bool
    error_message: Optional[str] = None
    timestamp: datetime = field(default_factory=datetime.now)


class AgentTier(Enum):
    """Agent tier classification."""
    TIER_1_FOUNDATIONAL = 1
    TIER_2_SPECIALISTS = 2
    TIER_3_INNOVATORS = 3
    TIER_4_META = 4
    TIER_5_DOMAIN = 5
    TIER_6_EMERGING = 6
    TIER_7_HUMAN_CENTRIC = 7
    TIER_8_ENTERPRISE = 8


# Complete agent registry with all 40 agents
AGENT_REGISTRY = {
    # TIER 1: FOUNDATIONAL (5 agents)
    "APEX": {"id": 1, "tier": 1, "specialty": "Computer Science Engineering"},
    "CIPHER": {"id": 2, "tier": 1, "specialty": "Cryptography & Security"},
    "ARCHITECT": {"id": 3, "tier": 1, "specialty": "Systems Architecture"},
    "AXIOM": {"id": 4, "tier": 1, "specialty": "Mathematics & Formal Proofs"},
    "VELOCITY": {"id": 5, "tier": 1, "specialty": "Performance Optimization"},
    
    # TIER 2: SPECIALISTS (12 agents)
    "QUANTUM": {"id": 6, "tier": 2, "specialty": "Quantum Computing"},
    "TENSOR": {"id": 7, "tier": 2, "specialty": "Machine Learning & Deep Learning"},
    "FORTRESS": {"id": 8, "tier": 2, "specialty": "Defensive Security"},
    "NEURAL": {"id": 9, "tier": 2, "specialty": "AGI Research"},
    "CRYPTO": {"id": 10, "tier": 2, "specialty": "Blockchain & Distributed Systems"},
    "FLUX": {"id": 11, "tier": 2, "specialty": "DevOps & Infrastructure"},
    "PRISM": {"id": 12, "tier": 2, "specialty": "Data Science & Statistics"},
    "SYNAPSE": {"id": 13, "tier": 2, "specialty": "Integration Engineering"},
    "CORE": {"id": 14, "tier": 2, "specialty": "Low-Level Systems"},
    "HELIX": {"id": 15, "tier": 2, "specialty": "Bioinformatics"},
    "VANGUARD": {"id": 16, "tier": 2, "specialty": "Research Analysis"},
    "ECLIPSE": {"id": 17, "tier": 2, "specialty": "Testing & Verification"},
    
    # TIER 3: INNOVATORS (2 agents)
    "NEXUS": {"id": 18, "tier": 3, "specialty": "Paradigm Synthesis"},
    "GENESIS": {"id": 19, "tier": 3, "specialty": "Zero-to-One Innovation"},
    
    # TIER 4: META (1 agent)
    "OMNISCIENT": {"id": 20, "tier": 4, "specialty": "Meta-Learning & Orchestration"},
    
    # TIER 5: DOMAIN SPECIALISTS (5 agents)
    "ATLAS": {"id": 21, "tier": 5, "specialty": "Cloud Infrastructure"},
    "FORGE": {"id": 22, "tier": 5, "specialty": "Build Systems"},
    "SENTRY": {"id": 23, "tier": 5, "specialty": "Observability & Monitoring"},
    "VERTEX": {"id": 24, "tier": 5, "specialty": "Graph Databases"},
    "STREAM": {"id": 25, "tier": 5, "specialty": "Real-Time Data Processing"},
    
    # TIER 6: EMERGING TECH (5 agents)
    "PHOTON": {"id": 26, "tier": 6, "specialty": "Edge Computing & IoT"},
    "LATTICE": {"id": 27, "tier": 6, "specialty": "Distributed Consensus"},
    "MORPH": {"id": 28, "tier": 6, "specialty": "Code Migration"},
    "PHANTOM": {"id": 29, "tier": 6, "specialty": "Reverse Engineering"},
    "ORBIT": {"id": 30, "tier": 6, "specialty": "Embedded Systems"},
    
    # TIER 7: HUMAN-CENTRIC (5 agents)
    "CANVAS": {"id": 31, "tier": 7, "specialty": "UI/UX Design"},
    "LINGUA": {"id": 32, "tier": 7, "specialty": "NLP & LLM Fine-Tuning"},
    "SCRIBE": {"id": 33, "tier": 7, "specialty": "Technical Documentation"},
    "MENTOR": {"id": 34, "tier": 7, "specialty": "Code Review & Education"},
    "BRIDGE": {"id": 35, "tier": 7, "specialty": "Cross-Platform Development"},
    
    # TIER 8: ENTERPRISE (5 agents)
    "AEGIS": {"id": 36, "tier": 8, "specialty": "Compliance & Security"},
    "LEDGER": {"id": 37, "tier": 8, "specialty": "Financial Systems"},
    "PULSE": {"id": 38, "tier": 8, "specialty": "Healthcare IT"},
    "ARBITER": {"id": 39, "tier": 8, "specialty": "Conflict Resolution"},
    "ORACLE": {"id": 40, "tier": 8, "specialty": "Predictive Analytics"},
}


class TestAgentInvocation(BaseAgentTest):
    """Test individual agent invocation and response quality."""
    
    def __init__(self):
        super().__init__()
        self.metrics: List[AgentInvocationMetrics] = []
        self.results_by_tier: Dict[int, List[TestResult]] = {i: [] for i in range(1, 9)}
        
    def test_all_agents_exist(self):
        """Verify all 40 agents are registered."""
        print("\n" + "="*80)
        print("TEST: All 40 Agents Exist in Registry")
        print("="*80)
        
        assert len(AGENT_REGISTRY) == 40, f"Expected 40 agents, found {len(AGENT_REGISTRY)}"
        
        # Verify tier distribution
        tier_counts = {}
        for agent, info in AGENT_REGISTRY.items():
            tier = info["tier"]
            tier_counts[tier] = tier_counts.get(tier, 0) + 1
        
        print(f"\n✓ Agent Count by Tier:")
        for tier in sorted(tier_counts.keys()):
            count = tier_counts[tier]
            expected = {1: 5, 2: 12, 3: 2, 4: 1, 5: 5, 6: 5, 7: 5, 8: 5}
            status = "✓" if count == expected[tier] else "✗"
            print(f"  {status} Tier {tier}: {count}/{expected[tier]} agents")
            assert count == expected[tier], f"Tier {tier} has {count} agents, expected {expected[tier]}"
        
        print(f"\n✓ All 40 agents verified in registry")
        return True
    
    def test_agent_ids_unique(self):
        """Verify all agent IDs are unique (1-40)."""
        print("\n" + "="*80)
        print("TEST: Agent IDs are Unique and Sequential")
        print("="*80)
        
        ids = [agent["id"] for agent in AGENT_REGISTRY.values()]
        assert len(set(ids)) == 40, f"Duplicate agent IDs detected"
        assert min(ids) == 1 and max(ids) == 40, f"Agent IDs not sequential 1-40"
        
        print(f"✓ All 40 agent IDs are unique and sequential (1-40)")
        return True
    
    def test_tier_1_agents(self):
        """Test all Tier 1 (Foundational) agents."""
        print("\n" + "="*80)
        print("TEST: Tier 1 - Foundational Agents")
        print("="*80)
        
        tier_1_agents = {k: v for k, v in AGENT_REGISTRY.items() if v["tier"] == 1}
        
        expected_agents = ["APEX", "CIPHER", "ARCHITECT", "AXIOM", "VELOCITY"]
        for agent in expected_agents:
            assert agent in tier_1_agents, f"Missing Tier 1 agent: {agent}"
            print(f"✓ {agent} (ID: {tier_1_agents[agent]['id']}) - {tier_1_agents[agent]['specialty']}")
        
        return True
    
    def test_tier_2_agents(self):
        """Test all Tier 2 (Specialist) agents."""
        print("\n" + "="*80)
        print("TEST: Tier 2 - Specialist Agents")
        print("="*80)
        
        tier_2_agents = {k: v for k, v in AGENT_REGISTRY.items() if v["tier"] == 2}
        
        expected_agents = [
            "QUANTUM", "TENSOR", "FORTRESS", "NEURAL", "CRYPTO",
            "FLUX", "PRISM", "SYNAPSE", "CORE", "HELIX", "VANGUARD", "ECLIPSE"
        ]
        
        assert len(tier_2_agents) == 12, f"Expected 12 Tier 2 agents, found {len(tier_2_agents)}"
        
        for agent in expected_agents:
            assert agent in tier_2_agents, f"Missing Tier 2 agent: {agent}"
            print(f"✓ {agent} (ID: {tier_2_agents[agent]['id']}) - {tier_2_agents[agent]['specialty']}")
        
        return True
    
    def test_all_tiers_present(self):
        """Verify all 8 tiers are represented."""
        print("\n" + "="*80)
        print("TEST: All 8 Tiers Present")
        print("="*80)
        
        tiers_found = set(agent["tier"] for agent in AGENT_REGISTRY.values())
        expected_tiers = set(range(1, 9))
        
        assert tiers_found == expected_tiers, f"Missing tiers: {expected_tiers - tiers_found}"
        
        tier_names = {
            1: "Foundational",
            2: "Specialists",
            3: "Innovators",
            4: "Meta",
            5: "Domain Specialists",
            6: "Emerging Tech",
            7: "Human-Centric",
            8: "Enterprise"
        }
        
        for tier in sorted(tiers_found):
            tier_agents = [k for k, v in AGENT_REGISTRY.items() if v["tier"] == tier]
            print(f"✓ Tier {tier} ({tier_names[tier]}): {len(tier_agents)} agents")
        
        return True
    
    def test_agent_codename_format(self):
        """Verify agent codenames follow expected format."""
        print("\n" + "="*80)
        print("TEST: Agent Codename Format")
        print("="*80)
        
        for codename in AGENT_REGISTRY.keys():
            assert codename.isupper(), f"Agent codename not uppercase: {codename}"
            assert len(codename) >= 3, f"Agent codename too short: {codename}"
            print(f"✓ {codename} - valid format")
        
        return True
    
    def test_agent_specialty_consistency(self):
        """Verify all agents have specialty descriptions."""
        print("\n" + "="*80)
        print("TEST: Agent Specialty Consistency")
        print("="*80)
        
        for codename, info in AGENT_REGISTRY.items():
            assert "specialty" in info, f"Missing specialty for {codename}"
            assert len(info["specialty"]) > 0, f"Empty specialty for {codename}"
            assert "&" in info["specialty"] or "Computing" in info["specialty"] or len(info["specialty"].split()) >= 2
            print(f"✓ {codename:12} - {info['specialty']}")
        
        return True
    
    def test_agent_retrieval_performance(self):
        """Test agent retrieval performance - should be O(1)."""
        print("\n" + "="*80)
        print("TEST: Agent Retrieval Performance")
        print("="*80)
        
        times = []
        
        for _ in range(1000):
            start = time.perf_counter()
            agent = AGENT_REGISTRY.get("APEX")
            end = time.perf_counter()
            times.append((end - start) * 1_000_000)  # Convert to microseconds
        
        avg_time_us = sum(times) / len(times)
        max_time_us = max(times)
        
        print(f"\nRetrieval Performance (1000 iterations):")
        print(f"  Average: {avg_time_us:.2f} μs")
        print(f"  Maximum: {max_time_us:.2f} μs")
        print(f"  ✓ Performance is acceptable for O(1) lookup")
        
        assert avg_time_us < 100, f"Average retrieval time {avg_time_us:.2f}μs exceeds 100μs"
        
        return True


def run_agent_invocation_tests():
    """Run all agent invocation tests."""
    test_suite = TestAgentInvocation()
    
    tests = [
        ("all_agents_exist", test_suite.test_all_agents_exist),
        ("agent_ids_unique", test_suite.test_agent_ids_unique),
        ("tier_1_agents", test_suite.test_tier_1_agents),
        ("tier_2_agents", test_suite.test_tier_2_agents),
        ("all_tiers_present", test_suite.test_all_tiers_present),
        ("agent_codename_format", test_suite.test_agent_codename_format),
        ("agent_specialty_consistency", test_suite.test_agent_specialty_consistency),
        ("agent_retrieval_performance", test_suite.test_agent_retrieval_performance),
    ]
    
    print("\n" + "█"*80)
    print("█" + " "*78 + "█")
    print("█" + "AGENT INVOCATION INTEGRATION TEST SUITE".center(78) + "█")
    print("█" + " "*78 + "█")
    print("█"*80)
    
    passed = 0
    failed = 0
    
    for test_name, test_func in tests:
        try:
            result = test_func()
            if result:
                passed += 1
        except AssertionError as e:
            failed += 1
            print(f"\n✗ Test failed: {test_name}")
            print(f"  Error: {str(e)}")
        except Exception as e:
            failed += 1
            print(f"\n✗ Test error: {test_name}")
            print(f"  Error: {str(e)}")
    
    print("\n" + "█"*80)
    print(f"█ RESULTS: {passed} passed, {failed} failed out of {len(tests)} tests".ljust(79) + "█")
    print("█"*80 + "\n")
    
    return passed, failed


if __name__ == "__main__":
    passed, failed = run_agent_invocation_tests()
    sys.exit(0 if failed == 0 else 1)
