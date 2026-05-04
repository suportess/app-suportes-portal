package http

import (
	"encoding/json"
	"net/http"

	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/domain"
	"br.tec.suportes/portal/internal/service"
	chi "github.com/go-chi/chi/v5"
)

type CommandHandler struct {
	svc *service.CommandService
}

func NewCommandHandler(svc *service.CommandService) *CommandHandler {
	return &CommandHandler{svc: svc}
}

func (h *CommandHandler) RegisterRoutes(r chi.Router) {
	r.Route("/commands", func(r chi.Router) {
		r.Use(Auth)
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

func (h *CommandHandler) List(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key != "" {
		result, aerr := h.svc.FindByKeyPattern(key)
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

func (h *CommandHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, aerr := h.svc.GetByID(id)
	if aerr != nil {
		writeError(w, r, aerr)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (h *CommandHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateCommandRequest
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

func (h *CommandHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req domain.UpdateCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, apierr.New("payload inválido: "+err.Error(), nil))
		return
	}
	result, aerr := h.svc.Update(id, &req)
	if aerr != nil {
		writeError(w, r, aerr)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (h *CommandHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	key := r.URL.Query().Get("key")

	var aerr apierr.Detail
	if key != "" {
		_, aerr = h.svc.DeleteByKey(key)
	} else {
		aerr = h.svc.Delete(id)
	}
	if aerr != nil {
		writeError(w, r, aerr)
		return
	}
	writeSuccess(w, http.StatusNoContent, nil)
}
