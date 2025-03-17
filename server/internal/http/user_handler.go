package http

import (
	"encoding/json"
	"net/http"

	"notifuse/server/config"
	"notifuse/server/internal/service"
)

type UserHandler struct {
	userService service.UserServiceInterface
	config      *config.Config
}

func NewUserHandler(userService service.UserServiceInterface, cfg *config.Config) *UserHandler {
	return &UserHandler{
		userService: userService,
		config:      cfg,
	}
}

func (h *UserHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var input service.SignInInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteJSONError(w, "Invalid SignIn request body", http.StatusBadRequest)
		return
	}

	// In development mode, we'll return the magic code directly
	if h.config.IsDevelopment() {
		code, err := h.userService.SignInDev(r.Context(), input)
		if err != nil {
			WriteJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Magic code sent to your email",
			"code":    code,
		})
		return
	}

	if err := h.userService.SignIn(r.Context(), input); err != nil {
		WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Magic code sent to your email",
	})
}

func (h *UserHandler) VerifyCode(w http.ResponseWriter, r *http.Request) {
	var input service.VerifyCodeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteJSONError(w, "Invalid VerifyCode request body", http.StatusBadRequest)
		return
	}

	response, err := h.userService.VerifyCode(r.Context(), input)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/user.signin", h.SignIn)
	mux.HandleFunc("/api/user.verify", h.VerifyCode)
}
