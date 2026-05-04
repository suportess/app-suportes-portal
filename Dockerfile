# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM golang:1.24-bullseye AS builder

WORKDIR /build

# Cache de dependências antes de copiar o código-fonte
COPY app/go.mod app/go.sum ./
RUN go mod download

# Copia o restante do código e compila
# CGO_ENABLED=1 é obrigatório — godror (Oracle) usa ODPI-C (biblioteca C)
COPY app/ .
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-w -s" \
    -o portal \
    ./cmd/main.go

# ── Stage 2: imagem final ─────────────────────────────────────────────────────
# Usa debian-slim (tem glibc) em vez de scratch — necessário para CGO + libaio
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        libaio1 \
        wget \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /build/portal /portal

EXPOSE 8080

ENTRYPOINT ["/portal"]
