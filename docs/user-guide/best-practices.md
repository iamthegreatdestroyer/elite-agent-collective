# Best Practices

Tips and tricks for getting the most out of the Elite Agent Collective.

## General Principles

### 1. Right Agent for the Right Job

Each agent has specialized expertise. Using the right agent dramatically improves results:

| Task | Best Agent | Why |
|------|-----------|-----|
| Algorithm design | @APEX | Deep CS fundamentals |
| Security review | @CIPHER | Cryptography expertise |
| Cloud architecture | @ATLAS | Multi-cloud knowledge |
| Code optimization | @VELOCITY | Performance focus |
| Test writing | @ECLIPSE | Testing methodologies |

### 2. Provide Rich Context

The more context you provide, the better the response:

❌ **Poor:**
```
@APEX fix this bug
```

✅ **Better:**
```
@APEX fix the race condition in user-service.go line 45
Expected: concurrent writes should be serialized
Actual: data corruption when multiple goroutines write
```

### 3. Iterate and Refine

Complex solutions emerge through iteration:

```
Step 1: @APEX implement a basic solution
Step 2: @VELOCITY optimize for performance  
Step 3: @CIPHER harden for security
Step 4: @ECLIPSE add comprehensive tests
```

---

## Communication Patterns

### Be Specific About Output Format

```
@APEX provide the solution as:
1. Brief explanation
2. Code implementation
3. Example usage
4. Edge cases to consider
```

### Specify Language and Framework

```
@APEX implement in TypeScript using Express.js
@TENSOR use PyTorch, not TensorFlow
@FLUX use GitHub Actions, not Jenkins
```

### State Constraints Clearly

```
@APEX design a solution with:
- O(n log n) time complexity maximum
- O(1) space complexity
- Thread-safe operations
- No external dependencies
```

### Request Explanations

```
@APEX explain your reasoning step by step
@ARCHITECT justify each architectural decision
@CIPHER explain why this is secure
```

---

## Domain-Specific Tips

### For Software Engineering (@APEX, @ARCHITECT)

- Reference design patterns by name
- Specify scalability requirements (users, requests/sec)
- Mention existing architecture constraints
- Include relevant technology stack

```
@ARCHITECT design a notification system
Stack: Node.js, PostgreSQL, Redis, AWS
Requirements: 100k users, 1M notifications/day
Constraints: Must integrate with existing user service
```

### For Security (@CIPHER, @FORTRESS)

- Specify threat model
- Mention compliance requirements (GDPR, SOC2, HIPAA)
- Include attack vectors to consider
- Reference security standards

```
@CIPHER design authentication for a healthcare app
Compliance: HIPAA required
Threats: Consider insider threats and external attackers
Standards: Follow NIST guidelines
```

### For Machine Learning (@TENSOR, @PRISM)

- Describe the data available
- Specify accuracy vs latency tradeoffs
- Mention deployment constraints
- Include evaluation metrics

```
@TENSOR recommend a model for fraud detection
Data: 1M transactions, 0.1% fraud rate
Constraints: <10ms inference, interpretable
Metrics: Optimize for recall, minimize false positives
```

### For DevOps (@FLUX, @ATLAS)

- Specify cloud provider(s)
- Include scaling requirements
- Mention budget constraints
- Reference existing infrastructure

```
@FLUX design CI/CD for a monorepo
Cloud: AWS (existing account)
Scale: 50 engineers, 200 deploys/week
Tools: Prefer GitHub Actions
```

---

## Common Mistakes and Solutions

### Mistake 1: Too Broad a Request

❌ `@APEX build me an e-commerce platform`

✅ Break it down:
```
@ARCHITECT design the e-commerce architecture
@APEX implement the product catalog service
@APEX implement the shopping cart service
@CIPHER implement payment security
```

### Mistake 2: Wrong Agent Selection

❌ `@TENSOR review my Python code style`

✅ `@MENTOR review my Python code style`

### Mistake 3: Missing Context

❌ `@VELOCITY make this faster`

✅ `@VELOCITY optimize this SQL query that takes 5s to return 10k rows from the orders table with 1M total rows`

### Mistake 4: Ignoring Agent Expertise

❌ Asking @APEX for compliance advice

✅ `@AEGIS advise on GDPR compliance for this data processing`

---

## Maximizing Agent Collaboration

### Sequential Handoffs

Create a clear workflow:

```
@ARCHITECT → Design
@APEX → Implement
@ECLIPSE → Test
@CIPHER → Security review
@SCRIBE → Document
@FLUX → Deploy
```

### Cross-Domain Problems

Use @NEXUS for problems spanning multiple domains:

```
@NEXUS help me combine ML predictions with rule-based business logic
```

### When Stuck

Use @OMNISCIENT to coordinate:

```
@OMNISCIENT I need help with this complex distributed systems problem - which agents should I consult?
```

---

## Productivity Tips

### 1. Save Common Prompts

Create templates for frequent tasks:

```
Code Review Template:
@MENTOR review this code for:
- Correctness
- Performance
- Security
- Maintainability
- Test coverage
```

### 2. Use Agent Shortcuts

Develop muscle memory for common agents:

| Task | Quick Command |
|------|--------------|
| Quick implementation | `@APEX ` |
| Review security | `@CIPHER ` |
| Write tests | `@ECLIPSE ` |
| Document code | `@SCRIBE ` |

### 3. Leverage Auto-Activation

Agents auto-activate based on file type:

- Working in `.py` files → @APEX ready
- Working in `Dockerfile` → @FLUX ready
- Working in `.tf` files → @ATLAS ready
- Working in test files → @ECLIPSE ready

### 4. Chain of Thought Prompting

Ask agents to explain their reasoning:

```
@APEX explain your reasoning step by step, then provide the solution
```

This leads to better solutions and helps you learn.

---

## Quality Checklist

Before finalizing agent-generated solutions:

- [ ] Code compiles/runs without errors
- [ ] Edge cases are handled
- [ ] Error handling is present
- [ ] Security considerations addressed
- [ ] Performance is acceptable
- [ ] Tests are included or planned
- [ ] Documentation is clear
- [ ] Code follows project conventions

---

## Learning from Agents

Agents are teachers as well as assistants:

### Ask "Why" Questions

```
@APEX why is this approach better than using a simple array?
@CIPHER why is bcrypt preferred over SHA-256 for passwords?
```

### Request Comparisons

```
@ARCHITECT compare microservices vs monolith for this use case
@TENSOR compare CNN vs Transformer for image classification
```

### Explore Alternatives

```
@APEX show me three different ways to implement this
@ARCHITECT what are the tradeoffs of each approach?
```

---

## Anti-Patterns to Avoid

1. **Over-relying on a single agent** - Use specialists for specialized tasks
2. **Ignoring suggestions** - Agents provide best practices worth considering
3. **Not providing feedback** - Tell agents when responses don't meet needs
4. **Skipping review** - Always review and understand generated code
5. **Copying without understanding** - Use agents to learn, not just copy

---

## Next Steps

- Review the complete [Agent Reference](agent-reference.md)
- Learn about [Architecture](../developer-guide/architecture.md)
- Check [Troubleshooting](../troubleshooting/common-issues.md) if you encounter issues
