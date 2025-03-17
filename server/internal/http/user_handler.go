package http

import (
	"encoding/json"
	"net/http"

	"notifuse/server/internal/service"
)

type UserHandler struct {
	userService service.UserServiceInterface
}

func NewUserHandler(userService service.UserServiceInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var input service.SignInInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
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
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
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
