// internal/adapter/http/user_handler.go
package handler

import (
	"backend/internal/application"
	"backend/internal/handler/dto"
	"backend/internal/middleware"
	"encoding/json"
	"log"
	"net/http"
)

type UserHandler struct {
	userService *application.UserService
}

func NewUserHandler(userService *application.UserService) *UserHandler {
	return &UserHandler{userService}
}

// Register handler for user registration
func (h *UserHandler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {

	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.RegisterRequest

	// Decode the request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call the service to register the user
	if err := h.userService.Register(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

// Login handler for user login
func (h *UserHandler) LoginUserHandler(w http.ResponseWriter, r *http.Request) {

	// Only allow POST method
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req dto.LoginRequest

	// Decode the request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Call the service to login the user
	user, token, err := h.userService.Login(req)
	if err != nil {
		log.Printf("Login error: %v", err) // Add this line for debugging
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Prepare the response without token since it's written in the cookie
	response := dto.LoginResponse{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	// Set the token in a secure HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Logout handler for user logout
func (h *UserHandler) LogoutUserHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

// Fetch current user info
func (h *UserHandler) MeHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch user info from the service
	user, err := h.userService.FindByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Prepare and send the response
	response := dto.MeResponse{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	// Send the response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) RefreshJWTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("access_token")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	tokenStr := cookie.Value
	newToken, err := h.userService.RefreshToken(tokenStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Set the token in a secure HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    newToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Token refreshed successfully"})
}
