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

// AnonymousExecute executa um bloco anônimo PL/SQL (DECLARE ... BEGIN ... END).
// Parâmetros IN são vinculados via named binds (:nome).
// Parâmetros com tipo "out" são tratados como OUT binds e retornados no corpo da resposta.
// Se não houver parâmetros OUT, retorna {"status": "concluido"}.
type AnonymousExecute struct{}

func NewAnonymousExecute() *AnonymousExecute { return &AnonymousExecute{} }

func (e *AnonymousExecute) Type() string { return string(enum.CommandTypeAnonymous) }

func (e *AnonymousExecute) Run(ctx *ExecContext) (interface{}, apierr.Detail) {
	cmd := ctx.Command
	db := ctx.DB

	if cmd.SQL == "" {
		return nil, apierr.New("campo 'sql' com o bloco anônimo PL/SQL é obrigatório", nil)
	}

	// Mescla path/query params + body (body tem precedência sobre path)
	resolved := make(map[string]interface{})
	for k, v := range ctx.Params {
		resolved[k] = v
	}
	for k, v := range ctx.Body {
		resolved[k] = v
	}

	if err := validateRequiredParams(cmd.Parameters, resolved); err != nil {
		return nil, err
	}

	outParams := make(map[string]*string)
	args := make([]interface{}, 0, len(cmd.Parameters))

	for _, p := range cmd.Parameters {
		if p.Type == enum.ParamTypeOut {
			buf := strings.Repeat(" ", 4000)
			outParams[p.Name] = &buf
			args = append(args, sql.Named(p.Name, sql.Out{Dest: &buf, In: false}))
		} else {
			args = append(args, sql.Named(p.Name, resolved[p.Name]))
		}
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, err := db.ExecContext(dbCtx, cmd.SQL, args...)
	if err != nil {
		return nil, apierr.New(fmt.Sprintf("erro ao executar bloco anônimo: %s", err.Error()), nil)
	}

	if len(outParams) == 0 {
		return map[string]interface{}{"status": "concluido"}, nil
	}

	result := make(map[string]interface{})
	for name, ptr := range outParams {
		result[name] = strings.TrimSpace(*ptr)
	}
	return result, nil
}

