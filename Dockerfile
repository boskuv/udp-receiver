# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ussc ./cmd/ussc

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/ussc .

# Copy example config (user should mount their own config.yml)
COPY --from=builder /build/cmd/ussc/config.yml.example ./config.yml.example

# Expose Prometheus metrics port
EXPOSE 8080

# Expose UDP ports (these should be mapped at runtime)
# Example: -p 8830:8830/udp -p 8831:8831/udp

# Run the application
CMD ["./ussc", "-config", "/app/config.yml"]

