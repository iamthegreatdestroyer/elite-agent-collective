# System Requirements

This page outlines the requirements for using the Elite Agent Collective with GitHub Copilot.

## Prerequisites

### Required

| Requirement | Details |
|-------------|---------|
| **GitHub Account** | Active GitHub account |
| **GitHub Copilot Subscription** | Individual, Business, or Enterprise |
| **Supported IDE or Platform** | See [Supported Platforms](#supported-platforms) below |

### Recommended

| Requirement | Details |
|-------------|---------|
| **Internet Connection** | Stable connection for Copilot API calls |
| **Latest IDE Version** | For best compatibility and features |
| **GitHub Copilot Extension** | Latest version installed |

---

## Supported Platforms

### IDEs

| IDE | Minimum Version | Copilot Extension Required |
|-----|-----------------|---------------------------|
| VS Code | 1.80+ | GitHub Copilot & GitHub Copilot Chat |
| JetBrains IntelliJ IDEA | 2023.1+ | GitHub Copilot Plugin |
| JetBrains PyCharm | 2023.1+ | GitHub Copilot Plugin |
| JetBrains WebStorm | 2023.1+ | GitHub Copilot Plugin |
| JetBrains Other IDEs | 2023.1+ | GitHub Copilot Plugin |
| Visual Studio | 2022 17.6+ | GitHub Copilot Extension |
| Neovim | 0.6+ | copilot.vim or copilot.lua |

### Web & Mobile

| Platform | Notes |
|----------|-------|
| GitHub.com | Copilot Chat in browser |
| GitHub Mobile | iOS and Android apps |
| GitHub Codespaces | Full VS Code experience in browser |

---

## GitHub Copilot Plans

The Elite Agent Collective works with all GitHub Copilot subscription tiers:

| Plan | Availability | Notes |
|------|--------------|-------|
| Copilot Individual | ✅ Full support | Personal use |
| Copilot Business | ✅ Full support | Organization-wide |
| Copilot Enterprise | ✅ Full support | Advanced features |
| Copilot Free (Limited) | ⚠️ Limited | May have usage restrictions |

---

## Operating System Compatibility

| OS | Status | Notes |
|----|--------|-------|
| Windows 10/11 | ✅ Supported | Full functionality |
| macOS 11+ | ✅ Supported | Full functionality |
| Linux (Ubuntu 20.04+) | ✅ Supported | Full functionality |
| Linux (Other distros) | ✅ Supported | May require manual setup |
| ChromeOS | ✅ Supported | Via Codespaces or web |

---

## Browser Requirements (for GitHub.com)

| Browser | Minimum Version |
|---------|-----------------|
| Chrome | 90+ |
| Firefox | 88+ |
| Safari | 14+ |
| Edge | 90+ |

---

## Network Requirements

### Connectivity

- **Outbound HTTPS (443)**: Required for Copilot API
- **WebSocket support**: For real-time chat features
- **No VPN restrictions**: Some corporate VPNs may block Copilot

### Firewall/Proxy Considerations

If you're behind a corporate firewall or proxy, ensure these domains are accessible:

- `github.com`
- `api.github.com`
- `copilot.github.com`
- `*.githubcopilot.com`

---

## Storage Requirements

| Component | Size |
|-----------|------|
| Instructions file | ~50 KB |
| VS Code prompts (full) | ~500 KB |
| Cache/temp files | ~10 MB |

---

## Performance Considerations

For optimal performance:

1. **Stable Internet**: Latency affects response time
2. **Adequate RAM**: At least 8 GB recommended for IDE
3. **Modern CPU**: For IDE responsiveness
4. **SSD Storage**: For faster file operations

---

## Limitations

### Known Limitations

- **Rate Limits**: Subject to GitHub Copilot rate limits
- **Context Window**: Large codebases may exceed context limits
- **Response Time**: Complex requests may take longer
- **Offline Usage**: Requires internet connection

### Enterprise Considerations

- **Compliance**: Review with your security team
- **Data Residency**: Copilot processes data per GitHub policies
- **Audit Logs**: Available with Enterprise plan

---

## Checking Your Setup

### Verify GitHub Copilot

1. Open your IDE
2. Check the Copilot icon is active (green checkmark)
3. Try a simple completion to verify connectivity

### Verify Agent Installation

After installing the Elite Agent Collective:

```
@APEX help
```

If APEX responds with its capabilities, your setup is complete.

---

## Next Steps

Once you've verified requirements:

1. Follow the [Installation Guide](installation.md)
2. Complete the [Quick Start](quick-start.md) tutorial
3. Explore the [User Guide](../user-guide/invoking-agents.md)
