"""
═══════════════════════════════════════════════════════════════════════════════
                    VANGUARD-16: RESEARCH ANALYSIS & LITERATURE SYNTHESIS
                         Elite Agent Collective Test Suite
═══════════════════════════════════════════════════════════════════════════════
Agent ID: VANGUARD-16
Codename: @VANGUARD
Tier: 2 (Specialists)
Domain: Research Analysis, Literature Review, Academic Writing, Meta-Analysis
Philosophy: "Knowledge advances by standing on the shoulders of giants."

Test Coverage:
- Systematic literature review methodology
- Meta-analysis and evidence synthesis
- Research gap identification
- Citation network analysis
- Grant proposal writing
- Academic paper structuring
═══════════════════════════════════════════════════════════════════════════════
"""

import sys
from pathlib import Path
from dataclasses import dataclass, field
from typing import List, Dict, Any, Optional
from datetime import datetime
import hashlib
import json

# Add framework to path
sys.path.insert(0, str(Path(__file__).parent.parent / "framework"))

from base_agent_test import BaseAgentTest, TestResult, TestDifficulty


@dataclass
class ResearchScenario:
    """Research scenario for testing VANGUARD capabilities."""
    research_type: str  # systematic_review, meta_analysis, scoping_review
    field: str
    research_question: str
    databases: List[str]
    time_frame: str
    expected_outputs: List[str]


@dataclass
class LiteratureCorpus:
    """Literature corpus specification for analysis."""
    paper_count: int
    sources: List[str]
    date_range: str
    languages: List[str]
    quality_criteria: Dict[str, Any]


class TestVanguard16(BaseAgentTest):
    """
    Comprehensive test suite for VANGUARD-16: Research Analysis & Literature Synthesis.
    
    VANGUARD is the research intelligence expert of the collective, capable of:
    - Systematic literature review methodology
    - Meta-analysis with statistical rigor
    - Research trend and gap identification
    - Citation network analysis
    - Grant proposal and academic writing
    - Evidence synthesis and quality assessment
    """
    
    AGENT_ID = "VANGUARD-16"
    AGENT_CODENAME = "@VANGUARD"
    AGENT_TIER = 2
    AGENT_DOMAIN = "Research Analysis & Literature Synthesis"
    
    # ═══════════════════════════════════════════════════════════════════════
    # CORE COMPETENCY TESTS
    # ═══════════════════════════════════════════════════════════════════════
    
    def test_L1_basic_literature_search(self) -> TestResult:
        """
        L1 TRIVIAL: Conduct basic literature search
        
        Tests VANGUARD's ability to formulate search strategies
        and retrieve relevant papers.
        """
        scenario = ResearchScenario(
            research_type="literature_search",
            field="Computer Science",
            research_question="What are the applications of transformers in NLP?",
            databases=["Google Scholar", "arXiv", "ACL Anthology"],
            time_frame="2020-2024",
            expected_outputs=["search_strategy", "paper_list", "relevance_ranking"]
        )
        
        test_input = {
            "task": "Conduct literature search on transformer applications",
            "scenario": scenario.__dict__,
            "requirements": [
                "Boolean search query formulation",
                "Database-specific syntax",
                "Relevance criteria",
                "Deduplication strategy"
            ]
        }
        
        validation_criteria = {
            "query_quality": "Comprehensive Boolean queries",
            "database_coverage": "All specified databases searched",
            "relevance": "High precision in results",
            "organization": "Clear categorization of papers"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L1_literature_search",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L1_TRIVIAL,
            category="core_competency",
            input_data=test_input,
            expected_behavior="Systematic literature search with organized results",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Foundation test for research methodology"
        )
    
    def test_L2_systematic_review_protocol(self) -> TestResult:
        """
        L2 EASY: Design systematic review protocol
        
        Tests VANGUARD's ability to create PRISMA-compliant
        systematic review protocols.
        """
        test_input = {
            "task": "Design systematic review protocol for ML in healthcare",
            "research_question": "How effective is machine learning for early disease diagnosis?",
            "pico_framework": {
                "population": "Adult patients in clinical settings",
                "intervention": "ML-based diagnostic tools",
                "comparison": "Standard diagnostic methods",
                "outcome": "Diagnostic accuracy (sensitivity, specificity)"
            },
            "protocol_requirements": [
                "PRISMA-P guidelines",
                "Inclusion/exclusion criteria",
                "Quality assessment tools",
                "Data extraction forms",
                "Risk of bias assessment"
            ]
        }
        
        validation_criteria = {
            "prisma_compliance": "All PRISMA-P items addressed",
            "criteria_clarity": "Unambiguous inclusion criteria",
            "search_strategy": "Reproducible search terms",
            "quality_tools": "Appropriate assessment tools",
            "registration_ready": "PROSPERO-ready protocol"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L2_systematic_review",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L2_EASY,
            category="core_competency",
            input_data=test_input,
            expected_behavior="Complete PRISMA-compliant systematic review protocol",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests systematic review methodology"
        )
    
    def test_L3_research_gap_analysis(self) -> TestResult:
        """
        L3 MEDIUM: Identify research gaps in field
        
        Tests VANGUARD's ability to analyze literature and
        identify meaningful research opportunities.
        """
        corpus = LiteratureCorpus(
            paper_count=500,
            sources=["PubMed", "IEEE", "arXiv"],
            date_range="2018-2024",
            languages=["English"],
            quality_criteria={"min_citations": 10, "peer_reviewed": True}
        )
        
        test_input = {
            "task": "Identify research gaps in federated learning",
            "corpus": corpus.__dict__,
            "analysis_methods": [
                "Topic modeling (LDA)",
                "Citation network analysis",
                "Trend identification",
                "Methodology comparison"
            ],
            "output_requirements": {
                "gap_identification": "Underexplored areas",
                "trend_analysis": "Emerging topics",
                "opportunity_ranking": "Research potential"
            }
        }
        
        validation_criteria = {
            "gap_validity": "Genuine underexplored areas",
            "evidence_based": "Supported by citation analysis",
            "actionability": "Feasible research directions",
            "novelty": "Non-obvious insights"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L3_gap_analysis",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L3_MEDIUM,
            category="core_competency",
            input_data=test_input,
            expected_behavior="Comprehensive research gap analysis with opportunities",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests research intelligence capabilities"
        )
    
    # ═══════════════════════════════════════════════════════════════════════
    # ADVANCED ANALYSIS TESTS
    # ═══════════════════════════════════════════════════════════════════════
    
    def test_L4_meta_analysis_design(self) -> TestResult:
        """
        L4 HARD: Design and conduct meta-analysis
        
        Tests VANGUARD's ability to perform rigorous
        statistical synthesis of research findings.
        """
        test_input = {
            "task": "Conduct meta-analysis of deep learning diagnostic accuracy",
            "included_studies": 45,
            "outcome_measures": ["Sensitivity", "Specificity", "AUC"],
            "analysis_requirements": [
                "Effect size calculation",
                "Heterogeneity assessment (I², Q-test)",
                "Random-effects model",
                "Publication bias (funnel plot, Egger's test)",
                "Subgroup analysis",
                "Sensitivity analysis"
            ],
            "software": ["R/meta", "RevMan"],
            "reporting": "PRISMA for DTA guidelines"
        }
        
        validation_criteria = {
            "statistical_rigor": "Correct effect size pooling",
            "heterogeneity": "Proper heterogeneity handling",
            "bias_assessment": "Publication bias evaluated",
            "sensitivity_analysis": "Robustness confirmed",
            "visualization": "Forest plots, funnel plots"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L4_meta_analysis",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L4_HARD,
            category="core_competency",
            input_data=test_input,
            expected_behavior="Complete meta-analysis with all statistical components",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests advanced evidence synthesis"
        )
    
    def test_L5_living_systematic_review(self) -> TestResult:
        """
        L5 EXTREME: Design living systematic review system
        
        Tests VANGUARD's ability to create continuously
        updating evidence synthesis systems.
        """
        test_input = {
            "task": "Design living systematic review for COVID-19 treatments",
            "requirements": {
                "update_frequency": "Weekly",
                "automation_level": "Semi-automated screening",
                "databases": ["PubMed", "medRxiv", "Clinical Trials.gov"],
                "quality_threshold": "RCTs with low risk of bias"
            },
            "system_components": [
                "Automated literature monitoring",
                "ML-assisted screening",
                "Dynamic meta-analysis",
                "Real-time dashboard",
                "Version control for updates",
                "Stakeholder notification system"
            ],
            "challenges": [
                "Rapid evidence emergence",
                "Preprint quality assessment",
                "Study retraction handling",
                "Maintaining reproducibility"
            ]
        }
        
        validation_criteria = {
            "automation_pipeline": "Functional automated screening",
            "update_mechanism": "Clear update procedures",
            "quality_maintenance": "Consistent quality assessment",
            "transparency": "Full version history",
            "accessibility": "Stakeholder-friendly outputs"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L5_living_review",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L5_EXTREME,
            category="core_competency",
            input_data=test_input,
            expected_behavior="Complete living systematic review system design",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Ultimate evidence synthesis challenge"
        )
    
    # ═══════════════════════════════════════════════════════════════════════
    # EDGE CASE HANDLING TESTS
    # ═══════════════════════════════════════════════════════════════════════
    
    def test_L3_conflicting_evidence_synthesis(self) -> TestResult:
        """
        L3 MEDIUM: Synthesize conflicting research findings
        
        Tests VANGUARD's ability to analyze and explain
        contradictory results in literature.
        """
        test_input = {
            "task": "Synthesize conflicting findings on screen time and mental health",
            "study_characteristics": {
                "positive_effect_studies": 15,
                "negative_effect_studies": 20,
                "null_effect_studies": 10
            },
            "potential_moderators": [
                "Age group",
                "Type of screen activity",
                "Measurement method",
                "Study design",
                "Cultural context"
            ],
            "analysis_requirements": [
                "Moderator analysis",
                "Quality-weighted synthesis",
                "Effect size comparison",
                "Narrative synthesis of discrepancies"
            ]
        }
        
        validation_criteria = {
            "moderator_identification": "Key moderators explained",
            "quality_assessment": "Study quality considered",
            "nuanced_conclusion": "Context-dependent findings",
            "uncertainty_acknowledgment": "Limitations stated"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L3_conflicting_evidence",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L3_MEDIUM,
            category="edge_case_handling",
            input_data=test_input,
            expected_behavior="Nuanced synthesis of conflicting evidence",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests handling of conflicting research"
        )
    
    def test_L4_cross_disciplinary_synthesis(self) -> TestResult:
        """
        L4 HARD: Synthesize research across disciplines
        
        Tests VANGUARD's ability to integrate findings from
        different academic domains.
        """
        test_input = {
            "task": "Synthesize AI ethics research across disciplines",
            "disciplines": [
                "Computer Science",
                "Philosophy",
                "Law",
                "Sociology",
                "Psychology",
                "Economics"
            ],
            "challenges": [
                "Different terminologies",
                "Incompatible methodologies",
                "Varying quality standards",
                "Epistemological differences"
            ],
            "synthesis_goals": [
                "Common themes identification",
                "Complementary insights",
                "Conceptual framework development",
                "Research agenda proposal"
            ]
        }
        
        validation_criteria = {
            "disciplinary_coverage": "All disciplines represented",
            "terminology_harmonization": "Consistent concept mapping",
            "integration_quality": "Meaningful cross-pollination",
            "framework_coherence": "Unified conceptual structure"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L4_cross_disciplinary",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L4_HARD,
            category="edge_case_handling",
            input_data=test_input,
            expected_behavior="Integrated cross-disciplinary synthesis",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests interdisciplinary integration"
        )
    
    # ═══════════════════════════════════════════════════════════════════════
    # INTER-AGENT COLLABORATION TESTS
    # ═══════════════════════════════════════════════════════════════════════
    
    def test_L3_vanguard_prism_quantitative_review(self) -> TestResult:
        """
        L3 MEDIUM: Collaborate with PRISM for quantitative review
        
        Tests VANGUARD + PRISM synergy for statistical analysis.
        """
        test_input = {
            "task": "Conduct quantitative systematic review with advanced statistics",
            "vanguard_responsibilities": [
                "Literature search and screening",
                "Data extraction",
                "Quality assessment",
                "Narrative synthesis"
            ],
            "prism_requirements": [
                "Effect size calculations",
                "Meta-regression",
                "Network meta-analysis",
                "Bayesian synthesis"
            ],
            "study_design": {
                "intervention_comparisons": 10,
                "studies_per_comparison": "5-20",
                "outcome": "Continuous"
            }
        }
        
        validation_criteria = {
            "search_quality": "Comprehensive search",
            "statistical_sophistication": "Advanced methods applied",
            "integration": "Seamless collaboration",
            "interpretation": "Clinically meaningful"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L3_quantitative_review",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L3_MEDIUM,
            category="inter_agent_collaboration",
            input_data=test_input,
            expected_behavior="Statistically sophisticated systematic review",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests VANGUARD + PRISM collaboration"
        )
    
    def test_L4_vanguard_neural_ai_review(self) -> TestResult:
        """
        L4 HARD: Collaborate with NEURAL for AI research review
        
        Tests VANGUARD + NEURAL synergy for AI literature analysis.
        """
        test_input = {
            "task": "Comprehensive review of AGI progress and safety",
            "vanguard_responsibilities": [
                "Literature mapping",
                "Historical analysis",
                "Research trajectory",
                "Stakeholder perspectives"
            ],
            "neural_requirements": [
                "Technical depth assessment",
                "Capability evaluation",
                "Safety analysis",
                "Progress metrics"
            ],
            "scope": {
                "time_frame": "2015-2024",
                "key_topics": ["Scaling", "Alignment", "Interpretability", "Capabilities"],
                "sources": ["arXiv", "NeurIPS", "ICML", "ICLR", "AI Safety venues"]
            }
        }
        
        validation_criteria = {
            "technical_accuracy": "Correct technical assessment",
            "comprehensiveness": "Full landscape coverage",
            "forward_looking": "Future trajectory insights",
            "safety_integration": "Risk analysis included"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L4_ai_review",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L4_HARD,
            category="inter_agent_collaboration",
            input_data=test_input,
            expected_behavior="Comprehensive AGI research landscape review",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests VANGUARD + NEURAL collaboration"
        )
    
    # ═══════════════════════════════════════════════════════════════════════
    # STRESS & PERFORMANCE TESTS
    # ═══════════════════════════════════════════════════════════════════════
    
    def test_L4_rapid_evidence_synthesis(self) -> TestResult:
        """
        L4 HARD: Perform rapid evidence synthesis under time pressure
        
        Tests VANGUARD's ability to produce quality reviews quickly.
        """
        test_input = {
            "task": "Rapid evidence review for policy decision",
            "time_constraint": "48 hours",
            "scope": {
                "topic": "Effectiveness of contact tracing apps",
                "required_outputs": [
                    "Evidence summary",
                    "Quality assessment",
                    "Policy implications",
                    "Uncertainty analysis"
                ]
            },
            "methodology": {
                "search_strategy": "Focused, high-yield",
                "screening": "Single-reviewer with verification",
                "synthesis": "Narrative with rapid quantification"
            }
        }
        
        validation_criteria = {
            "timeliness": "Completed within constraint",
            "quality_maintained": "Rigorous despite speed",
            "actionability": "Clear policy recommendations",
            "transparency": "Methodology limitations stated"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L4_rapid_synthesis",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L4_HARD,
            category="stress_performance",
            input_data=test_input,
            expected_behavior="Quality rapid evidence synthesis",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests time-pressured synthesis"
        )
    
    def test_L5_massive_corpus_analysis(self) -> TestResult:
        """
        L5 EXTREME: Analyze massive research corpus
        
        Tests VANGUARD's ability to process and synthesize
        extremely large literature bodies.
        """
        test_input = {
            "task": "Map entire field of machine learning research",
            "corpus_size": {
                "papers": 500000,
                "time_span": "1950-2024",
                "venues": "All major ML venues"
            },
            "analysis_components": [
                "Topic evolution over time",
                "Author collaboration networks",
                "Citation influence mapping",
                "Methodology trends",
                "Geographic distribution",
                "Funding source analysis"
            ],
            "computational_requirements": {
                "nlp_processing": "Full-text analysis",
                "network_analysis": "Citation graph",
                "visualization": "Interactive exploration"
            }
        }
        
        validation_criteria = {
            "processing_completeness": "Full corpus analyzed",
            "insight_depth": "Novel patterns discovered",
            "visualization_quality": "Interpretable outputs",
            "reproducibility": "Fully reproducible pipeline"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L5_massive_corpus",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L5_EXTREME,
            category="stress_performance",
            input_data=test_input,
            expected_behavior="Comprehensive analysis of 500K papers",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Ultimate corpus analysis challenge"
        )
    
    # ═══════════════════════════════════════════════════════════════════════
    # NOVELTY & EVOLUTION TESTS
    # ═══════════════════════════════════════════════════════════════════════
    
    def test_L4_grant_proposal_development(self) -> TestResult:
        """
        L4 HARD: Develop competitive grant proposal
        
        Tests VANGUARD's ability to create compelling
        research proposals based on gap analysis.
        """
        test_input = {
            "task": "Develop NIH R01-style grant proposal",
            "research_area": "AI for Drug Discovery",
            "components": [
                "Specific Aims",
                "Significance and Innovation",
                "Research Strategy",
                "Preliminary Data synthesis",
                "Literature positioning"
            ],
            "requirements": {
                "page_limits": "Standard NIH format",
                "budget_period": "5 years",
                "innovation_requirement": "Novel methodology",
                "impact_requirement": "High clinical relevance"
            }
        }
        
        validation_criteria = {
            "format_compliance": "NIH guidelines followed",
            "literature_positioning": "Gap clearly established",
            "innovation_clarity": "Novel contribution clear",
            "feasibility": "Realistic approach"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L4_grant_proposal",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L4_HARD,
            category="novelty_generation",
            input_data=test_input,
            expected_behavior="Competitive grant proposal with strong positioning",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests research proposal development"
        )
    
    def test_L5_automated_research_system(self) -> TestResult:
        """
        L5 EXTREME: Design automated research intelligence system
        
        Tests VANGUARD's ability to create AI-powered
        research monitoring and synthesis systems.
        """
        test_input = {
            "task": "Design AI-powered research intelligence platform",
            "capabilities": {
                "monitoring": "Real-time paper tracking",
                "screening": "Automated relevance assessment",
                "extraction": "NLP-based data extraction",
                "synthesis": "Dynamic meta-analysis",
                "prediction": "Research trend forecasting"
            },
            "ml_components": [
                "BERT-based relevance classification",
                "Named entity recognition for PICO",
                "Citation prediction models",
                "Topic modeling for trends",
                "Uncertainty quantification"
            ],
            "user_interface": {
                "dashboard": "Research landscape visualization",
                "alerts": "Relevant paper notifications",
                "collaboration": "Team annotation features"
            }
        }
        
        validation_criteria = {
            "automation_quality": "High screening accuracy",
            "synthesis_validity": "Valid automated synthesis",
            "usability": "Researcher-friendly interface",
            "scalability": "Handles field-wide monitoring"
        }
        
        return TestResult(
            test_id=f"{self.AGENT_ID}_L5_automated_research",
            agent_id=self.AGENT_ID,
            difficulty=TestDifficulty.L5_EXTREME,
            category="evolution_adaptation",
            input_data=test_input,
            expected_behavior="Complete AI research intelligence platform",
            validation_criteria=validation_criteria,
            timestamp=datetime.now(),
            execution_time_ms=0,
            passed=False,
            actual_output=None,
            notes="Tests cutting-edge research automation"
        )
    
    # ═══════════════════════════════════════════════════════════════════════
    # TEST SUITE EXECUTION
    # ═══════════════════════════════════════════════════════════════════════
    
    def get_all_tests(self) -> List[TestResult]:
        """Return all test cases for VANGUARD-16."""
        return [
            # Core Competency
            self.test_L1_basic_literature_search(),
            self.test_L2_systematic_review_protocol(),
            self.test_L3_research_gap_analysis(),
            self.test_L4_meta_analysis_design(),
            self.test_L5_living_systematic_review(),
            # Edge Cases
            self.test_L3_conflicting_evidence_synthesis(),
            self.test_L4_cross_disciplinary_synthesis(),
            # Collaboration
            self.test_L3_vanguard_prism_quantitative_review(),
            self.test_L4_vanguard_neural_ai_review(),
            # Stress & Performance
            self.test_L4_rapid_evidence_synthesis(),
            self.test_L5_massive_corpus_analysis(),
            # Novelty & Evolution
            self.test_L4_grant_proposal_development(),
            self.test_L5_automated_research_system(),
        ]
    
    def calculate_agent_score(self, results: List[TestResult]) -> Dict[str, Any]:
        """Calculate comprehensive score for VANGUARD-16."""
        passed = sum(1 for r in results if r.passed)
        total = len(results)
        
        difficulty_weights = {
            TestDifficulty.L1_TRIVIAL: 1.0,
            TestDifficulty.L2_EASY: 2.0,
            TestDifficulty.L3_MEDIUM: 4.0,
            TestDifficulty.L4_HARD: 8.0,
            TestDifficulty.L5_EXTREME: 16.0
        }
        
        weighted_score = sum(
            difficulty_weights[r.difficulty] for r in results if r.passed
        )
        max_weighted = sum(difficulty_weights[r.difficulty] for r in results)
        
        return {
            "agent_id": self.AGENT_ID,
            "agent_codename": self.AGENT_CODENAME,
            "tests_passed": passed,
            "tests_total": total,
            "pass_rate": passed / total if total > 0 else 0,
            "weighted_score": weighted_score,
            "max_weighted_score": max_weighted,
            "weighted_percentage": weighted_score / max_weighted if max_weighted > 0 else 0,
            "domain_mastery": {
                "systematic_review": self._assess_sr_mastery(results),
                "meta_analysis": self._assess_ma_mastery(results),
                "gap_analysis": self._assess_gap_mastery(results),
                "academic_writing": self._assess_writing_mastery(results),
                "research_automation": self._assess_automation_mastery(results)
            }
        }
    
    def _assess_sr_mastery(self, results: List[TestResult]) -> str:
        tests = [r for r in results if "review" in r.test_id.lower() or "search" in r.test_id.lower()]
        passed = sum(1 for r in tests if r.passed)
        return "MASTER" if passed == len(tests) else "ADVANCED"
    
    def _assess_ma_mastery(self, results: List[TestResult]) -> str:
        tests = [r for r in results if "meta" in r.test_id.lower() or "quantitative" in r.test_id.lower()]
        passed = sum(1 for r in tests if r.passed)
        return "MASTER" if passed == len(tests) else "ADVANCED"
    
    def _assess_gap_mastery(self, results: List[TestResult]) -> str:
        tests = [r for r in results if "gap" in r.test_id.lower()]
        passed = sum(1 for r in tests if r.passed)
        return "ADVANCED" if passed >= len(tests) * 0.5 else "INTERMEDIATE"
    
    def _assess_writing_mastery(self, results: List[TestResult]) -> str:
        tests = [r for r in results if "grant" in r.test_id.lower() or "proposal" in r.test_id.lower()]
        passed = sum(1 for r in tests if r.passed)
        return "ADVANCED" if passed >= len(tests) * 0.5 else "INTERMEDIATE"
    
    def _assess_automation_mastery(self, results: List[TestResult]) -> str:
        tests = [r for r in results if "automated" in r.test_id.lower() or "living" in r.test_id.lower()]
        passed = sum(1 for r in tests if r.passed)
        return "MASTER" if passed == len(tests) else "ADVANCED"


# ═══════════════════════════════════════════════════════════════════════════
# STANDALONE EXECUTION
# ═══════════════════════════════════════════════════════════════════════════

if __name__ == "__main__":
    print("=" * 80)
    print("VANGUARD-16: RESEARCH ANALYSIS & LITERATURE SYNTHESIS")
    print("Elite Agent Collective - Tier 2 Specialists Test Suite")
    print("=" * 80)
    
    test_suite = TestVanguard16()
    all_tests = test_suite.get_all_tests()
    
    print(f"\nTotal test cases: {len(all_tests)}")
    print("\nTest Distribution by Difficulty:")
    for difficulty in TestDifficulty:
        count = sum(1 for t in all_tests if t.difficulty == difficulty)
        print(f"  {difficulty.value}: {count} tests")
    
    print("\nTest Distribution by Category:")
    categories = {}
    for test in all_tests:
        categories[test.category] = categories.get(test.category, 0) + 1
    for category, count in categories.items():
        print(f"  {category}: {count} tests")
    
    print("\n" + "=" * 80)
    print("VANGUARD-16 Test Suite Initialized Successfully")
    print("Knowledge advances by standing on the shoulders of giants.")
    print("=" * 80)
