# 🤖 Agente

**Agente** é um gateway universal e leve para execução de comandos em bancos de dados e chamadas a APIs REST externas. Reescrito do zero com arquitetura limpa, enums, validações robustas e payloads que fazem sentido.

---

## 🏗️ Estrutura do Projeto

```
agente/
├── app/                    # Código-fonte Go
│   ├── cmd/
│   │   └── main.go         # Ponto de entrada
│   ├── internal/
│   │   ├── apierr/         # Erros HTTP estruturados
│   │   ├── config/         # Configuração via env vars
│   │   ├── domain/         # Entidades + payloads de request com validação
│   │   ├── enum/           # Enums tipados (DBDriver, CommandType, HTTPMethod, …)
│   │   ├── execute/        # Executores (QUERY, INSERT, UPDATE, DELETE, PROCEDURE, POST, GET, PUT, REMOVE)
│   │   ├── http/           # Handlers HTTP (database, command, route, certificate, status)
│   │   ├── repository/     # Repositórios BoltDB (storm)
│   │   ├── service/        # Serviços de negócio
│   │   └── store/          # Wrapper BoltDB
│   ├── go.mod
│   ├── go.sum
│   └── .env.example
└── doc/
    └── insomnia_collection.json   # Coleção completa do Insomnia
```

---

## ⚡ Enums Disponíveis

### `DBDriver` — drivers suportados
| Valor        | Banco              |
|--------------|--------------------|
| `mysql`      | MySQL              |
| `postgres`   | PostgreSQL         |
| `oracle`     | Oracle DB          |
| `sqlserver`  | Microsoft SQL Server |

### `CommandType` — tipos de comando
| Valor       | Descrição                        |
|-------------|----------------------------------|
| `QUERY`     | SELECT SQL                       |
| `INSERT`    | INSERT SQL via `table`           |
| `UPDATE`    | UPDATE SQL via `table` ou `sql`  |
| `DELETE`    | DELETE SQL via `table` ou `sql`  |
| `PROCEDURE` | Stored procedure com OUT params  |
| `POST`      | HTTP POST para API externa       |
| `PUT`       | HTTP PUT para API externa        |
| `GET`       | HTTP GET para API externa        |
| `REMOVE`    | HTTP DELETE para API externa     |

### `HTTPMethod` — métodos HTTP de rotas dinâmicas
`GET` | `POST` | `PUT` | `DELETE` | `PATCH`

### `ParamType` — tipos de parâmetros SQL
`string` | `number` | `boolean` | `date` | `out` | `base64` | `array`

### `FieldType` — tipos de campos de body HTTP
`string` | `number` | `boolean` | `base64` | `array` | `object`

### `SQLOperator` — operadores de filtros dinâmicos
`=` | `<>` | `>` | `<` | `>=` | `<=` | `LIKE` | `NOT LIKE` | `IN` | `NOT IN` | `IS NULL` | `IS NOT NULL`

---

## 🚀 Como Executar

### Pré-requisitos
- Go 1.23+
- Oracle Instant Client (apenas se usar driver Oracle)

### Localmente
```bash
# 1. Configure o ambiente
cp .env.example .env
# Edite .env conforme necessário

# 2. Execute
go run cmd/main.go
```

A aplicação estará disponível em `http://localhost:8080`

### Build
```bash
go build -o agente cmd/main.go
./agente
```

---

## 🔐 Autenticação

Todos os endpoints de gerenciamento exigem o header:
```
Authorization: Bearer <AGENTE_API_KEY>
```

O valor padrão é `agente-default-api-key-2025`. Configure via env var `AGENTE_API_KEY`.

---

## 📡 Endpoints

### Databases — `/databases`
| Método | Path              | Descrição                          |
|--------|-------------------|------------------------------------|
| GET    | `/databases`      | Lista todas as conexões            |
| GET    | `/databases?key=` | Busca por key                      |
| GET    | `/databases/{id}` | Busca por ID                       |
| POST   | `/databases`      | Cria e testa nova conexão          |
| DELETE | `/databases/{id}` | Remove por ID                      |
| DELETE | `/databases/{id}?key=` | Remove por key              |

**Payload de criação:**
```json
{
  "key": "meu-banco",
  "driver": "postgres",
  "host": "localhost",
  "port": 5432,
  "dbName": "minha_base",
  "user": "postgres",
  "password": "senha",
  "pool": {
    "maxOpenConns": 10,
    "maxIdleConns": 5,
    "connMaxLifetimeSec": 90,
    "connMaxIdleTimeSec": 30
  }
}
```

---

### Commands — `/commands`
| Método | Path               | Descrição                          |
|--------|--------------------|------------------------------------|
| GET    | `/commands`        | Lista todos                        |
| GET    | `/commands?key=`   | Busca por key (pattern)            |
| GET    | `/commands/{id}`   | Busca por ID                       |
| POST   | `/commands`        | Cria novo comando                  |
| PUT    | `/commands/{id}`   | Atualiza (key é imutável)          |
| DELETE | `/commands/{id}`   | Remove por ID                      |
| DELETE | `/commands/{id}?key=` | Remove por key (pattern)       |

---

### Routes — `/routes`
| Método | Path              | Descrição                          |
|--------|-------------------|------------------------------------|
| GET    | `/routes`         | Lista todas as rotas               |
| GET    | `/routes?key=`    | Filtra por key                     |
| GET    | `/routes?method=` | Filtra por método HTTP             |
| GET    | `/routes?path=`   | Filtra por path                    |
| GET    | `/routes/{id}`    | Busca por ID                       |
| POST   | `/routes`         | Registra nova rota dinâmica        |
| DELETE | `/routes/{id}`    | Remove por ID                      |
| DELETE | `/routes/{id}?key=` | Remove por key (pattern)       |

---

### Certificates — `/certificates`
| Método | Path                    | Descrição                    |
|--------|-------------------------|------------------------------|
| GET    | `/certificates`         | Lista todos                  |
| GET    | `/certificates?name=`   | Busca por name               |
| GET    | `/certificates/{id}`    | Busca por ID                 |
| POST   | `/certificates`         | Cria certificado             |
| PUT    | `/certificates/{id}`    | Atualiza certificado         |
| DELETE | `/certificates/{id}`    | Remove por ID                |
| DELETE | `/certificates/{id}?name=` | Remove por name         |

---

### Status — sem autenticação
| Método | Path      | Descrição                              |
|--------|-----------|----------------------------------------|
| GET    | `/status` | Status da aplicação + conexões DB      |
| GET    | `/health` | Métricas de saúde (memória, goroutines)|

---

## 🔀 Fluxo de uma Rota Dinâmica

```
Cliente → POST /api/v1/pedidos/100/processar
             │
             ▼
         Pipeline (steps sequenciais ou paralelos)
             │
             ├── Step 1: QUERY → busca pedido no banco
             │
             └── Step 2: POST → notifica API externa com dados do step 1
             │
             ▼
         Resposta JSON consolidada
```

---

## 📦 Dependências Principais

| Pacote | Uso |
|--------|-----|
| `github.com/go-chi/chi/v5` | Roteamento HTTP |
| `github.com/go-chi/cors` | CORS |
| `github.com/asdine/storm/v3` | ORM sobre BoltDB |
| `go.etcd.io/bbolt` | BoltDB embutido |
| `github.com/jmoiron/sqlx` | Queries SQL tipadas |
| `github.com/godror/godror` | Driver Oracle |
| `github.com/lib/pq` | Driver PostgreSQL |
| `github.com/go-sql-driver/mysql` | Driver MySQL |
| `github.com/microsoft/go-mssqldb` | Driver SQL Server |
| `github.com/xwb1989/sqlparser` | Validação SQL |
| `github.com/caarlos0/env/v8` | Parse de env vars |
| `software.sslmate.com/src/go-pkcs12` | Certificados PFX/mTLS |

---

## 📝 Variáveis de Ambiente

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `SERVER_PORT` | `8080` | Porta HTTP |
| `DEFAULT_TIMEOUT` | `60` | Timeout de requests (segundos) |
| `STATUS` | `UP` | Status da aplicação |
| `AGENTE_API_KEY` | `agente-default-api-key-2025` | Bearer token de autenticação |
| `DATABASE` | `agente.db` | Nome do arquivo BoltDB |
| `DATABASE_PATH` | `db` | Diretório do arquivo BoltDB |
| `DATABASE_TIMEOUT` | `10` | Timeout de abertura do BoltDB (segundos) |
| `JAEGER_ENABLED` | `false` | Habilita tracing distribuído |
| `JAEGER_SERVICE_NAME` | `agente` | Nome do serviço no Jaeger |
| `JAEGER_SERVICE_VERSION` | `1.0.0` | Versão do serviço |
| `JAEGER_ENVIRONMENT` | `development` | Ambiente (dev/staging/prod) |
| `JAEGER_ENDPOINT` | `http://localhost:4318` | Endpoint OTLP do Jaeger |
