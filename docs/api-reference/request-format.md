# Request Format

This document describes the format of requests to Elite Agent Collective agents.

## Basic Request Structure

```
@AGENT_NAME [request_content]
```

### Components

| Component | Description | Required |
|-----------|-------------|----------|
| `@` | Agent invocation prefix | Yes |
| `AGENT_NAME` | Agent codename in UPPERCASE | Yes |
| `request_content` | The task, question, or instruction | Yes |

---

## Request Types

### Simple Query

Basic questions or simple tasks.

**Format:**
```
@AGENT_NAME [question or simple task]
```

**Examples:**
```
@APEX what is the time complexity of quicksort?
@CIPHER explain AES-256 encryption
@TENSOR what is a transformer architecture?
```

### Implementation Request

Request for code implementation.

**Format:**
```
@AGENT_NAME implement [description]
```

**Examples:**
```
@APEX implement a binary search tree
@FLUX implement a GitHub Actions workflow
@TENSOR implement a CNN classifier
```

### Design Request

Request for architectural or system design.

**Format:**
```
@AGENT_NAME design [description]
```

**Examples:**
```
@ARCHITECT design a microservices architecture
@SYNAPSE design a GraphQL schema
@ATLAS design a multi-region AWS deployment
```

### Review Request

Request for code or design review.

**Format:**
```
@AGENT_NAME review [target]
```

**Examples:**
```
@MENTOR review this code
@CIPHER review this security implementation
@ECLIPSE review the test coverage
```

### Analysis Request

Request for analysis or optimization.

**Format:**
```
@AGENT_NAME analyze/optimize [target]
```

**Examples:**
```
@VELOCITY optimize this database query
@PRISM analyze this dataset
@FORTRESS analyze this API for vulnerabilities
```

---

## Context Specification

### Code Context

Reference code in your request:

**Current File:**
```
@APEX optimize this function
```

**Selected Code:**
```
# Select code first, then:
@ECLIPSE write tests for the selected code
```

**Inline Code:**
```
@APEX explain this code:
\`\`\`python
def mystery(n):
    return n * (n + 1) // 2
\`\`\`
```

### Specification Context

Include requirements in your request:

```
@ARCHITECT design a notification system
Requirements:
- Support 1M users
- Real-time delivery
- Multi-channel (email, push, SMS)
- 99.9% uptime
Technologies: AWS, Node.js, PostgreSQL
```

### Constraint Context

Specify constraints:

```
@APEX implement a cache with:
- Maximum 10,000 entries
- LRU eviction
- Thread-safe
- O(1) operations
- No external dependencies
```

---

## Multi-Agent Requests

### Simultaneous Invocation

Invoke multiple agents for combined expertise:

**Format:**
```
@AGENT1 @AGENT2 [request]
```

**Examples:**
```
@APEX @ARCHITECT design a scalable system
@CIPHER @FORTRESS security audit
@TENSOR @VELOCITY optimize ML inference
```

### Sequential Collaboration

Chain requests for workflow:

```
Step 1: @ARCHITECT design the architecture
Step 2: @APEX implement the core
Step 3: @ECLIPSE add tests
Step 4: @SCRIBE document the API
```

---

## Request Modifiers

### Output Format

Specify desired output format:

```
@APEX explain hash tables as a bulleted list
@SCRIBE document this in OpenAPI format
@ARCHITECT describe using C4 model
```

### Language/Framework

Specify technology preferences:

```
@APEX implement in TypeScript
@FLUX use GitHub Actions
@TENSOR use PyTorch, not TensorFlow
```

### Detail Level

Request specific detail levels:

```
@APEX brief: summarize the algorithm
@APEX detailed: comprehensive explanation with examples
@APEX quick: simple implementation
```

### Scope

Limit the scope:

```
@APEX implement basic version first
@ARCHITECT high-level design only
@ECLIPSE unit tests only, not integration
```

---

## Best Practices

### Be Specific

❌ **Vague:**
```
@APEX make this better
```

✅ **Specific:**
```
@APEX optimize this function to reduce time complexity from O(n²) to O(n log n)
```

### Provide Context

❌ **Missing context:**
```
@ARCHITECT design an API
```

✅ **With context:**
```
@ARCHITECT design a REST API for user management
- CRUD operations
- JWT authentication
- Rate limiting: 100 req/min
- Target: Node.js/Express
```

### State Requirements

Include relevant requirements:

```
@APEX implement password validation with:
- Minimum 8 characters
- At least one uppercase
- At least one number
- At least one special character
- No common passwords
```

### Use Appropriate Agent

Match agent to task:

| Task | Appropriate Agent |
|------|------------------|
| Algorithm implementation | @APEX |
| Security review | @CIPHER, @FORTRESS |
| Cloud design | @ATLAS |
| API design | @SYNAPSE |
| Testing | @ECLIPSE |

---

## Request Validation

### Valid Requests

- Include at least one agent mention
- Have a clear task or question
- Are within the agent's expertise

### Invalid Requests

- Empty requests after agent name
- Requests to non-existent agents
- Requests violating platform guidelines

---

## Examples by Domain

### Software Engineering

```
@APEX implement a thread-safe singleton pattern in Java
@APEX design a rate limiter using sliding window algorithm
@APEX refactor this code to follow SOLID principles
```

### Security

```
@CIPHER implement secure password hashing with Argon2
@FORTRESS create a threat model for this system
@AEGIS review for GDPR compliance
```

### Machine Learning

```
@TENSOR design a model for sentiment analysis
@PRISM perform exploratory data analysis
@ORACLE build a demand forecasting model
```

### DevOps

```
@FLUX create a CI/CD pipeline with testing and deployment
@ATLAS design multi-region deployment on AWS
@SENTRY implement distributed tracing with OpenTelemetry
```

---

## Related Documentation

- [Response Format](response-format.md)
- [Endpoints](endpoints.md)
- [Best Practices](../user-guide/best-practices.md)
