# Adding New Agents

This guide explains how to create and add new specialized agents to the Elite Agent Collective.

## Overview

Adding a new agent involves:

1. Defining the agent's specialty and capabilities
2. Creating the agent definition file
3. Adding to the main instructions file
4. Creating VS Code prompt file (optional)
5. Adding documentation
6. Testing the agent

---

## Step 1: Define the Agent

Before creating the agent, answer these questions:

### Specialty
- What specific domain does this agent specialize in?
- Is it sufficiently different from existing agents?
- What unique expertise does it bring?

### Tier Placement
Determine which tier the agent belongs to:

| Tier | Description | Criteria |
|------|-------------|----------|
| 1 | Foundational | Core engineering disciplines |
| 2 | Specialists | Deep domain expertise |
| 3 | Innovators | Cross-domain synthesis |
| 4 | Meta | System orchestration |
| 5 | Domain | Infrastructure/data domains |
| 6 | Emerging Tech | Cutting-edge technology |
| 7 | Human-Centric | Developer experience |
| 8 | Enterprise | Business/compliance |

### Collaboration
- Which existing agents should this agent collaborate with?
- Are there natural handoff patterns?

---

## Step 2: Create Agent Definition

### Template

```markdown
#### @AGENT_NAME (ID) - Specialty Title

**Primary Function:** One-sentence description of what this agent does.

**Philosophy:** *"A guiding principle that shapes the agent's approach."*

**Invoke:** `@AGENT_NAME [task]`

**Capabilities:**
- Capability 1: Specific skill or knowledge area
- Capability 2: Another specific skill
- Capability 3: Technologies, tools, or frameworks
- Capability 4: Methodologies or approaches
- Capability 5: Additional expertise

**Methodology:** (Optional)
1. STEP_1 → Description
2. STEP_2 → Description
3. STEP_3 → Description

**Decision Matrix:** (Optional)
| Scenario | Approach | Trade-off |
|----------|----------|-----------|
| Case 1 | Recommended | Consideration |
| Case 2 | Alternative | Consideration |

**Example invocations:**
\`\`\`
@AGENT_NAME specific task example 1
@AGENT_NAME specific task example 2
@AGENT_NAME specific task example 3
\`\`\`

**Collaborates well with:** @AGENT1, @AGENT2, @AGENT3
```

### Naming Guidelines

- Use UPPERCASE for agent names
- Keep names short (6-10 characters)
- Make names memorable and descriptive
- Avoid names that conflict with existing agents

### Philosophy Guidelines

- Should be inspiring and memorable
- Captures the essence of the agent's approach
- Uses quotation format: *"..."*

### Capability Guidelines

- Be specific and actionable
- Include tools, technologies, and frameworks
- List 5-10 capabilities
- Use parallel structure

---

## Step 3: Add to Main Instructions

Add the agent definition to `.github/copilot-instructions.md`:

1. Find the appropriate tier section
2. Add the agent definition in ID order
3. Update the Agent Registry table
4. Add to auto-activation rules (if applicable)

### Registry Table Entry

```markdown
| ID | **AGENT_NAME** | Specialty Description | `@AGENT_NAME` |
```

### Auto-Activation (Optional)

If the agent should auto-activate on certain contexts:

```markdown
- **Context trigger** → @AGENT_NAME
```

---

## Step 4: Create VS Code Prompt File (Optional)

For VS Code integration, create a standalone prompt file:

### File Location
```
vscode-prompts/agents/AGENT_NAME.instructions.md
```

### Template

```markdown
---
mode: 'agent'
tools: ['codebase']
description: 'Brief one-line description of the agent'
---

# @AGENT_NAME - Specialty Title

You are **@AGENT_NAME**, the [specialty] agent of the Elite Agent Collective.

## Philosophy
*"The agent's guiding principle."*

## Primary Function
Detailed description of the agent's purpose and expertise.

## Capabilities
- Capability 1
- Capability 2
- ...

## Methodology
When approaching problems:
1. Step 1
2. Step 2
3. Step 3

## Response Format
- Provide clear, actionable responses
- Include code examples when relevant
- Explain reasoning and trade-offs
- Suggest next steps or related agents

## Collaboration
You collaborate effectively with:
- @AGENT1 for [purpose]
- @AGENT2 for [purpose]
```

---

## Step 5: Add Documentation

### Agent Reference Entry

Add to `docs/user-guide/agent-reference.md`:

```markdown
### @AGENT_NAME - Specialty Title

**ID:** XX | **Specialty:** Description

**Philosophy:** *"..."*

**When to use:**
- Use case 1
- Use case 2

**Capabilities:**
- Capability 1
- Capability 2

**Example invocations:**
\`\`\`
@AGENT_NAME example 1
@AGENT_NAME example 2
\`\`\`

**Collaborates well with:** @AGENT1, @AGENT2
```

### Quick Reference Table

Update the quick reference table in `agent-reference.md`:

```markdown
| XX | @AGENT_NAME | X | Specialty Description |
```

---

## Step 6: Test the Agent

### Local Testing

1. Copy updated instructions to your local setup:
   ```bash
   cp .github/copilot-instructions.md ~/.github/
   ```

2. Test basic invocation:
   ```
   @AGENT_NAME help
   ```

3. Test specific capabilities:
   ```
   @AGENT_NAME [specific task]
   ```

4. Test collaboration:
   ```
   @AGENT_NAME @OTHER_AGENT [combined task]
   ```

### Test Checklist

- [ ] Agent responds to basic help request
- [ ] Capabilities match definition
- [ ] Philosophy is reflected in responses
- [ ] Collaboration patterns work
- [ ] Auto-activation triggers (if defined)
- [ ] Examples in documentation work

---

## Example: Adding a New Agent

Let's walk through adding a hypothetical `@QUANTUM_SAFE` agent for post-quantum cryptography:

### Step 1: Define

- **Specialty**: Post-quantum cryptography transition
- **Tier**: 6 (Emerging Tech)
- **ID**: Could be 41 (next available)
- **Collaborates with**: @CIPHER, @QUANTUM

### Step 2: Agent Definition

```markdown
#### @QUANTUM_SAFE (41) - Post-Quantum Cryptography Transition

**Primary Function:** Guide organizations through the transition to quantum-resistant cryptographic systems.

**Philosophy:** *"Prepare today for the quantum threats of tomorrow—cryptographic agility is survival."*

**Invoke:** `@QUANTUM_SAFE [task]`

**Capabilities:**
- Post-quantum algorithm selection (CRYSTALS-Kyber, CRYSTALS-Dilithium, SPHINCS+)
- Cryptographic inventory assessment
- Hybrid classical-quantum schemes
- NIST PQC standards implementation
- Migration planning and timelines
- Backward compatibility strategies

**Example invocations:**
\`\`\`
@QUANTUM_SAFE assess our cryptographic inventory for PQC readiness
@QUANTUM_SAFE implement hybrid TLS with Kyber
@QUANTUM_SAFE create a PQC migration roadmap
\`\`\`

**Collaborates well with:** @CIPHER, @QUANTUM, @AEGIS
```

### Step 3: Add to Instructions

Add to the Tier 6 section in the main instructions file.

### Step 4: Create VS Code File

Create `vscode-prompts/agents/QUANTUM_SAFE.instructions.md`.

### Step 5: Document

Add to agent reference documentation.

### Step 6: Test

Verify the agent works as expected.

---

## Guidelines for Quality Agents

### Do

- ✅ Focus on a specific, well-defined domain
- ✅ Include practical, actionable capabilities
- ✅ Provide clear examples
- ✅ Define meaningful collaboration patterns
- ✅ Create a memorable philosophy

### Don't

- ❌ Overlap significantly with existing agents
- ❌ Be too broad or generic
- ❌ List vague capabilities
- ❌ Skip documentation
- ❌ Forget to test

---

## Submitting Your Agent

1. Create a branch for your agent
2. Add all required files
3. Test thoroughly
4. Submit a pull request
5. Respond to review feedback

See [Contributing Guide](contributing.md) for the full PR process.

---

## Questions?

If you have questions about adding agents:

- Check existing [agent definitions](../../.github/copilot-instructions.md)
- Review the [architecture documentation](architecture.md)
- Open a [discussion](https://github.com/iamthegreatdestroyer/elite-agent-collective/discussions)
