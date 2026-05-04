# SPEC — Rotas da API

> Todas as rotas abaixo (exceto `/status` e `/health`) exigem header:
> `Authorization: Bearer <GATEWAY_API_KEY>`

---

## /status — Status da aplicação

### GET /status
Retorna o status configurado da aplicação e o estado das conexões de banco cadastradas.

**Autenticação:** não exigida

**Resposta 200:**
```json
{
  "status": "UP",
  "uptime": "2h34m12s",
  "databases": {
    "dbamv": "conectado",
    "legado": "desconectado"
  }
}
```

---

### GET /health
Retorna informações de saúde da instância em execução (goroutines, memória).

**Autenticação:** não exigida

**Resposta 200:**
```json
{
  "status": "UP",
  "uptime": "2h34m12s",
  "goroutines": 14,
  "memStats": {
    "allocMB": 8,
    "totalAllocMB": 42,
    "sysMB": 22
  }
}
```

---

## /databases — Conexões de banco de dados

### GET /databases
Lista todas as conexões cadastradas. A senha **não** é retornada.

**Query params:**
| Parâmetro | Tipo   | Descrição                                  |
|-----------|--------|--------------------------------------------|
| `key`     | string | Opcional. Filtra pelo key exato da conexão |

**Resposta 200:** array de `DatabaseResponse`
**Resposta 204:** nenhuma conexão cadastrada

---

### GET /databases/{id}
Retorna uma conexão pelo ID numérico interno.

**Path params:**
| Parâmetro | Tipo    | Descrição           |
|-----------|---------|---------------------|
| `id`      | integer | ID interno do banco |

**Resposta 200:** objeto `DatabaseResponse`
**Resposta 404:** não encontrado

---

### POST /databases
Cadastra uma nova conexão de banco e **testa a conectividade** imediatamente.
Se o teste de conexão falhar, a conexão **não** é persistida.

**Body:** `CreateDatabaseRequest`
```json
{
  "key": "dbamv",
  "driver": "postgres",
  "host": "localhost",
  "port": 5435,
  "dbName": "dbamv",
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

**Campos obrigatórios:** `key`, `driver`, `host`, `port`, `dbName`, `user`, `password`

**Drivers suportados:** `mysql`, `postgres`, `oracle`, `sqlserver`

**Campos Oracle opcionais (mutuamente exclusivos, em ordem de prioridade):**
| Campo           | Descrição                                    |
|-----------------|----------------------------------------------|
| `connectString` | DSN completo (passado direto ao godror)       |
| `sid`           | SID Oracle → `host:port:SID`                |
| `serviceName`   | Service name → `host:port/serviceName`       |
| `dbName`        | Usado como service name se nenhum acima for informado |

**Resposta 201:** `DatabaseResponse` (sem senha, com `dsnUsed`)
**Resposta 400/422:** erro de validação
**Resposta 409:** `key` já cadastrada
**Resposta 500:** falha ao testar conexão

---

### DELETE /databases/{id}
Remove uma conexão pelo ID ou pelo `key` (query param).

**Path params:** `id` (integer)
**Query params:** `key` (string) — se informado, ignora o `id`

**Resposta 204:** removido com sucesso
**Resposta 404:** não encontrado

---

## /commands — Comandos reutilizáveis

### GET /commands
Lista todos os comandos cadastrados.

**Query params:**
| Parâmetro | Tipo   | Descrição                                              |
|-----------|--------|--------------------------------------------------------|
| `key`     | string | Opcional. Filtra por padrão de chave (busca por prefixo) |

**Resposta 200:** array de `Command`
**Resposta 204:** nenhum comando

---

### GET /commands/{id}
Retorna um comando pelo ID numérico interno.

**Resposta 200:** objeto `Command`
**Resposta 404:** não encontrado

---

### POST /commands
Cadastra um novo comando.

**Body:** `CreateCommandRequest`
```json
{
  "chave": "buscar-produtos",
  "tipo": "QUERY",
  "tipoBanco": "postgres",
  "sql": "SELECT * FROM PRODUTO WHERE CD_PRODUTO = :cd_produto",
  "descricao": "Busca produto por código",
  "parametros": [
    { "nome": "cd_produto", "tipo": "number", "obrigatorio": true }
  ]
}
```

**Campo `tipo` — valores aceitos:**
| Valor       | Categoria | Descrição                                    |
|-------------|-----------|----------------------------------------------|
| `QUERY`     | SQL       | SELECT — retorna linhas                      |
| `INSERT`    | SQL       | INSERT — insere registro                     |
| `UPDATE`    | SQL       | UPDATE — atualiza registro                   |
| `DELETE`    | SQL       | DELETE — remove registro                     |
| `PROCEDURE` | SQL       | Chamada de stored procedure                  |
| `POST`      | HTTP      | Requisição HTTP POST para serviço externo    |
| `GET`       | HTTP      | Requisição HTTP GET para serviço externo     |
| `PUT`       | HTTP      | Requisição HTTP PUT para serviço externo     |
| `REMOVE`    | HTTP      | Requisição HTTP DELETE para serviço externo  |

**Campos por tipo:**
| Tipo         | Campo obrigatório       | Campo opcional               |
|--------------|-------------------------|------------------------------|
| QUERY/INSERT/UPDATE/DELETE | `sql` ou `tabela` | `tipoBanco`, `parametros`, `ordenacao`, `paginacao` |
| PROCEDURE    | `sql`                   | `tipoBanco`, `parametros`    |
| POST/PUT/GET/REMOVE | `rota`           | `tipoConteudo`, `nomeCertificado`, `corpo`, `consulta` |

**Resposta 201:** objeto `Command` criado
**Resposta 400/422:** erro de validação
**Resposta 409:** `chave` já cadastrada

---

### PUT /commands/{id}
Atualiza um comando existente. A `chave` é imutável.

**Body:** `UpdateCommandRequest` — mesmos campos do POST, todos opcionais.

**Resposta 200:** objeto `Command` atualizado
**Resposta 404:** não encontrado

---

### DELETE /commands/{id}
Remove um comando pelo ID ou pelo `key`.

**Query params:** `key` (string) — se informado, remove todos que casam com o padrão
**Resposta 204:** removido
**Resposta 404:** não encontrado

---

## /routes — Rotas dinâmicas

### GET /routes
Lista todas as rotas registradas.

**Query params (mutuamente exclusivos):**
| Parâmetro | Tipo   | Descrição                          |
|-----------|--------|------------------------------------|
| `key`     | string | Filtra por padrão de chave         |
| `method`  | string | Filtra por método HTTP (`GET`, `POST`, etc.) |
| `path`    | string | Filtra por caminho exato           |

**Resposta 200:** array de `Route`
**Resposta 204:** nenhuma rota

---

### GET /routes/{id}
Retorna uma rota pelo ID numérico interno.

**Resposta 200:** objeto `Route`
**Resposta 404:** não encontrado

---

### POST /routes
Registra uma nova rota dinâmica. A rota é imediatamente disponibilizada no servidor.

**Body:** `CreateRouteRequest`
```json
{
  "chave": "listar-produtos",
  "caminho": "/api/produtos",
  "metodo": "GET",
  "resposta": { "status": 200 },
  "servico": {
    "threadUnica": false,
    "resultadoUnico": false,
    "passos": [
      {
        "alias": "produtos",
        "abortarSemDados": false,
        "comando": {
          "tipo": "QUERY",
          "database": "dbamv",
          "nome": "buscar-produtos",
          "retornarResultado": true,
          "parametros": []
        }
      }
    ]
  }
}
```

**Métodos HTTP aceitos nas rotas:** `GET`, `POST`, `PUT`, `DELETE`, `PATCH`

**Campo `database`:** chave da conexão cadastrada em `/databases`. Obrigatório para comandos SQL.

**Resposta 201:** objeto `Route` registrado
**Resposta 400/422:** erro de validação
**Resposta 409:** `chave` já cadastrada

---

### DELETE /routes/{id}
Remove e desregistra uma rota pelo ID ou pelo padrão de `key`.

**Query params:** `key` (string) — remove todas as rotas que casam com o padrão
**Resposta 204:** removido por ID
**Resposta 200:** removido por key (retorna contagem)
```json
{
  "mensagem": "rotas removidas com sucesso",
  "chave": "listar-produtos",
  "totalRemovido": 1
}
```
**Resposta 404:** não encontrado

---

## /certificates — Certificados TLS/mTLS

### GET /certificates
Lista todos os certificados cadastrados.

**Query params:**
| Parâmetro | Tipo   | Descrição                          |
|-----------|--------|------------------------------------|
| `name`    | string | Filtra pelo nome exato do certificado |

**Resposta 200:** array de `Certificate`
**Resposta 204:** nenhum certificado

---

### GET /certificates/{id}
Retorna um certificado pelo ID.

**Resposta 200:** objeto `Certificate`
**Resposta 404:** não encontrado

---

### POST /certificates
Cadastra um novo certificado TLS/mTLS. Os arquivos devem ser enviados em **base64**.

**Body:** `CreateCertificateRequest`
```json
{
  "nome": "cert-banco",
  "descricao": "Certificado mTLS do banco legado",
  "arquivoCert": "<PEM base64>",
  "arquivoChave": "<chave privada base64>",
  "arquivoPfx": "<PFX base64>",
  "senha": "senha-do-pfx",
  "arquivoCACert": "<CA base64>",
  "dataExpiracao": "2027-12-31"
}
```

**Regra:** `arquivoCert` **ou** `arquivoPfx` deve ser informado.
**`dataExpiracao`:** formato `YYYY-MM-DD`.

**Resposta 201:** objeto `Certificate` criado (`ativo: true` por padrão)
**Resposta 400/422:** erro de validação
**Resposta 409:** `nome` já cadastrado

---

### PUT /certificates/{id}
Atualiza um certificado. O `nome` é imutável.

**Body:** `UpdateCertificateRequest` — todos os campos opcionais:
```json
{
  "descricao": "Nova descrição",
  "ativo": false,
  "dataExpiracao": "2028-06-30"
}
```

**Resposta 200:** objeto `Certificate` atualizado
**Resposta 404:** não encontrado

---

### DELETE /certificates/{id}
Remove um certificado pelo ID ou pelo `name`.

**Query params:** `name` (string) — se informado, ignora o `id`
**Resposta 204:** removido
**Resposta 404:** não encontrado
