# Elite Agent Collective - Deployment Guide

Complete guide for deploying the Elite Agent Collective to various platforms.

## 📋 Deployment Options

### Local Development
- **Platform**: Windows/macOS/Linux
- **Setup Time**: 5-10 minutes
- **Cost**: Free
- **Best For**: Development, testing, learning

### Docker (Local)
- **Platform**: Any (with Docker installed)
- **Setup Time**: 10-15 minutes
- **Cost**: Free
- **Best For**: Consistent environments, team development

### Cloud Platforms
- **AWS (EC2, ECS, Fargate)**
- **Azure (App Service, Container Instances)**
- **Google Cloud (Cloud Run, Compute Engine)**
- **DigitalOcean (App Platform, Droplets)**
- **Heroku (Platform as a Service)**

## 🏠 Local Development Setup

### Prerequisites
- Go 1.21 or later
- Git
- VS Code (optional)

### Steps

```bash
# 1. Clone repository
git clone https://github.com/iamthegreatdestroyer/elite-agent-collective.git
cd elite-agent-collective

# 2. Navigate to backend
cd backend

# 3. Install dependencies
make deps

# 4. Run server
make run

# Server starts on http://localhost:8080

# 5. Test health endpoint
curl http://localhost:8080/health

# 6. In another terminal, test agent listing
curl http://localhost:8080/agents | jq '.[] | .codename'
```

### Configuration

Create `.env` file in `backend/` directory:

```bash
PORT=8080
LOG_LEVEL=info
OIDC_DISABLED=true  # For local development
```

### Verification

```bash
# Health check
curl http://localhost:8080/health
# Expected: {"status": "healthy", "agents_loaded": 40, ...}

# List agents
curl http://localhost:8080/agents | jq length
# Expected: 40

# Invoke agent
curl -X POST http://localhost:8080/agent \
  -H "Content-Type: application/json" \
  -d '{
    "codename": "APEX",
    "task": "implement sliding window rate limiter"
  }' | jq '.response' | head -c 200
```

## 🐳 Docker Deployment

### Prerequisites
- Docker 20.10+
- Docker Compose 1.29+ (optional)

### Single Container

```bash
# 1. Build image
docker build -t elite-agents:latest backend/

# 2. Run container
docker run -p 8080:8080 elite-agents:latest

# 3. Test
curl http://localhost:8080/health
```

### Docker Compose

```bash
# 1. Create docker-compose.yml (see below)

# 2. Start services
cd backend
docker-compose up -d

# 3. Check logs
docker-compose logs -f elite-agents

# 4. Test
curl http://localhost:8080/health

# 5. Stop services
docker-compose down
```

**docker-compose.yml** (already exists in `backend/`):

```yaml
version: '3.8'

services:
  elite-agents:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: elite-agents
    ports:
      - "8080:8080"
    environment:
      PORT: 8080
      LOG_LEVEL: info
      REDIS_URL: redis://redis:6379
      DATABASE_URL: postgresql://elite:elite@postgres:5432/elite_agents
    depends_on:
      - redis
      - postgres
    volumes:
      - ../.github/agents:/app/.github/agents:ro
    networks:
      - elite
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    image: redis:7-alpine
    container_name: elite-redis
    ports:
      - "6379:6379"
    networks:
      - elite
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes

  postgres:
    image: postgres:15-alpine
    container_name: elite-postgres
    environment:
      POSTGRES_DB: elite_agents
      POSTGRES_USER: elite
      POSTGRES_PASSWORD: elite
    ports:
      - "5432:5432"
    networks:
      - elite
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  redis_data:
  postgres_data:

networks:
  elite:
    driver: bridge
```

## ☁️ Cloud Deployment

### AWS (Elastic Container Service)

#### Option 1: Fargate (Serverless)

```bash
# 1. Push image to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin YOUR_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com

docker tag elite-agents:latest YOUR_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/elite-agents:latest
docker push YOUR_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/elite-agents:latest

# 2. Create ECS task definition (elite-agents-task.json)
# See below

# 3. Register task definition
aws ecs register-task-definition --cli-input-json file://elite-agents-task.json

# 4. Create ECS service
aws ecs create-service \
  --cluster production \
  --service-name elite-agents \
  --task-definition elite-agents:1 \
  --desired-count 3 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx,subnet-yyy],securityGroups=[sg-xxx],assignPublicIp=ENABLED}" \
  --load-balancers targetGroupArn=arn:aws:elasticloadbalancing:...,containerName=elite-agents,containerPort=8080
```

**elite-agents-task.json**:

```json
{
  "family": "elite-agents",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "1024",
  "memory": "2048",
  "containerDefinitions": [
    {
      "name": "elite-agents",
      "image": "YOUR_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/elite-agents:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "hostPort": 8080,
          "protocol": "tcp"
        }
      ],
      "essential": true,
      "environment": [
        {
          "name": "PORT",
          "value": "8080"
        },
        {
          "name": "LOG_LEVEL",
          "value": "info"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/elite-agents",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

#### Option 2: EC2

```bash
# 1. Launch EC2 instance (Amazon Linux 2)
aws ec2 run-instances \
  --image-id ami-0c55b159cbfafe1f0 \
  --instance-type t3.medium \
  --security-groups elite-agents \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=elite-agents}]'

# 2. SSH into instance
ssh -i your-key.pem ec2-user@your-instance-ip

# 3. Install Docker
sudo amazon-linux-extras install docker
sudo systemctl start docker
sudo usermod -a -G docker ec2-user

# 4. Pull and run image
docker pull YOUR_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/elite-agents:latest
docker run -d -p 8080:8080 \
  --name elite-agents \
  YOUR_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/elite-agents:latest

# 5. Verify
curl http://localhost:8080/health
```

### Azure (Container Instances)

```bash
# 1. Create resource group
az group create \
  --name elite-agents-rg \
  --location eastus

# 2. Create ACR
az acr create \
  --resource-group elite-agents-rg \
  --name eliteagents \
  --sku Basic

# 3. Build and push image
az acr build \
  --registry eliteagents \
  --image elite-agents:latest \
  backend/

# 4. Create container instance
az container create \
  --resource-group elite-agents-rg \
  --name elite-agents \
  --image eliteagents.azurecr.io/elite-agents:latest \
  --cpu 1 \
  --memory 2 \
  --registry-login-server eliteagents.azurecr.io \
  --registry-username <username> \
  --registry-password <password> \
  --ports 8080 \
  --protocol TCP \
  --dns-name-label elite-agents \
  --environment-variables PORT=8080 LOG_LEVEL=info

# 5. Get public IP
az container show \
  --resource-group elite-agents-rg \
  --name elite-agents \
  --query ipAddress.fqdn

# 6. Test
curl http://elite-agents.eastus.azurecontainer.io:8080/health
```

### Google Cloud (Cloud Run)

```bash
# 1. Set project
export PROJECT_ID=your-project-id
gcloud config set project $PROJECT_ID

# 2. Build with Cloud Build
gcloud builds submit \
  --tag gcr.io/$PROJECT_ID/elite-agents:latest \
  backend/

# 3. Deploy to Cloud Run
gcloud run deploy elite-agents \
  --image gcr.io/$PROJECT_ID/elite-agents:latest \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --memory 2Gi \
  --cpu 1 \
  --timeout 3600 \
  --set-env-vars "PORT=8080,LOG_LEVEL=info"

# 4. Get service URL
gcloud run services describe elite-agents --platform managed --region us-central1

# 5. Test
curl https://elite-agents-xxx.run.app/health
```

### DigitalOcean (App Platform)

```bash
# 1. Create app.yaml
cat > app.yaml << 'EOF'
name: elite-agents
services:
- name: api
  github:
    branch: main
    repo: iamthegreatdestroyer/elite-agent-collective
  build_command: cd backend && make build
  run_command: ./elite-agents-server
  source_dir: backend
  http_port: 8080
  envs:
  - key: PORT
    value: "8080"
  - key: LOG_LEVEL
    value: "info"
  health_check:
    http_path: /health
  min_instance_count: 1
  instance_count_max: 3
  instance_size_slug: basic-s

databases:
- name: redis
  engine: REDIS
  version: "7"
- name: postgres
  engine: PG
  version: "15"
EOF

# 2. Deploy
doctl apps create --spec app.yaml

# 3. Monitor deployment
doctl apps list
doctl apps get-logs <app-id>
```

## 🔒 Production Deployment Checklist

### Before Deployment

- [ ] All tests passing (`make test-all`)
- [ ] No sensitive data in code or config
- [ ] Environment variables properly configured
- [ ] Database migrations run successfully
- [ ] Logging configured for the environment
- [ ] Rate limiting configured
- [ ] CORS properly configured
- [ ] HTTPS/TLS configured

### Security Configuration

```bash
# Generate secure secrets
openssl rand -hex 32  # For JWT_SECRET
openssl rand -hex 16  # For other secrets

# Environment variables (use secrets manager)
# AWS Secrets Manager / Azure Key Vault / Google Cloud Secret Manager
```

### Monitoring Setup

```bash
# Application Insights / CloudWatch / Stackdriver
# Set up alerts for:
# - High error rate (> 1%)
# - High latency (p99 > 5s)
# - Low throughput (< 100 req/min)
# - Service restarts
```

### Backup & Disaster Recovery

```bash
# Database backups
# - Automated daily backups
# - Retention: 30 days
# - Test restore procedures

# Configuration backups
# - Version control all config
# - Document all secrets management
# - Plan recovery time < 1 hour
```

### Load Testing

```bash
# Before production launch, test under load
# Using tools like:
# - Apache JMeter
# - k6
# - Locust
# - hey

# Test scenarios:
# - 1000 concurrent users
# - 100 req/sec sustained
# - 50th, 95th, 99th percentile latencies
# - Error rate under 0.1%
```

## 🔧 Troubleshooting Deployments

### Container Won't Start

```bash
# Check logs
docker logs <container-id>

# Verify image
docker inspect <image-name>

# Test locally first
docker run -it <image-name> /bin/sh
```

### High Memory Usage

```bash
# Profile memory
docker stats <container-id>

# Check for leaks
go tool pprof http://localhost:6060/debug/pprof/heap

# Reduce cache size
MEMORY_CACHE_SIZE=5000
```

### Slow Responses

```bash
# Check metrics
curl http://localhost:8080/metrics

# Profile CPU
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Check database performance
# Run queries with EXPLAIN
```

## 📊 Monitoring & Observability

### Prometheus Metrics

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'elite-agents'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

### Grafana Dashboards

```
Import dashboards:
- Go Application Dashboard
- HTTP Request Rate
- Agent Performance
- Memory System Health
```

### Alerting Rules

```yaml
# alert.rules.yml
groups:
  - name: elite-agents
    rules:
      - alert: HighErrorRate
        expr: rate(elite_http_errors_total[5m]) > 0.01
        for: 5m
        
      - alert: HighLatency
        expr: histogram_quantile(0.99, elite_http_request_duration_seconds) > 5
        for: 5m
```

## 🚀 Advanced Deployments

### Kubernetes Deployment

See [backend/k8s/deployment.yaml](../backend/k8s/deployment.yaml) for:
- Deployment manifests
- Service configuration
- ConfigMap for configuration
- Secrets for sensitive data
- HorizontalPodAutoscaler for auto-scaling
- Ingress for routing

### Multi-Region Deployment

```
┌─────────────────┐       ┌─────────────────┐
│   US Region     │       │  EU Region      │
│ (Primary)       │       │ (Failover)      │
│  3 replicas     │◄─────►│  3 replicas     │
│  PostgreSQL     │       │  PostgreSQL     │
│  Redis          │       │  Redis          │
└─────────────────┘       └─────────────────┘
        ▲                        ▲
        │ DNS Round-Robin        │
        └────────┬───────────────┘
                 │
            CDN / Global LB
                 │
            Users Worldwide
```

## 📚 Additional Resources

- [Deployment Strategies](https://martinfowler.com/bliki/BlueGreenDeployment.html)
- [12 Factor App](https://12factor.net/)
- [Container Best Practices](https://snyk.io/blog/10-docker-image-security-best-practices/)
- [Kubernetes Basics](https://kubernetes.io/docs/concepts/overview/)

## 🆘 Support

For deployment issues:

1. Check [TROUBLESHOOTING.md](../docs/TROUBLESHOOTING.md)
2. Review logs: `docker logs` or cloud provider logs
3. File GitHub issue with deployment details
4. Contact support through GitHub issues

---

**Last Updated**: December 11, 2025  
**Version**: 2.0.0
