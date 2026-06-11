# syntax=docker/dockerfile:1

# ---------- Stage 1: build ----------
# Compile a fully static binary so the final image can stay tiny.
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy module files first to leverage Docker layer caching:
# dependencies are re-downloaded only when go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source.
COPY . .

# Generate the Swagger/OpenAPI docs package (ymmo/docs) at build time, so the
# image is self-contained and `docker compose up` needs no manual step.
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.4
RUN swag init -g cmd/api/main.go -o docs

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /ymmo-api ./cmd/api

# ---------- Stage 2: runtime ----------
# Minimal Alpine image: small attack surface, includes wget for the healthcheck.
FROM alpine:3.20

# Run as a non-root user (security best practice).
RUN addgroup -S ymmo && adduser -S ymmo -G ymmo

WORKDIR /app
COPY --from=builder /ymmo-api /app/ymmo-api

USER ymmo
EXPOSE 8080

# Container-level health probe hitting our /health endpoint.
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/ymmo-api"]
