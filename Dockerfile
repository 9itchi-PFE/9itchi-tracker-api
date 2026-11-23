# syntax=docker/dockerfile:1.4

ARG GO_VERSION=1.26
ARG ALPINE_VERSION=3.19

# ===========================================
# Build stage
# ===========================================
FROM golang:${GO_VERSION}-alpine AS builder

RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    wget

WORKDIR /src

# Copy go modules manifests first for better caching
COPY go.mod go.sum* ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

COPY . .

RUN go mod tidy

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOFLAGS="-trimpath"

RUN --mount=type=cache,target=/root/.cache/go-build \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /app-binary \
    ./main.go

RUN test -f /app-binary && chmod +x /app-binary

# ===========================================
# Production stage
# ===========================================
FROM alpine:${ALPINE_VERSION}

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wget \
    curl && \
    addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

WORKDIR /app

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder --chown=appuser:appuser /app-binary /app/service

USER appuser:appuser

EXPOSE 8086

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["curl", "-f", "http://localhost:8086/health"]

ENTRYPOINT ["/app/service"]
