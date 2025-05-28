package http

import (
	"net/http"

	"github.com/Notifuse/notifuse/internal/service"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// DemoHandler handles HTTP requests for demo operations
type DemoHandler struct {
	service *service.DemoService
	logger  logger.Logger
}

// NewDemoHandler creates a new demo handler
func NewDemoHandler(service *service.DemoService, logger logger.Logger) *DemoHandler {
	return &DemoHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers the demo HTTP endpoints
func (h *DemoHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/demo.reset", h.handleResetDemo)
}

// handleResetDemo handles the GET request to reset demo data
func (h *DemoHandler) handleResetDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get HMAC from query string
	providedHMAC := r.URL.Query().Get("hmac")
	if providedHMAC == "" {
		WriteJSONError(w, "Missing HMAC parameter", http.StatusBadRequest)
		return
	}

	// Verify HMAC using the service
	if !h.service.VerifyRootEmailHMAC(providedHMAC) {
		h.logger.WithField("provided_hmac", providedHMAC).Warn("Invalid HMAC provided for demo reset")
		WriteJSONError(w, "Invalid authentication", http.StatusUnauthorized)
		return
	}

	// Reset demo data
	if err := h.service.ResetDemo(r.Context()); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to reset demo data")
		WriteJSONError(w, "Failed to reset demo data", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Demo data reset successfully",
	})
}
