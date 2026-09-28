# ---------- Build ----------
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Cache de dependências
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Binário estático (sem CGO) para rodar em imagem mínima
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /fechou-brick-api ./cmd/fechou-brick-api

# ---------- Runtime ----------
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /fechou-brick-api ./fechou-brick-api
# As migrations rodam automaticamente no start (golang-migrate, source file://db/migrations)
COPY --from=builder /app/db/migrations ./db/migrations

USER app

# Configuração via ambiente (ver SPEC.md):
# DATABASE_URL (obrigatória) | JWT_SECRET (obrigatória)
# ADMIN_USERNAME (default admin) | ADMIN_INITIAL_PASSWORD (primeira subida)
# PORT (default 8085) | MAX_UPLOAD_MB (default 10)
EXPOSE 8085

ENTRYPOINT ["./fechou-brick-api"]
