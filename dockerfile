# Build
FROM golang:1.25-alpine3.22 AS builder
# Install only required build dependencies
RUN apk update && apk add --no-cache git ca-certificates tzdata \
    && adduser -D -g '' appuser
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code (use .dockerignore to exclude sensitive files)
COPY . .

# Build with security flags
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o asmm8 .

RUN go install github.com/projectdiscovery/alterx/cmd/alterx@v0.0.6 && \
    go install github.com/projectdiscovery/dnsx/cmd/dnsx@v1.2.2 && \
    go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@v2.9.0

# Release
FROM alpine:3.22

# Security updates and minimal runtime dependencies
RUN apk --no-cache add ca-certificates tzdata \
    && apk --no-cache upgrade \
    && rm -rf /var/cache/apk/* \
    && adduser -D -g '' -s /bin/sh appuser


COPY --from=builder --chown=appuser:appuser /app/asmm8 /usr/local/bin/
COPY --from=builder --chown=appuser:appuser /go/bin/alterx /usr/local/bin/
COPY --from=builder --chown=appuser:appuser /go/bin/dnsx /usr/local/bin/
COPY --from=builder --chown=appuser:appuser /go/bin/subfinder /usr/local/bin/

# Security: Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD asmm8 --help || exit 1

# Metadata
LABEL maintainer="i@deifzar.me" \
    version="1.0" \
    description="ASMM8 - Hardened Hostname Discovery Scanner (Runtime)" \
    security.scan="required-non-root-privileges"

# Expose port (document the port used)
# EXPOSE 8000

# Use exec form for better signal handling
CMD ["asmm8", "help"]