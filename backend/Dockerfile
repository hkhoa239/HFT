# Multi-stage build for QuantAlpha Backend
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 for a static binary that works in alpine
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Final stage
FROM alpine:latest

# Add certificates for TLS/HTTPS and curl for healthchecks
RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

# Default port for the backend
EXPOSE 8080

# Run the binary
CMD ["./main"]
