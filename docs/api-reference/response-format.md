# Response Format

This document describes the format of responses from Elite Agent Collective agents.

## Response Structure

Agent responses follow a consistent structure optimized for clarity and usefulness.

### Standard Response Components

1. **Acknowledgment**: Brief confirmation of understanding
2. **Content**: Main response body
3. **Code**: Implementation (when applicable)
4. **Explanation**: Context and reasoning
5. **Next Steps**: Suggestions for follow-up (optional)

---

## Response Types

### Explanatory Response

For questions and conceptual queries.

**Structure:**
```markdown
[Brief introduction or acknowledgment]

[Main explanation organized with headings, lists, or paragraphs]

[Examples if helpful]

[Summary or key takeaways]
```

**Example:**
```
Quicksort has an average time complexity of **O(n log n)**.

## How It Works
1. Select a pivot element
2. Partition array around pivot
3. Recursively sort sub-arrays

## Complexity Analysis
- **Best/Average Case**: O(n log n)
- **Worst Case**: O(n²) - occurs with sorted input and poor pivot selection
- **Space**: O(log n) for recursion stack

## Key Insight
The efficiency depends heavily on pivot selection. Random pivot or median-of-three reduces worst-case likelihood.
```

### Code Response

For implementation requests.

**Structure:**
```markdown
[Brief description of the solution]

\`\`\`language
[Implementation code]
\`\`\`

[Explanation of key parts]

[Usage examples]

[Complexity or performance notes]
```

**Example:**
```
Here's a Python implementation of binary search:

\`\`\`python
def binary_search(arr: list[int], target: int) -> int:
    """
    Binary search for target in sorted array.
    
    Args:
        arr: Sorted list of integers
        target: Value to find
        
    Returns:
        Index of target, or -1 if not found
    """
    left, right = 0, len(arr) - 1
    
    while left <= right:
        mid = left + (right - left) // 2  # Avoid overflow
        
        if arr[mid] == target:
            return mid
        elif arr[mid] < target:
            left = mid + 1
        else:
            right = mid - 1
            
    return -1
\`\`\`

## Key Points
- Uses `left + (right - left) // 2` to avoid integer overflow
- Works only on sorted arrays
- Time complexity: O(log n)
- Space complexity: O(1)

## Usage
\`\`\`python
arr = [1, 3, 5, 7, 9, 11]
print(binary_search(arr, 7))  # Output: 3
print(binary_search(arr, 4))  # Output: -1
\`\`\`
```

### Design Response

For architecture and design requests.

**Structure:**
```markdown
[Design overview]

## Architecture Diagram (if applicable)
[ASCII diagram or description]

## Components
[Component descriptions]

## Data Flow
[How data moves through the system]

## Trade-offs
[Design decisions and their implications]

## Recommendations
[Specific technology or implementation suggestions]
```

**Example:**
```
Here's a microservices architecture for an e-commerce platform:

## Architecture Overview

\`\`\`
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│  API Gateway │────▶│   Services  │
└─────────────┘     └─────────────┘     └─────────────┘
                           │
        ┌──────────────────┼──────────────────┐
        ▼                  ▼                  ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Product   │     │    Cart     │     │    Order    │
│   Service   │     │   Service   │     │   Service   │
└─────────────┘     └─────────────┘     └─────────────┘
\`\`\`

## Components

### API Gateway
- Request routing
- Authentication
- Rate limiting

### Product Service
- Catalog management
- Search and filtering
- PostgreSQL database

### Cart Service
- Session management
- Redis for caching
- Cart persistence

### Order Service
- Order processing
- Payment integration
- Event publishing

## Trade-offs

| Decision | Benefit | Trade-off |
|----------|---------|-----------|
| Microservices | Independent scaling | Operational complexity |
| Event-driven | Loose coupling | Eventual consistency |
| API Gateway | Centralized concerns | Single point of failure |
```

### Review Response

For code or design review requests.

**Structure:**
```markdown
[Overall assessment]

## Strengths
[Positive aspects]

## Areas for Improvement
[Issues found, organized by severity]

## Specific Recommendations
[Actionable suggestions with examples]

## Summary
[Final thoughts]
```

**Example:**
```
## Code Review Summary

Overall, the implementation is functional but has several areas for improvement.

### ✅ Strengths
- Clear function naming
- Good separation of concerns
- Appropriate use of type hints

### ⚠️ Areas for Improvement

**High Priority:**
1. **SQL Injection Risk** (line 45)
   - Current: `f"SELECT * FROM users WHERE id = {user_id}"`
   - Fix: Use parameterized queries

**Medium Priority:**
2. **Missing error handling** (line 23)
   - Add try/catch for database operations

**Low Priority:**
3. **Magic numbers** (line 67)
   - Extract `86400` to `SECONDS_PER_DAY` constant

### Recommendations

Before:
\`\`\`python
query = f"SELECT * FROM users WHERE id = {user_id}"
\`\`\`

After:
\`\`\`python
query = "SELECT * FROM users WHERE id = %s"
cursor.execute(query, (user_id,))
\`\`\`

### Summary
Address the SQL injection issue before deployment. Other items can be tackled in follow-up PRs.
```

---

## Response Formatting

### Markdown Elements

Agents use standard Markdown formatting:

- **Headers**: For section organization
- **Bold/Italic**: For emphasis
- **Code blocks**: For code with syntax highlighting
- **Tables**: For structured data
- **Lists**: For steps or items
- **Links**: For references (when applicable)

### Code Block Languages

Code blocks include language identifiers:

```python
# Python code
```

```typescript
// TypeScript code
```

```bash
# Shell commands
```

```yaml
# YAML configuration
```

### Diagrams

ASCII diagrams for architecture:

```
┌───────────┐     ┌───────────┐
│  Client   │────▶│  Server   │
└───────────┘     └───────────┘
```

---

## Agent-Specific Response Patterns

### @APEX
- Clean code with comments
- Complexity analysis
- Alternative approaches

### @CIPHER
- Security considerations
- Threat analysis
- Compliance notes

### @ARCHITECT
- System diagrams
- Component breakdown
- Trade-off analysis

### @ECLIPSE
- Test structure
- Coverage considerations
- Edge cases

### @SCRIBE
- Formatted documentation
- Examples
- API specifications

---

## Response Quality Indicators

### Good Responses Include

- ✅ Direct answer to the question
- ✅ Relevant code examples
- ✅ Explanation of reasoning
- ✅ Consideration of edge cases
- ✅ Actionable next steps

### Response Limitations

Agents may indicate:
- Uncertainty about specific details
- Need for clarification
- Suggestions for alternative agents
- Scope limitations

---

## Error Responses

When agents cannot fully address a request:

```markdown
I can help with [related aspect], but for [specific need], 
you might want to consult @AGENT_NAME who specializes in that area.

Here's what I can provide:
[Partial response]

For the full solution, consider:
1. [Suggestion 1]
2. [Suggestion 2]
```

---

## Related Documentation

- [Request Format](request-format.md)
- [Endpoints](endpoints.md)
- [Best Practices](../user-guide/best-practices.md)
