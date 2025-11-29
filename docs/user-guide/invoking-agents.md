# Invoking Agents

Learn how to effectively invoke and communicate with the 40 specialized agents in the Elite Agent Collective.

## Basic Invocation Syntax

The standard syntax for invoking an agent is:

```
@AGENT_NAME [your request]
```

### Examples

```
@APEX implement a red-black tree
@CIPHER explain AES-256 encryption
@TENSOR design a transformer model
```

---

## Agent Invocation Patterns

### Simple Request

Direct question or task:

```
@APEX what is the time complexity of quicksort?
```

### Code Context

Reference the current file or selection:

```
@APEX optimize this function
@ECLIPSE write tests for the selected code
@MENTOR review this implementation
```

### Detailed Specification

Provide comprehensive requirements:

```
@ARCHITECT design a microservices architecture for an e-commerce platform
Requirements:
- Handle 10,000 concurrent users
- 99.9% uptime
- Event-driven communication
- PostgreSQL for persistence
```

### Iterative Refinement

Build on previous responses:

```
@APEX implement a rate limiter

[APEX provides implementation]

@APEX add support for distributed rate limiting with Redis
```

---

## Context Awareness

Agents automatically consider:

### Current File
```
// In user-auth.ts
@CIPHER review this authentication logic
```
CIPHER analyzes the file you're viewing.

### Selected Code
1. Highlight code in your editor
2. Invoke an agent
3. The agent focuses on the selection

### Project Structure
Agents understand your repository layout and can reference other files.

---

## Agent Prefixes and Modifiers

### Urgency Modifiers

```
@APEX quick: give me a simple hash function
@APEX detailed: explain hash tables comprehensively
```

### Output Format

```
@APEX explain binary search as a bullet list
@SCRIBE document this function in JSDoc format
@ARCHITECT describe this system in C4 model format
```

### Language Specification

```
@APEX implement quicksort in Python
@APEX implement quicksort in Rust
@APEX implement quicksort in TypeScript
```

---

## Best Practices for Invocation

### 1. Choose the Right Agent

| Task | Primary Agent | Alternative |
|------|---------------|-------------|
| Algorithm implementation | @APEX | @VELOCITY |
| Security review | @CIPHER | @FORTRESS |
| Architecture design | @ARCHITECT | @ATLAS |
| Machine learning | @TENSOR | @PRISM |
| Testing | @ECLIPSE | @APEX |
| Documentation | @SCRIBE | @VANGUARD |

### 2. Be Specific

❌ Vague:
```
@APEX make this faster
```

✅ Specific:
```
@APEX optimize this database query to reduce execution time from 5s to under 500ms
```

### 3. Provide Context

❌ Missing context:
```
@ARCHITECT design an API
```

✅ With context:
```
@ARCHITECT design a RESTful API for a todo application
- CRUD operations for tasks
- User authentication via JWT
- Rate limiting
- OpenAPI documentation
```

### 4. State Constraints

```
@APEX implement a cache with:
- Maximum 1000 entries
- LRU eviction policy
- Thread-safe operations
- O(1) lookup time
```

---

## Advanced Invocation Techniques

### Chained Requests

Ask for step-by-step solutions:

```
@APEX walk me through implementing a B-tree:
1. First explain the data structure
2. Then show the insert operation
3. Then show the search operation
4. Finally, discuss time complexity
```

### Comparative Analysis

```
@APEX compare Redis vs Memcached for session storage
@ARCHITECT compare microservices vs monolith for this use case
@TENSOR compare CNN vs ViT for image classification
```

### Code Review Focus

```
@MENTOR review this code focusing on:
- Error handling
- Code style
- Performance
- Security
```

### Follow-up Questions

```
@APEX what are the edge cases for this algorithm?
@APEX how would this handle concurrent access?
@APEX what tests should I write for this?
```

---

## Agent Response Types

Agents can provide various response types:

### Code Solutions
```
@APEX implement a priority queue
```
Returns working code with explanations.

### Explanations
```
@AXIOM explain the CAP theorem
```
Returns conceptual explanations.

### Reviews
```
@MENTOR review this pull request
```
Returns feedback and suggestions.

### Designs
```
@ARCHITECT design a notification system
```
Returns architecture diagrams and specifications.

### Documentation
```
@SCRIBE document this API
```
Returns formatted documentation.

---

## Common Mistakes to Avoid

### 1. Wrong Agent for Task

❌ `@TENSOR review my Python code style`
✅ `@MENTOR review my Python code style`

### 2. Insufficient Context

❌ `@APEX fix the bug`
✅ `@APEX fix the null pointer exception in the getUserById function`

### 3. Overly Complex Requests

Break large requests into smaller parts:

❌ `@APEX build me a complete authentication system with OAuth, MFA, session management, password reset, email verification, and audit logging`

✅ Start with:
```
@APEX design the authentication system architecture
```
Then proceed step by step.

---

## Platform-Specific Notes

### VS Code
- Use the Copilot Chat panel
- Highlight code for context
- Use `/` commands in combination with agents

### JetBrains
- Access via the Copilot tool window
- Right-click for context menu options

### GitHub.com
- Use Copilot Chat in the browser
- Works in issues, PRs, and code view

---

## Next Steps

- Learn about [Multi-Agent Collaboration](multi-agent-collaboration.md)
- See the complete [Agent Reference](agent-reference.md)
- Read [Best Practices](best-practices.md) for advanced usage
