package http

import (
	"encoding/json"
	"net/http"

	"aidanwoods.dev/go-paseto"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/http/middleware"
	"github.com/Notifuse/notifuse/internal/service"
)

// WorkspaceServiceInterface is already defined in workspace_handler.go
// So no need to define it again here

type UserHandler struct {
	userService      service.UserServiceInterface
	workspaceService WorkspaceServiceInterface
	config           *config.Config
	publicKey        paseto.V4AsymmetricPublicKey
}

func NewUserHandler(userService service.UserServiceInterface, workspaceService WorkspaceServiceInterface, cfg *config.Config, publicKey paseto.V4AsymmetricPublicKey) *UserHandler {
	return &UserHandler{
		userService:      userService,
		workspaceService: workspaceService,
		config:           cfg,
		publicKey:        publicKey,
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

// GetCurrentUser returns the authenticated user and their workspaces
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	authUser, ok := r.Context().Value(middleware.AuthUserKey).(*middleware.AuthenticatedUser)
	if !ok || authUser == nil {
		WriteJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user details
	user, err := h.userService.GetUserByID(r.Context(), authUser.ID)
	if err != nil {
		WriteJSONError(w, "User not found", http.StatusNotFound)
		return
	}

	// Get user's workspaces
	workspaces, err := h.workspaceService.ListWorkspaces(r.Context(), authUser.ID)
	if err != nil {
		WriteJSONError(w, "Failed to retrieve workspaces", http.StatusInternalServerError)
		return
	}

	// Combine user and workspaces in response
	response := map[string]interface{}{
		"user":       user,
		"workspaces": workspaces,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	// Public routes (no auth required)
	mux.HandleFunc("/api/user.signin", h.SignIn)
	mux.HandleFunc("/api/user.verify", h.VerifyCode)

	// Protected routes (auth required)
	// Create auth middleware if we have a userService that implements the AuthServiceInterface
	authService, ok := h.userService.(middleware.AuthServiceInterface)
	if ok {
		// Create auth middleware with the public key
		authMiddleware := middleware.NewAuthMiddleware(h.publicKey)
		requireAuth := authMiddleware.RequireAuth(authService)

		// Register protected routes
		mux.Handle("/api/user.me", requireAuth(http.HandlerFunc(h.GetCurrentUser)))
	}
}
