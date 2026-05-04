package service

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/domain"
	"br.tec.suportes/portal/internal/execute"
	"br.tec.suportes/portal/internal/repository"
	chilib "github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type RouteService struct {
	repo      *repository.RouteRepo
	router    chilib.Router
	dbSvc     *DatabaseService
	cmdSvc    *CommandService
	certSvc   *CertificateService
	executors map[string]execute.Executor
}

func NewRouteService(
	repo *repository.RouteRepo,
	router chilib.Router,
	dbSvc *DatabaseService,
	cmdSvc *CommandService,
	certSvc *CertificateService,
) *RouteService {
	certProvider := func(name string) (*domain.Certificate, apierr.Detail) {
		return certSvc.GetByName(name)
	}

	executors := map[string]execute.Executor{
		"QUERY":     execute.NewQueryExecute(),
		"INSERT":    execute.NewInsertExecute(),
		"UPDATE":    execute.NewUpdateExecute(),
		"DELETE":    execute.NewDeleteExecute(),
		"PROCEDURE": execute.NewProcedureExecute(),
		"ANONYMOUS": execute.NewAnonymousExecute(),
		"POST":      execute.NewPostExecute(certProvider),
		"PUT":       execute.NewPutExecute(certProvider),
		"GET":       execute.NewGetExecute(certProvider),
		"REMOVE":    execute.NewRemoveExecute(certProvider),
	}

	svc := &RouteService{
		repo:      repo,
		router:    router,
		dbSvc:     dbSvc,
		cmdSvc:    cmdSvc,
		certSvc:   certSvc,
		executors: executors,
	}
	svc.loadPersistedRoutes()
	return svc
}

func (s *RouteService) loadPersistedRoutes() {
	routes, err := s.repo.FindAll()
	if err != nil {
		log.Printf("[route-service] erro ao carregar rotas: %v", err)
		return
	}
	for _, r := range routes {
		if aerr := s.mount(&r); aerr != nil {
			log.Printf("[route-service] erro ao montar rota '%s': %v", r.Key, aerr.Error())
		}
	}
}

func (s *RouteService) Register(req *domain.CreateRouteRequest) (*domain.Route, apierr.Detail) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if s.repo.ExistsByKey(req.Key) {
		return nil, apierr.Conflict("rota com chave '" + req.Key + "' já existe")
	}

	route := req.ToDomain()

	if aerr := s.mount(route); aerr != nil {
		return nil, aerr
	}
	if err := s.repo.Create(route); err != nil {
		return nil, apierr.New("falha ao persistir rota: "+err.Error(), nil)
	}

	log.Printf("[route-service] rota '%s' registrada em %s %s", route.Key, route.Method, route.Path)
	return route, nil
}

func (s *RouteService) Unregister(id string) apierr.Detail {
	route, err := s.repo.FindByID(id)
	if err != nil {
		return apierr.NotFound(err.Error())
	}
	// Register a no-op handler to shadow the old one
	s.router.Method(string(route.Method), route.Path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))

	if err := s.repo.Delete(route); err != nil {
		return apierr.New(err.Error(), nil)
	}
	log.Printf("[route-service] rota '%s' removida", route.Key)
	return nil
}

func (s *RouteService) UnregisterByKey(key string) (int, apierr.Detail) {
	routes, err := s.repo.FindByKeyPattern(key)
	if err != nil {
		return 0, apierr.New(err.Error(), nil)
	}
	if len(routes) == 0 {
		return 0, apierr.NotFound("nenhuma rota encontrada com chave: " + key)
	}
	count, err2 := s.repo.DeleteByKey(key)
	if err2 != nil {
		return 0, apierr.New(err2.Error(), nil)
	}
	return count, nil
}

func (s *RouteService) List() ([]domain.Route, apierr.Detail) {
	routes, err := s.repo.FindAll()
	if err != nil {
		return nil, apierr.New("falha ao listar rotas: "+err.Error(), nil)
	}
	return routes, nil
}

func (s *RouteService) GetByID(id string) (*domain.Route, apierr.Detail) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		return nil, apierr.NotFound(err.Error())
	}
	return r, nil
}

func (s *RouteService) FindByKeyPattern(pattern string) ([]domain.Route, apierr.Detail) {
	r, err := s.repo.FindByKeyPattern(pattern)
	if err != nil {
		return nil, apierr.New(err.Error(), nil)
	}
	return r, nil
}

func (s *RouteService) FindByMethod(method string) ([]domain.Route, apierr.Detail) {
	r, err := s.repo.FindByMethod(method)
	if err != nil {
		return nil, apierr.New(err.Error(), nil)
	}
	return r, nil
}

func (s *RouteService) FindByPath(path string) ([]domain.Route, apierr.Detail) {
	r, err := s.repo.FindByPath(path)
	if err != nil {
		return nil, apierr.New(err.Error(), nil)
	}
	return r, nil
}

// mount builds the pipeline and registers the net/http handler on chi.
func (s *RouteService) mount(route *domain.Route) apierr.Detail {
	steps := make([]execute.StepConfig, 0, len(route.Service.Steps))
	for _, step := range route.Service.Steps {
		params := make([]execute.StepParam, 0, len(step.Command.Parameters))
		for _, p := range step.Command.Parameters {
			params = append(params, execute.StepParam{
				Name:    p.Name,
				Type:    p.Type,
				Extract: p.Extract,
				Value:   p.Value,
				Field:   p.Field,
			})
		}
		steps = append(steps, execute.StepConfig{
			Alias:        step.Alias,
			CommandKey:   step.Command.Name,
			CommandType:  step.Command.Type,
			DatabaseKey:  step.Command.Database,
			ReturnResult: step.Command.ReturnResult,
			AbortNoData:  step.AbortNoData,
			SingleResult: route.Service.SingleResult,
			Parameters:   params,
		})
	}

	dbProvider := func(key string) (*sqlx.DB, apierr.Detail) {
		return s.dbSvc.GetConnection(key)
	}
	cmdProvider := func(key string) (*domain.Command, apierr.Detail) {
		return s.cmdSvc.GetByKey(key)
	}

	pipeline := execute.NewPipeline(steps, s.executors, route.Service.SingleThread, dbProvider, cmdProvider)
	pathParams := extractPathParams(route.Path)
	status := route.Response.Status
	if status == 0 {
		status = http.StatusOK
	}

	s.router.Method(string(route.Method), route.Path, buildRouteHandler(pipeline, pathParams, status))
	return nil
}

func extractPathParams(path string) []string {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(path, -1)
	var result []string
	for _, m := range matches {
		result = append(result, m[1])
	}
	return result
}

func buildRouteHandler(p *execute.Pipeline, pathParams []string, status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := make(map[string]interface{})
		for _, key := range pathParams {
			params[key] = chilib.URLParam(r, key)
		}
		for k, v := range r.URL.Query() {
			params[k] = v[0]
		}

		var body map[string]interface{}
		if r.Body != nil && !strings.EqualFold(r.Method, "GET") {
			_ = json.NewDecoder(r.Body).Decode(&body)
		}

		headers := make(map[string]string)
		for k, v := range r.Header {
			headers[k] = v[0]
		}

		result, aerr := p.Execute(params, body, headers)
		if aerr != nil {
			routeWriteError(w, r, aerr)
			return
		}
		routeWriteSuccess(w, status, result)
	}
}

func routeWriteError(w http.ResponseWriter, r *http.Request, aerr apierr.Detail) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(aerr.HTTPStatus())
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  aerr.HTTPStatus(),
		"message": aerr.Error(),
		"path":    r.URL.Path,
		"trace":   aerr.Trace(),
	})
}

func routeWriteSuccess(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if body == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
