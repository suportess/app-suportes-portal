package http

import (
	"encoding/json"
	"net/http"

	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/domain"
	"br.tec.suportes/portal/internal/service"
	chi "github.com/go-chi/chi/v5"
)

type RouteHandler struct {
	svc *service.RouteService
}

func NewRouteHandler(svc *service.RouteService) *RouteHandler {
	return &RouteHandler{svc: svc}
}

func (h *RouteHandler) RegisterRoutes(r chi.Router) {
	r.Route("/routes", func(r chi.Router) {
		r.Use(Auth)
		r.Get("/", h.List)
		r.Post("/", h.Register)
		r.Get("/{id}", h.GetByID)
		r.Delete("/{id}", h.Unregister)
	})
}

func (h *RouteHandler) List(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	method := r.URL.Query().Get("method")
	path := r.URL.Query().Get("path")

	var result interface{}
	var aerr apierr.Detail

	switch {
	case key != "":
		result, aerr = h.svc.FindByKeyPattern(key)
	case method != "":
		result, aerr = h.svc.FindByMethod(method)
	case path != "":
		result, aerr = h.svc.FindByPath(path)
	default:
		result, aerr = h.svc.List()
	}

	if aerr != nil {
		writeError(w, r, aerr)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (h *RouteHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, aerr := h.svc.GetByID(id)
	if aerr != nil {
		writeError(w, r, aerr)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (h *RouteHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, apierr.New("payload inválido: "+err.Error(), nil))
		return
	}
	result, aerr := h.svc.Register(&req)
	if aerr != nil {
		writeError(w, r, aerr)
		return
	}
	writeSuccess(w, http.StatusCreated, result)
}

func (h *RouteHandler) Unregister(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	key := r.URL.Query().Get("key")

	if key != "" {
		count, aerr := h.svc.UnregisterByKey(key)
		if aerr != nil {
			writeError(w, r, aerr)
			return
		}
		writeSuccess(w, http.StatusOK, map[string]interface{}{
			"mensagem":      "rotas removidas com sucesso",
			"chave":         key,
			"totalRemovido": count,
		})
		return
	}

	if aerr := h.svc.Unregister(id); aerr != nil {
		writeError(w, r, aerr)
		return
	}
	writeSuccess(w, http.StatusNoContent, nil)
}
