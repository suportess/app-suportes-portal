package execute

import (
	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/domain"
	"github.com/jmoiron/sqlx"
)

// ExecContext carries everything an executor needs to run a step.
type ExecContext struct {
	Command      *domain.Command
	DB           *sqlx.DB
	Params       map[string]interface{}
	Body         map[string]interface{}
	Headers      map[string]string
	SingleResult bool
}

// Executor is implemented by every execute strategy (QUERY, INSERT, …, POST, GET, …).
type Executor interface {
	Type() string
	Run(ctx *ExecContext) (interface{}, apierr.Detail)
}
