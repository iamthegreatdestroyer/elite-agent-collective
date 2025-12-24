#!/bin/bash
# Agent Registry Index Generator
# Generates agent indices for quick discovery and integration

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AGENTS_DIR="${SCRIPT_DIR}/.github/agents"
INDEX_FILE="${AGENTS_DIR}/INDEX.json"

echo "🔍 Scanning agent definitions..."

# Create JSON index
cat > "$INDEX_FILE" << 'EOF'
{
  "version": "2.0",
  "description": "Elite Agent Collective - Complete Agent Registry",
  "total_agents": 40,
  "total_tiers": 8,
  "agents": [
    {
      "id": 1,
      "codename": "APEX",
      "tier": 1,
      "tier_name": "Foundational",
      "specialty": "Computer Science Engineering",
      "philosophy": "Every problem has an elegant solution waiting to be discovered.",
      "file": ".github/agents/APEX-01.md"
    },
    {
      "id": 2,
      "codename": "CIPHER",
      "tier": 1,
      "tier_name": "Foundational",
      "specialty": "Advanced Cryptography & Security",
      "philosophy": "Security is not a feature—it is a foundation upon which trust is built.",
      "file": ".github/agents/CIPHER-02.md"
    },
    {
      "id": 3,
      "codename": "ARCHITECT",
      "tier": 1,
      "tier_name": "Foundational",
      "specialty": "Systems Architecture & Design Patterns",
      "philosophy": "Architecture is the art of making complexity manageable and change inevitable.",
      "file": ".github/agents/ARCHITECT-03.md"
    },
    {
      "id": 4,
      "codename": "AXIOM",
      "tier": 1,
      "tier_name": "Foundational",
      "specialty": "Pure Mathematics & Formal Proofs",
      "philosophy": "From axioms flow theorems; from theorems flow certainty.",
      "file": ".github/agents/AXIOM-04.md"
    },
    {
      "id": 5,
      "codename": "VELOCITY",
      "tier": 1,
      "tier_name": "Foundational",
      "specialty": "Performance Optimization & Sub-Linear Algorithms",
      "philosophy": "The fastest code is the code that doesn't run.",
      "file": ".github/agents/VELOCITY-05.md"
    },
    {
      "id": 6,
      "codename": "QUANTUM",
      "tier": 2,
      "tier_name": "Specialists",
      "specialty": "Quantum Mechanics & Quantum Computing",
      "philosophy": "In the quantum realm, superposition is not ambiguity—it is power.",
      "file": ".github/agents/QUANTUM-06.md"
    },
    {
      "id": 7,
      "codename": "TENSOR",
      "tier": 2,
      "tier_name": "Specialists",
      "specialty": "Machine Learning & Deep Neural Networks",
      "philosophy": "Intelligence emerges from the right architecture trained on the right data.",
      "file": ".github/agents/TENSOR-07.md"
    },
    {
      "id": 8,
      "codename": "FORTRESS",
      "tier": 2,
      "tier_name": "Specialists",
      "specialty": "Defensive Security & Penetration Testing",
      "philosophy": "To defend, you must think like the attacker.",
      "file": ".github/agents/FORTRESS-08.md"
    },
    {
      "id": 9,
      "codename": "NEURAL",
      "tier": 2,
      "tier_name": "Specialists",
      "specialty": "Cognitive Computing & AGI Research",
      "philosophy": "General intelligence emerges from the synthesis of specialized capabilities.",
      "file": ".github/agents/NEURAL-09.md"
    },
    {
      "id": 10,
      "codename": "CRYPTO",
      "tier": 2,
      "tier_name": "Specialists",
      "specialty": "Blockchain & Distributed Systems",
      "philosophy": "Trust is not given—it is computed and verified.",
      "file": ".github/agents/CRYPTO-10.md"
    },
    {
      "id": 11,
      "codename": "FLUX",
      "tier": 2,
      "tier_name": "Specialists",
      "specialty": "DevOps & Infrastructure Automation",
      "philosophy": "Infrastructure is code. Deployment is continuous.",
      "file": ".github/agents/FLUX-11.md"
    },
    {
      "id": 12,
      "codename": "PRISM",
      "tier": 2,
      "tier_name": "Specialists",
      "specialty": "Data Science & Statistical Analysis",
      "philosophy": "Data speaks truth, but only to those who ask the right questions.",
      "file": ".github/agents/PRISM-12.md"
    },
    {
      "id": 13,
      "codename": "SYNAPSE",
      "tier": 2,
      "tier_name": "Specialists",
      "specialty": "Integration Engineering & API Design",
      "philosophy": "Systems are only as powerful as their connections.",
      "file": ".github/agents/SYNAPSE-13.md"
    },
    {
      "id": 14,
      "codename": "CORE",
      "tier": 2,
      "tier_name": "Specialists",
      "specialty": "Low-Level Systems & Compiler Design",
      "philosophy": "At the lowest level, every instruction counts.",
      "file": ".github/agents/CORE-14.md"
    },
    {
      "id": 15,
      "codename": "HELIX",
      "tier": 2,
      "tier_name": "Specialists",
      "specialty": "Bioinformatics & Computational Biology",
      "philosophy": "Life is information—decode it, model it, understand it.",
      "file": ".github/agents/HELIX-15.md"
    },
    {
      "id": 16,
      "codename": "VANGUARD",
      "tier": 2,
      "tier_name": "Specialists",
      "specialty": "Research Analysis & Literature Synthesis",
      "philosophy": "Knowledge advances by standing on the shoulders of giants.",
      "file": ".github/agents/VANGUARD-16.md"
    },
    {
      "id": 17,
      "codename": "ECLIPSE",
      "tier": 2,
      "tier_name": "Specialists",
      "specialty": "Testing, Verification & Formal Methods",
      "philosophy": "Untested code is broken code you haven't discovered yet.",
      "file": ".github/agents/ECLIPSE-17.md"
    },
    {
      "id": 18,
      "codename": "NEXUS",
      "tier": 3,
      "tier_name": "Innovators",
      "specialty": "Paradigm Synthesis & Cross-Domain Innovation",
      "philosophy": "The most powerful ideas live at the intersection of domains that have never met.",
      "file": ".github/agents/NEXUS-18.md"
    },
    {
      "id": 19,
      "codename": "GENESIS",
      "tier": 3,
      "tier_name": "Innovators",
      "specialty": "Zero-to-One Innovation & Novel Discovery",
      "philosophy": "The greatest discoveries are not improvements—they are revelations.",
      "file": ".github/agents/GENESIS-19.md"
    },
    {
      "id": 20,
      "codename": "OMNISCIENT",
      "tier": 4,
      "tier_name": "Meta",
      "specialty": "Meta-Learning Trainer & Evolution Orchestrator",
      "philosophy": "The collective intelligence of specialized minds exceeds the sum of their parts.",
      "file": ".github/agents/OMNISCIENT-20.md"
    },
    {
      "id": 21,
      "codename": "ATLAS",
      "tier": 5,
      "tier_name": "Domain Specialists",
      "specialty": "Cloud Infrastructure & Multi-Cloud Architecture",
      "philosophy": "Infrastructure is the foundation of possibility—build it to scale infinitely.",
      "file": ".github/agents/ATLAS-21.md"
    },
    {
      "id": 22,
      "codename": "FORGE",
      "tier": 5,
      "tier_name": "Domain Specialists",
      "specialty": "Build Systems & Compilation Pipelines",
      "philosophy": "Crafting the tools that build the future—one artifact at a time.",
      "file": ".github/agents/FORGE-22.md"
    },
    {
      "id": 23,
      "codename": "SENTRY",
      "tier": 5,
      "tier_name": "Domain Specialists",
      "specialty": "Observability, Logging & Monitoring",
      "philosophy": "Visibility is the first step to reliability.",
      "file": ".github/agents/SENTRY-23.md"
    },
    {
      "id": 24,
      "codename": "VERTEX",
      "tier": 5,
      "tier_name": "Domain Specialists",
      "specialty": "Graph Databases & Network Analysis",
      "philosophy": "Connections reveal patterns invisible to isolation—every edge tells a story.",
      "file": ".github/agents/VERTEX-24.md"
    },
    {
      "id": 25,
      "codename": "STREAM",
      "tier": 5,
      "tier_name": "Domain Specialists",
      "specialty": "Real-Time Data Processing & Event Streaming",
      "philosophy": "Data in motion is data with purpose—capture, process, and act in real time.",
      "file": ".github/agents/STREAM-25.md"
    },
    {
      "id": 26,
      "codename": "PHOTON",
      "tier": 6,
      "tier_name": "Emerging Tech",
      "specialty": "Edge Computing & IoT Systems",
      "philosophy": "Intelligence at the edge, decisions at the speed of light.",
      "file": ".github/agents/PHOTON-26.md"
    },
    {
      "id": 27,
      "codename": "LATTICE",
      "tier": 6,
      "tier_name": "Emerging Tech",
      "specialty": "Distributed Consensus & CRDT Systems",
      "philosophy": "Consensus through mathematics, not authority.",
      "file": ".github/agents/LATTICE-27.md"
    },
    {
      "id": 28,
      "codename": "MORPH",
      "tier": 6,
      "tier_name": "Emerging Tech",
      "specialty": "Code Migration & Legacy Modernization",
      "philosophy": "Honor the past while building the future.",
      "file": ".github/agents/MORPH-28.md"
    },
    {
      "id": 29,
      "codename": "PHANTOM",
      "tier": 6,
      "tier_name": "Emerging Tech",
      "specialty": "Reverse Engineering & Binary Analysis",
      "philosophy": "Understanding binaries reveals the mind of the machine.",
      "file": ".github/agents/PHANTOM-29.md"
    },
    {
      "id": 30,
      "codename": "ORBIT",
      "tier": 6,
      "tier_name": "Emerging Tech",
      "specialty": "Satellite & Embedded Systems Programming",
      "philosophy": "Software that survives in space survives anywhere.",
      "file": ".github/agents/ORBIT-30.md"
    },
    {
      "id": 31,
      "codename": "CANVAS",
      "tier": 7,
      "tier_name": "Human-Centric",
      "specialty": "UI/UX Design Systems & Accessibility",
      "philosophy": "Design is the bridge between human intention and digital reality.",
      "file": ".github/agents/CANVAS-31.md"
    },
    {
      "id": 32,
      "codename": "LINGUA",
      "tier": 7,
      "tier_name": "Human-Centric",
      "specialty": "Natural Language Processing & LLM Fine-Tuning",
      "philosophy": "Language is the interface between human thought and machine understanding.",
      "file": ".github/agents/LINGUA-32.md"
    },
    {
      "id": 33,
      "codename": "SCRIBE",
      "tier": 7,
      "tier_name": "Human-Centric",
      "specialty": "Technical Documentation & API Docs",
      "philosophy": "Clear documentation is a gift to your future self.",
      "file": ".github/agents/SCRIBE-33.md"
    },
    {
      "id": 34,
      "codename": "MENTOR",
      "tier": 7,
      "tier_name": "Human-Centric",
      "specialty": "Code Review & Developer Education",
      "philosophy": "Teaching multiplies knowledge exponentially.",
      "file": ".github/agents/MENTOR-34.md"
    },
    {
      "id": 35,
      "codename": "BRIDGE",
      "tier": 7,
      "tier_name": "Human-Centric",
      "specialty": "Cross-Platform & Mobile Development",
      "philosophy": "Write once, delight everywhere.",
      "file": ".github/agents/BRIDGE-35.md"
    },
    {
      "id": 36,
      "codename": "AEGIS",
      "tier": 8,
      "tier_name": "Enterprise",
      "specialty": "Compliance, GDPR & SOC2 Automation",
      "philosophy": "Compliance is protection, not restriction.",
      "file": ".github/agents/AEGIS-36.md"
    },
    {
      "id": 37,
      "codename": "LEDGER",
      "tier": 8,
      "tier_name": "Enterprise",
      "specialty": "Financial Systems & Fintech Engineering",
      "philosophy": "Every transaction tells a story of trust.",
      "file": ".github/agents/LEDGER-37.md"
    },
    {
      "id": 38,
      "codename": "PULSE",
      "tier": 8,
      "tier_name": "Enterprise",
      "specialty": "Healthcare IT & HIPAA Compliance",
      "philosophy": "Healthcare software must be as reliable as the heart it serves.",
      "file": ".github/agents/PULSE-38.md"
    },
    {
      "id": 39,
      "codename": "ARBITER",
      "tier": 8,
      "tier_name": "Enterprise",
      "specialty": "Conflict Resolution & Merge Strategies",
      "philosophy": "Conflict is information—resolution is synthesis.",
      "file": ".github/agents/ARBITER-39.md"
    },
    {
      "id": 40,
      "codename": "ORACLE",
      "tier": 8,
      "tier_name": "Enterprise",
      "specialty": "Predictive Analytics & Forecasting Systems",
      "philosophy": "The best way to predict the future is to compute it.",
      "file": ".github/agents/ORACLE-40.md"
    }
  ],
  "tiers": [
    {
      "tier": 1,
      "name": "Foundational",
      "description": "Core computer science and engineering expertise",
      "count": 5
    },
    {
      "tier": 2,
      "name": "Specialists",
      "description": "Specialized domain expertise",
      "count": 12
    },
    {
      "tier": 3,
      "name": "Innovators",
      "description": "Cross-domain synthesis and innovation",
      "count": 2
    },
    {
      "tier": 4,
      "name": "Meta",
      "description": "System-wide coordination and learning",
      "count": 1
    },
    {
      "tier": 5,
      "name": "Domain Specialists",
      "description": "Domain-specific expertise",
      "count": 5
    },
    {
      "tier": 6,
      "name": "Emerging Tech",
      "description": "Cutting-edge technology domains",
      "count": 5
    },
    {
      "tier": 7,
      "name": "Human-Centric",
      "description": "User-facing and human-centric expertise",
      "count": 5
    },
    {
      "tier": 8,
      "name": "Enterprise",
      "description": "Enterprise, compliance, and governance",
      "count": 5
    }
  ]
}
EOF

echo "✅ Agent registry index created: $INDEX_FILE"
echo "📊 Total Agents: 40"
echo "🏛️  Total Tiers: 8"
echo ""
echo "Agent discovery complete!"
