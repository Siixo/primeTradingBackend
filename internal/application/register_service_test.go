package application

import (
	"backend/internal/domain/model"
	validationErrors "backend/internal/errors"
	"context"
	stdErrors "errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

const (
	notFoundErr    = "not found"
	aliceEmail     = "alice@example.com"
	strongPassword = "ValidStrongPassword123!"
	usernameAlice  = "alice"
)

func TestRegisterReturnsValidationErrors(t *testing.T) {
	repo := &fakeUserRepository{}
	svc := NewUserService(repo, "register-test-secret")

	err := svc.Register(context.Background(), RegisterInput{
		Username:  "ab",
		Email:     "bad-email",
		Password:  "123",
		Password2: "1234",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	var vErrs validationErrors.ValidationErrors
	if !stdErrors.As(err, &vErrs) {
		t.Fatalf("expected ValidationErrors type, got %T (%v)", err, err)
	}
}

func TestRegisterReturnsUsernameAlreadyTaken(t *testing.T) {
	repo := &fakeUserRepository{
		findByUsernameOrEmailFn: func(identifier string) (model.User, error) {
			if identifier == usernameAlice {
				return model.User{ID: 1, Username: usernameAlice}, nil
			}
			return model.User{}, stdErrors.New(notFoundErr)
		},
	}
	svc := NewUserService(repo, "register-test-secret")

	err := svc.Register(context.Background(), RegisterInput{
		Username:  usernameAlice,
		Email:     aliceEmail,
		Password:  strongPassword,
		Password2: strongPassword,
	})
	if err == nil || err.Error() != "username already taken" {
		t.Fatalf("error = %v, want username already taken", err)
	}
}

func TestRegisterReturnsEmailAlreadyExists(t *testing.T) {
	repo := &fakeUserRepository{
		findByUsernameOrEmailFn: func(identifier string) (model.User, error) {
			if identifier == aliceEmail {
				return model.User{ID: 2, Email: aliceEmail}, nil
			}
			return model.User{}, stdErrors.New(notFoundErr)
		},
	}
	svc := NewUserService(repo, "register-test-secret")

	err := svc.Register(context.Background(), RegisterInput{
		Username:  usernameAlice,
		Email:     aliceEmail,
		Password:  strongPassword,
		Password2: strongPassword,
	})
	if err == nil || err.Error() != "an account with this email already exists" {
		t.Fatalf("error = %v, want duplicate email error", err)
	}
}

func TestRegisterHashesPasswordBeforeSave(t *testing.T) {
	repo := &fakeUserRepository{
		findByUsernameOrEmailFn: func(identifier string) (model.User, error) {
			return model.User{}, stdErrors.New(notFoundErr)
		},
	}
	svc := NewUserService(repo, "register-test-secret")

	plain := strongPassword
	err := svc.Register(context.Background(), RegisterInput{
		Username:  usernameAlice,
		Email:     aliceEmail,
		Password:  plain,
		Password2: plain,
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if len(repo.savedUsers) != 1 {
		t.Fatalf("saved users = %d, want 1", len(repo.savedUsers))
	}
	saved := repo.savedUsers[0]
	if saved.Password == plain {
		t.Fatal("password should be hashed, got plaintext")
	}
	if strings.TrimSpace(saved.Password) == "" {
		t.Fatal("saved password hash is empty")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(saved.Password), []byte(plain)); err != nil {
		t.Fatalf("saved hash does not match plaintext password: %v", err)
	}
}
