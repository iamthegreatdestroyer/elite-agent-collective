# Quick Start

Get up and running with the Elite Agent Collective in just a few minutes.

## Your First Agent Invocation

After [installation](installation.md), open Copilot Chat and try:

```
@APEX help me understand your capabilities
```

APEX will respond with information about its software engineering expertise.

---

## Basic Agent Invocation Syntax

```
@AGENT_NAME [your request here]
```

### Examples

```
@APEX implement a binary search algorithm
@CIPHER design a JWT authentication system
@TENSOR recommend a neural network architecture for image classification
@FLUX help me set up a Kubernetes deployment
```

---

## Finding the Right Agent

Use this quick reference to find the best agent for your task:

| Problem Type | Recommended Agent(s) |
|--------------|---------------------|
| Algorithm design & code | @APEX, @VELOCITY |
| Security review | @CIPHER, @FORTRESS |
| Cloud architecture | @ATLAS, @FLUX |
| ML/AI development | @TENSOR, @NEURAL |
| API design | @SYNAPSE |
| Documentation | @SCRIBE |
| Code review | @MENTOR |
| Testing | @ECLIPSE |
| Performance optimization | @VELOCITY |
| Database design | @VERTEX |
| DevOps/CI-CD | @FLUX |

---

## Multi-Agent Collaboration

For complex problems, combine multiple agents:

```
@APEX @ARCHITECT design a scalable microservices system
```

```
@CIPHER @FORTRESS perform a comprehensive security audit
```

```
@TENSOR @VELOCITY optimize this ML inference pipeline
```

---

## Context-Aware Responses

Agents automatically consider your current context:

- **Current file**: Agents see the file you're working on
- **Selected code**: Highlight code for focused analysis
- **Project structure**: Agents understand your project layout

### Example Workflow

1. Open a Python file
2. Highlight a function
3. Ask: `@APEX optimize this function for performance`
4. APEX analyzes and suggests improvements

---

## Agent Categories Quick Reference

### Tier 1: Foundational (5 agents)
Core engineering capabilities: @APEX, @CIPHER, @ARCHITECT, @AXIOM, @VELOCITY

### Tier 2: Specialists (12 agents)
Deep domain expertise: @QUANTUM, @TENSOR, @FORTRESS, @NEURAL, @CRYPTO, @FLUX, @PRISM, @SYNAPSE, @CORE, @HELIX, @VANGUARD, @ECLIPSE

### Tier 3: Innovators (2 agents)
Cross-domain innovation: @NEXUS, @GENESIS

### Tier 4: Meta (1 agent)
System coordination: @OMNISCIENT

### Tier 5-8: Domain Specialists (20 agents)
Specialized domains including cloud (@ATLAS), build systems (@FORGE), observability (@SENTRY), and more.

---

## Common First Tasks

### 1. Get a Code Review

```
@MENTOR review this code and suggest improvements
```

### 2. Design an API

```
@SYNAPSE design a RESTful API for a user management system
```

### 3. Write Tests

```
@ECLIPSE write unit tests for this function
```

### 4. Optimize Performance

```
@VELOCITY analyze and optimize this database query
```

### 5. Document Code

```
@SCRIBE generate API documentation for this module
```

---

## Tips for Better Results

1. **Be specific**: "Design a rate limiter with sliding window" > "Make a rate limiter"

2. **Provide context**: Include relevant constraints, requirements, or preferences

3. **Use the right agent**: Each agent has specialized knowledge

4. **Combine agents**: Complex problems benefit from multiple perspectives

5. **Iterate**: Ask follow-up questions to refine solutions

---

## Next Steps

- Explore the [Agent Reference](../user-guide/agent-reference.md) for all 40 agents
- Learn about [Multi-Agent Collaboration](../user-guide/multi-agent-collaboration.md)
- Read [Best Practices](../user-guide/best-practices.md) for advanced usage

---

## Example Session

Here's a complete example workflow:

```
You: @ARCHITECT I need to design a real-time chat application. What architecture would you recommend?

ARCHITECT: [Provides architecture recommendations with components, data flow, and technology choices]

You: @APEX Can you implement the WebSocket handler based on that architecture?

APEX: [Provides implementation code with best practices]

You: @ECLIPSE Write tests for this WebSocket handler

ECLIPSE: [Provides comprehensive test suite]

You: @SCRIBE Document this API

SCRIBE: [Generates API documentation]
```

This demonstrates how agents can work together on a complete feature from design to documentation.
