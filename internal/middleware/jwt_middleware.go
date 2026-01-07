package middleware

import (
	"backend/internal/auth"
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDCtxKey contextKey = "userID"
const roleKey contextKey = "role"

func GetUserIDFromContext(ctx context.Context) (uint, bool) {
	id, ok := ctx.Value(userIDCtxKey).(uint)
	return id, ok
}

func GetUserRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleKey).(string)
	return role, ok
}

func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract and validate JWT token from Authorization header
		// If valid, extract user ID and set it in the request context
		// If invalid, respond with 401 Unauthorized
		var tokenString string
		if c, err := r.Cookie("access_token"); err == nil {
			tokenString = c.Value
		}

		if tokenString == "" {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := auth.VerifyJWTToken(tokenString)
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDCtxKey, claims.ID)
		ctx = context.WithValue(ctx, roleKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
