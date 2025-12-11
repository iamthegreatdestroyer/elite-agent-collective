# Backend Agent Loading System - Quick Reference

## Overview

The Elite Agent Collective backend now supports **dynamic agent loading** from distributed `.agent.md` files. This eliminates the need to modify backend code when adding, updating, or removing agents.

### How It Works

```
.github/agents/*.agent.md
         ↓
LoadAllAgentsFromDirectory()
         ↓
Parse YAML + Extract Markdown
         ↓
Populate models.Agent (12 fields)
         ↓
Register in HTTP routes
         ↓
Agents available via API
```

## Agent File Format

Each agent is defined in a single `.agent.md` file in the `.github/agents/` directory.

### File Naming

```
[CODENAME].agent.md
Example: APEX.agent.md, CIPHER.agent.md, ORACLE.agent.md
```

### File Structure

```yaml
---
name: Agent Full Name
description: Brief description of what agent does
codename: APEX
tier: 1
id: "01"
category: "Foundational"
keywords:
  - keyword1
  - keyword2
  - keyword3
---

# Agent Name

**Philosophy:** _"Quote about the agent's core principle."_

## Primary Function

[Description of what the agent does and its expertise areas]

**Invoke:** `@CODENAME [task description]`

## Capabilities

List of key capabilities:

- Capability 1 description
- Capability 2 description
- Capability 3 description
- More capabilities...

## Core Methodology

Steps or approach the agent takes to solve problems:

1. STEP_ONE → Brief description
2. STEP_TWO → Brief description
3. STEP_THREE → Brief description

## Collaboration Patterns

### Works Best With
- @AGENT1 - brief reason
- @AGENT2 - brief reason
- @AGENT3 - brief reason

### Typical Workflows
[Describe common usage patterns]

## Invocation Examples

```

@CODENAME implement specific feature description
@CODENAME design new architecture pattern
@CODENAME analyze this problem
@CODENAME review code with focus area

````

## Technical Stack

- Languages: [List supported languages]
- Frameworks: [Key frameworks/tools]
- Methodologies: [Key approaches]

## Decision Matrix / Selection Guide

| Scenario | Recommendation | Alternative |
|----------|---|---|
| Use Case 1 | Recommended | Alternative if condition |
| Use Case 2 | Recommended | Alternative if condition |

## Memory & Learning

The agent leverages MNEMONIC memory system to:
- Store successful solutions from past tasks
- Share strategies with tier-mate agents
- Access breakthrough discoveries from other tiers
- Continuously improve through fitness-based evolution

Past experiences inform faster, more informed decisions on similar tasks.

## Integration Notes

- HTTP endpoint: `/agent` (POST with agent codename in body)
- Handler: `handlers.NewBaseAgent(agentDef)` or `handlers.NewApexAgent()` for APEX
- Authentication: OAuth2 via OIDC (if configured)
- Rate limiting: Applied per agent endpoint

## Troubleshooting

### Agent Not Loading
1. Check file exists at `.github/agents/CODENAME.agent.md`
2. Verify YAML frontmatter is valid (---...---)
3. Check YAML field types match schema
4. Run: `go test ./internal/agents/...`

### Fields Not Populating
1. Check markdown section headers match expected patterns
2. Verify regex patterns in agent_loader.go match your markdown
3. View extracted content: Look at test output or add logging

### Registry Issues
1. Verify `findAgentsDirectory()` returns correct path
2. Check permissions on `.github/agents/` directory
3. Ensure YAML dependency available: `gopkg.in/yaml.v3`

## Development Workflow

### Adding a New Agent

1. **Create file** in `.github/agents/`:
   ```bash
   cp .github/agents/APEX.agent.md .github/agents/NEWAGENT.agent.md
````

2. **Update metadata**:

   - Change `codename` (uppercase)
   - Assign unique `id` (must be 01-40)
   - Update `name`, `description`
   - Set correct `tier`

3. **Update content**:

   - Philosophy statement (what drives the agent)
   - Capabilities (what it can do)
   - Methodology (how it works)
   - Examples (how to invoke it)
   - Collaboration patterns

4. **Test loading**:

   ```bash
   cd backend
   go test -v ./internal/agents/...
   ```

5. **Verify in server**:
   ```bash
   go run ./cmd/server
   # Check logs for: "Loaded X agents from ..."
   ```

### Updating an Agent

1. Edit the `.agent.md` file
2. Change content directly (no backend code changes needed)
3. Restart server to reload changes
4. Test via API or Copilot UI

### Removing an Agent

1. Delete the `.agent.md` file
2. Restart server
3. Server will load remaining agents

## Directory Resolution

When backend starts, it searches for `.github/agents/` in this order:

1. `../.github/agents` - Standard relative path
2. `../../.github/agents` - For nested directory structures
3. `../../../.github/agents` - For deeply nested structures
4. `./.github/agents` - Current working directory
5. `/app/.github/agents` - Docker container default
6. `$HOME/elite-agent-collective-1/.github/agents` - Home directory fallback

If directory not found, backend **falls back to hardcoded definitions** and continues running (no crash).

## Testing

### Unit Tests

```bash
cd backend
go test -v ./internal/agents/...
```

### Load All Agents

```bash
cd backend
go run ./cmd/server
# Logs should show: "Loaded 40 agents from C:\...\agents"
```

### Verify Agent Fields

```go
// In tests, verify:
assert.NotEmpty(t, agent.Name)
assert.NotEmpty(t, agent.Philosophy)
assert.NotEmpty(t, agent.Examples)
assert.NotEmpty(t, agent.Collaborators)
```

## Performance Characteristics

| Operation                 | Complexity | Time   |
| ------------------------- | ---------- | ------ |
| Load all 40 agents        | O(n)       | ~287ms |
| Get single agent          | O(1)       | ~0.02s |
| Parse YAML frontmatter    | O(1)       | ~1ms   |
| Extract markdown sections | O(m)       | ~2ms   |
| Register in router        | O(1)       | <1ms   |

**m** = markdown content size

## Future Enhancements

- [ ] Hot reload: Watch file changes without restart
- [ ] Agent validation: Schema validation before registration
- [ ] Markdown parsing: More robust markdown extractors
- [ ] Agent templates: Scaffolding tool for new agents
- [ ] Versioning: Support multiple agent versions
- [ ] Agent marketplace: Dynamic agent installation

## References

- [Agent Loader Implementation](../backend/internal/agents/agent_loader.go)
- [Agent Model Definition](../backend/pkg/models/agent.go)
- [Agent Registry](../backend/internal/agents/registry.go)
- [Handlers](../backend/internal/agents/handlers/)
- [YAML Parser](https://gopkg.in/yaml.v3)

## Support

For issues or questions:

1. Check the [Troubleshooting](#troubleshooting) section
2. Review [Development Workflow](#development-workflow)
3. Examine test files: `backend/internal/agents/*_test.go`
4. Check GitHub Issues for similar problems
