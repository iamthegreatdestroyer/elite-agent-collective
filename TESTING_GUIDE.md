# Elite Agent Collective - Testing Guide

## Overview

The Elite Agent Collective includes a comprehensive testing framework that validates all 40 agents, their collaborative capabilities, learning mechanisms, and performance characteristics.

## Test Architecture

### Test Hierarchy

```
tests/
├── tier_1_foundational/           # Tier 1 agents (5 agents)
│   └── test_*.py                 # Individual tier tests
├── tier_2_specialists/            # Tier 2 agents (12 agents)
│   └── test_*.py
├── tier_3_innovators/             # Tier 3 agents (2 agents)
│   └── test_*.py
├── tier_4_meta/                   # Tier 4 agents (1 agent)
│   └── test_*.py
├── integration/                   # Integration & advanced tests
│   ├── test_agent_invocation.py   # All 40 agents
│   ├── test_collective_problem_solving.py
│   ├── test_evolution_protocols.py
│   ├── test_mnemonic_memory.py
│   ├── test_performance_benchmarks.py
│   └── test_github_actions_workflow.py
├── framework/                     # Test infrastructure
│   ├── base_agent_test.py
│   ├── test_runner.py
│   ├── difficulty_engine.py
│   ├── omniscient_aggregator.py
│   └── documentation_generator.py
├── run_all_tests.py              # Master test orchestrator
└── README.md                     # Test documentation
```

### Test Categories

1. **Tier Tests** (Foundational → Meta)

   - Agent capability validation
   - Tier-specific protocols
   - Hierarchical verification

2. **Integration Tests** (Advanced)

   - Multi-agent collaboration
   - Cross-tier communication
   - Performance benchmarking

3. **Memory Tests**

   - Experience storage/retrieval
   - Sub-linear techniques (Bloom, LSH, HNSW)
   - ReMem control loop (RETRIEVE→THINK→ACT→REFLECT→EVOLVE)

4. **Performance Tests**
   - Latency percentiles (P50, P95, P99)
   - Throughput measurement (RPS)
   - Memory scaling analysis

## Running Tests

### Quick Start

**Windows (PowerShell):**

```powershell
# Run all tests
.\run_tests.ps1

# Run specific test type
.\run_tests.ps1 -TestType tier1
.\run_tests.ps1 -TestType comprehensive
.\run_tests.ps1 -TestType performance

# With verbose output and report generation
.\run_tests.ps1 -TestType all -Verbose -GenerateReport
```

**Unix/Linux/macOS (Bash):**

```bash
# Make script executable (first time only)
chmod +x run_tests.sh

# Run all tests
./run_tests.sh

# Run specific test type
./run_tests.sh tier1
./run_tests.sh comprehensive
./run_tests.sh performance

# With verbose output
./run_tests.sh all --verbose
```

**Python Direct:**

```bash
# Run all tests
python tests/run_all_tests.py

# Run integration tests only
python tests/run_all_tests.py --test-type integration

# Generate JSON report
python tests/run_all_tests.py --report results/test-report.json
```

### Test Types

| Type            | Tests    | Coverage               | Time        |
| --------------- | -------- | ---------------------- | ----------- |
| `tier1`         | 125      | Foundational agents    | ~2 min      |
| `tier2`         | 180      | Specialist agents      | ~3 min      |
| `tier3`         | 50       | Innovator agents       | ~1 min      |
| `tier4`         | 40       | Meta agents            | ~1 min      |
| `integration`   | 95       | Multi-agent tests      | ~2 min      |
| `comprehensive` | 150      | Full integration suite | ~3 min      |
| `performance`   | 80       | Benchmarks             | ~5 min      |
| `memory`        | 110      | Memory system          | ~2 min      |
| `all`           | **820+** | **Full suite**         | **~15 min** |

## Test Modules

### 1. Agent Invocation Tests (`test_agent_invocation.py`)

**Purpose:** Validate all 40 agents respond correctly to invocations

**Coverage:**

- Tier 1: APEX, CIPHER, ARCHITECT, AXIOM, VELOCITY (5 agents)
- Tier 2: QUANTUM, TENSOR, FORTRESS, NEURAL, CRYPTO, FLUX, PRISM, SYNAPSE, CORE, HELIX, VANGUARD, ECLIPSE (12 agents)
- Tier 3: NEXUS, GENESIS (2 agents)
- Tier 4: OMNISCIENT (1 agent)
- Tiers 5-8: Domain specialists, emerging tech, human-centric, enterprise (20 agents)

**Test Methods:**

```python
def test_foundational_agents()
def test_specialist_agents()
def test_innovator_agents()
def test_meta_agents()
def test_domain_specialists()
def test_emerging_tech_agents()
def test_human_centric_agents()
def test_enterprise_agents()
```

**Success Criteria:**

- All agents return valid metadata
- Agent codenames match registry
- Capabilities properly documented
- Philosophy statements present

### 2. Collective Problem Solving (`test_collective_problem_solving.py`)

**Purpose:** Test multi-agent collaboration across tiers

**Coverage:**

- Tier 1 → Tier 2 collaboration
- Tier 2 → Tier 3 collaboration
- Cross-domain collaboration
- Breakthrough propagation
- Collective synthesis

**Test Methods:**

```python
def test_tier_1_to_tier_2_collaboration()
def test_tier_2_to_tier_3_collaboration()
def test_cross_domain_collaboration()
def test_breakthrough_propagation()
def test_collective_problem_solving()
```

**Success Criteria:**

- Agents communicate successfully
- Knowledge transfers between tiers
- Breakthroughs propagate efficiently
- Latency < 500ms per round

### 3. Evolution Protocols (`test_evolution_protocols.py`)

**Purpose:** Validate learning mechanisms and fitness evolution

**Coverage:**

- Fitness scoring (0.0-1.0)
- Breakthrough detection (threshold: 0.90)
- Experience persistence
- Tier-based performance distribution
- Temporal decay (λ=0.99)

**Test Methods:**

```python
def test_fitness_scoring()
def test_breakthrough_detection()
def test_fitness_evolution()
def test_tier_fitness_distribution()
def test_experience_persistence()
```

**Success Criteria:**

- Fitness scores increase with learning
- Breakthroughs detected correctly
- Fitness preserved across sessions
- Exponential decay working properly

### 4. MNEMONIC Memory System (`test_mnemonic_memory.py`)

**Purpose:** Comprehensive memory system validation

**Coverage:**

- Experience storage (dataclass-based)
- Sub-linear retrieval techniques:
  - Bloom Filter (O(1), ~1% FP rate)
  - LSH Index (O(1) expected)
  - HNSW Graph (O(log n))
  - Count-Min Sketch (O(1))
  - Cuckoo Filter (O(1))
- ReMem control loop (5 phases)
- Cross-agent knowledge sharing
- Breakthrough discovery
- Memory efficiency (~1.2× compression)

**Test Methods:**

```python
def test_memory_store_creation()
def test_experience_retrieval()
def test_fitness_based_ranking()
def test_temporal_decay()
def test_sub_linear_retrieval_techniques()
def test_remem_control_loop()
def test_cross_agent_knowledge_sharing()
def test_breakthrough_discovery()
def test_memory_efficiency()
```

**Success Criteria:**

- All retrieval techniques functional
- ReMem cycle completes in <500ms
- Memory overhead < 1.5×
- Cross-agent knowledge sharing works
- Breakthrough promotion occurs

### 5. Performance Benchmarks (`test_performance_benchmarks.py`)

**Purpose:** Load testing and performance measurement

**Coverage:**

- Latency percentiles (P50, P95, P99)
- Concurrent request handling
- Cache performance (70% hit rate)
- Memory scaling (100 → 50,000 items)
- Query optimization (4 levels)
- Inference throughput

**Test Methods:**

```python
def benchmark_agent_latency()
def benchmark_concurrent_requests()
def benchmark_cache_performance()
def benchmark_memory_scaling()
def benchmark_query_optimization()
def benchmark_inference_throughput()
```

**Performance Targets:**

- P50 latency: < 50ms
- P95 latency: < 150ms
- P99 latency: < 300ms
- Throughput: > 100 RPS
- Cache hit rate: 60-70%
- Memory growth: Linear or sub-linear

### 6. GitHub Actions Workflow (`test_github_actions_workflow.py`)

**Purpose:** CI/CD pipeline validation

**Coverage:**

- Workflow file structure (YAML)
- Trigger configurations
- Job dependency analysis
- Matrix strategies
- Artifact handling
- Environment variable management
- Secret reference validation
- Deployment safety gates
- Test result reporting

**Test Methods:**

```python
def test_workflow_file_structure()
def test_workflow_yaml_syntax()
def test_workflow_triggers()
def test_job_configuration()
def test_matrix_strategy()
def test_artifact_handling()
def test_environment_variables()
def test_conditional_execution()
def test_deployment_safety_gates()
def test_test_result_reporting()
```

**Success Criteria:**

- Valid YAML syntax
- All triggers defined
- Jobs properly configured
- Artifacts handled correctly
- Secrets not exposed

## Test Results Interpretation

### Result Summary Fields

```python
TestSuiteResult(
    execution_id: str              # Unique execution ID
    timestamp: str                 # ISO timestamp
    total_tests: int              # Total test count
    tests_passed: int             # Passed test count
    tests_failed: int             # Failed test count
    pass_rate: float              # Pass rate (0.0-1.0)
    tier_results: Dict            # Per-tier breakdown
    integration_results: Dict     # Integration test results
    agent_scores: Dict            # Per-agent scores
    execution_time_seconds: float # Total duration
    recommendations: List[str]    # Improvement recommendations
)
```

### Pass Rate Interpretation

| Rate    | Status     | Action                    |
| ------- | ---------- | ------------------------- |
| 95-100% | Excellent  | Continue monitoring       |
| 85-94%  | Good       | Minor improvements needed |
| 75-84%  | Acceptable | Plan improvements         |
| < 75%   | Critical   | Immediate action required |

### Example Output

```
================================================================================
TEST EXECUTION SUMMARY
================================================================================

Execution ID: EXEC-20240115-143022
Duration: 847.32 seconds

Total Tests: 820
Passed: 792 (96.6%)
Failed: 28

────────────────────────────────────────────────────────────────────────────────
TIER BREAKDOWN
────────────────────────────────────────────────────────────────────────────────

TIER_1_FOUNDATIONAL:
  Agents Tested: 5
  Tests: 123/125 (98.4%)

TIER_2_SPECIALISTS:
  Agents Tested: 12
  Tests: 174/180 (96.7%)

... [additional tiers] ...

────────────────────────────────────────────────────────────────────────────────
INTEGRATION TESTS
────────────────────────────────────────────────────────────────────────────────

Total: 92/95 (96.8%)

  Agent Invocation: 40/40 (100.0%)
  Collective Problem Solving: 22/23 (95.7%)
  Evolution Protocols: 18/20 (90.0%)
  MNEMONIC Memory: 9/9 (100.0%)
  Performance Benchmarks: 2/2 (100.0%)
  GitHub Actions Workflow: 1/1 (100.0%)

────────────────────────────────────────────────────────────────────────────────
RECOMMENDATIONS
────────────────────────────────────────────────────────────────────────────────

• [EXCELLENT] All agents performing at or above target levels.
  Continue monitoring and consider expanding test coverage.
```

## Troubleshooting

### Common Issues

#### 1. Python Not Found

```
Error: Python not found
Solution:
  - Install Python 3.8+ from python.org
  - Set PYTHON_EXE environment variable
  - Verify: python --version
```

#### 2. Test Runner Not Found

```
Error: Test runner not found: .../run_all_tests.py
Solution:
  - Verify tests/ directory exists
  - Check tests/run_all_tests.py exists
  - Run from repository root
```

#### 3. Module Import Errors

```
Error: ModuleNotFoundError: No module named 'xyz'
Solution:
  - Install dependencies: pip install -r requirements.txt
  - Check Python path
  - Verify tests/ is in PYTHONPATH
```

#### 4. Permission Denied (Unix/Linux)

```
Error: Permission denied: './run_tests.sh'
Solution:
  - Make executable: chmod +x run_tests.sh
  - Run: ./run_tests.sh
```

#### 5. Test Timeout

```
Error: Test timeout exceeded
Solution:
  - Run individual test types: ./run_tests.sh tier1
  - Check system resources
  - Reduce concurrent test count in config
```

## Performance Benchmarking

### Running Benchmarks

```bash
# Run all performance tests
./run_tests.sh performance

# Run with detailed output
./run_tests.sh performance --verbose

# Save results
./run_tests.sh performance --report-dir ./perf-results
```

### Benchmark Metrics

Each benchmark test produces:

- Min/Max/Mean/Median latencies (ms)
- P95 and P99 percentiles
- Throughput (requests per second)
- Peak memory usage (MB)
- Success rate (%)

### Performance Targets

**Latency:**

- Tier 1 agents: P50 < 30ms, P95 < 100ms, P99 < 200ms
- Tier 2 agents: P50 < 50ms, P95 < 150ms, P99 < 300ms
- Tier 3 agents: P50 < 40ms, P95 < 120ms, P99 < 250ms
- Tier 4 agents: P50 < 60ms, P95 < 180ms, P99 < 350ms

**Throughput:**

- Single agent: 100+ RPS
- Concurrent (5 agents): 50+ RPS
- Cache with 70% hit rate: 500+ RPS

**Memory:**

- Per-agent baseline: 1-5 MB
- Experience storage: 1 KB per experience
- Total overhead: < 1.5× experience data size

## Continuous Integration

### GitHub Actions Integration

The test suite integrates with GitHub Actions:

```yaml
# .github/workflows/integration-tests.yml
name: Integration Tests
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-python@v4
      - run: ./run_tests.sh all
      - uses: actions/upload-artifact@v3
        if: always()
        with:
          name: test-results
          path: test-reports/
```

### Success Criteria for CI

- Overall pass rate: ≥ 95%
- No critical failures
- Performance within targets
- All tier tests passing
- Integration tests passing

## Best Practices

### For Developers

1. **Run tests before committing:**

   ```bash
   ./run_tests.sh all
   ```

2. **Run relevant tests when modifying:**

   - Agent code: `./run_tests.sh tier1` (or specific tier)
   - Integration: `./run_tests.sh integration`
   - Memory: `./run_tests.sh memory`
   - Performance: `./run_tests.sh performance`

3. **Check test reports:**

   ```bash
   cat test-reports/test-results-all-*.json | python -m json.tool
   ```

4. **Monitor performance:**
   - Collect baseline metrics
   - Track P95/P99 latencies over time
   - Alert if performance degrades

### For CI/CD

1. **Run on every commit:** Ensures quality gate
2. **Archive results:** For historical analysis
3. **Alert on failures:** Notify team immediately
4. **Track trends:** Monitor pass rates over time

## Contributing

To add new tests:

1. **Create test class:**

   ```python
   from framework.base_agent_test import BaseAgentTest

   class TestNewFeature(BaseAgentTest):
       def __init__(self):
           super().__init__()

       def test_feature_one(self):
           # Test implementation
           return True
   ```

2. **Run validation:**

   ```bash
   python -m pytest tests/test_new_feature.py
   ```

3. **Update documentation:**
   - Add test to README.md
   - Document test coverage
   - Update performance targets if needed

## Support

For testing issues:

- Check [docs/troubleshooting/common-issues.md](../docs/troubleshooting/common-issues.md)
- Review test output logs
- File GitHub issue with test output
- Contact development team

---

**Last Updated:** January 2024
**Framework Version:** 2.0
**Python Minimum:** 3.8+
