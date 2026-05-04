package http

import (
	"encoding/json"
	"net/http"

	"br.tec.suportes/portal/internal/apierr"
	"br.tec.suportes/portal/internal/domain"
	"br.tec.suportes/portal/internal/service"
	chi "github.com/go-chi/chi/v5"
)

type CertificateHandler struct {
	svc *service.CertificateService
}

func NewCertificateHandler(svc *service.CertificateService) *CertificateHandler {
	return &CertificateHandler{svc: svc}
}

func (h *CertificateHandler) RegisterRoutes(r chi.Router) {
	r.Route("/certificates", func(r chi.Router) {
		r.Use(Auth)
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

func (h *CertificateHandler) List(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name != "" {
		result, aerr := h.svc.GetByName(name)
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

func (h *CertificateHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, aerr := h.svc.GetByID(id)
	if aerr != nil {
		writeError(w, r, aerr)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (h *CertificateHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateCertificateRequest
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

func (h *CertificateHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req domain.UpdateCertificateRequest
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

func (h *CertificateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	name := r.URL.Query().Get("name")

	var aerr apierr.Detail
	if name != "" {
		aerr = h.svc.DeleteByName(name)
	} else {
		aerr = h.svc.Delete(id)
	}
	if aerr != nil {
		writeError(w, r, aerr)
		return
	}
	writeSuccess(w, http.StatusNoContent, nil)
}
