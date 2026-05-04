# SPEC — Visão Geral do Projeto PORTAL

## Propósito
Gateway HTTP dinâmico que funciona como intermediário entre sistemas externos e bancos de dados relacionais.
Permite cadastrar conexões de banco, registrar comandos SQL/HTTP reutilizáveis e montar rotas HTTP dinamicamente —
tudo via API REST, sem necessidade de redeploy.

## Tecnologias
| Componente       | Tecnologia                          |
|------------------|-------------------------------------|
| Linguagem        | Go 1.23                             |
| Módulo           | `br.tec.suportes/portal`            |
| HTTP Router      | go-chi/chi v5                       |
| Persistência     | BoltDB via asdine/storm v3          |
| Drivers SQL      | MySQL, PostgreSQL, Oracle (godror), SQL Server |
| Autenticação     | Bearer token (env `GATEWAY_API_KEY`) |
| Pool de conexões | sqlx + configuração por banco        |

## Estrutura de pastas
```
portal/
├── .github/specs/          ← documentação técnica (este diretório)
└── app/
    ├── go.mod
    ├── .env.example
    ├── cmd/
    │   └── main.go         ← entrypoint: carrega config, monta rotas, inicia servidor
    └── internal/
        ├── apierr/         ← tipos de erro da API (Detail interface)
        ├── config/         ← variáveis de ambiente (Config struct)
        ├── domain/         ← entidades e request/response: certificate, command, database, route
        ├── enum/           ← enumerações: DBDriver, CommandType, HTTPMethod, ParamType, etc.
        ├── execute/        ← executores: query, insert, update, delete, procedure, post, get, put, remove
        ├── http/           ← handlers HTTP: certificate, command, database, route, status
        ├── repository/     ← acesso ao BoltDB: certificate, command, database, route
        ├── service/        ← lógica de negócio: certificate, command, database, route
        └── store/          ← abertura e configuração do BoltDB
```

## Variáveis de ambiente
| Variável               | Padrão                        | Descrição                                      |
|------------------------|-------------------------------|------------------------------------------------|
| `SERVER_PORT`          | `8080`                        | Porta HTTP do servidor                         |
| `DEFAULT_TIMEOUT`      | `60`                          | Timeout padrão em segundos para requisições    |
| `STATUS`               | `UP`                          | Status da aplicação (`UP`, `DOWN`, `MAINTENANCE`) |
| `GATEWAY_API_KEY`      | `gateway-default-api-key-2025`| Bearer token exigido em todas as rotas de admin|
| `DATABASE`             | `gateway.db`                  | Nome do arquivo BoltDB                         |
| `DATABASE_PATH`        | `db`                          | Diretório do arquivo BoltDB                    |
| `DATABASE_TIMEOUT`     | `10`                          | Timeout em segundos para operações no BoltDB   |
| `JAEGER_ENABLED`       | `false`                       | Habilita tracing distribuído (Jaeger/OTEL)     |
| `JAEGER_SERVICE_NAME`  | `gateway`                     | Nome do serviço no Jaeger                      |
| `JAEGER_ENDPOINT`      | `http://localhost:4318`       | Endpoint OTLP do Jaeger                        |

## Como executar
```powershell
# Instalar dependências
go mod tidy

# Executar
go run ./cmd/main.go

# Build
go build -o portal.exe ./cmd/main.go
```

## Autenticação
Todas as rotas de administração (`/certificates`, `/commands`, `/databases`, `/routes`) exigem:
```
Authorization: Bearer <GATEWAY_API_KEY>
```
As rotas `/status` e `/health` são **públicas** (sem autenticação).
As **rotas dinâmicas** registradas via `/routes` também são **públicas** por padrão.

## Formato de erro padrão
```json
{
  "timestamp": "2026-05-01T21:00:00-03:00",
  "status": 400,
  "error": "Bad Request",
  "message": "campo 'chave' é obrigatório",
  "path": "/commands",
  "trace": null
}
```

## Arquivos de spec
| Arquivo                   | Conteúdo                                                              |
|---------------------------|-----------------------------------------------------------------------|
| `01-rotas-api.md`         | Todas as rotas disponíveis, parâmetros, payloads e respostas          |
| `02-regras-negocio.md`    | Regras de validação, pipeline de execução, tipos de comando e rota    |
| `03-exemplos-payloads.md` | Exemplos completos de JSON para cada recurso                          |
