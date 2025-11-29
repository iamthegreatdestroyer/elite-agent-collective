# Getting Support

This guide explains how to get help with the Elite Agent Collective.

---

## Self-Service Resources

Before reaching out, try these resources:

### 1. Documentation

- [Getting Started Guide](../getting-started/installation.md)
- [User Guide](../user-guide/invoking-agents.md)
- [Common Issues](common-issues.md)
- [Agent Reference](../user-guide/agent-reference.md)

### 2. Search Existing Issues

Many questions have already been answered:

[Search Issues](https://github.com/iamthegreatdestroyer/elite-agent-collective/issues?q=is%3Aissue)

### 3. Check GitHub Discussions

Community discussions and Q&A:

[Discussions](https://github.com/iamthegreatdestroyer/elite-agent-collective/discussions)

---

## Reporting Issues

### Bug Reports

For bugs and unexpected behavior:

1. Go to [GitHub Issues](https://github.com/iamthegreatdestroyer/elite-agent-collective/issues/new?template=bug_report.md)
2. Use the Bug Report template
3. Include:
   - **Description**: What happened?
   - **Steps to Reproduce**: How can we see the issue?
   - **Expected Behavior**: What should have happened?
   - **Environment**: IDE, OS, Copilot version
   - **Agent(s) Involved**: Which agents were you using?

### Good Bug Report Example

```markdown
## Description
@APEX doesn't respond with specialized behavior when invoked.

## Steps to Reproduce
1. Install instructions file to ~/.github/
2. Open VS Code with Copilot Chat
3. Type: @APEX help me design a rate limiter
4. Response is generic, not APEX-specific

## Expected Behavior
APEX should respond with detailed software engineering guidance.

## Environment
- VS Code 1.85.0
- GitHub Copilot 1.143.0
- macOS 14.2
- Instructions file version: 2.0

## Additional Context
Tried reinstalling, restarting IDE. Same issue persists.
```

---

## Feature Requests

### Suggesting New Features

Have an idea? We'd love to hear it!

1. Go to [Feature Request](https://github.com/iamthegreatdestroyer/elite-agent-collective/issues/new?template=feature_request.md)
2. Describe:
   - **The Problem**: What need does this address?
   - **Proposed Solution**: What should we build?
   - **Alternatives Considered**: Other approaches you thought of
   - **Additional Context**: Screenshots, examples, etc.

### Suggesting New Agents

For new agent ideas:

1. Open a feature request
2. Include:
   - Agent name and specialty
   - Use cases it would address
   - Proposed tier placement
   - Example invocations

---

## Security Issues

### Reporting Security Vulnerabilities

**⚠️ Do NOT open public issues for security vulnerabilities.**

For security concerns:

1. Email: security@elite-agent-collective.dev
2. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We will respond within 48 hours and work with you on a fix.

---

## Community Support

### GitHub Discussions

For questions and community help:

- [Ask Questions](https://github.com/iamthegreatdestroyer/elite-agent-collective/discussions/categories/q-a)
- [Share Ideas](https://github.com/iamthegreatdestroyer/elite-agent-collective/discussions/categories/ideas)
- [Show & Tell](https://github.com/iamthegreatdestroyer/elite-agent-collective/discussions/categories/show-and-tell)

### Community Guidelines

When participating:

- Be respectful and helpful
- Provide context and details
- Search before posting
- Use appropriate categories
- Thank those who help you

---

## Contributing

Want to help fix issues?

1. See [Contributing Guide](../developer-guide/contributing.md)
2. Look for issues labeled:
   - `good first issue` - Great for newcomers
   - `help wanted` - Community contributions welcome
   - `documentation` - Docs improvements needed

---

## Response Times

### Issue Triage

- **Critical bugs**: Within 24 hours
- **Regular bugs**: Within 1 week
- **Feature requests**: Reviewed monthly
- **Security issues**: Within 48 hours

### Pull Requests

- Initial review: Within 1 week
- Follow-up reviews: Within 48 hours

---

## Staying Updated

### Watch the Repository

Click "Watch" on GitHub to get notified of:
- New releases
- Important discussions
- Issue updates

### Release Notes

Check [CHANGELOG.md](../../CHANGELOG.md) for updates.

---

## Frequently Asked Questions

### General

**Q: Is this an official GitHub product?**
A: No, Elite Agent Collective is a community project that works with GitHub Copilot.

**Q: Do I need a Copilot subscription?**
A: Yes, GitHub Copilot subscription is required.

**Q: Which IDEs are supported?**
A: VS Code, JetBrains IDEs, and GitHub.com Copilot Chat.

### Technical

**Q: Why doesn't my agent respond correctly?**
A: See [Common Issues](common-issues.md) for troubleshooting steps.

**Q: Can I customize agents?**
A: Yes, fork the repo and modify the instructions file.

**Q: How do I add my own agents?**
A: See [Adding Agents Guide](../developer-guide/adding-agents.md).

---

## Contact

For issues not covered above:

- **GitHub Issues**: [Open an Issue](https://github.com/iamthegreatdestroyer/elite-agent-collective/issues/new)
- **GitHub Discussions**: [Start a Discussion](https://github.com/iamthegreatdestroyer/elite-agent-collective/discussions/new)
- **Security**: security@elite-agent-collective.dev

---

Thank you for using the Elite Agent Collective! We're committed to making this project as helpful as possible.

*"The collective intelligence of specialized minds exceeds the sum of their parts."*
