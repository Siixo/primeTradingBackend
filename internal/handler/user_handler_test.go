package handler

import (
	"backend/internal/application"
	"backend/internal/domain/model"
	"backend/internal/middleware"
	"bytes"
	"encoding/json"
	stdErrors "errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

const (
	apiPathRegister   = "/api/register"
	apiPathLogin      = "/api/login"
	apiPathMe         = "/api/me"
	apiPathRefresh    = "/api/refresh"
	apiPathLogout     = "/api/logout"
	authTestSecret    = "handler-test-secret"
	testEmail         = "alice@example.com"
	testUsername      = "alice"
	strongPass        = "ValidStrongPassword123!"
	statusFormat      = "status = %d, want %d"
	accessTokenCookie = "access_token"
	notFoundMsg       = "not found"
)

type fakeUserRepoHandler struct {
	saveFn                  func(user model.User) error
	findByUsernameOrEmailFn func(identifier string) (model.User, error)
	findByIDFn              func(id uint) (model.User, error)

	savedUsers []model.User
}

func (f *fakeUserRepoHandler) Migrate() error { return nil }

func (f *fakeUserRepoHandler) Save(user model.User) error {
	f.savedUsers = append(f.savedUsers, user)
	if f.saveFn != nil {
		return f.saveFn(user)
	}
	return nil
}

func (f *fakeUserRepoHandler) FindByUsernameOrEmail(identifier string) (model.User, error) {
	if f.findByUsernameOrEmailFn != nil {
		return f.findByUsernameOrEmailFn(identifier)
	}
	return model.User{}, stdErrors.New(notFoundMsg)
}

func (f *fakeUserRepoHandler) FindByID(id uint) (model.User, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(id)
	}
	return model.User{}, stdErrors.New(notFoundMsg)
}

func newHandlerWithRepo(repo *fakeUserRepoHandler) *UserHandler {
	svc := application.NewUserService(repo)
	return NewUserHandler(svc)
}

func TestRegisterUserHandlerMethodNotAllowed(t *testing.T) {
	h := newHandlerWithRepo(&fakeUserRepoHandler{})

	req := httptest.NewRequest(http.MethodGet, apiPathRegister, nil)
	rr := httptest.NewRecorder()

	h.RegisterUserHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf(statusFormat, rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestRegisterUserHandlerSuccess(t *testing.T) {
	repo := &fakeUserRepoHandler{}
	h := newHandlerWithRepo(repo)

	body := map[string]string{
		"username":  testUsername,
		"email":     testEmail,
		"password":  strongPass,
		"password2": strongPass,
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, apiPathRegister, bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	h.RegisterUserHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf(statusFormat, rr.Code, http.StatusCreated)
	}
	if len(repo.savedUsers) != 1 {
		t.Fatalf("saved users = %d, want 1", len(repo.savedUsers))
	}
	if repo.savedUsers[0].Password == strongPass {
		t.Fatal("expected hashed password, got plaintext")
	}
}

func TestLoginUserHandlerSuccessSetsCookieAndResponse(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", authTestSecret)

	hash, err := bcrypt.GenerateFromPassword([]byte(strongPass), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash generation failed: %v", err)
	}

	repo := &fakeUserRepoHandler{
		findByUsernameOrEmailFn: func(identifier string) (model.User, error) {
			return model.User{ID: 42, Username: testUsername, Email: testEmail, Password: string(hash), Role: "admin"}, nil
		},
	}
	h := newHandlerWithRepo(repo)

	body := map[string]string{"identifier": testUsername, "password": strongPass}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, apiPathLogin, bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	h.LoginUserHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf(statusFormat, rr.Code, http.StatusOK)
	}

	res := rr.Result()
	defer res.Body.Close()

	var loginResp map[string]any
	if err := json.NewDecoder(res.Body).Decode(&loginResp); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if loginResp["username"] != testUsername {
		t.Fatalf("username = %v, want %s", loginResp["username"], testUsername)
	}

	foundCookie := false
	for _, c := range res.Cookies() {
		if c.Name == accessTokenCookie {
			foundCookie = true
			if !c.HttpOnly {
				t.Fatal("access token cookie must be HttpOnly")
			}
			if c.MaxAge != 3600 {
				t.Fatalf("cookie MaxAge = %d, want 3600", c.MaxAge)
			}
		}
	}
	if !foundCookie {
		t.Fatal("expected access_token cookie")
	}
}

func TestLoginUserHandlerInvalidCredentials(t *testing.T) {
	repo := &fakeUserRepoHandler{
		findByUsernameOrEmailFn: func(identifier string) (model.User, error) {
			return model.User{}, stdErrors.New(notFoundMsg)
		},
	}
	h := newHandlerWithRepo(repo)

	body := map[string]string{"identifier": testUsername, "password": "wrong"}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, apiPathLogin, bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	h.LoginUserHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf(statusFormat, rr.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rr.Body.String(), "invalid credentials") {
		t.Fatalf("body = %q, expected invalid credentials", rr.Body.String())
	}
}

func TestMeHandlerUnauthorizedWithoutJWTContext(t *testing.T) {
	h := newHandlerWithRepo(&fakeUserRepoHandler{})

	req := httptest.NewRequest(http.MethodGet, apiPathMe, nil)
	rr := httptest.NewRecorder()

	h.MeHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf(statusFormat, rr.Code, http.StatusUnauthorized)
	}
}

func TestMeHandlerSuccessThroughJWTMiddleware(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", authTestSecret)
	hash, err := bcrypt.GenerateFromPassword([]byte(strongPass), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash generation failed: %v", err)
	}

	repo := &fakeUserRepoHandler{
		findByUsernameOrEmailFn: func(identifier string) (model.User, error) {
			return model.User{ID: 77, Username: testUsername, Email: testEmail, Password: string(hash), Role: "user"}, nil
		},
		findByIDFn: func(id uint) (model.User, error) {
			return model.User{ID: id, Username: testUsername, Email: testEmail}, nil
		},
	}
	h := newHandlerWithRepo(repo)

	tokenSvc := application.NewUserService(repo)
	_, token, err := tokenSvc.Login(application.LoginInput{Identifier: testUsername, Password: strongPass})
	if err != nil {
		t.Fatalf("failed to generate login token: %v", err)
	}

	next := middleware.JWTAuthMiddleware(http.HandlerFunc(h.MeHandler))
	req := httptest.NewRequest(http.MethodGet, apiPathMe, nil)
	req.AddCookie(&http.Cookie{Name: accessTokenCookie, Value: token})
	rr := httptest.NewRecorder()

	next.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf(statusFormat, rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), testUsername) {
		t.Fatalf("body = %q, expected username", rr.Body.String())
	}
}

func TestRefreshJWTTokenHandlerMissingCookie(t *testing.T) {
	h := newHandlerWithRepo(&fakeUserRepoHandler{})

	req := httptest.NewRequest(http.MethodPost, apiPathRefresh, nil)
	rr := httptest.NewRecorder()

	h.RefreshJWTokenHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf(statusFormat, rr.Code, http.StatusUnauthorized)
	}
}

func TestLogoutUserHandlerClearsCookie(t *testing.T) {
	h := newHandlerWithRepo(&fakeUserRepoHandler{})

	req := httptest.NewRequest(http.MethodPost, apiPathLogout, nil)
	rr := httptest.NewRecorder()

	h.LogoutUserHandler(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf(statusFormat, rr.Code, http.StatusNoContent)
	}

	res := rr.Result()
	defer res.Body.Close()

	found := false
	for _, c := range res.Cookies() {
		if c.Name == accessTokenCookie {
			found = true
			if c.MaxAge != -1 {
				t.Fatalf("logout cookie MaxAge = %d, want -1", c.MaxAge)
			}
		}
	}
	if !found {
		t.Fatal("expected cleared access_token cookie")
	}
}
