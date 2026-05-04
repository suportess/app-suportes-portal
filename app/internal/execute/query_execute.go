package execute

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/domain"
	"br.tec.suportes/portal/internal/enum"
	"github.com/jmoiron/sqlx"
)

type QueryExecute struct{}

func NewQueryExecute() *QueryExecute { return &QueryExecute{} }

func (e *QueryExecute) Type() string { return string(enum.CommandTypeQuery) }

// PagedResult é a resposta quando paginação está ativa.
type PagedResult struct {
	Data     interface{} `json:"dados"`
	Page     int         `json:"pagina"`
	PageSize int         `json:"tamanhoPagina"`
}

func (e *QueryExecute) Run(ctx *ExecContext) (interface{}, apierr.Detail) {
	cmd := ctx.Command
	db := ctx.DB

	log.Printf("[query-execute] cmd=%s params=%v", cmd.Key, ctx.Params)
	if err := validateRequiredParams(cmd.Parameters, ctx.Params); err != nil {
		return nil, err
	}

	var pag *pageInfo
	if cmd.Pagination != nil {
		p, aerr := resolvePagination(cmd.Pagination, ctx.Params)
		if aerr != nil {
			return nil, aerr
		}
		pag = p
	}

	var result interface{}
	var aerr apierr.Detail
	if cmd.SupportsNamedParams() {
		result, aerr = executeNamedQuery(db, cmd, ctx.Params, ctx.SingleResult, pag)
	} else {
		result, aerr = executePositionalQuery(db, cmd, ctx.Params, ctx.SingleResult, pag)
	}
	if aerr != nil {
		return nil, aerr
	}

	if pag != nil {
		return &PagedResult{
			Data:     result,
			Page:     pag.page,
			PageSize: pag.pageSize,
		}, nil
	}
	return result, nil
}

type pageInfo struct {
	page     int
	pageSize int
}

func resolvePagination(cfg *domain.Pagination, params map[string]interface{}) (*pageInfo, apierr.Detail) {
	page := toInt(params[cfg.PageParam], 1)
	pageSize := toInt(params[cfg.PageSizeParam], cfg.DefaultPageSize)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = cfg.DefaultPageSize
	}
	if cfg.MaxPageSize > 0 && pageSize > cfg.MaxPageSize {
		return nil, apierr.UnprocessableEntity(
			fmt.Sprintf("tamanhoPagina máximo permitido é %d", cfg.MaxPageSize), nil,
		)
	}
	return &pageInfo{page: page, pageSize: pageSize}, nil
}

func toInt(v interface{}, fallback int) int {
	if v == nil {
		return fallback
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case string:
		n, err := strconv.Atoi(val)
		if err != nil {
			return fallback
		}
		return n
	}
	return fallback
}

func executePositionalQuery(db *sqlx.DB, cmd *domain.Command, params map[string]interface{}, singleResult bool, pag *pageInfo) (interface{}, apierr.Detail) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	args, filters := buildArgs(cmd.Parameters, params)
	query := buildQuery(cmd, filters)
	query = appendPagination(query, cmd, pag)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apierr.New(err.Error(), nil)
	}
	defer rows.Close()
	return extractRows(rows, singleResult), nil
}

func executeNamedQuery(db *sqlx.DB, cmd *domain.Command, params map[string]interface{}, singleResult bool, pag *pageInfo) (interface{}, apierr.Detail) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, filters := buildArgs(cmd.Parameters, params)
	query := buildQuery(cmd, filters)
	query = appendPagination(query, cmd, pag)

	rows, err := db.NamedQueryContext(ctx, query, params)
	if err != nil {
		return nil, apierr.New(err.Error(), nil)
	}
	defer rows.Close()
	return extractRows(rows.Rows, singleResult), nil
}

func extractRows(rows *sql.Rows, singleResult bool) interface{} {
	columns, _ := rows.Columns()
	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		rows.Scan(pointers...)
		row := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			if b, ok := values[i].([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = values[i]
			}
		}
		results = append(results, row)
	}

	if len(results) == 0 {
		return nil
	}
	if singleResult {
		return results[0]
	}
	return results
}

func buildArgs(params []domain.SQLParameter, values map[string]interface{}) ([]interface{}, []domain.SQLParameter) {
	var args []interface{}
	var filters []domain.SQLParameter

	for _, p := range params {
		val := values[p.Name]
		if s, ok := val.(string); ok && s == "" {
			val = nil
		}
		if val == nil {
			args = append(args, sql.Named(p.Name, nil))
			continue
		}
		if p.Type == enum.ParamTypeOut {
			buf := strings.Repeat(" ", 4000)
			args = append(args, sql.Named(p.Name, sql.Out{Dest: &buf, In: true}))
			continue
		}
		if p.Type == enum.ParamTypeBase64 {
			if file, ok := val.(UploadedFile); ok {
				data := base64.StdEncoding.EncodeToString(file.Content)
				uri := fmt.Sprintf("data:%s;base64,%s", file.ContentType, data)
				args = append(args, sql.Named(p.Name, uri))
			}
			continue
		}
		if reflect.TypeOf(val).Kind() == reflect.Slice {
			items := val.([]interface{})
			for i, v := range items {
				args = append(args, sql.Named(p.Name+"_"+strconv.Itoa(i), v))
			}
			pp := p
			pp.Operator = enum.SQLOperatorIn
			pp.Len = len(items)
			filters = append(filters, pp)
			continue
		}
		args = append(args, sql.Named(p.Name, val))
		filters = append(filters, p)
	}
	return args, filters
}

func buildQuery(cmd *domain.Command, filters []domain.SQLParameter) string {
	q := cmd.SQL
	for i, f := range filters {
		if !f.AlreadyAdded {
			if !cmd.HasFilter && i == 0 {
				q += " WHERE "
			} else {
				q += " AND "
			}
			q += f.FilterClause(cmd.ParamSymbol())
		}
		if f.IsIN() && f.Len > 0 {
			placeholder := cmd.ParamSymbol() + f.Name
			q = strings.ReplaceAll(q, placeholder, expandIN(placeholder, f))
		}
	}
	q += cmd.OrderClause()
	return q
}

func expandIN(base string, p domain.SQLParameter) string {
	var items []string
	for i := 0; i < p.Len; i++ {
		items = append(items, base+"_"+strconv.Itoa(i))
	}
	return strings.Join(items, ", ")
}

// appendPagination injeta a cláusula de paginação correta para o driver.
func appendPagination(query string, cmd *domain.Command, pag *pageInfo) string {
	if pag == nil {
		return query
	}
	offset := (pag.page - 1) * pag.pageSize
	switch cmd.DatabaseType {
	case enum.DBDriverOracle:
		// Oracle 12c+: OFFSET x ROWS FETCH NEXT y ROWS ONLY
		// Precisa de ORDER BY antes — se não tiver, adiciona por ROWNUM para ser determinístico
		q := query
		if !cmd.HasOrder() {
			q += " ORDER BY 1"
		}
		return fmt.Sprintf("%s OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", q, offset, pag.pageSize)
	case enum.DBDriverSQLServer:
		// SQL Server: ORDER BY requerido antes de OFFSET/FETCH
		q := query
		if !cmd.HasOrder() {
			q += " ORDER BY (SELECT NULL)"
		}
		return fmt.Sprintf("%s OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", q, offset, pag.pageSize)
	default:
		// MySQL e PostgreSQL: LIMIT / OFFSET
		return fmt.Sprintf("%s LIMIT %d OFFSET %d", query, pag.pageSize, offset)
	}
}

func validateRequiredParams(params []domain.SQLParameter, values map[string]interface{}) apierr.Detail {
	for _, p := range params {
		val := values[p.Name]
		log.Printf("[validate] param=%s required=%v val=%v (%T)", p.Name, p.Required, val, val)
		if p.Required && val == nil {
			return apierr.New(fmt.Sprintf("parâmetro obrigatório '%s' não informado", p.Name), nil)
		}
	}
	return nil
}

// UploadedFile represents an uploaded multipart file in memory.
type UploadedFile struct {
	FileName    string
	Content     []byte
	ContentType string
	IsMultipart bool
}
