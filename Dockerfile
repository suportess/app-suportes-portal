# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /build

# Cache de dependências antes de copiar o código-fonte
COPY app/go.mod app/go.sum ./
RUN go mod download

# Copia o restante do código e compila
COPY app/ .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o portal \
    ./cmd/main.go

# ── Stage 2: imagem final mínima ──────────────────────────────────────────────
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /build/portal /portal

EXPOSE 8080

ENTRYPOINT ["/portal"]
