package http

import (
	"net/http"
	"runtime"
	"time"

	"br.tec.suportes/portal/internal/service"
	chi "github.com/go-chi/chi/v5"
)

var startTime = time.Now()

type StatusHandler struct {
	dbSvc  *service.DatabaseService
	status string
}

func NewStatusHandler(dbSvc *service.DatabaseService, status string) *StatusHandler {
	return &StatusHandler{dbSvc: dbSvc, status: status}
}

func (h *StatusHandler) RegisterRoutes(r chi.Router) {
	r.Get("/status", h.Status)
	r.Get("/health", h.Health)
}

func (h *StatusHandler) Status(w http.ResponseWriter, r *http.Request) {
	dbStatus := h.dbSvc.ConnectionStatus()
	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"status":    h.status,
		"uptime":    time.Since(startTime).String(),
		"databases": dbStatus,
	})
}

func (h *StatusHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"status":   "UP",
		"uptime":   time.Since(startTime).String(),
		"goroutines": runtime.NumGoroutine(),
		"memStats": func() map[string]interface{} {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return map[string]interface{}{
				"allocMB":      m.Alloc / 1024 / 1024,
				"totalAllocMB": m.TotalAlloc / 1024 / 1024,
				"sysMB":        m.Sys / 1024 / 1024,
			}
		}(),
	})
}
