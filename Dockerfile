# Multi-stage build for optimal image size and security
# Step 1: Build the Go binary
FROM golang:1.26-alpine AS builder

# Install system dependencies needed for compiling Go
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy dependency files first to leverage Docker layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire workspace source code
COPY . .

# Build a static binary with network/security support
# Target cmd/worker/main.go as the primary entrypoint
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o notification-worker ./cmd/worker/main.go

# Step 2: Minimal runtime image for production
FROM alpine:3.19

# Install base certificates and timezone database
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy compiled static binary from builder stage
COPY --from=builder /app/notification-worker /app/notification-worker

# Copy migration scripts if they are needed for starting/migrating db
# (Check if the internal/database/migrations folder exists before running container)
COPY internal/database/migrations ./internal/database/migrations

# Expose Go application port
EXPOSE 8090

# Run the app
CMD ["/app/notification-worker"]
