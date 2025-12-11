# Troubleshooting Guide

Solutions to common problems with the Elite Agent Collective.

## Backend Server Issues

### Server Won't Start

**Problem:** `go run ./cmd/server` fails

**Solutions:**

1. **Check Go Version**

   ```bash
   go version
   # Must be 1.21 or higher
   ```

   → Update Go if needed

2. **Check Dependencies**

   ```bash
   cd backend
   go mod tidy
   go mod download
   ```

   → Reinstall dependencies

3. **Check for Typos**

   ```bash
   go build -v ./cmd/server
   ```

   → Look for compilation errors

4. **Check Port**
   ```bash
   lsof -i :8080  # Unix/Mac
   netstat -ano | findstr :8080  # Windows
   ```
   → Use different port if needed:
   ```bash
   PORT=9090 go run ./cmd/server
   ```

### Server Crashes on Startup

**Problem:** Server starts then immediately exits

**Solutions:**

1. **Check Logs**

   ```bash
   LOG_LEVEL=debug go run ./cmd/server
   ```

   → Look for error messages

2. **Verify Agent Directory**

   ```bash
   ls -la .github/agents/ | wc -l
   # Should show 40
   ```

   → If not 40, check git status: `git status`

3. **Check YAML Files**

   ```bash
   # Test YAML parsing
   cat .github/agents/APEX.agent.md | head -5
   ```

   → Look for syntax errors in frontmatter

4. **Check Working Directory**
   ```bash
   pwd  # Should be in backend/ directory
   cd backend
   go run ./cmd/server
   ```

### Agents Not Loading

**Problem:** "Loaded 0 agents" in logs

**Solutions:**

1. **Verify Agent Files Exist**

   ```bash
   ls ../.github/agents/ | grep -c .agent.md
   # Should be 40
   ```

2. **Check Directory Path**

   ```bash
   # Look at logs to see what path was tried
   LOG_LEVEL=debug go run ./cmd/server
   # Should show: "Loaded 40 agents from C:\..."
   ```

3. **Verify File Format**

   ```bash
   # Check for YAML frontmatter
   head -3 ../.github/agents/APEX.agent.md
   # Should start with: ---
   ```

4. **Check YAML Syntax**

   ```bash
   # Install yq tool: https://github.com/mikefarah/yq
   yq eval '.codename' ../.github/agents/APEX.agent.md
   # Should output: APEX
   ```

5. **Force Fallback to Hardcoded**
   - If directory not found, system falls back to hardcoded definitions
   - Check logs: "Loaded from hardcoded definitions"
   - This is normal if .github/agents/ not found

### Agents Loading but API Fails

**Problem:** GET /agents returns empty or error

**Solutions:**

1. **Check Server is Running**

   ```bash
   curl http://localhost:8080/health
   # Should return: {"status":"healthy", "agents_loaded":40, ...}
   ```

   → If fails, server not running

2. **Verify Agent Handler**

   ```bash
   # Check specific agent
   curl http://localhost:8080/agents/APEX
   # Should return agent info
   ```

   → If 404, agent not registered

3. **Check Registry**
   ```bash
   go test -v ./internal/agents/...
   # Look for: "Loaded 40 agents"
   ```
   → If test fails, check test output for details

## Agent-Related Issues

### Agent Returns Incomplete Information

**Problem:** Agent missing fields (Examples, Philosophy, etc.)

**Solutions:**

1. **Check Agent File Format**

   ```bash
   # Agent file must have proper sections
   cat .github/agents/APEX.agent.md | grep "Philosophy"
   cat .github/agents/APEX.agent.md | grep "Invocation"
   ```

2. **Verify Markdown Headers**

   ```bash
   # Must have exact headers for extraction
   grep "## Philosophy" ../.github/agents/*.agent.md
   grep "## Invocation Examples" ../.github/agents/*.agent.md
   ```

   → Update if headers don't match

3. **Check Regex Patterns**
   - Agent loader uses regex to extract sections
   - See `agent_loader.go` for patterns
   - If headers are different, patterns won't match

### Can't Invoke Agent

**Problem:** POST /agent returns error

**Solutions:**

1. **Check Request Format**

   ```bash
   curl -X POST http://localhost:8080/agent \
     -H "Content-Type: application/json" \
     -d '{
       "agent": "APEX",
       "messages": [{"role": "user", "content": "test"}]
     }'
   ```

   → Must be valid JSON

2. **Verify Agent Exists**

   ```bash
   curl http://localhost:8080/agents/APEX
   # Should return agent info
   ```

   → If 404, agent doesn't exist

3. **Check Request Body**

   ```bash
   # Required fields:
   # - agent (string, uppercase)
   # - messages (array of {role, content})
   ```

4. **Enable Request Logging**
   ```bash
   LOG_LEVEL=debug go run ./cmd/server
   # Will show request details
   ```

## Testing Issues

### Tests Fail

**Problem:** `go test` returns failures

**Solutions:**

1. **Run with Verbose Output**

   ```bash
   go test -v ./internal/agents/...
   ```

   → See detailed error messages

2. **Run Specific Test**

   ```bash
   go test -v -run TestRegistryGet ./internal/agents/...
   ```

3. **Check Dependencies**

   ```bash
   go mod tidy
   go test ./...
   ```

4. **Clear Test Cache**
   ```bash
   go clean -testcache
   go test ./...
   ```

### Tests Timeout

**Problem:** Tests hang indefinitely

**Solutions:**

1. **Set Timeout**

   ```bash
   go test -timeout 30s ./...
   ```

2. **Run with Verbose**

   ```bash
   go test -v -timeout 30s ./...
   ```

   → See which test times out

3. **Check for Deadlocks**
   - Tests may be waiting for goroutine
   - Check test code for unbuffered channels
   - Review registry locking logic

## Build Issues

### Build Fails with Compilation Error

**Problem:** `go build` returns error

**Solutions:**

1. **Check Error Message**

   ```bash
   go build -v ./cmd/server
   # Read the error carefully
   ```

2. **Common Errors:**

   **Unused Variable:**

   ```bash
   # Remove unused: variableName := unused
   # Don't use: _ = unused (this is fine)
   ```

   **Undefined Function/Type:**

   ```bash
   # Check spelling
   # Check package imports
   go mod tidy
   ```

   **Type Mismatch:**

   ```bash
   # Check function signatures
   # May need type conversion
   ```

3. **Format Code**
   ```bash
   go fmt ./...
   go vet ./...
   ```

### Dependencies Missing

**Problem:** `go: missing go.sum entry`

**Solutions:**

1. **Download Dependencies**

   ```bash
   go mod download
   go mod tidy
   ```

2. **Clear and Rebuild**

   ```bash
   rm go.sum
   go mod download
   ```

3. **Check go.mod**
   ```bash
   cat go.mod | grep yaml
   # Should show: gopkg.in/yaml.v3
   ```

## Docker Issues

### Docker Build Fails

**Problem:** `docker build` error

**Solutions:**

1. **Check Dockerfile**

   ```bash
   cat Dockerfile
   # Should have proper Go build stages
   ```

2. **Build with Verbose**

   ```bash
   docker build -t elite-agent:debug --no-cache .
   # --no-cache forces rebuild
   ```

3. **Check Docker Version**
   ```bash
   docker --version
   # Should be recent version
   ```

### Container Won't Start

**Problem:** `docker run` fails

**Solutions:**

1. **Check Logs**

   ```bash
   docker logs <container_id>
   # See error messages
   ```

2. **Run Interactively**

   ```bash
   docker run -it elite-agent:latest /bin/bash
   # Start shell to debug
   ```

3. **Check Port Binding**
   ```bash
   docker run -p 8080:8080 elite-agent:latest
   # Ensure port is exposed
   ```

### Docker Compose Fails

**Problem:** `docker-compose up` error

**Solutions:**

1. **Check File Format**

   ```bash
   cat docker-compose.yml
   # Valid YAML syntax?
   ```

2. **Check Services**

   ```bash
   docker-compose ps
   # See what services exist
   ```

3. **View Logs**
   ```bash
   docker-compose logs -f
   # Follow service logs
   ```

## API Integration Issues

### Cross-Origin (CORS) Error

**Problem:** Browser CORS error when calling API

**Solutions:**

1. **Check CORS Headers** - Backend may need CORS middleware

2. **Development Workaround**

   - Use curl or Postman instead
   - Use VS Code REST Client extension
   - Disable CORS in browser (for development only)

3. **Production Solution**
   - Backend adds CORS middleware
   - Frontend proxies requests
   - Use API gateway

### Rate Limiting

**Problem:** 429 Too Many Requests

**Solutions:**

1. **Wait and Retry**

   ```bash
   # Check retry-after header
   curl -i http://localhost:8080/agent
   # Look for: Retry-After: 60
   ```

2. **Implement Backoff**

   ```bash
   # Exponential backoff for retries
   # Wait 1s, then 2s, then 4s, etc.
   ```

3. **Reduce Request Rate**
   - Space out requests
   - Batch related requests

## Security Issues

### Authentication Failed

**Problem:** 401 Unauthorized error

**Solutions:**

1. **Check Token**

   ```bash
   # Verify OIDC token is valid
   echo $OIDC_TOKEN | jq .
   ```

2. **Check Environment**

   ```bash
   env | grep OIDC
   # Verify variables set
   ```

3. **Debug OIDC**
   ```bash
   LOG_LEVEL=debug go run ./cmd/server
   # Check OIDC validation logs
   ```

### Signature Verification Failed

**Problem:** GitHub Copilot webhook signature invalid

**Solutions:**

1. **Check Secret**

   ```bash
   echo $GITHUB_WEBHOOK_SECRET
   # Must match GitHub configured secret
   ```

2. **Verify Request**

   - Check X-Hub-Signature-256 header
   - Verify body hasn't been modified

3. **Debug Signature**
   ```bash
   LOG_LEVEL=debug go run ./cmd/server
   # Check signature validation logs
   ```

## Performance Issues

### Slow Agent Responses

**Problem:** Agent takes >30 seconds to respond

**Solutions:**

1. **Check Backend Load**

   ```bash
   top  # Unix/Mac
   Task Manager  # Windows
   # Check CPU and memory usage
   ```

2. **Check Network**

   ```bash
   curl -w "@curl-format.txt" http://localhost:8080/agents
   # Check response times
   ```

3. **Check LLM Service**

   - May be slow external API call
   - Check Anthropic API status
   - Try test request to LLM service

4. **Profile Performance**
   ```bash
   go test -cpuprofile=cpu.prof ./...
   go tool pprof cpu.prof
   ```

### Memory Leak

**Problem:** Memory usage keeps growing

**Solutions:**

1. **Monitor Memory**

   ```bash
   watch -n 1 'ps aux | grep server'
   # Check RSS memory column
   ```

2. **Check for Goroutine Leaks**

   ```go
   // In test:
   before := runtime.NumGoroutine()
   // ... run code
   after := runtime.NumGoroutine()
   if after > before {
       t.Fatalf("Goroutine leak: %d -> %d", before, after)
   }
   ```

3. **Profile Memory**
   ```bash
   go test -memprofile=mem.prof ./...
   go tool pprof mem.prof
   ```

## Miscellaneous

### File Permission Denied

**Problem:** Can't read/write files

**Solutions:**

1. **Check Permissions**

   ```bash
   ls -la .github/agents/
   # Should be readable
   ```

2. **Fix Permissions** (Unix/Mac)

   ```bash
   chmod 755 .github/agents/
   chmod 644 .github/agents/*.agent.md
   ```

3. **Check Owner**
   ```bash
   chown -R $USER .github/agents/
   ```

### Git Issues

**Problem:** Git command fails

**Solutions:**

1. **Check Git Status**

   ```bash
   git status
   # See what's changed
   ```

2. **Restore Agent Files**

   ```bash
   git checkout .github/agents/
   # Restore from git
   ```

3. **Clean Repository**
   ```bash
   git clean -fd
   # Remove untracked files
   ```

### Configuration Issues

**Problem:** Config values not being used

**Solutions:**

1. **Check Environment Variables**

   ```bash
   env | grep PORT
   env | grep LOG_LEVEL
   ```

2. **Check Config File**

   ```bash
   cat backend/config.yaml
   ```

3. **Priority Order** (highest to lowest):
   1. Command-line flags
   2. Environment variables
   3. Config files
   4. Defaults

## Getting Help

### Resources

- **Logs** - Enable `LOG_LEVEL=debug` for detailed output
- **Tests** - Run tests to verify components: `go test -v ./...`
- **Documentation** - Check [docs/](../) directory
- **GitHub Issues** - Search existing issues or create new
- **GitHub Discussions** - Ask questions in community

### Reporting Issues

When reporting a problem, include:

```markdown
## Environment

- Go version: (go version)
- OS: (Windows/Mac/Linux)
- Agent Collective version: (git log -1)

## Problem Description

[Detailed description]

## Steps to Reproduce

1. [Step 1]
2. [Step 2]
3. [Step 3]

## Expected Behavior

[What should happen]

## Actual Behavior

[What actually happens]

## Logs

[Relevant log output]

## Error Messages

[Full error messages]
```

### Common Solutions Checklist

- [ ] Go version is 1.21+
- [ ] Dependencies installed: `go mod download`
- [ ] Code formatted: `go fmt ./...`
- [ ] Tests pass: `go test ./...`
- [ ] Server builds: `go build -v ./cmd/server`
- [ ] Agent files present: `ls .github/agents | wc -l` (should be 40)
- [ ] YAML valid: Check file format
- [ ] Log level set to debug: `LOG_LEVEL=debug`
- [ ] Port not in use: `lsof -i :8080`
- [ ] Correct working directory: `pwd` (should be backend/)

---

**Still stuck?** Create an issue on GitHub with the information above!
