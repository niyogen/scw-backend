# ==========================================
# Stage 1: Build Stage
# ==========================================
FROM golang:alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies and git
RUN apk add --no-cache git ca-certificates tzdata

# Copy dependency manifests
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy source code
COPY . .

# Compile optimized static binary (CGO disabled for portability)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/server ./cmd/api

# ==========================================
# Stage 2: Production Minimal Image
# ==========================================
FROM alpine:3.19 AS runner

# Install essential certificates & tzdata
RUN apk add --no-cache ca-certificates tzdata wget

# Create non-root user for security best practices
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/server /app/server

# Set ownership to non-root user
RUN chown -R appuser:appgroup /app

USER appuser

# Expose API port
EXPOSE 8080

# Container Healthcheck
HEALTHCHECK --interval=20s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Start API service
CMD ["/app/server"]
