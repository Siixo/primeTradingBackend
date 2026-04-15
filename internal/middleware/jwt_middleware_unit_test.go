package middleware

import (
	"backend/internal/auth"
	"net/http"
	"net/http/httptest"
	"testing"
)

const jwtMiddlewareSecret = "jwt-middleware-secret"

func TestJWTAuthMiddlewareMissingTokenReturnsUnauthorized(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	rr := httptest.NewRecorder()

	NewJWTAuthMiddleware(jwtMiddlewareSecret)(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf(statusMismatch, rr.Code, http.StatusUnauthorized)
	}
	if nextCalled {
		t.Fatal("next should not be called")
	}
}

func TestJWTAuthMiddlewareValidCookieInjectsClaimsInContext(t *testing.T) {
	token, err := auth.GenerateJWTToken([]byte(jwtMiddlewareSecret), 11, "alice", "admin")
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := GetUserIDFromContext(r.Context())
		if !ok || id != 11 {
			t.Fatalf("user id from context = %d, ok=%v", id, ok)
		}
		role, ok := GetUserRoleFromContext(r.Context())
		if !ok || role != "admin" {
			t.Fatalf("role from context = %q, ok=%v", role, ok)
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	rr := httptest.NewRecorder()

	NewJWTAuthMiddleware(jwtMiddlewareSecret)(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf(statusMismatch, rr.Code, http.StatusOK)
	}
}
