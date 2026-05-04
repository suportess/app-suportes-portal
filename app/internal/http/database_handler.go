package http

import (
	"encoding/json"
	"net/http"

	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/domain"
	"br.tec.suportes/portal/internal/service"
	"github.com/go-chi/chi/v5"
)

type DatabaseHandler struct {
	svc *service.DatabaseService
}

func NewDatabaseHandler(svc *service.DatabaseService) *DatabaseHandler {
	return &DatabaseHandler{svc: svc}
}

func (h *DatabaseHandler) RegisterRoutes(r chi.Router) {
	r.Route("/databases", func(r chi.Router) {
		r.Use(Auth)
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Delete("/{id}", h.Delete)
	})
}

func (h *DatabaseHandler) List(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key != "" {
		result, aerr := h.svc.GetByKey(key)
		if aerr != nil {
			writeError(w, r, aerr)
			return
		}
		writeSuccess(w, http.StatusOK, result)
		return
	}
	result, aerr := h.svc.List()
	if aerr != nil {
		writeError(w, r, aerr)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (h *DatabaseHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, aerr := h.svc.GetByID(id)
	if aerr != nil {
		writeError(w, r, aerr)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (h *DatabaseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateDatabaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, apierr.New("payload inválido: "+err.Error(), nil))
		return
	}
	result, aerr := h.svc.Create(&req)
	if aerr != nil {
		writeError(w, r, aerr)
		return
	}
	writeSuccess(w, http.StatusCreated, result)
}

func (h *DatabaseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	key := r.URL.Query().Get("key")

	var aerr apierr.Detail
	if key != "" {
		aerr = h.svc.DeleteByKey(key)
	} else {
		aerr = h.svc.Delete(id)
	}
	if aerr != nil {
		writeError(w, r, aerr)
		return
	}
	writeSuccess(w, http.StatusNoContent, nil)
}
