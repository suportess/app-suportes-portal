package domain

import (
	"fmt"
	"regexp"
	"strings"

	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/enum"
)

// Command é a entidade persistida no BoltDB representando um comando SQL ou HTTP.
type Command struct {
	ID              int              `storm:"id,increment" json:"id"`
	Key             string           `storm:"unique,index" json:"chave"`
	Description     string           `json:"descricao,omitempty"`
	Type            enum.CommandType `json:"tipo"`
	DatabaseType    enum.DBDriver    `json:"tipoBanco,omitempty"`
	SQL             string           `json:"sql,omitempty"`
	Table           string           `json:"tabela,omitempty"`
	Route           string           `json:"rota,omitempty"`
	ContentType     string           `json:"tipoConteudo,omitempty"`
	CertificateName string           `json:"nomeCertificado,omitempty"`
	HasFilter       bool             `json:"temFiltro"`
	Body            CommandBody      `json:"corpo,omitempty"`
	Query           CommandBody      `json:"consulta,omitempty"`
	Parameters      []SQLParameter   `json:"parametros,omitempty"`
	Order           SQLOrder         `json:"ordenacao,omitempty"`
	Pagination      *Pagination      `json:"paginacao,omitempty"`
}

// Pagination define as configurações de paginação de um comando QUERY.
// Ao definir, o execute vai injetar LIMIT/OFFSET automaticamente
// usando os parâmetros informados na chamada da rota.
type Pagination struct {
	// PageParam é o nome do parâmetro de query/path que representa a página (default: "page")
	PageParam string `json:"paramPagina,omitempty"`
	// PageSizeParam é o nome do parâmetro que representa o tamanho da página (default: "pageSize")
	PageSizeParam string `json:"paramTamanhoPagina,omitempty"`
	// DefaultPageSize é o tamanho de página quando o parâmetro não for informado (default: 20)
	DefaultPageSize int `json:"tamanhoPaginaPadrao,omitempty"`
	// MaxPageSize limita o tamanho máximo de página (default: 100; 0 = sem limite)
	MaxPageSize int `json:"tamanhoPaginaMaximo,omitempty"`
}

// Defaults preenche valores omissos de Pagination.
func (p *Pagination) Defaults() {
	if p.PageParam == "" {
		p.PageParam = "page"
	}
	if p.PageSizeParam == "" {
		p.PageSizeParam = "pageSize"
	}
	if p.DefaultPageSize <= 0 {
		p.DefaultPageSize = 20
	}
	if p.MaxPageSize <= 0 {
		p.MaxPageSize = 100
	}
}

// CommandBody descreve os campos de um body HTTP ou query.
type CommandBody struct {
	Fields []BodyField `json:"campos,omitempty"`
}

// BodyField descreve um campo individual de um body ou query.
type BodyField struct {
	Name      string         `json:"nome"`
	Type      enum.FieldType `json:"tipo"`
	Required  bool           `json:"obrigatorio"`
	Maximum   int            `json:"maximo,omitempty"`
	Minimum   int            `json:"minimo,omitempty"`
	ToType    string         `json:"converterPara,omitempty"`
	KeyName   string         `json:"nomeChave,omitempty"`
	ParamType string         `json:"tipoParametro,omitempty"`
}

// SQLParameter descreve um parâmetro de filtro SQL.
type SQLParameter struct {
	Name         string           `json:"nome"`
	Type         enum.ParamType   `json:"tipo,omitempty"`
	Operator     enum.SQLOperator `json:"operador,omitempty"`
	Required     bool             `json:"obrigatorio"`
	AlreadyAdded bool             `json:"jaAdicionado,omitempty"`
	Len          int              `json:"-"`
}

// SQLOrder descreve a ordenação de resultados.
type SQLOrder struct {
	ColumnName string `json:"nomeColuna,omitempty"`
	Desc       bool   `json:"decrescente"`
}

// --------- helpers ---------

func (c *Command) SupportsNamedParams() bool {
	return c.DatabaseType == enum.DBDriverSQLServer || c.DatabaseType == enum.DBDriverOracle || c.DatabaseType == enum.DBDriverPostgres
}

func (c *Command) ParamSymbol() string {
	if c.DatabaseType == enum.DBDriverSQLServer {
		return "@"
	}
	return ":"
}

func (c *Command) HasOrder() bool {
	return c.Order.ColumnName != ""
}

func (c *Command) OrderClause() string {
	if !c.HasOrder() {
		return ""
	}
	dir := "ASC"
	if c.Order.Desc {
		dir = "DESC"
	}
	return fmt.Sprintf(" ORDER BY %s %s ", c.Order.ColumnName, dir)
}

func (c *Command) ProcessParameters() {
	if strings.Contains(strings.ToUpper(c.SQL), "WHERE") {
		c.HasFilter = true
	} else {
		c.HasFilter = false
	}
	c.syncParametersFromSQL()
	c.markAlreadyAddedParams()
}

func (c *Command) syncParametersFromSQL() {
	// Strip single-quoted string literals before scanning to avoid false bind-variable
	// matches inside SQL string content, e.g. '"ERRO":true' containing ':true'.
	reLiteral := regexp.MustCompile(`'[^']*'`)
	stripped := reLiteral.ReplaceAllString(c.SQL, "''")

	re := regexp.MustCompile(regexp.QuoteMeta(c.ParamSymbol()) + `(\w+)`)
	matches := re.FindAllStringSubmatch(stripped, -1)
	for _, m := range matches {
		name := m[1]
		if !c.hasParam(name) {
			c.Parameters = append(c.Parameters, SQLParameter{Name: name, Required: true})
		}
	}
}

func (c *Command) markAlreadyAddedParams() {
	for i, p := range c.Parameters {
		if strings.Contains(c.SQL, c.ParamSymbol()+p.Name) {
			c.Parameters[i].AlreadyAdded = true
		}
	}
}

func (c *Command) hasParam(name string) bool {
	for _, p := range c.Parameters {
		if p.Name == name {
			return true
		}
	}
	return false
}

func (p *SQLParameter) FilterClause(symbol string) string {
	op := p.GetOperator()
	return fmt.Sprintf("%s %s %s%s", p.Name, op, symbol, p.Name)
}

func (p *SQLParameter) GetOperator() string {
	if p.Operator == "" {
		return "="
	}
	return string(p.Operator)
}

func (p *SQLParameter) IsIN() bool {
	return p.Operator == enum.SQLOperatorIn
}

// --------- Request/Response ---------

// CreateCommandRequest é o payload de criação de um comando.
type CreateCommandRequest struct {
	Key             string           `json:"chave"`
	Description     string           `json:"descricao,omitempty"`
	Type            enum.CommandType `json:"tipo"`
	DatabaseType    enum.DBDriver    `json:"tipoBanco,omitempty"`
	SQL             string           `json:"sql,omitempty"`
	Table           string           `json:"tabela,omitempty"`
	Route           string           `json:"rota,omitempty"`
	ContentType     string           `json:"tipoConteudo,omitempty"`
	CertificateName string           `json:"nomeCertificado,omitempty"`
	Body            CommandBody      `json:"corpo,omitempty"`
	Query           CommandBody      `json:"consulta,omitempty"`
	Parameters      []SQLParameter   `json:"parametros,omitempty"`
	Order           SQLOrder         `json:"ordenacao,omitempty"`
	Pagination      *Pagination      `json:"paginacao,omitempty"`
}

func (r *CreateCommandRequest) Validate() apierr.Detail {
	if r.Key == "" {
		return apierr.New("campo 'chave' é obrigatório", nil)
	}
	if !r.Type.IsValid() {
		return apierr.UnprocessableEntity(
			fmt.Sprintf("tipo '%s' inválido. Valores aceitos: %v", r.Type, enum.ValidCommandTypes()),
			nil,
		)
	}
	if r.Type.IsSQL() && r.Type != enum.CommandTypeProcedure {
		if r.SQL == "" && r.Table == "" {
			return apierr.New("para comandos SQL, 'sql' ou 'tabela' é obrigatório", nil)
		}
		if r.DatabaseType != "" && !r.DatabaseType.IsValid() {
			return apierr.UnprocessableEntity(
				fmt.Sprintf("tipoBanco '%s' inválido. Valores aceitos: %v", r.DatabaseType, enum.ValidDBDrivers()),
				nil,
			)
		}
	}
	if r.Type.IsHTTP() {
		if r.Route == "" {
			return apierr.New("para comandos HTTP, 'rota' é obrigatório", nil)
		}
	}
	for _, f := range r.Body.Fields {
		if !f.Type.IsValid() {
			return apierr.UnprocessableEntity(
				fmt.Sprintf("tipo do campo '%s' inválido para o campo '%s'", f.Type, f.Name),
				nil,
			)
		}
	}
	for _, p := range r.Parameters {
		if p.Type != "" && !p.Type.IsValid() {
			return apierr.UnprocessableEntity(
				fmt.Sprintf("tipo do parâmetro '%s' inválido para o parâmetro '%s'", p.Type, p.Name),
				nil,
			)
		}
		if p.Operator != "" && !p.Operator.IsValid() {
			return apierr.UnprocessableEntity(
				fmt.Sprintf("operador '%s' inválido para o parâmetro '%s'", p.Operator, p.Name),
				nil,
			)
		}
	}
	return nil
}

func (r *CreateCommandRequest) ToDomain() *Command {
	pag := r.Pagination
	if pag != nil {
		pag.Defaults()
	}
	return &Command{
		Key:             r.Key,
		Description:     r.Description,
		Type:            r.Type,
		DatabaseType:    r.DatabaseType,
		SQL:             r.SQL,
		Table:           r.Table,
		Route:           r.Route,
		ContentType:     r.ContentType,
		CertificateName: r.CertificateName,
		Body:            r.Body,
		Query:           r.Query,
		Parameters:      r.Parameters,
		Order:           r.Order,
		Pagination:      pag,
	}
}

// UpdateCommandRequest é o payload de atualização de um comando (chave imutável).
type UpdateCommandRequest struct {
	Description     string           `json:"descricao,omitempty"`
	Type            enum.CommandType `json:"tipo,omitempty"`
	DatabaseType    enum.DBDriver    `json:"tipoBanco,omitempty"`
	SQL             string           `json:"sql,omitempty"`
	Table           string           `json:"tabela,omitempty"`
	Route           string           `json:"rota,omitempty"`
	ContentType     string           `json:"tipoConteudo,omitempty"`
	CertificateName string           `json:"nomeCertificado,omitempty"`
	Body            CommandBody      `json:"corpo,omitempty"`
	Query           CommandBody      `json:"consulta,omitempty"`
	Parameters      []SQLParameter   `json:"parametros,omitempty"`
	Order           SQLOrder         `json:"ordenacao,omitempty"`
	Pagination      *Pagination      `json:"paginacao,omitempty"`
}

func (r *UpdateCommandRequest) Validate() apierr.Detail {
	if r.Type != "" && !r.Type.IsValid() {
		return apierr.UnprocessableEntity(
			fmt.Sprintf("tipo '%s' inválido. Valores aceitos: %v", r.Type, enum.ValidCommandTypes()),
			nil,
		)
	}
	if r.DatabaseType != "" && !r.DatabaseType.IsValid() {
		return apierr.UnprocessableEntity(
			fmt.Sprintf("tipoBanco '%s' inválido. Valores aceitos: %v", r.DatabaseType, enum.ValidDBDrivers()),
			nil,
		)
	}
	return nil
}

func (r *UpdateCommandRequest) ApplyTo(cmd *Command) {
	if r.Description != "" {
		cmd.Description = r.Description
	}
	if r.Type != "" {
		cmd.Type = r.Type
	}
	if r.DatabaseType != "" {
		cmd.DatabaseType = r.DatabaseType
	}
	if r.SQL != "" {
		cmd.SQL = r.SQL
	}
	if r.Table != "" {
		cmd.Table = r.Table
	}
	if r.Route != "" {
		cmd.Route = r.Route
	}
	if r.ContentType != "" {
		cmd.ContentType = r.ContentType
	}
	if r.CertificateName != "" {
		cmd.CertificateName = r.CertificateName
	}
	if len(r.Body.Fields) > 0 {
		cmd.Body = r.Body
	}
	if len(r.Query.Fields) > 0 {
		cmd.Query = r.Query
	}
	if len(r.Parameters) > 0 {
		cmd.Parameters = r.Parameters
	}
	if r.Order.ColumnName != "" {
		cmd.Order = r.Order
	}
	if r.Pagination != nil {
		r.Pagination.Defaults()
		cmd.Pagination = r.Pagination
	}
}
