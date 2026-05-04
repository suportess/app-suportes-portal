package execute

import (
	"context"
	"fmt"
	"strings"
	"time"

	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/enum"
)

type UpdateExecute struct{}

func NewUpdateExecute() *UpdateExecute { return &UpdateExecute{} }

func (e *UpdateExecute) Type() string { return string(enum.CommandTypeUpdate) }

func (e *UpdateExecute) Run(ctx *ExecContext) (interface{}, apierr.Detail) {
	cmd := ctx.Command
	db := ctx.DB

	if err := validateRequiredParams(cmd.Parameters, ctx.Params); err != nil {
		return nil, err
	}

	if cmd.Table == "" {
		return nil, apierr.New("campo 'tabela' ou 'sql' é obrigatório para UPDATE", nil)
	}

	body := ctx.Body
	if len(body) == 0 {
		return nil, apierr.New("body com campos a atualizar é obrigatório", nil)
	}

	sets := make([]string, 0, len(body))
	vals := make([]interface{}, 0, len(body))
	for k, v := range body {
		sets = append(sets, fmt.Sprintf("%s = ?", k))
		vals = append(vals, v)
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

	query := fmt.Sprintf("UPDATE %s SET %s", cmd.Table, strings.Join(sets, ", "))
	if len(whereParts) > 0 {
		query += " WHERE " + strings.Join(whereParts, " AND ")
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := db.ExecContext(dbCtx, query, vals...)
	if err != nil {
		return nil, apierr.New("erro ao executar UPDATE: "+err.Error(), nil)
	}

	rowsAffected, _ := result.RowsAffected()
	return map[string]interface{}{"linhasAfetadas": rowsAffected}, nil
}
