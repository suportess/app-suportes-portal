package execute

import (
	"context"
	"fmt"
	"strings"
	"time"

	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/enum"
)

type DeleteExecute struct{}

func NewDeleteExecute() *DeleteExecute { return &DeleteExecute{} }

func (e *DeleteExecute) Type() string { return string(enum.CommandTypeDelete) }

func (e *DeleteExecute) Run(ctx *ExecContext) (interface{}, apierr.Detail) {
	cmd := ctx.Command
	db := ctx.DB

	if err := validateRequiredParams(cmd.Parameters, ctx.Params); err != nil {
		return nil, err
	}

	var query string
	var vals []interface{}

	if cmd.SQL != "" {
		query = cmd.SQL
		for _, p := range cmd.Parameters {
			if v := ctx.Params[p.Name]; v != nil {
				vals = append(vals, v)
			}
		}
	} else {
		if cmd.Table == "" {
			return nil, apierr.New("campo 'tabela' ou 'sql' é obrigatório para DELETE", nil)
		}
		whereParts := make([]string, 0)
		for _, p := range cmd.Parameters {
			v := ctx.Params[p.Name]
			if p.Required && v == nil {
				return nil, apierr.New(fmt.Sprintf("parâmetro '%s' é obrigatório", p.Name), nil)
			}
			if v != nil {
				whereParts = append(whereParts, fmt.Sprintf("%s = ?", p.Name))
				vals = append(vals, v)
			}
		}
		query = fmt.Sprintf("DELETE FROM %s", cmd.Table)
		if len(whereParts) > 0 {
			query += " WHERE " + strings.Join(whereParts, " AND ")
		}
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := db.ExecContext(dbCtx, query, vals...)
	if err != nil {
		return nil, apierr.New("erro ao executar DELETE: "+err.Error(), nil)
	}

	rowsAffected, _ := result.RowsAffected()
	return map[string]interface{}{"linhasAfetadas": rowsAffected}, nil
}
