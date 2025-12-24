# Marketplace Directory - Guide

This directory contains all assets and documentation required for GitHub Marketplace listing and submission.

## 📁 Directory Structure

```
marketplace/
├── README.md                      # This file
├── listing.md                     # Marketplace listing description
├── privacy-policy.md              # Privacy policy for marketplace
├── terms-of-service.md            # Terms of service
├── SUBMISSION_CHECKLIST.md         # Pre-submission checklist
├── ICON_GUIDELINES.md             # Icon creation guidelines
├── BANNER_GUIDELINES.md           # Banner creation guidelines
└── screenshots/
    ├── README.md                  # Screenshot guidelines
    ├── 1-agent-invocation.png    # Example: Invoking an agent
    ├── 2-multi-agent-collab.png  # Example: Multi-agent collaboration
    └── 3-response-example.png    # Example: Agent response
```

## 🎯 Quick Start

### For Marketplace Submission:

1. **Review Requirements**: Check [SUBMISSION_CHECKLIST.md](SUBMISSION_CHECKLIST.md)
2. **Prepare Assets**:
   - Icon (256×256 PNG)
   - Banner (1280×640 PNG)
   - Screenshots (1280×640 PNG, minimum 3)
3. **Review Policies**:
   - Read [privacy-policy.md](privacy-policy.md)
   - Read [terms-of-service.md](terms-of-service.md)
4. **Submit**: Follow [../GITHUB_MARKETPLACE_GUIDE.md](../GITHUB_MARKETPLACE_GUIDE.md)

### For Development/Updates:

1. **Update Listing**: Edit [listing.md](listing.md) if features change
2. **Update Policies**: Keep privacy policy and ToS current
3. **Update Screenshots**: Capture new screenshots if UI changes
4. **Test Locally**: Use [../MCP_SERVER_CONFIG.md](../MCP_SERVER_CONFIG.md) for local testing

## 📝 Asset Specifications

### Icon (logo-256x256.png)

**Specifications:**

- Size: 256×256 pixels
- Format: PNG with transparency
- Background: Transparent or solid color
- Style: Clear, recognizable, simple
- Complexity: Avoid fine details (readable at 32×32)

**Design Tips:**

- Show the collective/agent concept
- Use distinct colors for branding
- Ensure clarity at small sizes
- Include text only if critical to identity

**Examples of good icons:**

- Abstract nodes/connections (representing agent network)
- Brain/intelligence symbol
- Stacked agents or team concept
- Gear with neural elements

### Banner (banner-1280x640.png)

**Specifications:**

- Size: 1280×640 pixels (2:1 aspect ratio)
- Format: PNG
- Background: Can be gradient or image
- Placement: Safe area 100px from edges

**Design Tips:**

- Highlight key benefit (e.g., "40 Expert Agents")
- Show sample agents (e.g., @APEX @TENSOR @CIPHER)
- Include call-to-action (e.g., "Get Started", "Learn More")
- Maintain brand consistency with icon
- Ensure text is readable at 50% opacity

**Suggested Layout:**

```
┌────────────────────────────────────────────┐
│ 100px margin  Elite Agent Collective  100px│
│               40 Specialized AI Agents      │
│               for GitHub Copilot           │
│                                            │
│ @APEX  @TENSOR  @CIPHER  @ARCHITECT       │
│                                            │
│              🚀 Get Started 🚀             │
└────────────────────────────────────────────┘
```

### Screenshots (1280×640 PNG)

**Specifications:**

- Size: 1280×640 pixels minimum
- Format: PNG
- Count: Minimum 3, recommended 5
- Content: Real product screenshots

**Suggested Screenshots:**

1. **Agent Invocation** (screenshot-1-agent.png)

   - Show: Copilot chat window with agent invocation
   - Example: `@APEX implement rate limiter`
   - Show: Agent response in chat
   - Highlight: Quick access to specialized agent

2. **Multi-Agent Collaboration** (screenshot-2-multi.png)

   - Show: Multiple agents invoked together
   - Example: `@CIPHER @FORTRESS security audit`
   - Show: Different agent perspectives
   - Highlight: Collaboration capabilities

3. **Agent Response** (screenshot-3-response.png)

   - Show: Detailed agent response with code
   - Show: Code quality and completeness
   - Show: Memory system context (if visible)
   - Highlight: Production-quality output

4. **Agent Browser** (optional-screenshot-4.png)

   - Show: List of 40 agents with descriptions
   - Show: Tier organization (Foundational, Specialist, etc.)
   - Highlight: Breadth of capabilities

5. **Integration** (optional-screenshot-5.png)
   - Show: Multiple platforms (VS Code, JetBrains, GitHub.com)
   - Show: Consistent experience across platforms
   - Highlight: Everywhere you code

**Screenshot Creation Guide:**

```bash
# 1. Run local server
cd backend && make run

# 2. Open VS Code with GitHub Copilot extension
# 3. Invoke @APEX with a task
# 4. Capture window (avoid sensitive data)
# 5. Crop/resize to 1280×640
# 6. Add 20px padding with branded color
# 7. Add caption/annotation if helpful
```

**Screenshot Best Practices:**

- ✅ Show real product (not mockups)
- ✅ Use realistic but clean examples
- ✅ Avoid sensitive data or credentials
- ✅ Ensure text is readable
- ✅ Maintain brand consistency
- ✅ Show value immediately
- ❌ Don't use outdated interface
- ❌ Don't include personal information
- ❌ Don't use blurry or pixelated images

## 📄 Asset Files

### listing.md

**Purpose**: Complete marketplace listing description  
**Content**:

- Product name and tagline
- Feature overview with tier table
- Key capabilities with examples
- Multi-agent collaboration examples
- Privacy & security statement
- Permissions explanation
- Platform support
- Documentation links
- Support information
- License

**Editing**: Update whenever:

- New agents added
- Features changed
- Capabilities expanded
- Support channels change

### privacy-policy.md

**Purpose**: Legal privacy policy for users  
**Content**:

- Data collection practices
- Data usage & storage
- GDPR/CCPA compliance
- Cookie usage (if applicable)
- Third-party services
- Data retention policies
- User rights
- Contact information

**Legal Review**: Should be reviewed by legal counsel before publication

### terms-of-service.md

**Purpose**: Legal terms for using the extension  
**Content**:

- Acceptable use policy
- Restrictions and limitations
- Liability limitations
- Warranty disclaimers
- Termination conditions
- Dispute resolution
- Changes to terms
- Jurisdiction

**Legal Review**: Should be reviewed by legal counsel before publication

## ✅ Pre-Submission Checklist

### Assets Required

- [ ] Icon: 256×256 PNG
- [ ] Banner: 1280×640 PNG
- [ ] Screenshots: 3+ (1280×640 PNG each)
- [ ] Listing description (in listing.md)
- [ ] Privacy policy (in privacy-policy.md)
- [ ] Terms of service (in terms-of-service.md)

### Repository Structure

- [ ] All assets in `marketplace/` directory
- [ ] `/copilot-extension.json` at root
- [ ] `/.github/agents/*.agent.md` (40 agents)
- [ ] `/.github/copilot-instructions.md`
- [ ] `/docs/` directory with docs
- [ ] `/README.md` at root
- [ ] `/LICENSE` file
- [ ] `/CHANGELOG.md` file

### Policy Compliance

- [ ] Privacy policy is current and accurate
- [ ] Terms of service reviewed by legal
- [ ] No prohibited content in listings
- [ ] No misleading claims about capabilities
- [ ] Links all work and are current

### Testing Complete

- [ ] All agents load and respond
- [ ] Extension works in VS Code
- [ ] Extension works in JetBrains
- [ ] Extension works on GitHub.com
- [ ] No error messages
- [ ] Rate limiting works
- [ ] Authentication works

### Documentation Complete

- [ ] QUICK_START.md written
- [ ] INSTALLATION.md complete
- [ ] API_REFERENCE.md documented
- [ ] Examples provided in multiple languages
- [ ] Troubleshooting guide available
- [ ] Architecture documented

## 🔄 Updating Assets

### Icon Update

```bash
# Replace icon
cp new-icon-256x256.png marketplace/logo-256x256.png

# Verify size
identify marketplace/logo-256x256.png
# Should show: 256x256 PNG

# Commit
git add marketplace/logo-256x256.png
git commit -m "assets: update application icon"
```

### Banner Update

```bash
# Replace banner
cp new-banner-1280x640.png marketplace/banner-1280x640.png

# Verify size
identify marketplace/banner-1280x640.png
# Should show: 1280x640 PNG

# Commit
git add marketplace/banner-1280x640.png
git commit -m "assets: update marketplace banner"
```

### Screenshot Update

```bash
# Replace screenshot
cp new-screenshot.png marketplace/screenshots/1-agent-invocation.png

# Verify size (should be 1280x640 minimum)
identify marketplace/screenshots/1-agent-invocation.png

# Commit
git add marketplace/screenshots/
git commit -m "assets: update marketplace screenshots"
```

### Listing Update

```bash
# Edit listing
vim marketplace/listing.md

# Preview (if you have markdown viewer)
# Review for accuracy

# Commit
git add marketplace/listing.md
git commit -m "docs: update marketplace listing description"
```

## 📞 Support

For questions about marketplace submission:

1. **GitHub Docs**: [GitHub Marketplace Developer Guide](https://docs.github.com/en/developers/github-marketplace)
2. **GitHub Support**: Contact GitHub support for submission questions
3. **Repository Issues**: File issues at the main repository
4. **Documentation**: See [../docs/](../docs/) directory

## 🚀 Next Steps

1. **Prepare Assets**: Follow asset specifications above
2. **Review Checklist**: Complete [SUBMISSION_CHECKLIST.md](SUBMISSION_CHECKLIST.md)
3. **Follow Guide**: Use [../GITHUB_MARKETPLACE_GUIDE.md](../GITHUB_MARKETPLACE_GUIDE.md)
4. **Test Thoroughly**: Follow testing procedures before submission
5. **Submit**: Follow GitHub's submission process
6. **Iterate**: Address any feedback and resubmit if needed

---

**Last Updated**: December 11, 2025  
**Version**: 2.0.0  
**Status**: Ready for Marketplace Submission
