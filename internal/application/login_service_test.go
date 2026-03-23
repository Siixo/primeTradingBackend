package application

import (
	"backend/internal/domain/model"
	stdErrors "errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

const correctPassword = "correct-password"

func TestLoginRejectsEmptyIdentifier(t *testing.T) {
	svc := NewUserService(&fakeUserRepository{})

	_, _, err := svc.Login(LoginInput{Password: "abc"})
	if err == nil || err.Error() != "user or email is required" {
		t.Fatalf("error = %v, want user or email is required", err)
	}
}

func TestLoginRejectsEmptyPassword(t *testing.T) {
	svc := NewUserService(&fakeUserRepository{})

	_, _, err := svc.Login(LoginInput{Identifier: "alice"})
	if err == nil || err.Error() != "password is required" {
		t.Fatalf("error = %v, want password is required", err)
	}
}

func TestLoginReturnsInvalidCredentialsWhenUserMissing(t *testing.T) {
	repo := &fakeUserRepository{
		findByUsernameOrEmailFn: func(identifier string) (model.User, error) {
			return model.User{}, stdErrors.New("not found")
		},
	}
	svc := NewUserService(repo)

	_, _, err := svc.Login(LoginInput{Identifier: "alice", Password: "password"})
	if err == nil || err.Error() != "invalid credentials" {
		t.Fatalf("error = %v, want invalid credentials", err)
	}
}

func TestLoginReturnsInvalidCredentialsWhenPasswordMismatch(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash test password: %v", err)
	}

	repo := &fakeUserRepository{
		findByUsernameOrEmailFn: func(identifier string) (model.User, error) {
			return model.User{ID: 1, Username: "alice", Password: string(hash), Role: "user"}, nil
		},
	}
	svc := NewUserService(repo)

	_, _, err = svc.Login(LoginInput{Identifier: "alice", Password: "wrong-password"})
	if err == nil || err.Error() != "invalid credentials" {
		t.Fatalf("error = %v, want invalid credentials", err)
	}
}

func TestLoginSuccessReturnsTokenAndUser(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "app-test-secret")

	hash, err := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash test password: %v", err)
	}

	repo := &fakeUserRepository{
		findByUsernameOrEmailFn: func(identifier string) (model.User, error) {
			return model.User{ID: 42, Username: "alice", Password: string(hash), Role: "admin"}, nil
		},
	}
	svc := NewUserService(repo)

	user, token, err := svc.Login(LoginInput{Identifier: "alice", Password: correctPassword})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if user.ID != 42 || user.Username != "alice" {
		t.Fatalf("unexpected user returned: %+v", user)
	}
	if strings.TrimSpace(token) == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestRefreshTokenRejectsInvalidToken(t *testing.T) {
	svc := NewUserService(&fakeUserRepository{})

	_, err := svc.RefreshToken("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
