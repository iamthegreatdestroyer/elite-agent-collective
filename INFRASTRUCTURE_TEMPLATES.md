# 🏗️ Phase 6 Infrastructure Templates & Configuration

**Status:** 🟢 INITIATED | **Last Updated:** December 11, 2025  
**Purpose:** Provide production-ready infrastructure as code templates for Phase 6 deployment

---

## 📁 Infrastructure Directory Structure

```
infrastructure/
├── docker/
│   ├── Dockerfile                  # Multi-stage production build
│   ├── .dockerignore              # Build optimization
│   └── docker-compose.prod.yml    # Production compose
├── kubernetes/
│   ├── base/
│   │   ├── deployment.yaml        # Core deployment config
│   │   ├── service.yaml           # Service definition
│   │   ├── ingress.yaml           # HTTP/HTTPS routing
│   │   └── configmap.yaml         # Configuration
│   ├── overlays/
│   │   ├── development/           # Dev overrides
│   │   ├── staging/               # Staging overrides
│   │   └── production/            # Production overrides
│   └── helm/
│       ├── Chart.yaml             # Helm chart metadata
│       ├── values.yaml            # Default values
│       ├── values-prod.yaml       # Production values
│       └── templates/             # K8s manifest templates
├── cloud/
│   ├── aws/
│   │   ├── terraform/
│   │   │   ├── main.tf            # ECS/EKS setup
│   │   │   ├── variables.tf       # Input variables
│   │   │   ├── outputs.tf         # Outputs
│   │   │   ├── rds.tf             # PostgreSQL
│   │   │   ├── elasticache.tf     # Redis
│   │   │   └── networking.tf      # VPC/security groups
│   │   └── cloudformation/        # CFN templates (alternative)
│   ├── azure/
│   │   ├── bicep/
│   │   │   ├── main.bicep         # AKS deployment
│   │   │   ├── variables.bicep    # Parameters
│   │   │   ├── database.bicep     # PostgreSQL
│   │   │   ├── cache.bicep        # Redis
│   │   │   └── networking.bicep   # VNet/NSGs
│   │   └── arm/                   # ARM templates (alternative)
│   └── gcp/
│       ├── terraform/
│       │   ├── main.tf            # GKE setup
│       │   ├── variables.tf       # Input variables
│       │   ├── cloud_sql.tf       # PostgreSQL
│       │   ├── memorystore.tf     # Redis
│       │   └── networking.tf      # VPC/firewall rules
│       └── deployment_manager/    # DM templates (alternative)
├── monitoring/
│   ├── prometheus/
│   │   ├── prometheus.yml         # Prometheus config
│   │   └── rules/
│   │       ├── alerts.yml         # Alert rules (30+)
│   │       └── recording.yml      # Recording rules
│   ├── grafana/
│   │   ├── datasources.yml        # Prometheus datasource
│   │   ├── dashboards/            # 15+ dashboards
│   │   └── provisioning/
│   ├── jaeger/
│   │   └── jaeger-deployment.yaml # Jaeger setup
│   └── logging/
│       ├── elasticsearch/         # ELK stack
│       ├── kibana/
│       └── filebeat/              # Log shipping
├── security/
│   ├── tls/
│   │   ├── cert-manager/
│   │   │   └── issuer.yaml        # Let's Encrypt
│   │   └── certificates.yaml      # TLS certs
│   ├── oauth/
│   │   ├── oauth-config.yaml      # OIDC setup
│   │   └── jwt-setup.md           # JWT configuration
│   └── secrets/
│       ├── vault-setup.sh         # HashiCorp Vault
│       └── secret-templates/      # Secret examples
├── database/
│   ├── postgresql/
│   │   ├── init-scripts/          # Database setup
│   │   ├── backup.sh              # Backup scripts
│   │   ├── replication.yml        # Replication config
│   │   └── pgbouncer.conf         # Connection pooling
│   └── redis/
│       ├── redis.conf             # Redis configuration
│       ├── cluster-setup.sh       # Cluster scripts
│       └── persistence.yml        # RDB/AOF config
├── scripts/
│   ├── install-tools.sh           # Install CLI tools
│   ├── setup-cloud-credentials.sh # Cloud auth
│   ├── init-phase-6.sh            # Initialize Phase 6
│   ├── deploy-all.sh              # Deploy to all clouds
│   ├── validate-deployment.sh     # Validation script
│   └── teardown.sh                # Cleanup
└── docs/
    ├── DOCKER_GUIDE.md            # Docker instructions
    ├── KUBERNETES_GUIDE.md        # K8s setup
    ├── HELM_GUIDE.md              # Helm instructions
    ├── AWS_DEPLOYMENT.md          # AWS guide
    ├── AZURE_DEPLOYMENT.md        # Azure guide
    ├── GCP_DEPLOYMENT.md          # GCP guide
    ├── MONITORING_SETUP.md        # Monitoring guide
    ├── SECURITY_SETUP.md          # Security guide
    └── DATABASE_SETUP.md          # Database guide
```

---

## 🐳 Docker Configuration (6.1.1)

### Multi-Stage Dockerfile Template

```dockerfile
# Stage 1: Builder
FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-s -w" -o server ./cmd/server

# Stage 2: Runtime
FROM alpine:3.18
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/server .
RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/server", "health"]
CMD ["/app/server"]
```

**Benefits:**

- ✅ Minimal final image size (~50MB vs 1GB+)
- ✅ Security: Non-root user, minimal base
- ✅ Health checks built-in
- ✅ Multi-stage isolation
- ✅ Layer caching optimization

---

## ☸️ Kubernetes Configuration (6.1.2)

### Deployment Manifest Template

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: elite-agent-collective
  namespace: production
  labels:
    app: elite-agent
    version: v2.0.0
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: elite-agent
  template:
    metadata:
      labels:
        app: elite-agent
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: elite-agent
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 100
              podAffinityTerm:
                labelSelector:
                  matchExpressions:
                    - key: app
                      operator: In
                      values:
                        - elite-agent
                topologyKey: kubernetes.io/hostname
      containers:
        - name: elite-agent
          image: registry.example.com/elite-agent:v2.0.0
          imagePullPolicy: IfNotPresent
          ports:
            - name: http
              containerPort: 8080
              protocol: TCP
            - name: metrics
              containerPort: 8081
              protocol: TCP
          env:
            - name: ENVIRONMENT
              value: "production"
            - name: LOG_LEVEL
              value: "info"
            - name: DATABASE_HOST
              valueFrom:
                secretKeyRef:
                  name: database-credentials
                  key: host
          resources:
            requests:
              cpu: 500m
              memory: 512Mi
            limits:
              cpu: 1000m
              memory: 1Gi
          livenessProbe:
            httpGet:
              path: /health
              port: http
            initialDelaySeconds: 10
            periodSeconds: 30
            timeoutSeconds: 5
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /ready
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
            timeoutSeconds: 3
            failureThreshold: 2
          securityContext:
            runAsNonRoot: true
            runAsUser: 1000
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop:
                - ALL
```

### Service Definition

```yaml
apiVersion: v1
kind: Service
metadata:
  name: elite-agent
  namespace: production
  labels:
    app: elite-agent
spec:
  type: LoadBalancer
  ports:
    - name: http
      port: 80
      targetPort: 8080
      protocol: TCP
    - name: metrics
      port: 8081
      targetPort: 8081
      protocol: TCP
  selector:
    app: elite-agent
  sessionAffinity: ClientIP
```

### Ingress Configuration

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: elite-agent
  namespace: production
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "1000"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - api.example.com
      secretName: elite-agent-tls
  rules:
    - host: api.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: elite-agent
                port:
                  number: 80
```

### NetworkPolicy (Security)

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: elite-agent-network-policy
  namespace: production
spec:
  podSelector:
    matchLabels:
      app: elite-agent
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: ingress-nginx
      ports:
        - protocol: TCP
          port: 8080
  egress:
    - to:
        - namespaceSelector: {}
      ports:
        - protocol: TCP
          port: 443
    - to:
        - podSelector:
            matchLabels:
              app: postgres
      ports:
        - protocol: TCP
          port: 5432
    - to:
        - podSelector:
            matchLabels:
              app: redis
      ports:
        - protocol: TCP
          port: 6379
```

### PodDisruptionBudget (Availability)

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: elite-agent-pdb
  namespace: production
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: elite-agent
```

---

## 📦 Helm Chart (6.1.3)

### Chart.yaml

```yaml
apiVersion: v2
name: elite-agent-collective
description: Elite Agent Collective production deployment
type: application
version: 2.0.0
appVersion: "2.0.0"
home: https://github.com/iamthegreatdestroyer/elite-agent-collective
sources:
  - https://github.com/iamthegreatdestroyer/elite-agent-collective
maintainers:
  - name: Infrastructure Team
    email: infra@example.com
keywords:
  - elite-agent
  - kubernetes
  - production
```

### values.yaml

```yaml
# Default values for elite-agent-collective

replicaCount: 3

image:
  repository: registry.example.com/elite-agent
  pullPolicy: IfNotPresent
  tag: "2.0.0"

imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""

serviceAccount:
  create: true
  annotations: {}
  name: ""

podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"

podSecurityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 1000

securityContext:
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  capabilities:
    drop:
      - ALL

service:
  type: LoadBalancer
  port: 80
  targetPort: 8080

ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  hosts:
    - host: api.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: elite-agent-tls
      hosts:
        - api.example.com

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70

nodeSelector: {}

tolerations: []

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchExpressions:
              - key: app
                operator: In
                values:
                  - elite-agent
          topologyKey: kubernetes.io/hostname

database:
  host: postgres.default.svc.cluster.local
  port: 5432
  name: elite_agent
  sslMode: require

redis:
  host: redis.default.svc.cluster.local
  port: 6379
  password: ""

environment:
  LOG_LEVEL: info
  ENVIRONMENT: production
```

---

## 🌍 Cloud Deployment Templates

### AWS Terraform (6.2.1)

**main.tf** structure:

```hcl
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  backend "s3" {
    bucket         = "elite-agent-terraform-state"
    key            = "production/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-locks"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Environment = "production"
      Project     = "elite-agent-collective"
      ManagedBy   = "terraform"
    }
  }
}

# EKS Cluster
module "eks" {
  source = "terraform-aws-modules/eks/aws"

  cluster_name    = "elite-agent-production"
  cluster_version = "1.27"

  cluster_endpoint_private_access = true
  cluster_endpoint_public_access  = true

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets
}

# RDS PostgreSQL
module "rds" {
  source = "terraform-aws-modules/rds/aws"

  identifier = "elite-agent-postgres"
  engine     = "postgres"
  engine_version = "15.4"
  family     = "postgres15"

  allocated_storage = 100
  multi_az          = true

  db_name  = "elite_agent"
  username = "postgres"
}

# ElastiCache Redis
resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "elite-agent-redis"
  engine              = "redis"
  engine_version      = "7.0"
  node_type          = "cache.r7g.xlarge"
  num_cache_nodes     = 3
  parameter_group_name = "default.redis7"
  port                = 6379
  multi_az_enabled    = true
  automatic_failover_enabled = true
}
```

### Azure Bicep (6.2.2)

**main.bicep** structure:

```bicep
metadata description = 'Elite Agent Collective on Azure'

param location string = resourceGroup().location
param environmentName string = 'production'
param projectName string = 'elite-agent'

// AKS Cluster
resource akCluster 'Microsoft.ContainerService/managedClusters@2023-06-01' = {
  name: '${projectName}-aks'
  location: location
  identity: {
    type: 'SystemAssigned'
  }
  properties: {
    kubernetesVersion: '1.27.0'
    dnsPrefix: '${projectName}-dns'
    agentPoolProfiles: [
      {
        name: 'agentpool'
        count: 3
        vmSize: 'Standard_D4s_v3'
        osType: 'Linux'
        type: 'VirtualMachineScaleSets'
      }
    ]
    networkProfile: {
      networkPlugin: 'azure'
      serviceCidr: '10.0.0.0/16'
      dnsServiceIP: '10.0.0.10'
      dockerBridgeCidr: '172.17.0.1/16'
    }
  }
}

// Azure Database for PostgreSQL
resource postgresServer 'Microsoft.DBforPostgreSQL/servers@2017-12-01' = {
  name: '${projectName}-postgres'
  location: location
  properties: {
    createMode: 'Default'
    version: '12'
    administratorLogin: 'postgres'
    sslEnforcement: 'ENABLED'
  }
}

// Azure Cache for Redis
resource redisCache 'Microsoft.Cache/redis@2023-04-01' = {
  name: '${projectName}-redis'
  location: location
  properties: {
    sku: {
      name: 'Premium'
      family: 'P'
      capacity: 3
    }
    enableNonSslPort: false
    minimumTlsVersion: '1.2'
  }
}
```

### GCP Terraform (6.2.3)

**main.tf** structure:

```hcl
terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
}

# GKE Cluster
resource "google_container_cluster" "primary" {
  name     = "elite-agent-production"
  location = var.gcp_region

  initial_node_count = 3

  node_config {
    machine_type = "n1-standard-4"
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]
  }

  workload_identity_config {
    workload_pool = "${var.gcp_project_id}.svc.id.goog"
  }
}

# Cloud SQL PostgreSQL
resource "google_sql_database_instance" "postgres" {
  name             = "elite-agent-postgres"
  database_version = "POSTGRES_15"
  region           = var.gcp_region

  settings {
    tier              = "db-custom-4-16384"
    availability_type = "REGIONAL"
    backup_configuration {
      enabled                        = true
      point_in_time_recovery_enabled = true
      backup_retention_days          = 30
    }
  }
}

# Memorystore Redis
resource "google_redis_instance" "cache" {
  name           = "elite-agent-redis"
  memory_size_gb = 5
  tier           = "STANDARD_HA"
  region         = var.gcp_region
}
```

---

## 📊 Monitoring Configuration

### Prometheus Configuration (6.3.1)

```yaml
global:
  scrape_interval: 15s
  scrape_timeout: 10s
  evaluation_interval: 15s

alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - alertmanager:9093

rule_files:
  - "/etc/prometheus/rules/*.yml"

scrape_configs:
  - job_name: "elite-agent"
    static_configs:
      - targets: ["localhost:8080"]
    scrape_interval: 5s

  - job_name: "kubernetes-apiservers"
    kubernetes_sd_configs:
      - role: endpoints
    relabel_configs:
      - source_labels:
          [
            __meta_kubernetes_namespace,
            __meta_kubernetes_service_name,
            __meta_kubernetes_endpoint_port_name,
          ]
        action: keep
        regex: default;kubernetes;https
```

### Alert Rules (6.3.4)

```yaml
groups:
  - name: elite-agent
    interval: 30s
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"

      - alert: HighLatency
        expr: histogram_quantile(0.95, http_request_duration_seconds_bucket) > 0.15
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High API latency detected"

      - alert: LowAvailability
        expr: up{job="elite-agent"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Elite Agent instance down"
```

---

## 🔐 Security Configuration (6.4)

### TLS Certificate Setup

```bash
# Using cert-manager with Let's Encrypt
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

### Secret Management Example

```bash
# Create sealed secret for database credentials
kubectl create secret generic database-credentials \
  --from-literal=host=postgres.default.svc.cluster.local \
  --from-literal=user=postgres \
  --from-literal=password=secretpassword \
  --dry-run=client -o yaml | \
  kubeseal --format yaml > sealed-secrets.yaml
```

---

## 💾 Database Configuration (6.5)

### PostgreSQL High Availability

```yaml
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: elite-agent-postgres
spec:
  instances: 3
  postgresql:
    parameters:
      shared_buffers: "256MB"
      max_connections: "200"
  bootstrap:
    initdb:
      database: elite_agent
      owner: postgres
  primaryUpdateStrategy: unsupervised
  monitoring:
    enabled: true
  backup:
    barmanObjectStore:
      destinationPath: "s3://backup-bucket/postgres"
      s3Credentials:
        secretKeyRef:
          name: aws-creds
          key: accessKey
```

### Redis Cluster Setup

```yaml
apiVersion: redis.opstreelabs.in/v1beta1
kind: Redis
metadata:
  name: elite-agent-redis
spec:
  mode: cluster
  replicas: 1
  storage:
    size: 5Gi
  persistence:
    enabled: true
    storageClassName: fast-ssd
```

---

## 🚀 Deployment Scripts

### Quick Deployment Script

```bash
#!/bin/bash
set -euo pipefail

ENVIRONMENT=${1:-production}
CLOUD=${2:-aws}  # aws, azure, gcp

echo "🚀 Deploying Elite Agent to $CLOUD ($ENVIRONMENT)"

# 1. Validate prerequisites
./scripts/validate-prerequisites.sh

# 2. Provision infrastructure
case $CLOUD in
  aws)
    cd infrastructure/cloud/aws/terraform
    terraform init
    terraform apply -var-file="$ENVIRONMENT.tfvars"
    ;;
  azure)
    cd infrastructure/cloud/azure/bicep
    az deployment group create \
      --resource-group elite-agent-$ENVIRONMENT \
      --template-file main.bicep
    ;;
  gcp)
    cd infrastructure/cloud/gcp/terraform
    terraform init
    terraform apply -var-file="$ENVIRONMENT.tfvars"
    ;;
esac

# 3. Deploy Kubernetes resources
kubectl apply -f infrastructure/kubernetes/

# 4. Deploy Helm chart
helm install elite-agent infrastructure/kubernetes/helm \
  -f infrastructure/kubernetes/helm/values-$ENVIRONMENT.yaml

# 5. Deploy monitoring stack
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack

# 6. Validate deployment
./scripts/validate-deployment.sh

echo "✅ Deployment complete!"
```

---

## ✅ Validation Checklist

- [ ] Dockerfile builds successfully
- [ ] Image size is minimal (<100MB)
- [ ] Health checks work
- [ ] Kubernetes manifests are valid
- [ ] Helm chart installs cleanly
- [ ] AWS/Azure/GCP resources provision
- [ ] Database connectivity works
- [ ] Redis cluster operational
- [ ] Monitoring stack online
- [ ] TLS certificates issued
- [ ] All alerts firing correctly
- [ ] Load tests passing
- [ ] Security scan clean

---

## 📚 Next Steps

1. **Create Infrastructure Repository**

   ```bash
   mkdir -p infrastructure/{docker,kubernetes,cloud,monitoring,security,database,scripts,docs}
   ```

2. **Initialize Terraform**

   ```bash
   cd infrastructure/cloud/aws/terraform
   terraform init
   ```

3. **Build and Test Docker Image**

   ```bash
   docker build -t elite-agent:v2.0.0 -f infrastructure/docker/Dockerfile .
   docker run --rm elite-agent:v2.0.0
   ```

4. **Test Kubernetes Locally**

   ```bash
   helm template elite-agent infrastructure/kubernetes/helm | kubectl apply -f -
   ```

5. **Begin Cloud Deployment**
   Follow AWS_DEPLOYMENT.md, AZURE_DEPLOYMENT.md, or GCP_DEPLOYMENT.md

---

**Infrastructure Templates Ready! 🏗️**

For detailed implementation guides, see Phase 6 documentation:

- [Docker Guide](DOCKER_KUBERNETES_GUIDE.md)
- [Kubernetes Guide](DOCKER_KUBERNETES_GUIDE.md)
- [AWS Deployment](CLOUD_DEPLOYMENT_GUIDE.md)
- [Azure Deployment](CLOUD_DEPLOYMENT_GUIDE.md)
- [GCP Deployment](CLOUD_DEPLOYMENT_GUIDE.md)
