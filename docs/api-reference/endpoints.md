# API Endpoints

This document describes the interaction endpoints for the Elite Agent Collective.

## Overview

The Elite Agent Collective operates through GitHub Copilot's chat interface. While there are no traditional REST API endpoints, this document describes the interaction model and invocation patterns.

---

## Invocation Interface

### Agent Invocation

**Pattern:** `@AGENT_NAME [request]`

**Description:** Invokes a specific agent with a request.

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| AGENT_NAME | string | Yes | The agent's codename (e.g., APEX, CIPHER) |
| request | string | Yes | The task or question for the agent |

**Example:**
```
@APEX implement a binary search algorithm in Python
```

---

### Multi-Agent Invocation

**Pattern:** `@AGENT1 @AGENT2 [request]`

**Description:** Invokes multiple agents simultaneously for collaborative responses.

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| AGENT1, AGENT2, ... | string | Yes | One or more agent codenames |
| request | string | Yes | The task or question |

**Example:**
```
@APEX @ARCHITECT design a scalable caching system
```

---

## Agent Registry

### Available Agents

| ID | Codename | Tier | Invocation |
|----|----------|------|------------|
| 01 | APEX | 1 | `@APEX` |
| 02 | CIPHER | 1 | `@CIPHER` |
| 03 | ARCHITECT | 1 | `@ARCHITECT` |
| 04 | AXIOM | 1 | `@AXIOM` |
| 05 | VELOCITY | 1 | `@VELOCITY` |
| 06 | QUANTUM | 2 | `@QUANTUM` |
| 07 | TENSOR | 2 | `@TENSOR` |
| 08 | FORTRESS | 2 | `@FORTRESS` |
| 09 | NEURAL | 2 | `@NEURAL` |
| 10 | CRYPTO | 2 | `@CRYPTO` |
| 11 | FLUX | 2 | `@FLUX` |
| 12 | PRISM | 2 | `@PRISM` |
| 13 | SYNAPSE | 2 | `@SYNAPSE` |
| 14 | CORE | 2 | `@CORE` |
| 15 | HELIX | 2 | `@HELIX` |
| 16 | VANGUARD | 2 | `@VANGUARD` |
| 17 | ECLIPSE | 2 | `@ECLIPSE` |
| 18 | NEXUS | 3 | `@NEXUS` |
| 19 | GENESIS | 3 | `@GENESIS` |
| 20 | OMNISCIENT | 4 | `@OMNISCIENT` |
| 21 | ATLAS | 5 | `@ATLAS` |
| 22 | FORGE | 5 | `@FORGE` |
| 23 | SENTRY | 5 | `@SENTRY` |
| 24 | VERTEX | 5 | `@VERTEX` |
| 25 | STREAM | 5 | `@STREAM` |
| 26 | PHOTON | 6 | `@PHOTON` |
| 27 | LATTICE | 6 | `@LATTICE` |
| 28 | MORPH | 6 | `@MORPH` |
| 29 | PHANTOM | 6 | `@PHANTOM` |
| 30 | ORBIT | 6 | `@ORBIT` |
| 31 | CANVAS | 7 | `@CANVAS` |
| 32 | LINGUA | 7 | `@LINGUA` |
| 33 | SCRIBE | 7 | `@SCRIBE` |
| 34 | MENTOR | 7 | `@MENTOR` |
| 35 | BRIDGE | 7 | `@BRIDGE` |
| 36 | AEGIS | 8 | `@AEGIS` |
| 37 | LEDGER | 8 | `@LEDGER` |
| 38 | PULSE | 8 | `@PULSE` |
| 39 | ARBITER | 8 | `@ARBITER` |
| 40 | ORACLE | 8 | `@ORACLE` |

---

## Context Endpoints

The system automatically captures context from:

### Current File Context

When in an IDE, the active file is automatically included in the context.

```
# Context automatically includes current file
@APEX optimize this function
```

### Selection Context

Selected code is prioritized in the context.

```
# Select code, then invoke
@ECLIPSE write tests for the selected code
```

### Repository Context

With appropriate permissions, repository structure informs responses.

---

## Response Endpoints

### Standard Response

Agents return responses in Copilot Chat format with:

- **Explanation**: Description of the solution
- **Code**: Implementation (when applicable)
- **References**: Related resources or next steps

### Code Response

For code-related requests:

```markdown
Here's the implementation:

\`\`\`python
def binary_search(arr, target):
    left, right = 0, len(arr) - 1
    while left <= right:
        mid = (left + right) // 2
        if arr[mid] == target:
            return mid
        elif arr[mid] < target:
            left = mid + 1
        else:
            right = mid - 1
    return -1
\`\`\`

This implementation has O(log n) time complexity...
```

---

## Error Handling

### Unrecognized Agent

If an agent name is not recognized, the system falls back to general Copilot behavior.

### Ambiguous Request

If a request is unclear, agents ask clarifying questions.

### Out-of-Scope Request

If a request is outside an agent's expertise, the agent may:
1. Provide a partial response
2. Suggest a more appropriate agent
3. Indicate the limitation

---

## Rate Limits

Agent invocations are subject to GitHub Copilot rate limits:

- Standard Copilot subscription limits apply
- No additional limits from the Elite Agent Collective
- Complex multi-agent requests count as single interactions

---

## Integration Points

### GitHub Copilot Chat

Primary integration through Copilot Chat in:
- VS Code
- JetBrains IDEs
- GitHub.com
- GitHub Mobile

### Instructions File Integration

The system integrates via:
- `.github/copilot-instructions.md` (repository or global)
- VS Code user prompts (per-agent files)

---

## Versioning

Current version: **2.0**

The API is versioned through the instructions file. Updates to agent capabilities are reflected in the instructions file version.

---

## Related Documentation

- [Request Format](request-format.md)
- [Response Format](response-format.md)
- [Agent Reference](../user-guide/agent-reference.md)
