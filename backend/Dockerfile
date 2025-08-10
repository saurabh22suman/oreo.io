# Multi-stage build for Go backend
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies for building
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY backend/go.mod backend/go.sum ./

# Download dependencies
RUN go mod download

# Copy backend source code
COPY backend/ .

# Build the server application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o server cmd/server/main.go

# Build the migration tool
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o migrate cmd/migrate/main.go

# Final stage - minimal runtime image
FROM alpine:latest

# Install ca-certificates, timezone data, and postgresql-client for migrations
RUN apk --no-cache add ca-certificates tzdata postgresql-client && \
    adduser -D -s /bin/sh -u 1001 appuser

# Set working directory
WORKDIR /app

# Copy the binaries from builder stage
COPY --from=builder /app/server .
COPY --from=builder /app/migrate .
COPY --from=builder /app/.env.example ./.env

# Copy migration files from the repository root
COPY database ./database

# Copy startup script
COPY --from=builder /app/docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

# Change ownership to non-root user
RUN chown -R appuser:appuser /app
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the startup script
CMD ["./docker-entrypoint.sh"]
