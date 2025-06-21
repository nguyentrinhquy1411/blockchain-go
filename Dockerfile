# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Run go mod tidy to update dependencies
RUN go mod tidy

# Build the application
RUN go build -o blockchain-node ./cmd/node

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/blockchain-node .

# Create data directory for LevelDB
RUN mkdir -p /app/data

# Expose ports
EXPOSE 50051 8080

# Run the application
CMD ["./blockchain-node"]
