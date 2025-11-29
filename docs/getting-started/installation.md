# Installation

This guide covers all the ways to install and use the Elite Agent Collective with GitHub Copilot.

## Installation Methods

### Option 1: GitHub Marketplace (Recommended)

The easiest way to get started is through the GitHub Marketplace:

1. Visit the [Elite Agent Collective](https://github.com/marketplace/elite-agent-collective) on GitHub Marketplace
2. Click "Install"
3. Select the organization/account to install to
4. Authorize the required permissions
5. Start using agents in Copilot Chat!

### Option 2: Global Installation

Copy the instructions to your home directory for **all repositories**:

**Windows (PowerShell):**
```powershell
# Create directory if it doesn't exist
New-Item -ItemType Directory -Path "$env:USERPROFILE\.github" -Force

# Copy the instructions file
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/iamthegreatdestroyer/elite-agent-collective/main/.github/copilot-instructions.md" -OutFile "$env:USERPROFILE\.github\copilot-instructions.md"
```

**Windows (CMD):**
```cmd
mkdir %USERPROFILE%\.github 2>nul
curl -o %USERPROFILE%\.github\copilot-instructions.md https://raw.githubusercontent.com/iamthegreatdestroyer/elite-agent-collective/main/.github/copilot-instructions.md
```

**macOS/Linux:**
```bash
mkdir -p ~/.github
curl -o ~/.github/copilot-instructions.md https://raw.githubusercontent.com/iamthegreatdestroyer/elite-agent-collective/main/.github/copilot-instructions.md
```

### Option 3: Per-Repository Installation

For project-specific installation:

1. Clone or download the Elite Agent Collective repository
2. Copy `.github/copilot-instructions.md` to your repository's `.github/` folder
3. Commit and push the changes

```bash
# From your repository root
mkdir -p .github
curl -o .github/copilot-instructions.md https://raw.githubusercontent.com/iamthegreatdestroyer/elite-agent-collective/main/.github/copilot-instructions.md
git add .github/copilot-instructions.md
git commit -m "Add Elite Agent Collective instructions"
```

### Option 4: VS Code User Prompts

For VS Code integration with individual agent files:

**Windows (PowerShell):**
```powershell
# Create directories
New-Item -ItemType Directory -Path "$env:APPDATA\Code\User\prompts\agents" -Force

# Clone repository and copy files
git clone https://github.com/iamthegreatdestroyer/elite-agent-collective.git temp-eac
Copy-Item -Path "temp-eac\vscode-prompts\ELITE_AGENT_COLLECTIVE.instructions.md" -Destination "$env:APPDATA\Code\User\prompts\"
Copy-Item -Path "temp-eac\vscode-prompts\agents\*.instructions.md" -Destination "$env:APPDATA\Code\User\prompts\agents\"
Remove-Item -Recurse -Force temp-eac
```

**macOS:**
```bash
# Create directories
mkdir -p ~/Library/Application\ Support/Code/User/prompts/agents/

# Clone repository and copy files
git clone https://github.com/iamthegreatdestroyer/elite-agent-collective.git temp-eac
cp temp-eac/vscode-prompts/ELITE_AGENT_COLLECTIVE.instructions.md ~/Library/Application\ Support/Code/User/prompts/
cp temp-eac/vscode-prompts/agents/*.instructions.md ~/Library/Application\ Support/Code/User/prompts/agents/
rm -rf temp-eac
```

**Linux:**
```bash
# Create directories
mkdir -p ~/.config/Code/User/prompts/agents/

# Clone repository and copy files
git clone https://github.com/iamthegreatdestroyer/elite-agent-collective.git temp-eac
cp temp-eac/vscode-prompts/ELITE_AGENT_COLLECTIVE.instructions.md ~/.config/Code/User/prompts/
cp temp-eac/vscode-prompts/agents/*.instructions.md ~/.config/Code/User/prompts/agents/
rm -rf temp-eac
```

---

## Permissions Required

| Permission | Purpose |
|------------|---------|
| Copilot Chat | Enable agent invocation in chat |
| Repository Content (Read) | Context-aware responses |

---

## Supported Platforms

| Platform | Status | Notes |
|----------|--------|-------|
| VS Code with GitHub Copilot | ✅ Supported | Full functionality |
| JetBrains IDEs with GitHub Copilot | ✅ Supported | Full functionality |
| GitHub.com Copilot Chat | ✅ Supported | Full functionality |
| GitHub Mobile | ✅ Supported | Limited UI |
| Neovim with Copilot | ✅ Supported | Via copilot.vim/copilot.lua |

---

## Verification

After installation, verify the setup by testing in Copilot Chat:

```
@APEX help
```

You should receive a response from the APEX agent explaining its capabilities.

### Troubleshooting Installation

**Agent not responding?**
- Ensure GitHub Copilot is active and authenticated
- Check that the instructions file is in the correct location
- Try restarting your IDE

**Wrong agent behavior?**
- Verify the instructions file is the latest version
- Check for syntax errors if you've modified the file

See [Common Issues](../troubleshooting/common-issues.md) for more troubleshooting help.

---

## Updating

To update to the latest version:

**Marketplace Installation:**
Updates are automatic.

**Manual Installation:**
Re-run the installation commands to download the latest version of the instructions file.

---

## Uninstallation

**Marketplace:**
1. Go to GitHub Settings → Applications
2. Find Elite Agent Collective
3. Click "Uninstall"

**Manual Installation:**
Remove the `copilot-instructions.md` file from your `.github/` directory or home directory.
