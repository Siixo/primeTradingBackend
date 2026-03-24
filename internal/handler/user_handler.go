// internal/handler/user_handler.go
package handler

import (
	"backend/internal/application"
	"backend/internal/domain/model"
	"backend/internal/handler/dto"
	"backend/internal/middleware"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

// UserServicePort defines the contract the handler depends on.
// This decouples the handler from the concrete *application.UserService.
type UserServicePort interface {
	Login(req application.LoginInput) (model.User, string, error)
	Register(req application.RegisterInput) error
	FindByID(id uint) (model.User, error)
	RefreshToken(tokenString string) (string, error)
}

type UserHandler struct {
	userService UserServicePort
}

func NewUserHandler(userService UserServicePort) *UserHandler {
	return &UserHandler{userService}
}

// Register handler for user registration
func (h *UserHandler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.userService.Register(application.RegisterInput{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		Password2: req.Password2,
	}); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

// Login handler for user login
func (h *UserHandler) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, token, err := h.userService.Login(application.LoginInput{
		Identifier: req.Identifier,
		Password:   req.Password,
	})
	if err != nil {
		log.Printf("Login error: %v", err)
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secureCookieEnabled(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dto.LoginResponse{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}

// Logout handler for user logout
func (h *UserHandler) LogoutUserHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secureCookieEnabled(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

// MeHandler fetches the current authenticated user's info
func (h *UserHandler) MeHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.FindByID(userID)
	if err != nil {
		jsonError(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dto.MeResponse{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}

func (h *UserHandler) RefreshJWTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("access_token")
	if err != nil {
		if err == http.ErrNoCookie {
			jsonError(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		jsonError(w, "bad request", http.StatusBadRequest)
		return
	}

	newToken, err := h.userService.RefreshToken(cookie.Value)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    newToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secureCookieEnabled(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "token refreshed successfully"})
}

func secureCookieEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("COOKIE_SECURE")))
	return v == "1" || v == "true" || v == "yes"
}
