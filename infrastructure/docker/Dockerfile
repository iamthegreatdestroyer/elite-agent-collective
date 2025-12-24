# Multi-stage Dockerfile for Elite Agent Collective Backend
# Stage 1: Builder - Compile Go binary
# Stage 2: Runtime - Minimal production image

# ============================================================================
# STAGE 1: BUILDER
# ============================================================================
FROM golang:1.21-alpine3.18 AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set build arguments
ARG VERSION=2.0.0
ARG BUILD_DATE
ARG VCS_REF

# Set working directory
WORKDIR /build

# Copy go mod files
COPY backend/go.mod backend/go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY backend/ .

# Build the application
# - Disable CGO for static binary
# - Optimize binary size
# - Embed version info
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -a \
    -installsuffix cgo \
    -ldflags="\
    -w -s \
    -X main.Version=${VERSION} \
    -X main.BuildDate=${BUILD_DATE} \
    -X main.VcsRef=${VCS_REF}" \
    -o /build/app \
    ./cmd/server/main.go

# Verify binary was created
RUN test -f /build/app || (echo "Build failed" && exit 1)

# ============================================================================
# STAGE 2: RUNTIME
# ============================================================================
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user for security
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder --chown=appuser:appgroup /build/app /app/app

# Copy config files if needed
COPY --chown=appuser:appgroup backend/config/ /app/config/

# Health check endpoint configuration
# The app should expose /health endpoint on port 8080
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Drop unnecessary capabilities for security
RUN chmod -R 555 /app && \
    chmod 555 /app/app

# Use non-root user
USER appuser

# Expose ports
# 8080 - API server
# 8081 - Metrics (Prometheus)
EXPOSE 8080 8081

# Set environment variables
ENV PORT=8080 \
    LOG_LEVEL=info \
    ENVIRONMENT=production

# Run the application
# Use exec form to ensure signals are properly handled
ENTRYPOINT ["/app/app"]
CMD ["--port", "8080"]

# ============================================================================
# IMAGE METADATA
# ============================================================================
LABEL org.opencontainers.image.title="Elite Agent Collective Backend" \
    org.opencontainers.image.description="Production-grade microservice for AI agent orchestration" \
    org.opencontainers.image.version="${VERSION}" \
    org.opencontainers.image.created="${BUILD_DATE}" \
    org.opencontainers.image.revision="${VCS_REF}" \
    org.opencontainers.image.vendor="Elite Agent Collective" \
    org.opencontainers.image.source="https://github.com/iamthegreatdestroyer/elite-agent-collective" \
    org.opencontainers.image.documentation="https://github.com/iamthegreatdestroyer/elite-agent-collective/docs" \
    org.opencontainers.image.authors="Elite Agent Collective Team"
