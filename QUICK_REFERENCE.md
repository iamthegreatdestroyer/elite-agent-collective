# Elite Agent Collective - Quick Reference Card

## Testing Quick Start

### Windows (PowerShell)

```powershell
# Run all tests
.\run_tests.ps1

# Run specific test type
.\run_tests.ps1 -TestType tier1
.\run_tests.ps1 -TestType comprehensive
.\run_tests.ps1 -TestType performance

# Verbose output
.\run_tests.ps1 -Verbose
```

### Unix/Linux/macOS (Bash)

```bash
# Make executable (first time)
chmod +x run_tests.sh

# Run tests
./run_tests.sh all
./run_tests.sh tier1
./run_tests.sh comprehensive

# With options
./run_tests.sh performance --verbose
```

### Python Direct

```bash
python tests/run_all_tests.py
python tests/run_all_tests.py --test-type integration
```

## Test Types

| Type            | Duration  | Coverage                  |
| --------------- | --------- | ------------------------- |
| `all`           | 13-15 min | All agents + integration  |
| `tier1`         | 2 min     | Foundational (5 agents)   |
| `tier2`         | 3 min     | Specialists (12 agents)   |
| `tier3`         | 1 min     | Innovators (2 agents)     |
| `tier4`         | 1 min     | Meta agents (1 agent)     |
| `integration`   | 2 min     | Multi-agent collaboration |
| `comprehensive` | 3 min     | Full integration suite    |
| `performance`   | 5 min     | Benchmarks                |
| `memory`        | 2 min     | MNEMONIC system           |

## Key Files

```
tests/
├── integration/
│   ├── test_agent_invocation.py (40 agents)
│   ├── test_collective_problem_solving.py (collab)
│   ├── test_evolution_protocols.py (learning)
│   ├── test_mnemonic_memory.py (memory system)
│   ├── test_performance_benchmarks.py (perf)
│   └── test_github_actions_workflow.py (CI/CD)
├── run_all_tests.py (master orchestrator)
├── run_integration_tests.py (integration orchestrator)
├── framework/ (test infrastructure)
└── README.md

Root:
├── run_tests.ps1 (Windows runner)
├── run_tests.sh (Unix runner)
├── TESTING_GUIDE.md (full documentation)
└── TASK_5_COMPLETION.md (completion summary)
```

## Test Structure

Every test module follows this pattern:

```python
from framework.base_agent_test import BaseAgentTest

class Test[Name](BaseAgentTest):
    def __init__(self):
        super().__init__()
        self.results = []

    def test_feature(self):
        # Test implementation
        return True/False

def run_[module]_tests():
    suite = Test[Name]()
    # Run tests and aggregate results
    return passed, failed
```

## Performance Targets

### Latency

- P50: < 50ms
- P95: < 150ms
- P99: < 300ms

### Throughput

- Single agent: > 100 RPS
- Concurrent: > 50 RPS

### Success Rate

- Overall: ≥ 95%
- Integration: ≥ 90%

## Common Issues

### Python not found

```bash
# Install Python 3.8+
# OR set PYTHON_EXE environment variable
export PYTHON_EXE=/path/to/python3
```

### Permission denied (Unix)

```bash
chmod +x run_tests.sh
./run_tests.sh
```

### Test timeout

```bash
# Run individual tests instead
./run_tests.sh tier1
./run_tests.sh tier2
```

### Module not found

```bash
# Install dependencies
pip install -r requirements.txt
# OR manually
pip install pytest pyyaml
```

## Test Results

Results are printed to console and can be saved:

```powershell
# Windows - save to file
.\run_tests.ps1 -ReportDir ./my-results
```

```bash
# Unix - save to file
REPORT_DIR=./my-results ./run_tests.sh all
```

## Agent Coverage

**Tier 1** (Foundational - 5 agents)

- APEX, CIPHER, ARCHITECT, AXIOM, VELOCITY

**Tier 2** (Specialists - 12 agents)

- QUANTUM, TENSOR, FORTRESS, NEURAL, CRYPTO, FLUX
- PRISM, SYNAPSE, CORE, HELIX, VANGUARD, ECLIPSE

**Tier 3** (Innovators - 2 agents)

- NEXUS, GENESIS

**Tier 4** (Meta - 1 agent)

- OMNISCIENT

**Tiers 5-8** (Domain Specialists - 20 agents)

- ATLAS, FORGE, SENTRY, VERTEX, STREAM, PHOTON
- LATTICE, MORPH, PHANTOM, ORBIT, CANVAS, LINGUA
- SCRIBE, MENTOR, BRIDGE, AEGIS, LEDGER, PULSE, ARBITER, ORACLE

## Memory System Tests

### Sub-Linear Techniques

- Bloom Filter (O(1))
- LSH Index (O(1))
- HNSW Graph (O(log n))
- Count-Min Sketch (O(1))
- Cuckoo Filter (O(1))

### ReMem Phases

1. RETRIEVE - Fetch relevant experiences
2. THINK - Augment context
3. ACT - Execute with enhanced context
4. REFLECT - Evaluate results
5. EVOLVE - Update fitness scores

## CI/CD Integration

GitHub Actions automatically:

- Runs tests on push/PR to main/develop
- Tests on multiple OS (Ubuntu, Windows, macOS)
- Tests multiple Python versions (3.9, 3.11)
- Runs linting and formatting checks
- Collects code coverage
- Generates test reports
- Comments with results on PRs

## Developer Workflow

1. **Before coding:**

   ```bash
   ./run_tests.sh all
   ```

2. **During development:**

   ```bash
   # Test only relevant tier
   ./run_tests.sh tier1
   ```

3. **Before committing:**

   ```bash
   # Full suite
   ./run_tests.sh all

   # Lint check
   python -m flake8 tests/
   ```

4. **After committing:**
   - GitHub Actions automatically runs full suite
   - Review results in Actions tab
   - Check PR comments for summary

## Performance Profiling

To profile specific test:

```python
import cProfile
import pstats

pr = cProfile.Profile()
pr.enable()
# Run test code here
pr.disable()

ps = pstats.Stats(pr)
ps.sort_stats('cumulative')
ps.print_stats(10)  # Top 10 functions
```

## Debugging Tests

### Enable verbose logging

```bash
./run_tests.sh all --verbose
```

### Run specific test class

```bash
python -m pytest tests/integration/test_agent_invocation.py::TestAgentInvocation::test_foundational_agents -v
```

### Print intermediate results

```bash
# Edit test file and add print statements
print(f"[DEBUG] Result: {result}")
```

## Documentation Links

- **Full Guide:** [TESTING_GUIDE.md](TESTING_GUIDE.md)
- **Completion Report:** [TASK_5_COMPLETION.md](TASK_5_COMPLETION.md)
- **Agent Reference:** [docs/user-guide/agent-reference.md](docs/user-guide/agent-reference.md)
- **Architecture:** [docs/developer-guide/architecture.md](docs/developer-guide/architecture.md)

## Support

- **GitHub Issues:** Report test failures
- **GitHub Discussions:** Ask questions
- **Pull Requests:** Submit improvements
- **Email:** See repository contacts

---

**Last Updated:** January 2024  
**Framework:** Elite Agent Collective v2.0  
**Python:** 3.8+  
**Status:** Production Ready ✅
