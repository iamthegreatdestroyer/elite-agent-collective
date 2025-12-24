# Elite Agent Collective - MCP Server Configuration

## Overview

The Elite Agent Collective is available as both a GitHub Copilot Extension and as a Model Context Protocol (MCP) server. This document describes how to configure and use the MCP server implementation.

## MCP Server Details

### What is MCP?

The Model Context Protocol (MCP) is an open standard for connecting AI models to external tools and data sources. It enables:

- Direct agent invocation from Claude, GPT-4, and other LLMs
- Real-time access to the Elite Agent Collective system
- Secure authentication and authorization
- Structured tool invocation with request/response validation

### Server Architecture

```
┌─────────────────────────────────────┐
│      LLM / AI Model (Claude, GPT-4) │
└─────────────┬───────────────────────┘
              │
              │ MCP Protocol
              │ (WebSocket/HTTP)
              ▼
┌─────────────────────────────────────┐
│   MCP Server (Backend on port 8080) │
├─────────────────────────────────────┤
│  ├─ Agent Registry (40 agents)      │
│  ├─ Agent Loader (.agent.md files)  │
│  ├─ MNEMONIC Memory System          │
│  ├─ Request Handler & Validator     │
│  └─ Response Builder                │
├─────────────────────────────────────┤
│  Persistent Storage:                │
│  ├─ Redis (experience cache)        │
│  ├─ PostgreSQL (history logs)       │
│  └─ File system (.github/agents/)   │
└─────────────────────────────────────┘
```

## Configuration

### Environment Variables

```bash
# Server Configuration
PORT=8080                           # HTTP server port
LOG_LEVEL=info                      # Logging level: debug, info, warn, error

# Authentication
OIDC_ISSUER=https://accounts.google.com
OIDC_CLIENT_ID=your-client-id
OIDC_CLIENT_SECRET=your-client-secret
JWT_SECRET=your-jwt-secret

# Redis (Optional, for experience caching)
REDIS_URL=redis://localhost:6379
REDIS_DB=0

# PostgreSQL (Optional, for history logging)
DATABASE_URL=postgresql://user:password@localhost:5432/elite_agents
DATABASE_MAX_CONNECTIONS=20

# Memory System
MEMORY_CACHE_SIZE=10000             # Max cached experiences
MEMORY_BREAKTHROUGH_THRESHOLD=0.9   # Fitness threshold for breakthroughs
MEMORY_TEMPORAL_DECAY_LAMBDA=0.99   # Recency weighting

# MCP Configuration
MCP_ENABLED=true
MCP_PROTOCOL_VERSION=2024.01
MCP_AUTH_METHOD=oidc                # oidc, jwt, apikey
```

### Docker Compose Setup

```yaml
version: "3.8"

services:
  elite-agents:
    build:
      context: ./backend
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      PORT: 8080
      LOG_LEVEL: info
      MCP_ENABLED: "true"
    depends_on:
      - redis
      - postgres
    volumes:
      - ./.github/agents:/app/.github/agents:ro
    networks:
      - elite

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    networks:
      - elite
    command: redis-server --appendonly yes

  postgres:
    image: postgres:15-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_DB: elite_agents
      POSTGRES_USER: elite
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-elite}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - elite

volumes:
  postgres_data:

networks:
  elite:
    driver: bridge
```

## API Endpoints

### Health Check

```
GET /health

Response:
{
  "status": "healthy",
  "version": "2.0.0",
  "agents_loaded": 40,
  "memory_status": "operational",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### List Agents

```
GET /agents

Response:
[
  {
    "id": 1,
    "codename": "APEX",
    "tier": 1,
    "category": "Foundational",
    "specialty": "Elite Computer Science Engineering",
    "description": "Master-level software engineering, system design, and computational problem-solving"
  },
  ...
]
```

### Get Agent Details

```
GET /agents/{codename}

Response:
{
  "id": 1,
  "codename": "APEX",
  "tier": 1,
  "category": "Foundational",
  "specialty": "Elite Computer Science Engineering",
  "philosophy": "Every problem has an elegant solution waiting to be discovered.",
  "capabilities": [
    "Production-grade code generation",
    "Data structures & algorithms",
    "System design & distributed systems",
    ...
  ],
  "example_invocations": [
    "@APEX implement a rate limiter with sliding window",
    "@APEX design a distributed cache system"
  ]
}
```

### Invoke Agent

```
POST /agent

Request:
{
  "codename": "APEX",
  "task": "implement a rate limiter with sliding window",
  "context": {
    "language": "python",
    "framework": "FastAPI",
    "constraints": "must support concurrent requests"
  },
  "memory_enabled": true
}

Response:
{
  "agent": "APEX",
  "task": "implement a rate limiter with sliding window",
  "response": "Here's a production-grade sliding window rate limiter...",
  "memory": {
    "experience_id": "uuid-here",
    "fitness_score": 0.95,
    "retrieval_sources": ["past_implementation_1", "pattern_memory_2"]
  },
  "execution_time_ms": 1240,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## MCP Client Integration

### Python Client Example

```python
import aiohttp
import json
from typing import Optional

class EliteAgentMCPClient:
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.session: Optional[aiohttp.ClientSession] = None

    async def connect(self):
        self.session = aiohttp.ClientSession()

    async def disconnect(self):
        if self.session:
            await self.session.close()

    async def get_agents(self) -> list:
        """Get all available agents."""
        async with self.session.get(f"{self.base_url}/agents") as resp:
            return await resp.json()

    async def get_agent(self, codename: str) -> dict:
        """Get details for a specific agent."""
        async with self.session.get(f"{self.base_url}/agents/{codename}") as resp:
            return await resp.json()

    async def invoke_agent(
        self,
        codename: str,
        task: str,
        context: Optional[dict] = None,
        memory_enabled: bool = True
    ) -> dict:
        """Invoke an agent with a task."""
        payload = {
            "codename": codename,
            "task": task,
            "context": context or {},
            "memory_enabled": memory_enabled
        }
        async with self.session.post(
            f"{self.base_url}/agent",
            json=payload,
            headers={"Content-Type": "application/json"}
        ) as resp:
            return await resp.json()

    async def invoke_multiple(self, agents: list[str], task: str) -> dict:
        """Invoke multiple agents collaboratively."""
        results = {}
        for agent_codename in agents:
            result = await self.invoke_agent(agent_codename, task)
            results[agent_codename] = result
        return results

# Usage
async def main():
    client = EliteAgentMCPClient()
    await client.connect()

    try:
        # Get all agents
        agents = await client.get_agents()
        print(f"Available agents: {len(agents)}")

        # Invoke single agent
        response = await client.invoke_agent(
            "APEX",
            "implement a distributed cache with LRU eviction",
            context={"language": "go", "max_size": "1GB"}
        )
        print(f"APEX Response: {response['response']}")

        # Multi-agent collaboration
        results = await client.invoke_multiple(
            ["APEX", "VELOCITY"],
            "optimize this database query"
        )
        for agent, result in results.items():
            print(f"{agent}: {result['response']}")

    finally:
        await client.disconnect()

if __name__ == "__main__":
    import asyncio
    asyncio.run(main())
```

### JavaScript/Node.js Client Example

```javascript
class EliteAgentMCPClient {
  constructor(baseUrl = "http://localhost:8080") {
    this.baseUrl = baseUrl;
  }

  async getAgents() {
    const response = await fetch(`${this.baseUrl}/agents`);
    return response.json();
  }

  async getAgent(codename) {
    const response = await fetch(`${this.baseUrl}/agents/${codename}`);
    return response.json();
  }

  async invokeAgent(codename, task, context = {}, memoryEnabled = true) {
    const response = await fetch(`${this.baseUrl}/agent`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        codename,
        task,
        context,
        memory_enabled: memoryEnabled,
      }),
    });
    return response.json();
  }

  async invokeMultiple(agentCodenames, task) {
    const results = {};
    for (const codename of agentCodenames) {
      results[codename] = await this.invokeAgent(codename, task);
    }
    return results;
  }
}

// Usage
const client = new EliteAgentMCPClient();

// Single agent
const response = await client.invokeAgent(
  "APEX",
  "implement rate limiter with token bucket",
  { language: "javascript", framework: "express" }
);
console.log(response.response);

// Multiple agents
const results = await client.invokeMultiple(
  ["APEX", "VELOCITY", "CIPHER"],
  "design secure authentication system"
);
Object.entries(results).forEach(([agent, result]) => {
  console.log(`${agent}: ${result.response}`);
});
```

### cURL Examples

```bash
# Health check
curl http://localhost:8080/health

# List all agents
curl http://localhost:8080/agents

# Get specific agent
curl http://localhost:8080/agents/APEX

# Invoke agent
curl -X POST http://localhost:8080/agent \
  -H "Content-Type: application/json" \
  -d '{
    "codename": "APEX",
    "task": "implement distributed rate limiter",
    "context": {
      "language": "go",
      "peak_requests": 10000
    },
    "memory_enabled": true
  }'

# Multi-agent invocation (sequential)
for agent in APEX VELOCITY CIPHER; do
  curl -X POST http://localhost:8080/agent \
    -H "Content-Type: application/json" \
    -d "{
      \"codename\": \"$agent\",
      \"task\": \"optimize this algorithm\",
      \"memory_enabled\": true
    }"
done
```

## Deployment Options

### Local Development

```bash
cd backend
make run
# Server starts on http://localhost:8080
```

### Docker

```bash
docker build -t elite-agents:latest backend/
docker run -p 8080:8080 elite-agents:latest
```

### Kubernetes

```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/configmap.yaml
```

### Cloud Platforms

#### AWS (ECS/Fargate)

```bash
# Push to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin 123456789.dkr.ecr.us-east-1.amazonaws.com

docker tag elite-agents:latest 123456789.dkr.ecr.us-east-1.amazonaws.com/elite-agents:latest
docker push 123456789.dkr.ecr.us-east-1.amazonaws.com/elite-agents:latest

# Deploy to Fargate
aws ecs create-service --cluster production \
  --service-name elite-agents \
  --task-definition elite-agents:1 \
  --desired-count 3 \
  --load-balancers targetGroupArn=arn:aws:elasticloadbalancing:...
```

#### Azure (Container Instances / App Service)

```bash
az acr build --registry eliteagents --image elite-agents:latest .
az container create --resource-group production \
  --name elite-agents \
  --image eliteagents.azurecr.io/elite-agents:latest \
  --ports 8080 \
  --environment-variables PORT=8080 LOG_LEVEL=info
```

#### Google Cloud (Cloud Run)

```bash
gcloud run deploy elite-agents \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars "PORT=8080,LOG_LEVEL=info"
```

## Security Considerations

### Authentication

The MCP server supports multiple authentication methods:

1. **OIDC (Recommended)**

   - Integrates with Google, GitHub, Microsoft, Okta
   - Automatic token refresh
   - Multi-factor authentication support

2. **JWT Tokens**

   - Self-signed or third-party CA
   - Short-lived access tokens with refresh
   - Role-based access control (RBAC)

3. **API Keys**
   - Per-application keys
   - Rate limiting per key
   - Key rotation policy

### Input Validation

- All request payloads validated against JSON schema
- Agent codenames verified against registry
- Task descriptions sanitized and length-limited
- Context objects validated for known fields

### Output Encoding

- HTML/JavaScript escaping in responses
- JSON encoding with proper type handling
- No sensitive data in logs or responses
- Audit trail of all agent invocations

### Rate Limiting

```
- Per IP: 1000 requests/hour
- Per API Key: 10,000 requests/hour
- Per Agent: 100 concurrent invocations
- Per User (if authenticated): 5,000 requests/hour
```

## Monitoring & Observability

### Metrics (Prometheus)

```
# Agent invocation metrics
elite_agent_invocations_total{codename="APEX",status="success"}
elite_agent_invocation_duration_seconds{codename="APEX"}
elite_agent_errors_total{codename="APEX",error_type="timeout"}

# Memory system metrics
elite_memory_experiences_total
elite_memory_retrieval_latency_seconds
elite_memory_fitness_score_avg

# Server metrics
elite_http_requests_total
elite_http_request_duration_seconds
elite_http_errors_total
```

### Logging

All invocations are logged with:

- Timestamp
- Agent codename
- Task description (first 500 chars)
- Response status (success/error)
- Execution time
- Memory system activity
- User/API key (if authenticated)

### Tracing

OpenTelemetry integration for distributed tracing:

- Request path through components
- Latency breakdown by component
- Error context and stack traces

## Troubleshooting

### Server Won't Start

```bash
# Check logs
docker logs <container-id>

# Verify port availability
lsof -i :8080
netstat -tlnp | grep 8080

# Check configuration
env | grep ELITE
```

### Slow Agent Responses

```bash
# Check memory cache hit rate
curl http://localhost:8080/metrics | grep memory_retrieval

# Check agent load
curl http://localhost:8080/agents | jq length

# Profile server
go test -cpuprofile=cpu.prof -memprofile=mem.prof ./...
go tool pprof backend.bin cpu.prof
```

### Authentication Failures

```bash
# Verify OIDC configuration
curl https://$OIDC_ISSUER/.well-known/openid-configuration

# Test token validation
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/health

# Check JWT expiration
jwt.io (paste token)
```

## Additional Resources

- [API Reference Documentation](../docs/API_REFERENCE.md)
- [Architecture Guide](../docs/ARCHITECTURE.md)
- [Development Guide](../docs/DEVELOPER_GUIDE.md)
- [GitHub Copilot Extension Docs](https://docs.github.com/en/copilot/building-copilot-extensions)
- [Model Context Protocol Specification](https://modelcontextprotocol.io/)

## Support

- **Documentation**: See [docs/](../docs/)
- **Issues**: GitHub Issues at iamthegreatdestroyer/elite-agent-collective
- **Discussions**: GitHub Discussions for questions and ideas
- **Email Support**: support@elite-agents.dev (if applicable)
