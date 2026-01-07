package application

import (
	"backend/internal/domain/model"
	"backend/internal/handler/dto"
	stdErrors "errors"

	"backend/internal/auth"

	"log"

	"golang.org/x/crypto/bcrypt"
)

func (s *UserService) Login(req dto.LoginRequest) (model.User, string, error) {
	if req.Identifier == "" {
		return model.User{}, "", stdErrors.New("user or email is required")
	}
	if req.Password == "" {
		return model.User{}, "", stdErrors.New("password is required")
	}
	// TODO add validations step , length check basic format etc..

	user, err := s.userRepo.FindByUsernameOrEmail(req.Identifier)
	if err != nil {
		return model.User{}, "", stdErrors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return model.User{}, "", stdErrors.New("invalid credentials")
	}
	log.Printf("Password is correct for user: %s", user.Username)

	//Token generation
	token, err := auth.GenerateJWTToken(user.ID, user.Username, user.Role)
	if err != nil {
		log.Printf("Error generating JWT token: %v", err)
		return model.User{}, "", err
	}
	log.Printf("Token generated for user %s: %s", user.Username, token)

	return user, token, nil
}
func (s *UserService) RefreshToken(tokenString string) (string, error) {
	claims, err := auth.VerifyJWTToken(tokenString)
	if err != nil {
		return "", err
	}
	return auth.RefreshJWToken(claims)
}
