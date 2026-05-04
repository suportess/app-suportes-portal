package execute

import (
	"context"
	"fmt"
	"strings"
	"time"

	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/enum"
)

type InsertExecute struct{}

func NewInsertExecute() *InsertExecute { return &InsertExecute{} }

func (e *InsertExecute) Type() string { return string(enum.CommandTypeInsert) }

func (e *InsertExecute) Run(ctx *ExecContext) (interface{}, apierr.Detail) {
	cmd := ctx.Command
	db := ctx.DB
	body := ctx.Body

	if body == nil {
		return nil, apierr.New("body é obrigatório para INSERT", nil)
	}

	cols := make([]string, 0, len(body))
	vals := make([]interface{}, 0, len(body))
	placeholders := make([]string, 0, len(body))

	for k, v := range body {
		cols = append(cols, k)
		vals = append(vals, v)
		placeholders = append(placeholders, "?")
	}

	table := cmd.Table
	if table == "" {
		return nil, apierr.New("campo 'tabela' é obrigatório para INSERT", nil)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	dbCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := db.ExecContext(dbCtx, query, vals...)
	if err != nil {
		return nil, apierr.New("erro ao executar INSERT: "+err.Error(), nil)
	}

	lastID, _ := result.LastInsertId()
	rowsAffected, _ := result.RowsAffected()

	return map[string]interface{}{
		"ultimoIdInserido": lastID,
		"linhasAfetadas":   rowsAffected,
	}, nil
}
