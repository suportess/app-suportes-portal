package domain

import (
	"fmt"
	"strings"

	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/enum"
)

// Route é a entidade persistida que mapeia um path HTTP a uma sequência de passos.
type Route struct {
	ID       int             `storm:"id,increment" json:"id"`
	Key      string          `storm:"unique,index" json:"chave"`
	Path     string          `json:"caminho"`
	Method   enum.HTTPMethod `json:"metodo"`
	Response RouteResponse   `json:"resposta"`
	Service  RouteService    `json:"servico"`
}

type RouteResponse struct {
	Status      int    `json:"status"`
	ContentType string `json:"tipoConteudo,omitempty"`
}

type RouteService struct {
	SingleThread bool        `json:"threadUnica"`
	SingleResult bool        `json:"resultadoUnico"`
	Steps        []RouteStep `json:"passos"`
}

type RouteStep struct {
	Alias       string       `json:"alias,omitempty"`
	AbortNoData bool         `json:"abortarSemDados"`
	Command     RouteCommand `json:"comando"`
}

type RouteCommand struct {
	Type         enum.CommandType `json:"tipo"`
	Database     string           `json:"database,omitempty"`
	Name         string           `json:"nome"`
	ReturnResult bool             `json:"retornarResultado"`
	Parameters   []RouteParameter `json:"parametros,omitempty"`
}

type RouteParameter struct {
	Name    string      `json:"nome"`
	Type    string      `json:"tipo,omitempty"`
	Extract bool        `json:"extrair"`
	Value   interface{} `json:"valor,omitempty"`
	Field   string      `json:"campo,omitempty"`
}

// --------- Request/Response ---------

// CreateRouteRequest é o payload de criação de uma rota dinâmica.
type CreateRouteRequest struct {
	Key      string          `json:"chave"`
	Path     string          `json:"caminho"`
	Method   enum.HTTPMethod `json:"metodo"`
	Response RouteResponse   `json:"resposta"`
	Service  RouteService    `json:"servico"`
}

func (r *CreateRouteRequest) Validate() apierr.Detail {
	if r.Key == "" {
		return apierr.New("campo 'chave' é obrigatório", nil)
	}
	if r.Path == "" {
		return apierr.New("campo 'caminho' é obrigatório", nil)
	}
	if !strings.HasPrefix(r.Path, "/") {
		return apierr.New("campo 'caminho' deve começar com '/'", nil)
	}
	if !r.Method.IsValid() {
		return apierr.UnprocessableEntity(
			fmt.Sprintf("metodo '%s' inválido. Valores aceitos: %v", r.Method, enum.ValidHTTPMethods()),
			nil,
		)
	}
	if len(r.Service.Steps) == 0 {
		return apierr.New("a rota deve ter ao menos um 'passo'", nil)
	}
	for i, step := range r.Service.Steps {
		if step.Command.Name == "" {
			return apierr.New(fmt.Sprintf("passos[%d].comando.nome é obrigatório", i), nil)
		}
		if !step.Command.Type.IsValid() {
			return apierr.UnprocessableEntity(
				fmt.Sprintf("passos[%d].comando.tipo '%s' inválido. Valores aceitos: %v",
					i, step.Command.Type, enum.ValidCommandTypes()),
				nil,
			)
		}
		if step.Command.Type.IsSQL() && step.Command.Database == "" {
			return apierr.New(
				fmt.Sprintf("passos[%d].comando.database é obrigatório para comandos SQL", i),
				nil,
			)
		}
	}
	if r.Response.Status == 0 {
		r.Response.Status = 200
	}
	return nil
}

func (r *CreateRouteRequest) ToDomain() *Route {
	return &Route{
		Key:      r.Key,
		Path:     r.Path,
		Method:   r.Method,
		Response: r.Response,
		Service:  r.Service,
	}
}
