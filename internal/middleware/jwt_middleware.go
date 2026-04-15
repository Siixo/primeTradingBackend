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

func NewJWTAuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	secret := []byte(jwtSecret)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string
			if c, err := r.Cookie("access_token"); err == nil {
				tokenString = c.Value
			}

			if tokenString == "" {
				authHeader := r.Header.Get("Authorization")
				if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
					tokenString = after
				}
			}

			if tokenString == "" {
				writeJSONError(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := auth.VerifyJWTToken(secret, tokenString)
			if err != nil {
				writeJSONError(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDCtxKey, claims.ID)
			ctx = context.WithValue(ctx, roleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
