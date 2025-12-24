# Elite Agent Collective - Marketplace & GitHub Integration Guide

## Overview

This guide covers the complete integration of Elite Agent Collective with GitHub Marketplace and GitHub Copilot. It includes marketplace setup, submission requirements, GitHub App configuration, and deployment.

## 🏪 GitHub Marketplace Setup

### Step 1: Create GitHub App

1. Navigate to `https://github.com/settings/apps/new`
2. Fill in the app configuration:

```json
{
  "name": "Elite Agent Collective",
  "description": "40 Specialized AI Agents for GitHub Copilot - Unlock expert-level assistance across all domains of software engineering",
  "homepage_url": "https://github.com/iamthegreatdestroyer/elite-agent-collective",
  "webhook_url": "https://your-domain.com/webhooks/github",
  "webhook_active": true,
  "webhook_secret": "generate-secure-random-string",
  "permissions": {
    "contents": "read",
    "metadata": "read",
    "issues": "read",
    "pull_requests": "read"
  },
  "events": ["issues", "pull_request", "push", "repository"],
  "user_authorization_required": false,
  "single_file_name": null,
  "public": true
}
```

### Step 2: Configure as Copilot Extension

1. In GitHub App settings, navigate to **"Copilot Extension"**
2. Set **Extension Type**: `Agent`
3. Set **Extension URL**: `https://your-domain.com/agent` (or `http://localhost:8080/agent` for local testing)
4. Set **Default Model**: `gpt-4` or `claude-3-opus-20240229`
5. Configure **Tool Capabilities**:
   ```json
   {
     "tools": [
       {
         "name": "APEX",
         "description": "Elite Computer Science Engineering",
         "parameters": {
           "type": "object",
           "properties": {
             "task": {
               "type": "string",
               "description": "The engineering task to perform"
             }
           },
           "required": ["task"]
         }
       }
       // ... (repeat for all 40 agents)
     ]
   }
   ```

### Step 3: Set Up Marketplace Listing

1. Navigate to **"Marketplace listing"** in GitHub App settings
2. Upload required assets:

   - **Icon**: 200x200 PNG (square, clear, recognizable)
   - **Logo**: 256x256 PNG (must be PNG, not JPG)
   - **Banner**: 1280x640 PNG (landscape, showing key features)
   - **Screenshots**: At least 3 (1280x640 recommended)

3. Fill in listing details:

   ```
   Short description:
   "40 Specialized AI Agents for GitHub Copilot - Expert assistance for any engineering challenge"

   Full description:
   (See marketplace/listing.md for full content)

   Category: Developer tools
   Pricing: Free

   Documentation URL: https://github.com/iamthegreatdestroyer/elite-agent-collective/tree/main/docs
   Support URL: https://github.com/iamthegreatdestroyer/elite-agent-collective/issues
   Privacy Policy URL: https://github.com/iamthegreatdestroyer/elite-agent-collective/blob/main/PRIVACY.md
   Terms of Service URL: https://github.com/iamthegreatdestroyer/elite-agent-collective/blob/main/LICENSE
   ```

### Step 4: Configure Authentication

#### Option A: OAuth 2.0 with GitHub

```
1. Redirect URI: https://your-domain.com/auth/github/callback
2. Client ID: (provided by GitHub)
3. Client Secret: (store securely in environment variables)
```

#### Option B: OIDC (Recommended)

```
1. OIDC Issuer: https://github.com
2. OIDC Audience: https://your-domain.com
3. OIDC Client ID: (from GitHub App)
4. OIDC Client Secret: (from GitHub App)
```

## 📋 Marketplace Submission Checklist

### Pre-Submission Requirements

**Assets** (Must be present in `marketplace/` directory):

- [x] `logo-256x256.png` - Application icon (256x256, transparent background)
- [x] `banner-1280x640.png` - Marketplace banner (1280x640)
- [x] `screenshot-1.png` - First screenshot (1280x640)
- [x] `screenshot-2.png` - Second screenshot (1280x640)
- [x] `screenshot-3.png` - Third screenshot (1280x640)
- [x] `listing.md` - Marketplace listing (see marketplace/listing.md)
- [x] `privacy-policy.md` - Privacy policy
- [x] `terms-of-service.md` - Terms of service

**Documentation** (Must be present in repository):

- [x] `README.md` - Main readme with features and quick start
- [x] `docs/QUICK_START.md` - 5-minute setup guide
- [x] `docs/INSTALLATION.md` - Detailed installation
- [x] `docs/API_REFERENCE.md` - API documentation
- [x] `CONTRIBUTING.md` - Contribution guidelines (if applicable)
- [x] `LICENSE` - MIT License or equivalent
- [x] `CHANGELOG.md` - Version history

**Configuration** (Must be configured in GitHub App):

- [x] `copilot-extension.json` - Extension configuration
- [x] `.github/copilot-instructions.md` - Agent instructions
- [x] GitHub App webhook configured and tested
- [x] Permissions set correctly (contents: read, metadata: read)
- [x] Support URL configured (issues page)
- [x] Webhook secret configured

### Testing Before Submission

**Functionality Testing**:

- [ ] All 40 agents load correctly
- [ ] Each agent responds to invocation
- [ ] Memory system (MNEMONIC) functions properly
- [ ] Error handling works for invalid inputs
- [ ] Rate limiting works correctly

**Platform Testing**:

- [ ] Test in VS Code with GitHub Copilot
- [ ] Test in JetBrains IDE with GitHub Copilot (if available)
- [ ] Test on GitHub.com Copilot Chat
- [ ] Test authentication flow (OAuth/OIDC)
- [ ] Test on macOS, Windows, Linux

**Integration Testing**:

- [ ] Webhook signature verification works
- [ ] Request body validation works
- [ ] Response formatting is correct
- [ ] Error responses are informative
- [ ] Multi-agent invocation works

**Security Testing**:

- [ ] No sensitive data in logs
- [ ] No hardcoded credentials
- [ ] Request validation prevents injection
- [ ] Rate limiting prevents abuse
- [ ] Authentication enforced on protected endpoints

### Submission Process

1. **Publish GitHub App** (if not already public)

   ```bash
   curl -X PATCH https://api.github.com/app \
     -H "Authorization: Bearer <token>" \
     -H "Accept: application/vnd.github+json" \
     -d '{"public": true}'
   ```

2. **Apply to GitHub Marketplace**

   - Navigate to GitHub App settings → "Marketplace listing"
   - Click "Apply to list on GitHub Marketplace"
   - Review terms and conditions
   - Submit application

3. **Review Process** (typically 3-7 business days)

   - GitHub reviews functionality
   - GitHub reviews security
   - GitHub reviews documentation
   - GitHub may request changes

4. **Address Feedback** (if any)

   - Respond to feedback within 7 days
   - Make requested changes
   - Resubmit for review
   - Iterate until approved

5. **Launch!**
   - Once approved, app appears in marketplace
   - Announce on social media, community, etc.
   - Gather feedback and iterate

## 🔐 GitHub App Permissions & Security

### Required Permissions

| Permission              | Scope | Purpose                                         |
| ----------------------- | ----- | ----------------------------------------------- |
| **Repository Contents** | Read  | Access agent definitions from `.github/agents/` |
| **Repository Metadata** | Read  | Access repo metadata for context                |
| **Issues**              | Read  | Context from issue discussions                  |
| **Pull Requests**       | Read  | Context from pull request discussions           |

### Optional Permissions (for extended features)

| Permission        | Scope | Purpose                        |
| ----------------- | ----- | ------------------------------ |
| **Workflows**     | Read  | Analyze CI/CD pipelines        |
| **Discussions**   | Read  | Access discussion context      |
| **Code Scanning** | Read  | Context from security scanning |

### Security Best Practices

1. **Webhook Secret**: Generate cryptographically secure secret

   ```bash
   openssl rand -hex 32
   ```

2. **Signature Verification**: Always verify webhook signatures

   ```go
   import "github.com/bradleyfalzon/ghinstallation"

   payload := []byte("webhook payload")
   signature := req.Header.Get("X-Hub-Signature-256")

   if !ghinstallation.ValidateSignature(payload, signature, secret) {
       http.Error(w, "Invalid signature", http.StatusForbidden)
       return
   }
   ```

3. **Token Management**:

   - Use GitHub App Installation tokens (short-lived)
   - Never store user tokens
   - Implement token refresh logic
   - Rotate secrets quarterly

4. **Rate Limiting**:

   - GitHub allows 5,000 API requests/hour per app
   - Implement exponential backoff
   - Cache responses when possible

5. **Audit Logging**:
   - Log all agent invocations
   - Include user (if authenticated), timestamp, agent, task
   - Store logs securely (encrypted at rest)
   - Retain logs for compliance requirements

## 🚀 Deployment to GitHub

### Repository Structure for Marketplace

```
elite-agent-collective/
├── .github/
│   ├── agents/                 # Agent definitions
│   ├── copilot-instructions.md # Copilot system prompt
│   ├── workflows/              # CI/CD workflows
│   └── ISSUE_TEMPLATE/         # Issue templates
├── backend/                    # Go server
├── docs/                       # Documentation
├── marketplace/                # Marketplace assets
│   ├── listing.md
│   ├── privacy-policy.md
│   ├── terms-of-service.md
│   ├── screenshots/
│   └── SUBMISSION_CHECKLIST.md
├── vscode-prompts/            # VS Code integration
├── copilot-extension.json     # Extension manifest
├── README.md                  # Main readme
├── LICENSE                    # MIT License
├── CHANGELOG.md              # Version history
└── CONTRIBUTING.md           # Contribution guidelines
```

### CI/CD Pipeline

Create `.github/workflows/marketplace-validation.yml`:

```yaml
name: Marketplace Validation

on:
  push:
    paths:
      - "copilot-extension.json"
      - "marketplace/**"
      - ".github/agents/**"
      - ".github/copilot-instructions.md"
  pull_request:
    paths:
      - "copilot-extension.json"
      - "marketplace/**"
      - ".github/agents/**"
      - ".github/copilot-instructions.md"

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Validate copilot-extension.json
        run: |
          jq . copilot-extension.json

      - name: Check marketplace assets
        run: |
          test -f marketplace/listing.md
          test -f marketplace/privacy-policy.md
          test -f marketplace/terms-of-service.md
          test -f marketplace/logo-256x256.png
          test -f marketplace/banner-1280x640.png

      - name: Validate agents
        run: |
          count=$(ls .github/agents/*.agent.md | wc -l)
          if [ "$count" -ne 40 ]; then
            echo "Expected 40 agents, found $count"
            exit 1
          fi

      - name: Check documentation
        run: |
          test -f docs/QUICK_START.md
          test -f docs/INSTALLATION.md
          test -f docs/API_REFERENCE.md
          test -f README.md
          test -f LICENSE
```

### Version Management

Use semantic versioning and tag releases:

```bash
# Create release
git tag -a v2.0.0 -m "Release v2.0.0: Initial marketplace launch"
git push origin v2.0.0

# GitHub automatically creates release from tag
```

Update `copilot-extension.json` version:

```json
{
  "version": "2.0.0"
}
```

## 📊 Monitoring & Analytics

After marketplace launch, track:

1. **Installation metrics**

   - Total installations
   - Daily active installations
   - Uninstall rate

2. **Usage metrics**

   - Invocations per day
   - Agent popularity (which agents used most)
   - Average response satisfaction
   - Errors and failures

3. **Performance metrics**

   - Average response time
   - 95th percentile latency
   - Memory usage
   - Error rate

4. **User feedback**
   - GitHub issues and discussions
   - Ratings and reviews (if marketplace provides)
   - Feature requests
   - Bug reports

## 🔄 Maintenance & Updates

### Release Cycle

- **Bug fixes**: Release as patch (2.0.1, 2.0.2, etc.)
- **New features**: Release as minor (2.1.0)
- **Breaking changes**: Release as major (3.0.0)

### Update Process

1. Make code changes and test
2. Update `CHANGELOG.md`
3. Update version in `copilot-extension.json`
4. Create git tag: `git tag -a v2.0.1 -m "..."`
5. Push: `git push origin v2.0.1`
6. GitHub releases automatically from tag
7. Monitor for issues in new release

### Support & Communication

- **Announcements**: GitHub releases, Twitter, blog
- **Support**: GitHub issues with templates
- **Discussions**: GitHub discussions for feature requests
- **Feedback**: Regular community surveys

## Additional Resources

- [GitHub Marketplace Developer Guide](https://docs.github.com/en/developers/github-marketplace/github-marketplace-overview)
- [GitHub Apps Documentation](https://docs.github.com/en/developers/apps)
- [GitHub Copilot Extension Guide](https://docs.github.com/en/copilot/building-copilot-extensions)
- [Model Context Protocol (MCP)](https://modelcontextprotocol.io/)
- [Repository: elite-agent-collective](https://github.com/iamthegreatdestroyer/elite-agent-collective)

## Troubleshooting

### App Not Appearing in Marketplace

1. Verify app is public: `Settings → General → Visibility`
2. Check marketplace listing is complete and approved
3. Ensure all required assets are present
4. Verify no policy violations

### Installation Issues

1. Check GitHub App permissions
2. Verify webhook is configured
3. Check authentication method (OAuth vs OIDC)
4. Review error logs for detailed failures

### Integration Issues

1. Verify GitHub app token is valid
2. Check API rate limiting
3. Validate webhook payload signatures
4. Review GitHub API responses

For additional help, see the [Troubleshooting Guide](../docs/TROUBLESHOOTING.md).
