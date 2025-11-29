# Contributing to Elite Agent Collective

Thank you for your interest in contributing to the Elite Agent Collective! This document provides guidelines for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Ways to Contribute](#ways-to-contribute)
- [Getting Started](#getting-started)
- [Pull Request Process](#pull-request-process)
- [Style Guidelines](#style-guidelines)
- [Community](#community)

---

## Code of Conduct

### Our Standards

- Be respectful and inclusive
- Welcome newcomers
- Provide constructive feedback
- Focus on the work, not the person
- Assume good intentions

### Unacceptable Behavior

- Harassment or discrimination
- Trolling or insulting comments
- Personal or political attacks
- Publishing others' private information

---

## Ways to Contribute

### 1. Report Issues

Found a bug? [Open an issue](https://github.com/iamthegreatdestroyer/elite-agent-collective/issues/new?template=bug_report.md).

Include:
- Clear description
- Steps to reproduce
- Expected vs actual behavior
- Environment details

### 2. Suggest New Agents

Have an idea for a new agent? [Open a feature request](https://github.com/iamthegreatdestroyer/elite-agent-collective/issues/new?template=feature_request.md).

Include:
- Agent specialty
- Use cases
- Proposed tier
- Example invocations

### 3. Improve Documentation

Documentation improvements are always welcome:
- Fix typos
- Clarify language
- Add examples
- Improve tutorials

### 4. Enhance Existing Agents

Help make agents better:
- Add capabilities
- Improve knowledge
- Fix inaccuracies
- Add examples

### 5. Submit Code

PRs for new features and improvements:
- New agents
- Tooling improvements
- Test additions

---

## Getting Started

### 1. Fork the Repository

Click "Fork" on GitHub to create your copy.

### 2. Clone Your Fork

```bash
git clone https://github.com/YOUR_USERNAME/elite-agent-collective.git
cd elite-agent-collective
```

### 3. Create a Branch

```bash
git checkout -b feature/your-feature-name
```

### 4. Make Changes

Edit files as needed, following our style guidelines.

### 5. Test Changes

For agent instructions:
```bash
# Copy to local installation
cp .github/copilot-instructions.md ~/.github/

# Test in Copilot Chat
# @AGENT_NAME help
```

### 6. Commit Changes

```bash
git add .
git commit -m "type: brief description"
```

Commit types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `refactor`: Code refactoring
- `test`: Tests
- `chore`: Maintenance

### 7. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then open a Pull Request on GitHub.

---

## Pull Request Process

### Before Submitting

- [ ] Changes are focused and atomic
- [ ] Documentation is updated
- [ ] Agent format follows template
- [ ] Commit messages are clear

### PR Description

Include:
- Description of changes
- Type of change
- Related issues
- Testing performed

### Review Process

1. Maintainer reviews PR
2. Address feedback
3. Approval and merge

---

## Style Guidelines

### Agent Definition Format

```markdown
#### @AGENT_NAME (ID) - Specialty Title

**Primary Function:** Description

**Philosophy:** *"Guiding principle"*

**Invoke:** `@AGENT_NAME [task]`

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

### Markdown Guidelines

- Consistent heading levels
- Code blocks with language tags
- Tables for structured data
- Clear, concise language

### File Naming

- Agents: `AGENT_NAME.instructions.md`
- Docs: `lowercase-with-dashes.md`

---

## Adding New Agents

See [docs/developer-guide/adding-agents.md](docs/developer-guide/adding-agents.md) for detailed instructions.

Quick checklist:
1. Define specialty and tier
2. Create agent definition
3. Add to main instructions
4. Create VS Code file (optional)
5. Add documentation
6. Test thoroughly

---

## Community

### Getting Help

- [Documentation](docs/README.md)
- [GitHub Discussions](https://github.com/iamthegreatdestroyer/elite-agent-collective/discussions)
- [Existing Issues](https://github.com/iamthegreatdestroyer/elite-agent-collective/issues)

### Recognition

Contributors are recognized in:
- Changelog
- GitHub contributors list
- Release notes

---

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to the Elite Agent Collective!

*"The collective intelligence of specialized minds exceeds the sum of their parts."*
