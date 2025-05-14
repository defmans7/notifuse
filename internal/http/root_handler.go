package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Notifuse/notifuse/pkg/logger"
)

type RootHandler struct {
	consoleDir            string
	notificationCenterDir string
	logger                logger.Logger
	apiEndpoint           string
}

func NewRootHandler() *RootHandler {
	return &RootHandler{}
}

// NewRootHandlerWithConsole creates a root handler that also serves console static files
func NewRootHandlerWithConsole(consoleDir string, logger logger.Logger, apiEndpoint string) *RootHandler {
	return &RootHandler{
		consoleDir:  consoleDir,
		logger:      logger,
		apiEndpoint: apiEndpoint,
	}
}

// NewRootHandlerWithConsoleAndNotificationCenter creates a root handler that serves both console and notification center static files
func NewRootHandlerWithConsoleAndNotificationCenter(consoleDir string, notificationCenterDir string, logger logger.Logger, apiEndpoint string) *RootHandler {
	return &RootHandler{
		consoleDir:            consoleDir,
		notificationCenterDir: notificationCenterDir,
		logger:                logger,
		apiEndpoint:           apiEndpoint,
	}
}

func (h *RootHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Handle config.js request
	if r.URL.Path == "/config.js" {
		h.serveConfigJS(w, r)
		return
	}

	// Handle notification center requests
	if h.notificationCenterDir != "" && strings.HasPrefix(r.URL.Path, "/notification-center") {
		h.serveNotificationCenter(w, r)
		return
	}

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

// serveConfigJS generates and serves the config.js file with environment variables
func (h *RootHandler) serveConfigJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	configJS := fmt.Sprintf("window.API_ENDPOINT = %q;", h.apiEndpoint)
	w.Write([]byte(configJS))
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

// serveNotificationCenter handles serving notification center static files, with a fallback for SPA routing
func (h *RootHandler) serveNotificationCenter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Strip the prefix to match the file structure
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/notification-center")
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}

	// Create file server for notification center files
	fs := http.FileServer(http.Dir(h.notificationCenterDir))

	path := h.notificationCenterDir + r.URL.Path
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// If the requested file doesn't exist, serve index.html for SPA routing
		r.URL.Path = "/"
	}

	fs.ServeHTTP(w, r)
}

func (h *RootHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.Handle)
	mux.HandleFunc("/config.js", h.serveConfigJS)

	// If notification center directory is configured, register it
	if h.notificationCenterDir != "" {
		mux.HandleFunc("/notification-center/", h.Handle)
	}

	// If console directory is configured, add specific /api route
	if h.consoleDir != "" {
		mux.HandleFunc("/api", h.Handle)
	}
}
