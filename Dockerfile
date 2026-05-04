# ── Stage 1: Oracle Instant Client ───────────────────────────────────────────
# Usa oraclelinux:8 para obter as libs do IC via dnf oficial.
# O glob 'oracle-instantclient*-basic' funciona independente da versão minor.
FROM oraclelinux:8 AS oracle-ic
RUN dnf install -y oracle-instantclient-release-el8 && \
    dnf install -y 'oracle-instantclient*-basic' && \
    rm -rf /var/cache/dnf

# ── Stage 2: build ────────────────────────────────────────────────────────────
# godror/ODPI-C embute o código C internamente — Oracle IC NÃO é necessário
# em tempo de compilação, apenas em runtime (carregado via dlopen).
FROM golang:1.24-bullseye AS builder

WORKDIR /build

COPY app/go.mod app/go.sum ./
RUN go mod download

COPY app/ .
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-w -s" \
    -o portal \
    ./cmd/main.go

# ── Stage 3: imagem final ─────────────────────────────────────────────────────
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        libaio1 \
        wget \
    && rm -rf /var/lib/apt/lists/*

# Oracle Instant Client runtime — localiza libclntsh.so dinamicamente
COPY --from=oracle-ic /usr/lib/oracle /usr/lib/oracle
RUN find /usr/lib/oracle -maxdepth 4 -name "libclntsh.so" -type f 2>/dev/null \
        | head -1 | xargs -r dirname \
        | xargs -r sh -c 'echo "$1" > /etc/ld.so.conf.d/oracle.conf' -- && \
    ldconfig

COPY --from=builder /build/portal /portal

EXPOSE 8080

ENTRYPOINT ["/portal"]
