# ---- Build stage ----
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go.mod and go.sum first for better layer caching.
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source tree.
COPY . .

# Build both the server and worker binaries with optimizations.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /bin/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /bin/worker ./cmd/worker

# ---- Runtime stage ----
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata curl

# Create a non-root user for security.
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

WORKDIR /app

# Copy binaries from the build stage.
COPY --from=builder /bin/server /app/server
COPY --from=builder /bin/worker /app/worker

# Copy migration files.
COPY --from=builder /app/internal/db/migrations /app/migrations

# Set ownership to the non-root user.
RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

STOPSIGNAL SIGTERM

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
	CMD curl -f http://localhost:8080/health || exit 1

# Default command runs the server. Override with "worker" to run the worker.
CMD ["/app/server"]
