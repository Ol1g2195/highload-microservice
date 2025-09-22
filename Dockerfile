# Build stage
FROM golang:1.24 AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM scratch

# Set working directory
WORKDIR /app

# Copy SSL cert bundle from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary and migrations
COPY --from=builder /app/main /app/main
COPY --from=builder /app/internal/database/migrations.sql /app/migrations.sql

# Run as non-root (numeric id)
USER 1001:1001

# Expose port
EXPOSE 8080

# Note: healthcheck is defined in compose; scratch has no shell/wget

# Run the application
ENTRYPOINT ["/app/main"]

