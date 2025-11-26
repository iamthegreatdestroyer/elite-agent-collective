# 🧠 Elite Agent Collective

> **20 Specialized AI Agents for GitHub Copilot**

A comprehensive system of specialized AI agents designed to provide expert-level assistance across all domains of software engineering, research, and innovation.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Agents: 20](https://img.shields.io/badge/Agents-20-blue.svg)]()
[![Status: Active](https://img.shields.io/badge/Status-Active-green.svg)]()

---

## 🚀 Quick Start

### Option 1: Global Installation (Recommended)

Copy the instructions to your home directory for **all repositories**:

```bash
# Windows
copy .github\copilot-instructions.md %USERPROFILE%\.github\

# macOS/Linux
cp .github/copilot-instructions.md ~/.github/
```

### Option 2: Per-Repository

Copy `.github/copilot-instructions.md` to any repository's `.github/` folder.

### Option 3: VS Code Integration

Copy the `vscode-prompts/` folder contents to your VS Code user prompts directory.

#### Windows PowerShell (Recommended)

```powershell
# Create directories if they don't exist
New-Item -ItemType Directory -Path "$env:APPDATA\Code\User\prompts" -Force
New-Item -ItemType Directory -Path "$env:APPDATA\Code\User\prompts\agents" -Force

# Copy main file
Copy-Item -Path "vscode-prompts\ELITE_AGENT_COLLECTIVE.instructions.md" -Destination "$env:APPDATA\Code\User\prompts\"

# Copy all agent files
Copy-Item -Path "vscode-prompts\agents\*.instructions.md" -Destination "$env:APPDATA\Code\User\prompts\agents\"
```

#### Windows CMD

```cmd
REM Create directories if they don't exist
mkdir %APPDATA%\Code\User\prompts 2>nul
mkdir %APPDATA%\Code\User\prompts\agents 2>nul

REM Copy main file
copy vscode-prompts\ELITE_AGENT_COLLECTIVE.instructions.md %APPDATA%\Code\User\prompts\

REM Copy agent files individually
for %%f in (vscode-prompts\agents\*.instructions.md) do copy "%%f" %APPDATA%\Code\User\prompts\agents\
```

#### macOS/Linux

```bash
# Create directories if they don't exist
mkdir -p ~/Library/Application\ Support/Code/User/prompts/agents/

# Copy files
cp vscode-prompts/ELITE_AGENT_COLLECTIVE.instructions.md ~/Library/Application\ Support/Code/User/prompts/
cp vscode-prompts/agents/*.instructions.md ~/Library/Application\ Support/Code/User/prompts/agents/
```

#### Verify Installation

After copying, verify the files are in place:

**PowerShell:**
```powershell
Get-ChildItem "$env:APPDATA\Code\User\prompts\" -Name
Get-ChildItem "$env:APPDATA\Code\User\prompts\agents\" -Name
```

**CMD:**
```cmd
dir %APPDATA%\Code\User\prompts\
dir %APPDATA%\Code\User\prompts\agents\
```

**macOS/Linux:**
```bash
ls ~/Library/Application\ Support/Code/User/prompts/
ls ~/Library/Application\ Support/Code/User/prompts/agents/
```

You should see 1 file in the `prompts` directory and 20 agent files in the `agents` subdirectory.

---

## 📋 Agent Registry

### Tier 1: Foundational Agents

| ID  |   Codename    | Specialization                                   | Invocation   |
| :-: | :-----------: | :----------------------------------------------- | :----------- |
| 01  |   **APEX**    | Elite Computer Science Engineering               | `@APEX`      |
| 02  |  **CIPHER**   | Advanced Cryptography & Security                 | `@CIPHER`    |
| 03  | **ARCHITECT** | Systems Architecture & Design Patterns           | `@ARCHITECT` |
| 04  |   **AXIOM**   | Pure Mathematics & Formal Proofs                 | `@AXIOM`     |
| 05  | **VELOCITY**  | Performance Optimization & Sub-Linear Algorithms | `@VELOCITY`  |

### Tier 2: Specialist Agents

| ID  |   Codename   | Specialization                           | Invocation  |
| :-: | :----------: | :--------------------------------------- | :---------- |
| 06  | **QUANTUM**  | Quantum Mechanics & Quantum Computing    | `@QUANTUM`  |
| 07  |  **TENSOR**  | Machine Learning & Deep Neural Networks  | `@TENSOR`   |
| 08  | **FORTRESS** | Defensive Security & Penetration Testing | `@FORTRESS` |
| 09  |  **NEURAL**  | Cognitive Computing & AGI Research       | `@NEURAL`   |
| 10  |  **CRYPTO**  | Blockchain & Distributed Systems         | `@CRYPTO`   |
| 11  |   **FLUX**   | DevOps & Infrastructure Automation       | `@FLUX`     |
| 12  |  **PRISM**   | Data Science & Statistical Analysis      | `@PRISM`    |
| 13  | **SYNAPSE**  | Integration Engineering & API Design     | `@SYNAPSE`  |
| 14  |   **CORE**   | Low-Level Systems & Compiler Design      | `@CORE`     |
| 15  |  **HELIX**   | Bioinformatics & Computational Biology   | `@HELIX`    |
| 16  | **VANGUARD** | Research Analysis & Literature Synthesis | `@VANGUARD` |
| 17  | **ECLIPSE**  | Testing, Verification & Formal Methods   | `@ECLIPSE`  |

### Tier 3: Innovator Agents

| ID  |  Codename   | Specialization                               | Invocation |
| :-: | :---------: | :------------------------------------------- | :--------- |
| 18  |  **NEXUS**  | Paradigm Synthesis & Cross-Domain Innovation | `@NEXUS`   |
| 19  | **GENESIS** | Zero-to-One Innovation & Novel Discovery     | `@GENESIS` |

### Tier 4: Meta Agents

| ID  |    Codename    | Specialization                                 | Invocation    |
| :-: | :------------: | :--------------------------------------------- | :------------ |
| 20  | **OMNISCIENT** | Meta-Learning Trainer & Evolution Orchestrator | `@OMNISCIENT` |

---

## 🎯 Usage Examples

### Single Agent Invocation

```
@APEX implement a distributed rate limiter
@CIPHER design JWT authentication with refresh tokens
@ARCHITECT design event-driven microservices
@TENSOR design CNN architecture for image classification
@FLUX design CI/CD pipeline for Kubernetes
```

### Multi-Agent Collaboration

```
@APEX @ARCHITECT design a scalable caching system
@CIPHER @FORTRESS security audit for this API
@TENSOR @VELOCITY optimize ML inference pipeline
@NEXUS @GENESIS novel approach to this problem
@OMNISCIENT coordinate analysis of this system
```

---

## 📁 Repository Structure

```
elite-agent-collective/
├── README.md
├── LICENSE
├── .github/
│   └── copilot-instructions.md      # GitHub Copilot instructions
├── vscode-prompts/
│   ├── ELITE_AGENT_COLLECTIVE.instructions.md
│   └── agents/
│       ├── APEX.instructions.md
│       ├── CIPHER.instructions.md
│       ├── ARCHITECT.instructions.md
│       ├── AXIOM.instructions.md
│       ├── VELOCITY.instructions.md
│       ├── QUANTUM.instructions.md
│       ├── TENSOR.instructions.md
│       ├── FORTRESS.instructions.md
│       ├── NEURAL.instructions.md
│       ├── CRYPTO.instructions.md
│       ├── FLUX.instructions.md
│       ├── PRISM.instructions.md
│       ├── SYNAPSE.instructions.md
│       ├── CORE.instructions.md
│       ├── HELIX.instructions.md
│       ├── VANGUARD.instructions.md
│       ├── ECLIPSE.instructions.md
│       ├── NEXUS.instructions.md
│       ├── GENESIS.instructions.md
│       └── OMNISCIENT.instructions.md
└── profiles/
    ├── TIER-1-FOUNDATIONAL/
    ├── TIER-2-SPECIALISTS/
    ├── TIER-3-INNOVATORS/
    └── TIER-4-META/
```

---

## 🏛️ System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        ELITE AGENT COLLECTIVE v1.0                          │
├─────────────────────────────────────────────────────────────────────────────┤
│  TIER 1: FOUNDATIONAL    │  TIER 2: SPECIALISTS     │  TIER 3-4: INNOVATORS│
│  ────────────────────    │  ──────────────────────  │  ───────────────────  │
│  @APEX    CS Engineering │  @QUANTUM  Quantum       │  @NEXUS   Synthesis   │
│  @CIPHER  Cryptography   │  @TENSOR   ML/DL         │  @GENESIS Innovation  │
│  @ARCHITECT Systems      │  @FORTRESS Security      │  @OMNISCIENT Meta     │
│  @AXIOM   Mathematics    │  @NEURAL   AGI Research  │                       │
│  @VELOCITY Performance   │  @CRYPTO   Blockchain    │                       │
│                          │  @FLUX     DevOps        │                       │
│                          │  @PRISM    Data Science  │                       │
│                          │  @SYNAPSE  Integration   │                       │
│                          │  @CORE     Low-Level     │                       │
│                          │  @HELIX    Bioinformtic  │                       │
│                          │  @VANGUARD Research      │                       │
│                          │  @ECLIPSE  Testing       │                       │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 🔄 Auto-Activation

Agents automatically activate based on context:

| Context                  | Primary Agents     |
| ------------------------ | ------------------ |
| Security files/code      | @CIPHER, @FORTRESS |
| Architecture discussions | @ARCHITECT         |
| Performance issues       | @VELOCITY          |
| ML/AI code               | @TENSOR, @NEURAL   |
| DevOps/infrastructure    | @FLUX              |
| Testing files            | @ECLIPSE           |
| API design               | @SYNAPSE           |
| Research questions       | @VANGUARD          |
| Novel problems           | @GENESIS, @NEXUS   |

---

## 📜 License

MIT License - feel free to use, modify, and distribute.

---

## 🤝 Contributing

Contributions welcome! Feel free to:

- Add new agent specializations
- Improve existing agent capabilities
- Share usage patterns and examples

---

_"The collective intelligence of specialized minds exceeds the sum of their parts."_
