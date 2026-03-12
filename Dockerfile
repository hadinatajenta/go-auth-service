# Stage 1: Modules caching
FROM golang:1.25-alpine AS modules
WORKDIR /modules
COPY go.mod go.sum ./
RUN go mod download

# Stage 2: Builder
FROM golang:1.25-alpine AS builder
COPY --from=modules /go/pkg /go/pkg
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/app ./cmd/server/main.go

# Stage 3: Final Image
FROM alpine:3.21 AS final

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Install security updates and CA certificates
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /bin/app /app/auth-service
COPY .env.example /app/.env

# Use non-root user
USER appuser

# Expose port
EXPOSE 8080

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/auth-service"]
