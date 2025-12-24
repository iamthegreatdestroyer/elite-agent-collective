#!/bin/bash
# Elite Agent Collective - Phase 6 Deployment Script
# Orchestrates full infrastructure setup and deployment
# Usage: ./deploy.sh [environment] [action]

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
INFRASTRUCTURE_DIR="${PROJECT_ROOT}/infrastructure"
ENVIRONMENT="${1:-development}"
ACTION="${2:-deploy}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites for ${ENVIRONMENT} deployment..."
    
    local missing_tools=()
    
    # Check required tools
    for tool in docker kubectl helm git; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    # Check based on cloud provider
    case "${ENVIRONMENT}" in
        aws)
            if ! command -v aws &> /dev/null; then
                missing_tools+=("aws-cli")
            fi
            if ! command -v terraform &> /dev/null; then
                missing_tools+=("terraform")
            fi
            ;;
        azure)
            if ! command -v az &> /dev/null; then
                missing_tools+=("azure-cli")
            fi
            ;;
        gcp)
            if ! command -v gcloud &> /dev/null; then
                missing_tools+=("gcloud")
            fi
            if ! command -v terraform &> /dev/null; then
                missing_tools+=("terraform")
            fi
            ;;
    esac
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        return 1
    fi
    
    log_success "All prerequisites met"
    return 0
}

# Build Docker image
build_docker_image() {
    log_info "Building Docker image for ${ENVIRONMENT}..."
    
    local version=$(git describe --tags --always 2>/dev/null || echo "dev")
    local build_date=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
    local vcs_ref=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    
    docker build \
        -t "elite-agent:${version}" \
        -t "elite-agent:latest" \
        --build-arg VERSION="${version}" \
        --build-arg BUILD_DATE="${build_date}" \
        --build-arg VCS_REF="${vcs_ref}" \
        -f "${INFRASTRUCTURE_DIR}/docker/Dockerfile" \
        "${PROJECT_ROOT}/backend"
    
    if [ $? -eq 0 ]; then
        log_success "Docker image built successfully"
        
        # Show image info
        docker inspect "elite-agent:${version}" --format='Image ID: {{.ID}} | Size: {{.Size}} bytes'
        
        return 0
    else
        log_error "Docker image build failed"
        return 1
    fi
}

# Test Docker image locally
test_docker_image() {
    log_info "Testing Docker image locally..."
    
    # Run container with health check
    local container_id=$(docker run -d -p 8080:8080 --health-cmd="curl -f http://localhost:8080/health || exit 1" elite-agent:latest)
    
    log_info "Started container: ${container_id}"
    
    # Wait for container to be healthy
    local max_attempts=30
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        local health=$(docker inspect --format='{{.State.Health.Status}}' "${container_id}" 2>/dev/null || echo "")
        
        if [ "${health}" = "healthy" ]; then
            log_success "Container is healthy"
            
            # Test endpoint
            local response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)
            
            if [ "${response}" = "200" ]; then
                log_success "Health endpoint returning 200 OK"
                docker stop "${container_id}" > /dev/null
                docker rm "${container_id}" > /dev/null
                return 0
            fi
        fi
        
        attempt=$((attempt + 1))
        sleep 1
    done
    
    log_error "Container health check failed"
    docker stop "${container_id}" > /dev/null
    docker rm "${container_id}" > /dev/null
    return 1
}

# Deploy to Kubernetes (local minikube)
deploy_minikube() {
    log_info "Deploying to local Minikube cluster..."
    
    # Start minikube if not running
    if ! minikube status &> /dev/null; then
        log_warn "Minikube not running. Starting..."
        minikube start --cpus=4 --memory=8192 --disk-size=20g
    fi
    
    # Load image into minikube
    log_info "Loading Docker image into Minikube..."
    minikube image load elite-agent:latest
    
    # Apply Kubernetes manifests
    log_info "Applying Kubernetes manifests..."
    kubectl apply -f "${INFRASTRUCTURE_DIR}/kubernetes/namespace.yaml" 2>/dev/null || true
    kubectl apply -f "${INFRASTRUCTURE_DIR}/kubernetes/deployment.yaml"
    kubectl apply -f "${INFRASTRUCTURE_DIR}/kubernetes/service.yaml"
    kubectl apply -f "${INFRASTRUCTURE_DIR}/kubernetes/ingress.yaml"
    
    # Wait for deployment
    log_info "Waiting for deployment to be ready..."
    kubectl rollout status deployment/elite-agent-api --timeout=5m
    
    if [ $? -eq 0 ]; then
        log_success "Deployment successful"
        
        # Show deployment info
        kubectl get pods -l app=elite-agent-api
        kubectl get svc elite-agent-api
        kubectl get hpa elite-agent-api-hpa
        
        log_info "Port forwarding: kubectl port-forward svc/elite-agent-api 8080:80"
        
        return 0
    else
        log_error "Deployment failed"
        kubectl describe deployment elite-agent-api
        return 1
    fi
}

# Deploy to AWS
deploy_aws() {
    log_info "Deploying to AWS (EKS)..."
    
    local region="${AWS_REGION:-us-east-1}"
    local cluster_name="elite-agent-prod-eks"
    
    # Initialize Terraform
    log_info "Initializing Terraform for AWS..."
    cd "${INFRASTRUCTURE_DIR}/cloud/aws/terraform"
    terraform init
    
    # Plan and apply
    log_info "Planning AWS infrastructure..."
    terraform plan -var-file="production.tfvars" -out=tfplan
    
    log_info "Applying AWS infrastructure (this may take 15-20 minutes)..."
    terraform apply tfplan
    
    # Get kubeconfig
    log_info "Getting EKS kubeconfig..."
    aws eks update-kubeconfig --region "${region}" --name "${cluster_name}"
    
    # Deploy with Helm
    log_info "Deploying application to EKS..."
    helm install elite-agent "${INFRASTRUCTURE_DIR}/kubernetes/helm" \
        -f "${INFRASTRUCTURE_DIR}/kubernetes/helm/values-prod.yaml" \
        --wait --timeout=10m
    
    log_success "AWS deployment complete"
    return 0
}

# Deploy to Azure
deploy_azure() {
    log_info "Deploying to Azure (AKS)..."
    
    local resource_group="elite-agent-prod"
    local location="${AZURE_LOCATION:-eastus}"
    
    # Create resource group
    log_info "Creating Azure resource group..."
    az group create --name "${resource_group}" --location "${location}"
    
    # Deploy Bicep templates
    log_info "Deploying Azure infrastructure..."
    az deployment group create \
        --resource-group "${resource_group}" \
        --template-file "${INFRASTRUCTURE_DIR}/cloud/azure/bicep/main.bicep" \
        --parameters location="${location}"
    
    # Get AKS credentials
    log_info "Getting AKS credentials..."
    local aks_name=$(az aks list -g "${resource_group}" --query '[0].name' -o tsv)
    az aks get-credentials --resource-group "${resource_group}" --name "${aks_name}"
    
    # Deploy with Helm
    log_info "Deploying application to AKS..."
    helm install elite-agent "${INFRASTRUCTURE_DIR}/kubernetes/helm" \
        -f "${INFRASTRUCTURE_DIR}/kubernetes/helm/values-prod.yaml" \
        --wait --timeout=10m
    
    log_success "Azure deployment complete"
    return 0
}

# Deploy to GCP
deploy_gcp() {
    log_info "Deploying to GCP (GKE)..."
    
    local project_id="${GCP_PROJECT_ID}"
    local region="${GCP_REGION:-us-central1}"
    
    # Initialize Terraform
    log_info "Initializing Terraform for GCP..."
    cd "${INFRASTRUCTURE_DIR}/cloud/gcp/terraform"
    terraform init
    
    # Plan and apply
    log_info "Planning GCP infrastructure..."
    terraform plan -var-file="production.tfvars" -out=tfplan
    
    log_info "Applying GCP infrastructure..."
    terraform apply tfplan
    
    # Get GKE credentials
    log_info "Getting GKE credentials..."
    local cluster_name="elite-agent-prod-gke"
    gcloud container clusters get-credentials "${cluster_name}" --region="${region}" --project="${project_id}"
    
    # Deploy with Helm
    log_info "Deploying application to GKE..."
    helm install elite-agent "${INFRASTRUCTURE_DIR}/kubernetes/helm" \
        -f "${INFRASTRUCTURE_DIR}/kubernetes/helm/values-prod.yaml" \
        --wait --timeout=10m
    
    log_success "GCP deployment complete"
    return 0
}

# Validate deployment
validate_deployment() {
    log_info "Validating deployment..."
    
    local checks_passed=0
    local checks_total=0
    
    # Check pods are running
    checks_total=$((checks_total + 1))
    if kubectl get pods -l app=elite-agent-api | grep -q "Running"; then
        log_success "Pods are running"
        checks_passed=$((checks_passed + 1))
    else
        log_error "Pods are not running"
    fi
    
    # Check service is available
    checks_total=$((checks_total + 1))
    if kubectl get svc elite-agent-api &> /dev/null; then
        log_success "Service is available"
        checks_passed=$((checks_passed + 1))
    else
        log_error "Service is not available"
    fi
    
    # Check HPA is configured
    checks_total=$((checks_total + 1))
    if kubectl get hpa elite-agent-api-hpa &> /dev/null; then
        log_success "HPA is configured"
        checks_passed=$((checks_passed + 1))
    else
        log_error "HPA is not configured"
    fi
    
    # Summary
    log_info "Validation: ${checks_passed}/${checks_total} checks passed"
    
    if [ ${checks_passed} -eq ${checks_total} ]; then
        log_success "Deployment validation successful"
        return 0
    else
        log_error "Deployment validation failed"
        return 1
    fi
}

# Rollback deployment
rollback_deployment() {
    log_info "Rolling back deployment..."
    
    # Rollback Helm release
    if helm status elite-agent &> /dev/null; then
        log_info "Rolling back Helm release..."
        helm rollback elite-agent
        
        log_success "Rollback complete"
        return 0
    else
        log_warn "No Helm release to rollback"
        return 0
    fi
}

# Show usage
show_usage() {
    cat <<EOF
Usage: $0 [environment] [action]

Environments:
  development     Deploy to local Minikube cluster
  aws            Deploy to AWS EKS
  azure          Deploy to Azure AKS
  gcp            Deploy to GCP GKE

Actions:
  deploy          Full deployment (build, test, deploy)
  build           Build Docker image only
  test            Test Docker image locally
  validate        Validate existing deployment
  rollback        Rollback to previous deployment
  clean           Clean up (remove resources)

Examples:
  ./deploy.sh development deploy      # Full local deployment
  ./deploy.sh aws deploy              # Deploy to AWS
  ./deploy.sh azure deploy            # Deploy to Azure
  ./deploy.sh gcp deploy              # Deploy to GCP
  ./deploy.sh development test        # Test Docker image
  ./deploy.sh development rollback    # Rollback Minikube deployment

Environment Variables:
  AWS_REGION          AWS region (default: us-east-1)
  AZURE_LOCATION      Azure location (default: eastus)
  GCP_PROJECT_ID      GCP project ID (required for GCP)
  GCP_REGION          GCP region (default: us-central1)

EOF
}

# Main execution
main() {
    log_info "=== Elite Agent Collective - Phase 6 Deployment ==="
    log_info "Environment: ${ENVIRONMENT}"
    log_info "Action: ${ACTION}"
    log_info ""
    
    # Check prerequisites
    if ! check_prerequisites; then
        log_error "Prerequisites check failed"
        exit 1
    fi
    
    # Execute action
    case "${ACTION}" in
        deploy)
            build_docker_image || exit 1
            test_docker_image || exit 1
            
            case "${ENVIRONMENT}" in
                development)
                    deploy_minikube || exit 1
                    ;;
                aws)
                    deploy_aws || exit 1
                    ;;
                azure)
                    deploy_azure || exit 1
                    ;;
                gcp)
                    deploy_gcp || exit 1
                    ;;
                *)
                    log_error "Unknown environment: ${ENVIRONMENT}"
                    show_usage
                    exit 1
                    ;;
            esac
            
            validate_deployment || exit 1
            log_success "Deployment complete!"
            ;;
        
        build)
            build_docker_image || exit 1
            ;;
        
        test)
            test_docker_image || exit 1
            ;;
        
        validate)
            validate_deployment || exit 1
            ;;
        
        rollback)
            rollback_deployment || exit 1
            ;;
        
        *)
            log_error "Unknown action: ${ACTION}"
            show_usage
            exit 1
            ;;
    esac
    
    log_success "Operation complete!"
}

# Run main function
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi
