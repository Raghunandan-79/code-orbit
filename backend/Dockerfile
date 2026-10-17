# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source tree
COPY . .

# Build static executable
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o /app/bin/syncraft-api ./cmd/api

# Final Minimal Runtime Stage
FROM alpine:3.21

WORKDIR /app

# Install root CA certificates for outbound TLS if required
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -S syncraft && adduser -S syncraft -G syncraft

COPY --from=builder /app/bin/syncraft-api /app/syncraft-api
COPY --from=builder /app/.env.example /app/.env.example

USER syncraft:syncraft

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8080}/healthz || exit 1

ENTRYPOINT ["/app/syncraft-api"]
