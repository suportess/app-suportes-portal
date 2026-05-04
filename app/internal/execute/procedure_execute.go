package execute

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/enum"
)

type ProcedureExecute struct{}

func NewProcedureExecute() *ProcedureExecute { return &ProcedureExecute{} }

func (e *ProcedureExecute) Type() string { return string(enum.CommandTypeProcedure) }

func (e *ProcedureExecute) Run(ctx *ExecContext) (interface{}, apierr.Detail) {
	cmd := ctx.Command
	db := ctx.DB

	if err := validateRequiredParams(cmd.Parameters, ctx.Params); err != nil {
		return nil, err
	}

	outParams := make(map[string]*string)
	args := make([]interface{}, 0, len(cmd.Parameters))

	for _, p := range cmd.Parameters {
		val := ctx.Params[p.Name]
		if p.Type == enum.ParamTypeOut {
			buf := strings.Repeat(" ", 4000)
			outParams[p.Name] = &buf
			args = append(args, sql.Named(p.Name, sql.Out{Dest: &buf, In: true}))
		} else {
			args = append(args, sql.Named(p.Name, val))
		}
	}

	query := cmd.SQL
	if query == "" {
		return nil, apierr.New("campo 'sql' com a chamada da procedure é obrigatório", nil)
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, err := db.ExecContext(dbCtx, query, args...)
	if err != nil {
		return nil, apierr.New(fmt.Sprintf("erro ao executar procedure: %s", err.Error()), nil)
	}

	result := make(map[string]interface{})
	for name, ptr := range outParams {
		result[name] = strings.TrimSpace(*ptr)
	}

	if len(result) == 0 {
		return map[string]interface{}{"status": "concluido"}, nil
	}
	return result, nil
}
