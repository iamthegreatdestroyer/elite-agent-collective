# Multi-Agent Collaboration

Learn how to combine multiple agents for complex problems that require diverse expertise.

## Why Use Multiple Agents?

Complex software engineering challenges often span multiple domains:

- **System design** requires architecture + security + performance considerations
- **Full-stack development** needs frontend + backend + infrastructure expertise
- **ML deployment** combines data science + DevOps + optimization skills

By invoking multiple agents, you get comprehensive solutions that consider all relevant perspectives.

---

## Multi-Agent Invocation Syntax

### Simultaneous Invocation

```
@AGENT1 @AGENT2 [your request]
```

### Example

```
@APEX @ARCHITECT design a distributed caching system
```

Both agents contribute their expertise to the solution.

---

## Agent Collaboration Patterns

### 1. Design + Implementation

Start with architecture, then implement:

```
@ARCHITECT design a notification service architecture
```

Then:

```
@APEX implement the notification service based on the design above
```

### 2. Implementation + Testing

Build and verify:

```
@APEX implement a rate limiter
```

Then:

```
@ECLIPSE write comprehensive tests for this rate limiter
```

### 3. Implementation + Security

Build securely:

```
@APEX @CIPHER implement a user authentication system
```

Or sequentially:

```
@APEX implement password hashing
@CIPHER review the implementation for security vulnerabilities
```

### 4. Design + Documentation

Document as you design:

```
@ARCHITECT @SCRIBE design and document a REST API
```

---

## Common Multi-Agent Combinations

### Architecture & Development

| Combination | Use Case |
|-------------|----------|
| @APEX @ARCHITECT | System design with implementation details |
| @APEX @VELOCITY | Performance-optimized implementations |
| @APEX @ECLIPSE | Test-driven development |

### Security & Development

| Combination | Use Case |
|-------------|----------|
| @CIPHER @FORTRESS | Comprehensive security audit |
| @APEX @CIPHER | Secure code implementation |
| @CRYPTO @CIPHER | Blockchain security |

### ML & Data

| Combination | Use Case |
|-------------|----------|
| @TENSOR @PRISM | ML with statistical rigor |
| @TENSOR @VELOCITY | Optimized ML inference |
| @NEURAL @TENSOR | Advanced AI architectures |

### Infrastructure

| Combination | Use Case |
|-------------|----------|
| @FLUX @ATLAS | Cloud-native DevOps |
| @FLUX @SENTRY | Deployment with observability |
| @FLUX @FORTRESS | Secure infrastructure |

### Innovation

| Combination | Use Case |
|-------------|----------|
| @NEXUS @GENESIS | Breakthrough solutions |
| @NEXUS @AXIOM | Mathematically grounded innovation |
| @GENESIS @NEURAL | Novel AI approaches |

---

## Sequential vs Parallel Collaboration

### Sequential (Recommended for Complex Tasks)

Step-by-step collaboration:

```
Step 1: @ARCHITECT design the system architecture

Step 2: @APEX implement the core components

Step 3: @ECLIPSE create the test suite

Step 4: @CIPHER perform security review

Step 5: @SCRIBE generate documentation
```

### Parallel (For Broad Analysis)

Multiple perspectives at once:

```
@APEX @ARCHITECT @VELOCITY analyze this system for improvements
```

---

## The Agent Collaboration Matrix

Primary agents naturally collaborate with certain specialists:

| Primary Agent | Natural Collaborators |
|--------------|----------------------|
| @APEX | @ARCHITECT, @VELOCITY, @ECLIPSE |
| @CIPHER | @AXIOM, @FORTRESS, @QUANTUM |
| @ARCHITECT | @APEX, @FLUX, @SYNAPSE |
| @TENSOR | @AXIOM, @PRISM, @VELOCITY |
| @FLUX | @ATLAS, @SENTRY, @FORTRESS |
| @NEXUS | ALL AGENTS |
| @GENESIS | @AXIOM, @NEXUS, @NEURAL |

---

## Example Workflows

### Complete Feature Development

```
1. @ARCHITECT design the feature architecture
   - Component breakdown
   - Data flow
   - API contracts

2. @APEX implement the backend
   - Core logic
   - Database operations
   - API endpoints

3. @CANVAS implement the frontend
   - UI components
   - User experience

4. @ECLIPSE write tests
   - Unit tests
   - Integration tests

5. @CIPHER security review
   - Vulnerability scan
   - Security best practices

6. @FLUX deploy
   - CI/CD pipeline
   - Infrastructure

7. @SCRIBE document
   - API documentation
   - User guides
```

### Security-First Development

```
1. @CIPHER @FORTRESS threat model the system

2. @ARCHITECT design with security constraints

3. @APEX implement with security patterns

4. @ECLIPSE write security tests

5. @CIPHER final security audit
```

### ML Model Deployment

```
1. @TENSOR design the model architecture

2. @PRISM validate with statistical analysis

3. @VELOCITY optimize for inference

4. @FLUX containerize and deploy

5. @SENTRY add monitoring and observability
```

---

## The @OMNISCIENT Meta-Agent

For complex multi-agent coordination, use @OMNISCIENT:

```
@OMNISCIENT coordinate an analysis of this system's architecture, security, and performance
```

OMNISCIENT will:
- Identify relevant agents
- Orchestrate their collaboration
- Synthesize their insights
- Present a unified analysis

---

## Tips for Effective Multi-Agent Collaboration

### 1. Start with the Right Scope

```
@ARCHITECT give me a high-level design first, then we'll dive deeper
```

### 2. Chain Refinements

```
@APEX implement the basic version
@VELOCITY now optimize it for performance
@CIPHER now harden it for security
```

### 3. Request Explicit Handoffs

```
@ARCHITECT design this, then tell me which agents should implement each component
```

### 4. Use Summary Requests

```
@OMNISCIENT summarize the key considerations from @ARCHITECT, @CIPHER, and @VELOCITY for this system
```

### 5. Resolve Conflicts Explicitly

When agents have different recommendations:

```
@ARCHITECT recommends microservices, @VELOCITY suggests a monolith for performance. 
@OMNISCIENT help me decide based on my specific constraints
```

---

## Limitations

- **Response Length**: Multi-agent responses may be longer
- **Context Windows**: Very complex collaborations may exceed context limits
- **Focus**: Sometimes a single expert is better than multiple generalists

---

## Next Steps

- See the complete [Agent Reference](agent-reference.md)
- Read [Best Practices](best-practices.md) for advanced usage
- Explore [Architecture Documentation](../developer-guide/architecture.md)
