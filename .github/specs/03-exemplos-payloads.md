# SPEC — Exemplos de Payloads

---

## 1. Databases

### Criar conexão PostgreSQL
```json
POST /databases
{
  "key": "dbamv",
  "driver": "postgres",
  "host": "localhost",
  "port": 5435,
  "dbName": "dbamv",
  "user": "postgres",
  "password": "postgres",
  "pool": {
    "maxOpenConns": 10,
    "maxIdleConns": 5,
    "connMaxLifetimeSec": 90,
    "connMaxIdleTimeSec": 30
  }
}
```

### Criar conexão MySQL
```json
POST /databases
{
  "key": "mysql-legado",
  "driver": "mysql",
  "host": "192.168.1.50",
  "port": 3306,
  "dbName": "erp",
  "user": "app",
  "password": "senha123",
  "pool": {}
}
```

### Criar conexão SQL Server
```json
POST /databases
{
  "key": "sqlserver-mv",
  "driver": "sqlserver",
  "host": "srv-sql",
  "port": 1433,
  "dbName": "DBMV",
  "user": "sa",
  "password": "SenhaForte@2026",
  "pool": {
    "maxOpenConns": 20,
    "maxIdleConns": 5,
    "connMaxLifetimeSec": 120,
    "connMaxIdleTimeSec": 60
  }
}
```

### Criar conexão Oracle (por serviceName)
```json
POST /databases
{
  "key": "oracle-prod",
  "driver": "oracle",
  "host": "ora-srv",
  "port": 1521,
  "dbName": "ORCL",
  "serviceName": "ORCL.empresa.local",
  "user": "app_user",
  "password": "OraclePwd#1",
  "pool": {}
}
```

### Resposta (DatabaseResponse)
```json
{
  "id": 1,
  "key": "dbamv",
  "driver": "postgres",
  "host": "localhost",
  "port": 5435,
  "dbName": "dbamv",
  "user": "postgres",
  "pool": {
    "maxOpenConns": 10,
    "maxIdleConns": 5,
    "connMaxLifetimeSec": 90,
    "connMaxIdleTimeSec": 30
  },
  "dsnUsed": "host=localhost port=5435 user=postgres password=postgres dbname=dbamv sslmode=disable"
}
```

---

## 2. Commands

### Comando QUERY simples
```json
POST /commands
{
  "chave": "buscar-produto-por-cd",
  "descricao": "Busca produto pelo código",
  "tipo": "QUERY",
  "tipoBanco": "postgres",
  "sql": "SELECT cd_produto, nm_produto, sn_lote, sn_validade FROM \"DBAMV\".produto WHERE cd_produto = :cd_produto",
  "parametros": [
    { "nome": "cd_produto", "tipo": "number", "obrigatorio": true }
  ]
}
```

### Comando QUERY com filtros opcionais e paginação
```json
POST /commands
{
  "chave": "listar-entradas",
  "descricao": "Lista entradas com filtros e paginação",
  "tipo": "QUERY",
  "tipoBanco": "postgres",
  "sql": "SELECT e.cd_ent_pro, e.dt_ent_pro, f.nm_fornecedor, e.vl_total FROM \"DBAMV\".ent_pro e JOIN \"DBAMV\".fornecedor f ON f.cd_fornecedor = e.cd_fornecedor WHERE e.cd_estoque = :cd_estoque",
  "temFiltro": true,
  "parametros": [
    { "nome": "cd_estoque",    "tipo": "number", "obrigatorio": true,  "jaAdicionado": true },
    { "nome": "dt_ent_pro",    "tipo": "date",   "obrigatorio": false, "operador": ">="     },
    { "nome": "cd_fornecedor", "tipo": "number", "obrigatorio": false                       }
  ],
  "ordenacao": {
    "nomeColuna": "e.dt_ent_pro",
    "decrescente": true
  },
  "paginacao": {
    "paramPagina": "page",
    "paramTamanhoPagina": "pageSize",
    "tamanhoPaginaPadrao": 20,
    "tamanhoPaginaMaximo": 100
  }
}
```

### Comando QUERY — resultado único
```json
POST /commands
{
  "chave": "saldo-produto",
  "descricao": "Saldo atual de um produto em um estoque",
  "tipo": "QUERY",
  "tipoBanco": "postgres",
  "sql": "SELECT qt_estoque_atual FROM \"DBAMV\".est_pro WHERE cd_estoque = :cd_estoque AND cd_produto = :cd_produto",
  "parametros": [
    { "nome": "cd_estoque", "tipo": "number", "obrigatorio": true },
    { "nome": "cd_produto", "tipo": "number", "obrigatorio": true }
  ]
}
```

### Comando INSERT
```json
POST /commands
{
  "chave": "inserir-paciente",
  "descricao": "Insere um paciente",
  "tipo": "INSERT",
  "tipoBanco": "postgres",
  "tabela": "\"DBAMV\".paciente",
  "corpo": {
    "campos": [
      { "nome": "nm_paciente", "tipo": "string", "obrigatorio": true, "maximo": 150 }
    ]
  }
}
```

### Comando UPDATE
```json
POST /commands
{
  "chave": "concluir-entrada",
  "descricao": "Preenche DT_CONCLUSAO de uma entrada",
  "tipo": "UPDATE",
  "tipoBanco": "postgres",
  "tabela": "\"DBAMV\".ent_pro",
  "corpo": {
    "campos": [
      { "nome": "dt_conclusao", "tipo": "string", "obrigatorio": true }
    ]
  },
  "parametros": [
    { "nome": "cd_ent_pro", "tipo": "number", "obrigatorio": true }
  ]
}
```

### Comando DELETE
```json
POST /commands
{
  "chave": "excluir-item-entrada",
  "descricao": "Remove item de uma entrada",
  "tipo": "DELETE",
  "tipoBanco": "postgres",
  "sql": "DELETE FROM \"DBAMV\".itent_pro WHERE cd_itent_pro = :cd_itent_pro",
  "parametros": [
    { "nome": "cd_itent_pro", "tipo": "number", "obrigatorio": true }
  ]
}
```

### Comando PROCEDURE (Oracle)
```json
POST /commands
{
  "chave": "proc-atualiza-saldo",
  "descricao": "Chama procedure de atualização de saldo",
  "tipo": "PROCEDURE",
  "tipoBanco": "oracle",
  "sql": "BEGIN sp_atualiza_est_pro(:p_cd_estoque, :p_cd_produto, :p_cd_uni_pro, :p_qt_entrada, :p_sinal); END;",
  "parametros": [
    { "nome": "p_cd_estoque",  "tipo": "number", "obrigatorio": true },
    { "nome": "p_cd_produto",  "tipo": "number", "obrigatorio": true },
    { "nome": "p_cd_uni_pro",  "tipo": "number", "obrigatorio": true },
    { "nome": "p_qt_entrada",  "tipo": "number", "obrigatorio": true },
    { "nome": "p_sinal",       "tipo": "number", "obrigatorio": true }
  ]
}
```

### Comando GET HTTP
```json
POST /commands
{
  "chave": "consultar-cep",
  "descricao": "Consulta endereço por CEP via ViaCEP",
  "tipo": "GET",
  "rota": "https://viacep.com.br/ws/{cep}/json/"
}
```

### Comando POST HTTP com certificado mTLS
```json
POST /commands
{
  "chave": "enviar-nfe",
  "descricao": "Envia NF-e para o serviço SEFAZ",
  "tipo": "POST",
  "rota": "https://nfe.sefaz.rs.gov.br/ws/NfeRecepcao",
  "tipoConteudo": "application/json",
  "nomeCertificado": "cert-sefaz",
  "corpo": {
    "campos": [
      { "nome": "xml",   "tipo": "string",  "obrigatorio": true  },
      { "nome": "chave", "tipo": "string",  "obrigatorio": true  }
    ]
  }
}
```

### Resposta (Command criado)
```json
{
  "id": 5,
  "chave": "buscar-produto-por-cd",
  "descricao": "Busca produto pelo código",
  "tipo": "QUERY",
  "tipoBanco": "postgres",
  "sql": "SELECT cd_produto, nm_produto FROM \"DBAMV\".produto WHERE cd_produto = :cd_produto",
  "temFiltro": true,
  "parametros": [
    { "nome": "cd_produto", "tipo": "number", "obrigatorio": true, "operador": "", "jaAdicionado": true }
  ],
  "ordenacao": { "nomeColuna": "", "decrescente": false }
}
```

---

## 3. Routes

### Rota GET simples — buscar produto por código
```json
POST /routes
{
  "chave": "GET-produto",
  "caminho": "/api/produtos/{cd_produto}",
  "metodo": "GET",
  "resposta": { "status": 200 },
  "servico": {
    "threadUnica": false,
    "resultadoUnico": true,
    "passos": [
      {
        "alias": "produto",
        "abortarSemDados": true,
        "comando": {
          "tipo": "QUERY",
          "database": "dbamv",
          "nome": "buscar-produto-por-cd",
          "retornarResultado": true,
          "parametros": [
            { "nome": "cd_produto", "tipo": "path", "extrair": false }
          ]
        }
      }
    ]
  }
}
```

### Rota GET com paginação — listar entradas
```json
POST /routes
{
  "chave": "GET-entradas",
  "caminho": "/api/entradas",
  "metodo": "GET",
  "resposta": { "status": 200 },
  "servico": {
    "threadUnica": false,
    "resultadoUnico": false,
    "passos": [
      {
        "alias": "entradas",
        "abortarSemDados": false,
        "comando": {
          "tipo": "QUERY",
          "database": "dbamv",
          "nome": "listar-entradas",
          "retornarResultado": true,
          "parametros": [
            { "nome": "cd_estoque", "tipo": "query", "extrair": false },
            { "nome": "page",       "tipo": "query", "extrair": false },
            { "nome": "pageSize",   "tipo": "query", "extrair": false }
          ]
        }
      }
    ]
  }
}
```

### Rota POST — inserir paciente
```json
POST /routes
{
  "chave": "POST-paciente",
  "caminho": "/api/pacientes",
  "metodo": "POST",
  "resposta": { "status": 201 },
  "servico": {
    "threadUnica": false,
    "resultadoUnico": false,
    "passos": [
      {
        "alias": "insert",
        "abortarSemDados": false,
        "comando": {
          "tipo": "INSERT",
          "database": "dbamv",
          "nome": "inserir-paciente",
          "retornarResultado": true,
          "parametros": []
        }
      }
    ]
  }
}
```

### Rota com múltiplos passos — buscar entrada e seus itens
```json
POST /routes
{
  "chave": "GET-entrada-completa",
  "caminho": "/api/entradas/{cd_ent_pro}/completa",
  "metodo": "GET",
  "resposta": { "status": 200 },
  "servico": {
    "threadUnica": true,
    "resultadoUnico": false,
    "passos": [
      {
        "alias": "cabecalho",
        "abortarSemDados": true,
        "comando": {
          "tipo": "QUERY",
          "database": "dbamv",
          "nome": "buscar-entrada-por-cd",
          "retornarResultado": true,
          "parametros": [
            { "nome": "cd_ent_pro", "tipo": "path", "extrair": false }
          ]
        }
      },
      {
        "alias": "itens",
        "abortarSemDados": false,
        "comando": {
          "tipo": "QUERY",
          "database": "dbamv",
          "nome": "listar-itens-entrada",
          "retornarResultado": true,
          "parametros": [
            { "nome": "cd_ent_pro", "tipo": "path", "extrair": false }
          ]
        }
      }
    ]
  }
}
```

### Rota DELETE — remover item
```json
POST /routes
{
  "chave": "DELETE-item-entrada",
  "caminho": "/api/entradas/itens/{cd_itent_pro}",
  "metodo": "DELETE",
  "resposta": { "status": 204 },
  "servico": {
    "threadUnica": false,
    "resultadoUnico": false,
    "passos": [
      {
        "alias": "del",
        "abortarSemDados": false,
        "comando": {
          "tipo": "DELETE",
          "database": "dbamv",
          "nome": "excluir-item-entrada",
          "retornarResultado": false,
          "parametros": [
            { "nome": "cd_itent_pro", "tipo": "path", "extrair": false }
          ]
        }
      }
    ]
  }
}
```

### Resposta (Route criada)
```json
{
  "id": 3,
  "chave": "GET-produto",
  "caminho": "/api/produtos/{cd_produto}",
  "metodo": "GET",
  "resposta": { "status": 200, "tipoConteudo": "" },
  "servico": {
    "threadUnica": false,
    "resultadoUnico": true,
    "passos": [...]
  }
}
```

---

## 4. Certificates

### Criar certificado PFX
```json
POST /certificates
{
  "nome": "cert-sefaz-rs",
  "descricao": "Certificado A1 para SEFAZ RS",
  "arquivoPfx": "MIIJqAIBAzCCCWIGCSqGSIb3DQEHAaCC...",
  "senha": "senha-do-pfx",
  "dataExpiracao": "2027-06-30"
}
```

### Criar certificado PEM
```json
POST /certificates
{
  "nome": "cert-banco-legado",
  "descricao": "Certificado mTLS para integração com banco legado",
  "arquivoCert": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0t...",
  "arquivoChave": "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBL...",
  "arquivoCACert": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0t..."
}
```

### Atualizar certificado (desativar)
```json
PUT /certificates/1
{
  "ativo": false
}
```

### Atualizar data de expiração
```json
PUT /certificates/1
{
  "dataExpiracao": "2028-12-31"
}
```

### Resposta (Certificate)
```json
{
  "id": 1,
  "nome": "cert-sefaz-rs",
  "descricao": "Certificado A1 para SEFAZ RS",
  "arquivoPfx": "MIIJqAIBAz...",
  "ativo": true,
  "dataExpiracao": "2027-06-30",
  "criadoEm": "2026-05-01T21:00:00-03:00",
  "atualizadoEm": "2026-05-01T21:00:00-03:00"
}
```

---

## 5. Respostas dos executores

### QUERY — lista
```json
[
  { "cd_produto": 1, "nm_produto": "Dipirona 500mg", "sn_lote": "N" },
  { "cd_produto": 2, "nm_produto": "Amoxicilina 500mg", "sn_lote": "N" }
]
```

### QUERY — com paginação
```json
{
  "dados": [
    { "cd_ent_pro": 100, "dt_ent_pro": "2026-01-15", "vl_total": 1500.00 },
    { "cd_ent_pro": 101, "dt_ent_pro": "2026-01-16", "vl_total": 980.50 }
  ],
  "pagina": 1,
  "tamanhoPagina": 20
}
```

### INSERT
```json
{ "ultimoIdInserido": 51, "linhasAfetadas": 1 }
```

### UPDATE / DELETE
```json
{ "linhasAfetadas": 1 }
```

### PROCEDURE (sem out params)
```json
{ "status": "concluido" }
```

### PROCEDURE (com out params Oracle)
```json
{ "p_resultado": "OK", "p_mensagem": "Saldo atualizado com sucesso" }
```

### Erro de validação
```json
{
  "timestamp": "2026-05-01T21:15:00-03:00",
  "status": 422,
  "error": "Unprocessable Entity",
  "message": "tipoBanco 'sqlserver2' inválido. Valores aceitos: [mysql postgres oracle sqlserver]",
  "path": "/commands",
  "trace": null
}
```

### Erro de conflito
```json
{
  "timestamp": "2026-05-01T21:15:00-03:00",
  "status": 409,
  "error": "Conflict",
  "message": "comando com chave 'buscar-produto-por-cd' já existe",
  "path": "/commands",
  "trace": null
}
```

### Erro de autenticação
```json
{
  "status": 401,
  "message": "Authorization inválido ou ausente"
}
```
