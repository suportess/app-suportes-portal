package http

import (
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"br.tec.suportes/portal/internal/apierr"
)

type errorBody struct {
	Timestamp string      `json:"timestamp"`
	Status    int         `json:"status"`
	Error     string      `json:"error"`
	Message   string      `json:"message"`
	Path      string      `json:"path"`
	Trace     interface{} `json:"trace,omitempty"`
}

// Handler wraps a controller func and writes JSON responses.
func Handler(fn func(http.ResponseWriter, *http.Request) (interface{}, apierr.Detail), successStatus int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, aerr := fn(w, r)
		if aerr != nil {
			writeError(w, r, aerr)
			return
		}
		writeSuccess(w, successStatus, result)
	}
}

func writeError(w http.ResponseWriter, r *http.Request, aerr apierr.Detail) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(aerr.HTTPStatus())
	_ = json.NewEncoder(w).Encode(errorBody{
		Timestamp: time.Now().Format(time.RFC3339),
		Status:    aerr.HTTPStatus(),
		Error:     http.StatusText(aerr.HTTPStatus()),
		Message:   aerr.Error(),
		Path:      r.URL.Path,
		Trace:     aerr.Trace(),
	})
}

func writeSuccess(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")

	if body == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	v := reflect.ValueOf(body)
	if v.Kind() == reflect.Slice && v.Len() == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
