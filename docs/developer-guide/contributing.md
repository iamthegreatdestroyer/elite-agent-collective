# Contributing to Elite Agent Collective

Thank you for your interest in contributing to the Elite Agent Collective! This guide will help you get started.

## Table of Contents

- [Ways to Contribute](#ways-to-contribute)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contribution Workflow](#contribution-workflow)
- [Coding Standards](#coding-standards)
- [Documentation Standards](#documentation-standards)
- [Pull Request Process](#pull-request-process)
- [Community Guidelines](#community-guidelines)

---

## Ways to Contribute

### 1. Report Issues

Found a bug or have a suggestion? [Open an issue](https://github.com/iamthegreatdestroyer/elite-agent-collective/issues/new).

**Good issue reports include:**
- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Environment details (IDE, OS, Copilot version)
- Agent(s) involved

### 2. Suggest New Agents

Have an idea for a new specialized agent? We'd love to hear it!

- Open a [feature request](https://github.com/iamthegreatdestroyer/elite-agent-collective/issues/new?template=feature_request.md)
- Describe the agent's specialty
- Explain the use cases it would address
- Suggest which tier it belongs to

### 3. Improve Documentation

Documentation improvements are always welcome:

- Fix typos and clarify language
- Add examples and use cases
- Improve tutorials
- Translate documentation

### 4. Enhance Existing Agents

Help make agents better:

- Add new capabilities
- Improve response quality
- Add specialized knowledge
- Fix incorrect information

### 5. Add Features

Submit PRs for new features:

- New agent definitions
- Improved collaboration patterns
- Testing infrastructure
- Tooling improvements

---

## Getting Started

### Prerequisites

- Git
- A text editor (VS Code recommended)
- GitHub account
- Familiarity with Markdown

### Fork and Clone

```bash
# Fork the repository on GitHub, then:
git clone https://github.com/YOUR_USERNAME/elite-agent-collective.git
cd elite-agent-collective
```

### Create a Branch

```bash
git checkout -b feature/your-feature-name
```

---

## Development Setup

### Local Testing

To test agent instructions locally:

1. **Global Installation**:
   ```bash
   # Copy to your home directory
   cp .github/copilot-instructions.md ~/.github/
   ```

2. **VS Code Prompts**:
   ```bash
   # Copy to VS Code prompts directory
   mkdir -p ~/Library/Application\ Support/Code/User/prompts/agents/
   cp vscode-prompts/ELITE_AGENT_COLLECTIVE.instructions.md ~/Library/Application\ Support/Code/User/prompts/
   cp vscode-prompts/agents/*.instructions.md ~/Library/Application\ Support/Code/User/prompts/agents/
   ```

3. **Verify**:
   - Open Copilot Chat
   - Test: `@APEX help`

### Running Tests

```bash
# Navigate to tests directory
cd tests

# Run all tests
python run_all_tests.py
```

---

## Contribution Workflow

### 1. Choose an Issue

- Look for issues labeled `good first issue` or `help wanted`
- Comment on the issue to claim it
- Wait for maintainer confirmation

### 2. Make Changes

- Keep changes focused and atomic
- Follow the coding standards
- Update documentation as needed
- Add tests if applicable

### 3. Test Locally

- Verify your changes work as expected
- Test with Copilot Chat
- Run the test suite

### 4. Submit PR

- Push your changes
- Open a pull request
- Fill out the PR template
- Wait for review

---

## Coding Standards

### Agent Definition Format

Follow this template for agent definitions:

```markdown
#### @AGENT_NAME (ID) - Specialty Title

**Primary Function:** Clear description of the agent's purpose

**Philosophy:** *"A guiding principle in quotes"*

**Invoke:** `@AGENT_NAME [task]`

**Capabilities:**
- Capability 1
- Capability 2
- Capability 3

**Example invocations:**
\`\`\`
@AGENT_NAME example task 1
@AGENT_NAME example task 2
\`\`\`

**Collaborates well with:** @AGENT1, @AGENT2
```

### Markdown Guidelines

- Use consistent heading levels
- Include code blocks with language tags
- Use tables for structured data
- Keep lines under 120 characters when possible

### File Naming

- Agent files: `AGENT_NAME.instructions.md`
- Documentation: `lowercase-with-dashes.md`
- Use descriptive names

---

## Documentation Standards

### Writing Style

- Use clear, concise language
- Write for an international audience
- Avoid jargon without explanation
- Use active voice
- Include examples

### Structure

- Start with a brief overview
- Use headings to organize content
- Include a table of contents for long documents
- End with related links or next steps

### Examples

Every feature should have examples:

```markdown
### Feature Name

Description of the feature.

**Example:**
\`\`\`
@APEX implement a binary search tree
\`\`\`

**Result:** APEX provides a complete BST implementation with insert, search, and delete operations.
```

---

## Pull Request Process

### Before Submitting

- [ ] Changes are focused and atomic
- [ ] Documentation is updated
- [ ] Tests pass (if applicable)
- [ ] Code follows style guidelines
- [ ] Commit messages are clear

### PR Template

```markdown
## Description
Brief description of the changes.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Agent improvement

## Related Issue
Fixes #issue_number

## Testing
Describe how you tested the changes.

## Checklist
- [ ] I have read the contributing guide
- [ ] I have tested my changes
- [ ] I have updated documentation
```

### Review Process

1. A maintainer will review your PR
2. Address any requested changes
3. Once approved, it will be merged
4. Your contribution will be acknowledged

---

## Community Guidelines

### Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Provide constructive feedback
- Focus on the work, not the person
- Assume good intentions

### Communication

- Use GitHub issues for bugs and features
- Be clear and specific
- Respond to feedback promptly
- Thank contributors

### Recognition

Contributors are recognized in:
- The changelog
- GitHub contributors list
- Release notes

---

## Specific Contribution Types

### Adding a New Agent

See [Adding Agents Guide](adding-agents.md) for detailed instructions.

### Improving Agent Capabilities

1. Identify the agent to improve
2. Research the domain thoroughly
3. Add new capabilities with examples
4. Update collaboration relationships
5. Test with real-world scenarios

### Documentation Improvements

1. Identify gaps or unclear sections
2. Research the topic
3. Write clear, concise content
4. Add examples where helpful
5. Link to related documentation

---

## Questions?

- Check existing [issues](https://github.com/iamthegreatdestroyer/elite-agent-collective/issues)
- Open a [discussion](https://github.com/iamthegreatdestroyer/elite-agent-collective/discussions)
- Review the [documentation](../README.md)

---

Thank you for contributing to the Elite Agent Collective! Your efforts help make this project better for everyone.

*"The collective intelligence of specialized minds exceeds the sum of their parts."*
