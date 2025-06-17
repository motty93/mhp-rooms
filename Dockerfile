# Build stage
FROM golang:1.22.2-alpine AS builder

# Install git for go mod download
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Verify files are copied correctly (debug)
RUN ls -la static/ && ls -la static/css/ || echo "CSS directory not found"

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/server

# Final stage - minimal runtime image
FROM alpine:3.18

# Install ca-certificates for HTTPS requests and timezone data
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy templates
COPY --from=builder /app/templates ./templates

# Copy static files individually to ensure they're included
COPY --from=builder /app/static/css ./static/css
COPY --from=builder /app/static/images ./static/images
COPY --from=builder /app/static/js ./static/js

# Verify files are copied to final stage (debug)
RUN ls -la static/ && ls -la static/css/ || echo "CSS directory not found in final stage"

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user for security
USER appuser

# Expose port
EXPOSE 8080

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Run the application
CMD ["./main"]