# Build stage
FROM golang:1.25.0-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Runtime stage
FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/bin/api .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./api"]
