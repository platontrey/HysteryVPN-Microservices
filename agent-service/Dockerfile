# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git and other build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o agent cmd/agent/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and curl for health checks
RUN apk --no-cache add ca-certificates curl

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/agent .

# Copy configuration files
COPY --from=builder /app/configs ./configs

# Create logs directory
RUN mkdir -p /app/logs

# Expose ports
EXPOSE 50051 8080/udp

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:50051/health || exit 1

# Run the binary
CMD ["./agent"]