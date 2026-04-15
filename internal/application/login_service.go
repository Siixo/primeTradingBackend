package application

import (
	"backend/internal/domain/model"
	"context"
	stdErrors "errors"

	"backend/internal/auth"

	"log"

	"golang.org/x/crypto/bcrypt"
)

func (s *UserService) Login(ctx context.Context, req LoginInput) (model.User, string, error) {
	if req.Identifier == "" {
		return model.User{}, "", stdErrors.New("user or email is required")
	}
	if req.Password == "" {
		return model.User{}, "", stdErrors.New("password is required")
	}

	user, err := s.userRepo.FindByUsernameOrEmail(ctx, req.Identifier)
	if err != nil {
		return model.User{}, "", stdErrors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return model.User{}, "", stdErrors.New("invalid credentials")
	}

	token, err := auth.GenerateJWTToken(s.jwtSecret, user.ID, user.Username, user.Role)
	if err != nil {
		log.Printf("Error generating JWT token: %v", err)
		return model.User{}, "", err
	}
	log.Printf("Token generated for user %s", user.Username)

	return user, token, nil
}

func (s *UserService) RefreshToken(tokenString string) (string, error) {
	claims, err := auth.VerifyJWTToken(s.jwtSecret, tokenString)
	if err != nil {
		return "", err
	}
	return auth.RefreshJWToken(s.jwtSecret, claims)
}
