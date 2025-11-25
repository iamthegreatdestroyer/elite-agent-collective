"""
Elite Agent Collective - Difficulty Engine
==========================================
Dynamic difficulty scaling and calibration system.
"""

from dataclasses import dataclass
from enum import Enum
from typing import Dict, List, Any, Optional
import yaml
from pathlib import Path


class DifficultyEngine:
    """
    Manages difficulty scaling and calibration for agent tests.
    """

    def __init__(self, config_path: Optional[str] = None):
        self.config_path = config_path or "config/difficulty_matrices.yaml"
        self.difficulty_data: Dict[str, Any] = {}
        self.performance_benchmarks: Dict[str, Dict] = {}
        self.collaboration_matrix: Dict[str, List[str]] = {}
        self._load_config()

    def _load_config(self) -> None:
        """Load difficulty configuration from YAML."""
        config_file = Path(__file__).parent.parent / self.config_path
        if config_file.exists():
            with open(config_file, 'r') as f:
                config = yaml.safe_load(f)
                self.difficulty_data = config.get('difficulty_scaling', {})
                self.performance_benchmarks = config.get('performance_benchmarks', {})
                self.collaboration_matrix = config.get('collaboration_matrix', {})

    def get_difficulty_description(self, agent_codename: str, level: str) -> str:
        """Get difficulty description for a specific agent and level."""
        agent_key = f"{agent_codename}_{agent_codename.split('_')[0]}" if '_' not in agent_codename else agent_codename
        
        # Try to find the agent in difficulty data
        for key in self.difficulty_data:
            if agent_codename.upper() in key.upper():
                return self.difficulty_data[key].get(level, "No description available")
        
        return "No description available"

    def get_expected_performance(self, level: str) -> Dict[str, Any]:
        """Get expected performance benchmarks for a difficulty level."""
        return self.performance_benchmarks.get(level, {
            "max_execution_time_ms": 1000,
            "expected_pass_rate": 0.5
        })

    def get_collaboration_partners(self, agent_codename: str) -> List[str]:
        """Get recommended collaboration partners for an agent."""
        for key, partners in self.collaboration_matrix.items():
            if agent_codename.upper() in key.upper():
                return partners
        return []

    def calculate_adaptive_difficulty(
        self,
        agent_codename: str,
        historical_pass_rate: float,
        current_level: str
    ) -> str:
        """
        Calculate adaptive difficulty based on historical performance.
        
        Returns recommended difficulty level.
        """
        levels = ["L1", "L2", "L3", "L4", "L5"]
        current_idx = levels.index(current_level)
        
        if historical_pass_rate >= 0.9 and current_idx < 4:
            return levels[current_idx + 1]
        elif historical_pass_rate < 0.3 and current_idx > 0:
            return levels[current_idx - 1]
        
        return current_level

    def generate_difficulty_report(self, agent_results: Dict[str, Any]) -> str:
        """Generate a difficulty analysis report for an agent."""
        report = f"""# Difficulty Analysis Report

## Agent: {agent_results.get('agent_codename', 'Unknown')}

### Performance by Difficulty Level

"""
        for level in ["L1", "L2", "L3", "L4", "L5"]:
            level_data = agent_results.get('difficulty_breakdown', {}).get(level, {})
            if level_data:
                benchmark = self.get_expected_performance(level)
                report += f"""
#### {level} - {self._get_level_name(level)}

| Metric | Actual | Expected |
|--------|--------|----------|
| Pass Rate | {level_data.get('pass_rate', 0):.2%} | {benchmark.get('expected_pass_rate', 0):.2%} |
| Avg Time | {level_data.get('avg_time', 0):.2f}ms | {benchmark.get('max_execution_time_ms', 0)}ms |

"""
        
        return report

    def _get_level_name(self, level: str) -> str:
        """Get human-readable name for difficulty level."""
        names = {
            "L1": "TRIVIAL",
            "L2": "STANDARD",
            "L3": "ADVANCED",
            "L4": "EXPERT",
            "L5": "EXTREME"
        }
        return names.get(level, "UNKNOWN")

    def validate_test_difficulty(
        self,
        test_name: str,
        claimed_level: str,
        execution_time_ms: float,
        complexity_indicators: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Validate if a test's claimed difficulty is accurate.
        
        Returns validation result with suggested corrections.
        """
        benchmark = self.get_expected_performance(claimed_level)
        max_time = benchmark.get('max_execution_time_ms', 1000)
        
        # Calculate actual difficulty based on metrics
        time_factor = execution_time_ms / max_time
        complexity_factor = complexity_indicators.get('complexity_score', 0.5)
        
        suggested_level = claimed_level
        if time_factor > 2.0 and claimed_level != "L5":
            # Test is harder than claimed
            levels = ["L1", "L2", "L3", "L4", "L5"]
            current_idx = levels.index(claimed_level)
            suggested_level = levels[min(current_idx + 1, 4)]
        elif time_factor < 0.2 and claimed_level != "L1":
            # Test is easier than claimed
            levels = ["L1", "L2", "L3", "L4", "L5"]
            current_idx = levels.index(claimed_level)
            suggested_level = levels[max(current_idx - 1, 0)]
        
        return {
            "test_name": test_name,
            "claimed_level": claimed_level,
            "suggested_level": suggested_level,
            "is_accurate": claimed_level == suggested_level,
            "time_factor": time_factor,
            "complexity_factor": complexity_factor
        }
