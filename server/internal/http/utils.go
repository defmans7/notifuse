package http

import (
	"encoding/json"
	"net/http"
)

// WriteJSONError writes a JSON error response with the given message and status code.
// It sets the Content-Type header to application/json and automatically formats
// the response as {"error": "message"}.
func WriteJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
