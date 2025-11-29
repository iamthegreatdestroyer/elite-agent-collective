# Local Development

This guide explains how to set up a local development environment for contributing to the Elite Agent Collective.

## Prerequisites

### Required

- **Git**: For version control
- **Text Editor**: VS Code recommended (with Copilot for testing)
- **GitHub Account**: For forking and contributing

### Recommended

- **GitHub Copilot**: For testing agent instructions
- **Markdown Preview**: For documentation work
- **Node.js**: For running documentation tools (if needed)
- **Python 3.x**: For running tests

---

## Initial Setup

### 1. Fork the Repository

1. Go to [Elite Agent Collective](https://github.com/iamthegreatdestroyer/elite-agent-collective)
2. Click "Fork" in the top right
3. Select your account

### 2. Clone Your Fork

```bash
git clone https://github.com/YOUR_USERNAME/elite-agent-collective.git
cd elite-agent-collective
```

### 3. Add Upstream Remote

```bash
git remote add upstream https://github.com/iamthegreatdestroyer/elite-agent-collective.git
```

### 4. Verify Remotes

```bash
git remote -v
# Should show:
# origin    https://github.com/YOUR_USERNAME/elite-agent-collective.git
# upstream  https://github.com/iamthegreatdestroyer/elite-agent-collective.git
```

---

## Repository Structure

```
elite-agent-collective/
├── .github/
│   ├── copilot-instructions.md      # Main Copilot instructions
│   └── ISSUE_TEMPLATE/              # Issue templates
├── docs/                            # Documentation
│   ├── getting-started/
│   ├── user-guide/
│   ├── developer-guide/
│   ├── api-reference/
│   └── troubleshooting/
├── marketplace/                     # Marketplace assets
├── profiles/                        # Agent profiles by tier
├── tests/                           # Test suite
├── vscode-prompts/                  # VS Code prompt files
│   ├── ELITE_AGENT_COLLECTIVE.instructions.md
│   └── agents/                      # Individual agent files
├── README.md
├── CHANGELOG.md
├── CONTRIBUTING.md
└── LICENSE
```

---

## Development Workflow

### Creating a Feature Branch

```bash
# Ensure you're on main
git checkout main

# Pull latest changes
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name
```

### Making Changes

1. Edit files as needed
2. Test your changes (see Testing section)
3. Commit with clear messages:

```bash
git add .
git commit -m "type: brief description

Longer description if needed"
```

Commit types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance tasks

### Pushing Changes

```bash
git push origin feature/your-feature-name
```

### Creating a Pull Request

1. Go to your fork on GitHub
2. Click "Compare & pull request"
3. Fill out the PR template
4. Submit for review

---

## Testing Your Changes

### Testing Copilot Instructions

#### Global Installation (Recommended for Testing)

```bash
# macOS/Linux
cp .github/copilot-instructions.md ~/.github/

# Windows PowerShell
Copy-Item .github/copilot-instructions.md $env:USERPROFILE\.github\
```

#### VS Code Prompts

```bash
# macOS
cp vscode-prompts/ELITE_AGENT_COLLECTIVE.instructions.md \
   ~/Library/Application\ Support/Code/User/prompts/
cp vscode-prompts/agents/*.instructions.md \
   ~/Library/Application\ Support/Code/User/prompts/agents/

# Linux
cp vscode-prompts/ELITE_AGENT_COLLECTIVE.instructions.md \
   ~/.config/Code/User/prompts/
cp vscode-prompts/agents/*.instructions.md \
   ~/.config/Code/User/prompts/agents/

# Windows PowerShell
Copy-Item vscode-prompts/ELITE_AGENT_COLLECTIVE.instructions.md `
   $env:APPDATA/Code/User/prompts/
Copy-Item vscode-prompts/agents/*.instructions.md `
   $env:APPDATA/Code/User/prompts/agents/
```

#### Verification Tests

After copying, test in Copilot Chat:

```
@APEX help
@CIPHER help
@ARCHITECT help
```

Each should respond with appropriate agent-specific information.

### Running the Test Suite

```bash
cd tests

# Run all tests
python run_all_tests.py

# Run specific tier tests
python -m pytest tier_1_foundational/
```

### Testing Documentation

1. Preview Markdown files locally
2. Check all links work
3. Verify code examples are correct
4. Run any documentation linters

#### Using VS Code Markdown Preview

1. Open a `.md` file
2. Press `Cmd+Shift+V` (macOS) or `Ctrl+Shift+V` (Windows/Linux)
3. Preview opens in a new tab

---

## Development Tips

### Quick Iteration

For fast testing of instruction changes:

```bash
# Create a quick copy script
#!/bin/bash
cp .github/copilot-instructions.md ~/.github/
echo "Instructions updated!"
```

Save as `update-local.sh` and run after changes.

### Watching for Changes

Use file watching for automatic updates:

```bash
# Using fswatch (macOS)
fswatch -o .github/copilot-instructions.md | xargs -n1 -I{} cp .github/copilot-instructions.md ~/.github/

# Using inotifywait (Linux)
inotifywait -m -e modify .github/copilot-instructions.md | while read; do
  cp .github/copilot-instructions.md ~/.github/
done
```

### Validating Markdown

Use a linter to check Markdown quality:

```bash
# Install markdownlint-cli
npm install -g markdownlint-cli

# Lint documentation
markdownlint docs/
markdownlint .github/
```

---

## VS Code Setup

### Recommended Extensions

- **Markdown All in One**: Markdown editing
- **markdownlint**: Linting
- **GitHub Copilot**: For testing
- **GitLens**: Git visualization

### Settings

Add to `.vscode/settings.json`:

```json
{
  "editor.wordWrap": "on",
  "markdown.preview.breaks": true,
  "[markdown]": {
    "editor.defaultFormatter": "DavidAnson.vscode-markdownlint"
  }
}
```

---

## Keeping Your Fork Updated

### Regular Sync

```bash
# Fetch upstream changes
git fetch upstream

# Switch to main
git checkout main

# Merge upstream changes
git merge upstream/main

# Push to your fork
git push origin main
```

### Rebasing Feature Branches

```bash
# On your feature branch
git fetch upstream
git rebase upstream/main

# Force push if needed
git push origin feature/your-feature-name --force
```

---

## Troubleshooting

### Agent Not Responding

1. Verify Copilot is active
2. Check instructions file location
3. Restart IDE
4. Clear Copilot cache

### Merge Conflicts

```bash
# Fetch latest
git fetch upstream

# Rebase on main
git rebase upstream/main

# Resolve conflicts in each file
# Then continue
git rebase --continue
```

### Test Failures

1. Check test output for details
2. Verify test dependencies installed
3. Run individual tests to isolate issues
4. Check for environment differences

---

## Getting Help

- **Documentation**: Check the docs folder
- **Issues**: Search existing issues
- **Discussions**: Ask in GitHub Discussions
- **Code Review**: Request feedback on draft PRs

---

## Next Steps

- Read the [Contributing Guide](contributing.md)
- Explore [Adding Agents](adding-agents.md)
- Review the [Architecture](architecture.md)
