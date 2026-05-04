# ── Stage 1: Oracle Instant Client 21 ────────────────────────────────────────
# oraclelinux:8 está disponível no Docker Hub sem autenticação e já tem os
# repositórios Oracle configurados (ol8_appstream inclui oracle-instantclient-release-el8)
FROM oraclelinux:8 AS oracle-ic
RUN dnf install -y oracle-instantclient-release-el8 && \
    dnf install -y oracle-instantclient21-basic && \
    rm -rf /var/cache/dnf

# ── Stage 2: build ────────────────────────────────────────────────────────────
FROM golang:1.24-bullseye AS builder

# Oracle Instant Client necessário para compilar godror/ODPI-C com CGO
COPY --from=oracle-ic /usr/lib/oracle/21/client64 /usr/lib/oracle/21/client64
RUN echo /usr/lib/oracle/21/client64/lib > /etc/ld.so.conf.d/oracle.conf && ldconfig

WORKDIR /build

# Cache de dependências antes de copiar o código-fonte
COPY app/go.mod app/go.sum ./
RUN go mod download

# Copia o restante do código e compila
# CGO_ENABLED=1 é obrigatório — godror (Oracle) usa ODPI-C (biblioteca C)
COPY app/ .
RUN CGO_ENABLED=1 GOOS=linux \
    CGO_CFLAGS="-I/usr/lib/oracle/21/client64/include" \
    CGO_LDFLAGS="-L/usr/lib/oracle/21/client64/lib" \
    go build \
    -ldflags="-w -s" \
    -o portal \
    ./cmd/main.go

# ── Stage 3: imagem final ─────────────────────────────────────────────────────
# Usa debian-slim (tem glibc) em vez de scratch — necessário para CGO + libaio
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        libaio1 \
        wget \
    && rm -rf /var/lib/apt/lists/*

# Oracle Instant Client runtime — libclntsh.so e dependências
COPY --from=oracle-ic /usr/lib/oracle/21/client64/lib/ /usr/lib/oracle/21/client64/lib/
RUN echo /usr/lib/oracle/21/client64/lib > /etc/ld.so.conf.d/oracle.conf && ldconfig

COPY --from=builder /build/portal /portal

EXPOSE 8080

ENTRYPOINT ["/portal"]
