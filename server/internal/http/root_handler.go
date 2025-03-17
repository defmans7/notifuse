package http

import (
	"encoding/json"
	"net/http"
)

type RootHandler struct{}

func NewRootHandler() *RootHandler {
	return &RootHandler{}
}

func (h *RootHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "api running",
	})
}

func (h *RootHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.Handle)
}
