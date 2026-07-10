# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Cache dependencies first.
COPY go.mod go.sum ./
RUN go mod download

# Build the static binary. VERSION is stamped into GET /version — pass
# --build-arg VERSION=$(git describe --tags); defaults to "dev".
ARG VERSION=dev
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s -X main.version=${VERSION}" -o /app/bin/api ./cmd/api

# ---- Runtime stage ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app
WORKDIR /app

# Binary is self-contained; migrations are read at startup.
COPY --from=builder /app/bin/api ./api
COPY migrations ./migrations

USER app
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./api"]
