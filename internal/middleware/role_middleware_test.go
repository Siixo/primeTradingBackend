package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	adminRoute     = "/admin"
	statusMismatch = "status = %d, want %d"
)

func TestAdminRoleMiddlewareRejectsMissingRole(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, adminRoute, nil)
	rr := httptest.NewRecorder()

	AdminRoleMiddleware(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf(statusMismatch, rr.Code, http.StatusForbidden)
	}
	if nextCalled {
		t.Fatal("next should not be called")
	}
}

func TestAdminRoleMiddlewareRejectsNonAdmin(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, adminRoute, nil)
	ctx := context.WithValue(req.Context(), roleKey, "user")
	rr := httptest.NewRecorder()

	AdminRoleMiddleware(next).ServeHTTP(rr, req.WithContext(ctx))

	if rr.Code != http.StatusForbidden {
		t.Fatalf(statusMismatch, rr.Code, http.StatusForbidden)
	}
	if nextCalled {
		t.Fatal("next should not be called")
	}
}

func TestAdminRoleMiddlewareAllowsAdmin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, adminRoute, nil)
	ctx := context.WithValue(req.Context(), roleKey, "admin")
	rr := httptest.NewRecorder()

	AdminRoleMiddleware(next).ServeHTTP(rr, req.WithContext(ctx))

	if rr.Code != http.StatusOK {
		t.Fatalf(statusMismatch, rr.Code, http.StatusOK)
	}
}
