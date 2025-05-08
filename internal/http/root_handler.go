package http

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/Notifuse/notifuse/pkg/logger"
)

type RootHandler struct {
	consoleDir string
	logger     logger.Logger
}

func NewRootHandler() *RootHandler {
	return &RootHandler{}
}

// NewRootHandlerWithConsole creates a root handler that also serves console static files
func NewRootHandlerWithConsole(consoleDir string, logger logger.Logger) *RootHandler {
	return &RootHandler{
		consoleDir: consoleDir,
		logger:     logger,
	}
}

func (h *RootHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// If configured to serve console files and path doesn't start with /api
	if h.consoleDir != "" && !strings.HasPrefix(r.URL.Path, "/api") {
		h.serveConsole(w, r)
		return
	}

	// Default API root response
	if r.URL.Path == "/api" || r.URL.Path == "/api/" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "api running",
		})
	} else {
		// For unhandled API paths
		http.NotFound(w, r)
	}
}

// serveConsole handles serving static files, with a fallback for SPA routing
func (h *RootHandler) serveConsole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create file server for console files
	fs := http.FileServer(http.Dir(h.consoleDir))

	path := h.consoleDir + r.URL.Path
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// If the requested file doesn't exist, serve index.html for SPA routing
		r.URL.Path = "/"
	}

	fs.ServeHTTP(w, r)
}

func (h *RootHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.Handle)

	// If console directory is configured, add specific /api route
	if h.consoleDir != "" {
		mux.HandleFunc("/api", h.Handle)
	}
}
