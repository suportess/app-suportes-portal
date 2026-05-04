# SPEC — Regras de Negócio

---

## 1. Autenticação

- Todas as rotas de administração (`/certificates`, `/commands`, `/databases`, `/routes`) exigem header `Authorization: Bearer <token>`.
- O token é comparado com a variável de ambiente `GATEWAY_API_KEY`.
- Se `GATEWAY_API_KEY` não estiver definida, o padrão `gateway-default-api-key-2025` é usado.
- Rotas `/status` e `/health` são **públicas**.
- Rotas dinâmicas registradas via `/routes` são **públicas** (sem middleware de autenticação).
- Falha de autenticação retorna HTTP 401.

---

## 2. Persistência (BoltDB)

- Toda a configuração do portal (databases, commands, routes, certificates) é armazenada em um arquivo BoltDB local.
- O arquivo é configurado via `DATABASE` e `DATABASE_PATH`.
- **Não há banco relacional para o próprio portal** — o BoltDB é o único armazenamento interno.
- IDs são gerados automaticamente com auto-increment pelo storm.
- Chaves (`key` / `chave` / `nome`) são únicas por entidade — tentativa de duplicata retorna HTTP 409.

---

## 3. Conexões de Banco (Database)

- Ao criar uma conexão (`POST /databases`), o portal **testa a conectividade imediatamente** com timeout de 5 segundos.
  - Se o teste falhar, a conexão **não** é persistida e retorna erro 500.
- Conexões são mantidas em um **pool em memória** durante o ciclo de vida do processo.
  - Na inicialização, todas as conexões persistidas são carregadas e seus pools abertos.
- A senha é armazenada no BoltDB mas **nunca retornada** nas respostas da API (omitida em `DatabaseResponse`).
- **Configuração do pool (padrões):**
  | Campo              | Padrão | Descrição                                |
  |--------------------|--------|------------------------------------------|
  | `maxOpenConns`     | 10     | Máximo de conexões abertas               |
  | `maxIdleConns`     | 5      | Máximo de conexões ociosas               |
  | `connMaxLifetimeSec` | 90   | Tempo máximo de vida de uma conexão (s)  |
  | `connMaxIdleTimeSec` | 30   | Tempo máximo ocioso de uma conexão (s)   |
- **DSN por driver:**
  | Driver      | Formato                                                       |
  |-------------|---------------------------------------------------------------|
  | `mysql`     | `user:pass@tcp(host:port)/dbName?parseTime=true`             |
  | `postgres`  | `host=H port=P user=U password=P dbname=D sslmode=disable`   |
  | `oracle`    | `user="U" password="P" connectString="H:port/service"`       |
  | `sqlserver` | `server=H;user id=U;password=P;port=P;database=D;encrypt=disable` |

---

## 4. Comandos (Command)

### 4.1 Tipos de comando
- **SQL:** `QUERY`, `INSERT`, `UPDATE`, `DELETE`, `PROCEDURE`
- **HTTP:** `POST`, `GET`, `PUT`, `REMOVE`

### 4.2 Validações na criação
- `chave` é obrigatória e única.
- `tipo` deve ser um valor válido.
- Para SQL (exceto PROCEDURE): `sql` ou `tabela` é obrigatório.
- Para PROCEDURE: `sql` com a chamada da procedure é obrigatório.
- Para HTTP: `rota` é obrigatória.
- Se `tipoBanco` for informado, deve ser um driver válido.
- Tipos dos campos de `corpo` e `parametros` devem ser valores válidos.
- Operadores dos `parametros` devem ser válidos.

### 4.3 Processamento de parâmetros SQL
Ao criar/atualizar um comando SQL, o portal:
1. Verifica se o SQL contém `WHERE` → define `temFiltro = true/false`.
2. Extrai parâmetros do SQL via regex (`@param` para SQL Server, `:param` para demais).
3. Parâmetros encontrados no SQL que não estejam na lista `parametros` são adicionados automaticamente com `obrigatorio: true`.
4. Parâmetros já presentes no SQL recebem `jaAdicionado: true`.

### 4.4 Tipos de parâmetro SQL
| Tipo      | Descrição                                    |
|-----------|----------------------------------------------|
| `string`  | Texto                                        |
| `number`  | Numérico                                     |
| `boolean` | Booleano                                     |
| `date`    | Data                                         |
| `out`     | Parâmetro de saída (Oracle/SQL Server)       |
| `base64`  | Arquivo em base64 (upload)                   |
| `array`   | Lista de valores (expandido com `IN`)        |

### 4.5 Operadores SQL suportados
`=`, `<>`, `>`, `<`, `>=`, `<=`, `LIKE`, `NOT LIKE`, `IN`, `NOT IN`, `IS NULL`, `IS NOT NULL`

### 4.6 Paginação automática (QUERY)
Quando `paginacao` é definida em um comando QUERY, o portal injeta `LIMIT/OFFSET` automaticamente:
- **Parâmetros de paginação (defaults):**
  | Campo                 | Padrão     | Descrição                           |
  |-----------------------|------------|-------------------------------------|
  | `paramPagina`         | `page`     | Nome do param que indica a página   |
  | `paramTamanhoPagina`  | `pageSize` | Nome do param que indica o tamanho  |
  | `tamanhoPaginaPadrao` | `20`       | Tamanho padrão quando não informado |
  | `tamanhoPaginaMaximo` | `100`      | Limite máximo de registros          |
- Resposta quando paginação ativa:
  ```json
  { "dados": [...], "pagina": 1, "tamanhoPagina": 20 }
  ```
- Sintaxe por driver:
  - **MySQL / PostgreSQL:** `LIMIT {size} OFFSET {offset}`
  - **Oracle 12c+:** `OFFSET {offset} ROWS FETCH NEXT {size} ROWS ONLY`
  - **SQL Server:** `ORDER BY (SELECT NULL) OFFSET {offset} ROWS FETCH NEXT {size} ROWS ONLY`

### 4.7 Named params vs Positional params
- **SQL Server e Oracle** usam named params (`@param` / `:param`) → `SupportsNamedParams() = true`
- **MySQL e PostgreSQL** usam `?` posicional

### 4.8 Campos de body HTTP (`corpo` / `consulta`)
| Tipo      | Validação                                      |
|-----------|------------------------------------------------|
| `string`  | Valida tipo, `maximo` (len) e `minimo` (len)   |
| `number`  | Valida se é float64                            |
| `boolean` | Valida se é bool                               |
| `base64`  | Valida se é string base64 decodificável        |
| `array`   | Sem validação adicional                        |
| `object`  | Sem validação adicional                        |

---

## 5. Rotas Dinâmicas (Route)

### 5.1 Registro
- Uma rota é registrada no router em memória **imediatamente** após `POST /routes`.
- A rota é também persistida no BoltDB para recarregamento na inicialização.
- Na inicialização, todas as rotas persistidas são remontadas automaticamente.

### 5.2 Desregistro
- `DELETE /routes/{id}` substitui o handler por um `404 Not Found` (shadow handler).
  - O chi não permite remoção de rotas em runtime — o shadow handler é a solução adotada.
- A rota é removida do BoltDB após o shadow.

### 5.3 Validações
- `chave`, `caminho` e `metodo` são obrigatórios.
- `caminho` deve começar com `/`.
- `metodo` deve ser um método HTTP válido: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`.
- `servico.passos` deve ter ao menos 1 passo.
- Cada passo deve ter `comando.nome` e `comando.tipo` válidos.
- Passos com `comando.tipo` SQL devem ter `comando.database` preenchido.
- Se `resposta.status` não for informado, o padrão é `200`.

### 5.4 Estrutura de um passo (RouteStep)
| Campo             | Tipo    | Descrição                                                    |
|-------------------|---------|--------------------------------------------------------------|
| `alias`           | string  | Nome do resultado deste passo (usado para extração de campos)|
| `abortarSemDados` | bool    | Se true, aborta a pipeline se este passo não retornar dados  |
| `comando`         | objeto  | Definição do comando a executar                              |

### 5.5 Estrutura de um comando de passo (RouteCommand)
| Campo               | Tipo   | Descrição                                                    |
|---------------------|--------|--------------------------------------------------------------|
| `tipo`              | string | Tipo do comando (`QUERY`, `INSERT`, `POST`, etc.)            |
| `database`          | string | Chave da conexão de banco (obrigatório para SQL)             |
| `nome`              | string | Chave do comando cadastrado em `/commands`                   |
| `retornarResultado` | bool   | Se true, o resultado deste passo compõe a resposta final     |
| `parametros`        | array  | Mapeamento de parâmetros para o comando                      |

### 5.6 Parâmetros de rota (RouteParameter)
| Campo    | Tipo      | Descrição                                                           |
|----------|-----------|---------------------------------------------------------------------|
| `nome`   | string    | Nome do parâmetro no comando de destino                             |
| `tipo`   | string    | Tipo do valor (`path`, `query`, `body`, `header`, `alias`)          |
| `extrair`| bool      | Se true, extrai o valor do resultado de um passo anterior (via `campo`) |
| `valor`  | any       | Valor fixo (usado quando `extrair = false` e não há source dinâmico)|
| `campo`  | string    | Nome do campo a extrair do resultado de passo anterior              |

### 5.7 Pipeline de execução
1. Para cada passo da rota, na ordem definida:
   a. O comando correspondente é buscado no BoltDB pelo `nome`.
   b. Os parâmetros são resolvidos (path params, query params, body, extrações de passos anteriores).
   c. O executor correto é selecionado pelo `tipo` do comando.
   d. O executor roda e retorna o resultado.
   e. Se `abortarSemDados = true` e o resultado for vazio → pipeline abortada, retorna 204.
2. Resultados de passos com `retornarResultado = true` são coletados.
3. Se `resultadoUnico = true`: retorna apenas o primeiro item do resultado.
4. Se `threadUnica = true`: passos são executados sequencialmente (sem concorrência).
5. A resposta final usa o `status` e `tipoConteudo` configurados em `resposta`.

---

## 6. Certificados (Certificate)

- Certificados são usados por comandos HTTP para autenticação mTLS.
- O arquivo de certificado deve ser enviado em **base64**.
- Suporta PEM (`arquivoCert` + `arquivoChave`) ou PFX (`arquivoPfx` + `senha`).
- `arquivoCACert` é opcional — CA adicional para validação do servidor.
- `ativo` controla se o certificado pode ser usado (padrão: `true` na criação).
- `dataExpiracao` é informativa — não há bloqueio automático ao vencer.
- Na execução de comando HTTP com `nomeCertificado`, o portal:
  1. Busca o certificado pelo nome.
  2. Decodifica o PFX via `pkcs12.DecodeChain`.
  3. Monta um `tls.Config` com o certificado e CA.
  4. Cria um `http.Client` com o `Transport` TLS configurado.
  - `InsecureSkipVerify: true` — verificação de hostname desativada (comportamento atual).

---

## 7. Executores SQL

### QUERY
- Executa SELECT e retorna as linhas como array de objetos.
- Se `resultadoUnico = true` na rota: retorna apenas o primeiro objeto.
- Suporta filtros dinâmicos via `parametros` (injeta `WHERE` / `AND` automaticamente).
- Suporta paginação via `paginacao` (injeta `LIMIT/OFFSET` ou equivalente).
- Suporta ordenação via `ordenacao` (`nomeColuna` + `decrescente`).

### INSERT
- Requer `corpo` com os campos a inserir.
- `tabela` é obrigatória.
- Retorna:
  ```json
  { "ultimoIdInserido": 42, "linhasAfetadas": 1 }
  ```

### UPDATE
- Requer `corpo` com os campos a atualizar.
- `tabela` é obrigatória.
- Parâmetros de filtro definem o `WHERE`.
- Retorna:
  ```json
  { "linhasAfetadas": 1 }
  ```

### DELETE
- `sql` customizado **ou** `tabela` + parâmetros de filtro.
- Retorna:
  ```json
  { "linhasAfetadas": 1 }
  ```

### PROCEDURE
- Executa via `sql` (ex: `EXEC minha_proc @param1=:p1`).
- Parâmetros do tipo `out` são capturados e retornados.
- Se não houver parâmetros de saída, retorna `{ "status": "concluido" }`.

---

## 8. Executores HTTP

### POST / PUT / GET / REMOVE
- `rota` do comando é a URL do serviço externo (suporta placeholders `{param}`).
- `tipoConteudo` define como o body é enviado:
  | Valor                               | Formato              |
  |-------------------------------------|----------------------|
  | `application/json` (padrão)         | JSON                 |
  | `application/x-www-form-urlencoded` | Form URL encoded     |
  | `multipart/form-data`               | Multipart form data  |
- Se `nomeCertificado` for informado, usa mTLS via certificado cadastrado.
- Se a resposta externa for `application/json`, retorna deserializada; caso contrário, retorna como string.
- Respostas HTTP >= 400 do serviço externo são propagadas como erro.
