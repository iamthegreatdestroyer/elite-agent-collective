# Common Issues

Solutions to frequently encountered problems with the Elite Agent Collective.

---

## Installation Issues

### Agent Not Responding

**Symptoms:**
- Agent invocation returns generic Copilot response
- No specialized agent behavior observed

**Solutions:**

1. **Verify Installation Location**

   Check that the instructions file is in the correct location:

   **Global (Home Directory):**
   ```bash
   # macOS/Linux
   ls ~/.github/copilot-instructions.md
   
   # Windows
   dir %USERPROFILE%\.github\copilot-instructions.md
   ```

   **Repository:**
   ```bash
   ls .github/copilot-instructions.md
   ```

2. **Restart Your IDE**
   
   After installing, restart VS Code or your JetBrains IDE to reload configurations.

3. **Verify GitHub Copilot is Active**
   
   - Check Copilot icon in status bar
   - Ensure you have an active subscription
   - Try a regular Copilot completion to verify it's working

4. **Check File Syntax**
   
   Ensure the instructions file hasn't been corrupted:
   - Open in a text editor
   - Verify Markdown formatting is intact
   - Check for encoding issues

---

### VS Code Prompts Not Loading

**Symptoms:**
- Individual agent files not appearing in VS Code
- Agents not available in chat

**Solutions:**

1. **Verify File Location**

   ```bash
   # macOS
   ls ~/Library/Application\ Support/Code/User/prompts/agents/
   
   # Windows
   dir %APPDATA%\Code\User\prompts\agents\
   
   # Linux
   ls ~/.config/Code/User/prompts/agents/
   ```

2. **Check File Extension**
   
   Files must end with `.instructions.md`:
   ```
   APEX.instructions.md ✅
   APEX.md ❌
   ```

3. **Verify VS Code Version**
   
   User prompts require VS Code 1.80+. Update if needed.

4. **Reload Window**
   
   Press `Ctrl+Shift+P` / `Cmd+Shift+P` and run "Reload Window"

---

## Usage Issues

### Wrong Agent Responds

**Symptoms:**
- Invoked @APEX but got generic response
- Agent seems confused about its role

**Solutions:**

1. **Check Agent Name Spelling**
   
   Agent names are case-insensitive but must be exact:
   ```
   @APEX ✅
   @Apex ✅
   @APE ❌
   ```

2. **Verify Instructions Are Loaded**
   
   Test with: `@APEX describe your capabilities`
   
   If it doesn't respond with APEX-specific info, reinstall.

3. **Clear Copilot Cache**
   
   - Close all IDE windows
   - Clear VS Code cache if issues persist
   - Restart IDE

---

### Agent Gives Poor Quality Responses

**Symptoms:**
- Responses don't match agent expertise
- Missing specialized knowledge

**Solutions:**

1. **Be More Specific**
   
   ❌ `@APEX help me`
   ✅ `@APEX implement a thread-safe LRU cache in Python`

2. **Provide Context**
   
   Include:
   - Current technology stack
   - Constraints and requirements
   - What you've already tried

3. **Use the Right Agent**
   
   Consult the [Agent Reference](../user-guide/agent-reference.md) to find the best agent for your task.

4. **Break Down Complex Requests**
   
   ❌ `@APEX build me a complete authentication system`
   ✅ Start with: `@APEX design the authentication flow`
   Then: `@APEX implement password hashing`

---

### Multi-Agent Collaboration Not Working

**Symptoms:**
- Only one agent responds
- Response doesn't combine expertise

**Solutions:**

1. **Use Sequential Requests**
   
   Instead of:
   ```
   @APEX @ARCHITECT @ECLIPSE complete system design
   ```
   
   Try:
   ```
   @ARCHITECT design the architecture
   @APEX implement the core
   @ECLIPSE add tests
   ```

2. **Be Explicit About Collaboration**
   
   ```
   @APEX @ARCHITECT I need both architecture (ARCHITECT) 
   and implementation (APEX) perspectives on this caching system
   ```

---

## Context Issues

### Agent Ignores Selected Code

**Symptoms:**
- Selected code not referenced in response
- Agent responds generically

**Solutions:**

1. **Verify Code is Selected**
   
   Ensure code is highlighted before invoking agent.

2. **Reference Selection Explicitly**
   
   ```
   @APEX optimize the selected function
   ```
   
   Or paste the code:
   ```
   @APEX optimize this:
   \`\`\`python
   [paste code here]
   \`\`\`
   ```

3. **Check File Size**
   
   Very large files may exceed context limits. Select specific sections.

---

### Agent Doesn't See File Changes

**Symptoms:**
- Agent references old code
- Suggestions don't match current state

**Solutions:**

1. **Save the File**
   
   Ensure all changes are saved before invoking agent.

2. **Use Fresh Context**
   
   Start a new chat session for updated context.

3. **Reference Specific Lines**
   
   ```
   @APEX review the function at line 45-60
   ```

---

## Performance Issues

### Slow Responses

**Symptoms:**
- Long wait times for agent responses
- Timeouts

**Solutions:**

1. **Check Internet Connection**
   
   Copilot requires stable internet.

2. **Simplify Request**
   
   Break complex requests into smaller parts.

3. **Check GitHub Status**
   
   Visit [status.github.com](https://status.github.com) for service issues.

---

### Rate Limiting

**Symptoms:**
- Responses suddenly stop
- Error messages about limits

**Solutions:**

1. **Wait and Retry**
   
   Rate limits reset after a short period.

2. **Reduce Request Frequency**
   
   Batch multiple questions into single requests.

3. **Check Subscription**
   
   Verify your Copilot subscription is active.

---

## Platform-Specific Issues

### VS Code

**Copilot Chat Panel Missing:**
1. Install "GitHub Copilot Chat" extension
2. Reload VS Code
3. Check View → Copilot Chat

**Extension Conflicts:**
- Disable other AI assistants temporarily
- Check for extension updates

### JetBrains IDEs

**Plugin Not Working:**
1. Update to latest IDE version
2. Reinstall Copilot plugin
3. Clear IDE cache: `File → Invalidate Caches`

### GitHub.com

**Copilot Chat Not Available:**
- Ensure you have Copilot subscription
- Check if feature is enabled for your account
- Try a different browser

---

## Getting More Help

If these solutions don't resolve your issue:

1. **Check GitHub Status**: [status.github.com](https://status.github.com)

2. **Search Existing Issues**: [GitHub Issues](https://github.com/iamthegreatdestroyer/elite-agent-collective/issues)

3. **Open a New Issue**: Include:
   - Agent(s) involved
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (IDE, OS, versions)

4. **See [Support](support.md)** for additional help options.

---

## Quick Fixes Checklist

- [ ] GitHub Copilot is active and authenticated
- [ ] Instructions file is in correct location
- [ ] IDE has been restarted after installation
- [ ] Agent name is spelled correctly
- [ ] Request is specific and clear
- [ ] File is saved before invocation
- [ ] Internet connection is stable
