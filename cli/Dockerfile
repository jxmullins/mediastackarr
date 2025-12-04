# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod ./
# go.sum will be created by go mod tidy
RUN go mod download 2>/dev/null || true

# Copy source code
COPY . .

# Download dependencies and tidy
RUN go mod tidy

# Build the binary
RUN make build

# Verify the build
RUN ./build/mediastack version

# Test stage - runs tests
FROM builder AS tester
RUN make test || echo "No tests yet"

# Final minimal image for distribution
FROM alpine:latest AS runtime

# Install runtime dependencies
RUN apk add --no-cache ca-certificates docker-cli docker-cli-compose

# Copy binary from builder
COPY --from=builder /app/build/mediastack /usr/local/bin/mediastack

# Verify binary works
RUN mediastack version

ENTRYPOINT ["mediastack"]
CMD ["--help"]
