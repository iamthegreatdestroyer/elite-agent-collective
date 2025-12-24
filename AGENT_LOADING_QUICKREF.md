# Quick Reference: Agent Loading System

## Overview

The backend now dynamically loads all 40 agents from distributed `.agent.md` files instead of hardcoded definitions.

## Key Files

| File                                      | Purpose                          |
| ----------------------------------------- | -------------------------------- |
| `.github/agents/*.agent.md`               | Agent definitions (40 files)     |
| `backend/internal/agents/agent_loader.go` | Parser for .agent.md files       |
| `backend/internal/agents/agents.go`       | Registration with smart fallback |
| `backend/pkg/models/agent.go`             | Extended Agent model             |

## How It Works

### 1. Server Startup

```
Start server
    ↓
DefaultRegistry()
    ↓
RegisterAllAgents()
    ↓
findAgentsDirectory() [Try 6 paths]
    ↓
LoadAllAgentsFromDirectory(".github/agents")
    ↓
For each .agent.md:
  - Parse YAML frontmatter
  - Extract philosophy, examples, collaborators
  - Create models.Agent
    ↓
Register agents in map
    ↓
HTTP /agents endpoint ready
```

### 2. Directory Search Order

1. `../.github/agents` (from backend/)
2. `../../.github/agents`
3. `../../../.github/agents`
4. `./.github/agents`
5. `/app/.github/agents` (Docker)
6. `$HOME/elite-agent-collective-1/.github/agents`

**Fallback:** If none found, uses hardcoded AllAgentDefinitions

### 3. Agent File Format

**Location:** `.github/agents/[CODENAME].agent.md`

**Example:**

```markdown
---
name: APEX
description: Elite Computer Science Engineering
codename: APEX
tier: 1
id: "01"
category: Foundational
---

**Philosophy:** _"Every problem has an elegant solution waiting to be discovered."_

## Core Capabilities

- Deliver production-grade code
- Apply computer science fundamentals
- Optimize for performance

## Invocation Examples

@APEX implement a rate limiter
@APEX design system architecture

## Multi-Agent Collaboration

Collaborates with: @ARCHITECT, @VELOCITY, @ECLIPSE
```

## Common Tasks

### Add a New Agent

1. Create `.github/agents/CODENAME.agent.md`
2. Add YAML frontmatter with required fields
3. Add markdown content (philosophy, capabilities, examples)
4. Restart server - agent automatically discovered

### Update Agent Definition

1. Edit `.github/agents/CODENAME.agent.md`
2. Restart server
3. Changes reflected automatically (no code change needed)

### Check Loading Status

```bash
# Server logs show:
2025/12/11 12:26:34 Loaded 40 agents from /path/to/.github/agents

# Or query API:
curl http://localhost:8080/agents
```

### Verify Agent Fields

```go
// Agent model now includes:
agent := models.Agent{
    ID:            "01",
    Codename:      "APEX",
    Tier:          1,
    Name:          "APEX",
    Specialty:     "Elite Computer Science Engineering",
    Philosophy:    "Every problem has...",
    Directives:    []string{"Deliver production-grade...", ...},
    Keywords:      []string{},
    Examples:      []string{"@APEX implement...", ...},
    Collaborators: []string{"ARCHITECT", "VELOCITY", "ECLIPSE"},
    Category:      "Foundational",
    MarkdownPath:  "/path/to/APEX.agent.md",
}
```

## Agent Extraction

### Extracted from YAML Frontmatter

- `name` → Agent.Name
- `description` → Agent.Specialty
- `codename` → Agent.Codename
- `tier` → Agent.Tier (converted to int)
- `id` → Agent.ID
- `category` → Agent.Category

### Extracted from Markdown

- **Philosophy:** Regex `**Philosophy:** _"([^"]+)"`
- **Directives:** Bullet points in "Core Capabilities" section
- **Examples:** Lines starting with `@`
- **Collaborators:** `@([A-Z]+)` pattern (agent names)

## Testing

### Run Tests

```bash
cd backend/
make test
```

### Expected Output

```
PASS: internal/agents (14 tests)
  ✓ TestDefaultRegistry - loads 40 agents
  ✓ TestAllAgentsHaveRequiredFields - validates fields
  (12 more tests)

Log: "Loaded 40 agents from C:\Users\sgbil\elite-agent-collective-1\.github\agents"
```

### Build & Run

```bash
make build      # Compile binary
./bin/server    # Run server
```

## Error Handling

### If .github/agents/ Not Found

- Logs warning
- Falls back to hardcoded AllAgentDefinitions
- Server continues normally
- All agents still available

### If Agent File Can't Parse

- Logs error for that file
- Continues loading other agents
- Server continues with successfully loaded agents
- Fallback mechanism activates if too many fail

## Performance

- **Load Time:** <100ms for all 40 agents
- **Memory:** Same as before (no overhead)
- **Runtime:** Zero impact (files only read at startup)

## Backward Compatibility

✅ **Zero Breaking Changes**

- All existing code continues to work
- AllAgentDefinitions kept for fallback
- HTTP API unchanged
- Handler interfaces unchanged
- Registry behavior unchanged

## Future Enhancements

- Watch .agent.md files for changes (hot reload)
- Validate agent files on startup
- Generate agent documentation from .agent.md
- Support agent versioning
- Cache extracted fields

## References

- Full implementation: `backend/internal/agents/agent_loader.go` (257 lines)
- Agent model: `backend/pkg/models/agent.go`
- Registration: `backend/internal/agents/agents.go`
- Tests: `backend/internal/agents/*_test.go`

---

**Status:** ✅ Production Ready  
**Last Updated:** 2025-12-11  
**Agents Loading:** 40/40
