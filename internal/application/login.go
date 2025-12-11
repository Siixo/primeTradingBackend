package application

import (
	"backend/internal/domain/model"
	stdErrors "errors"

	"golang.org/x/crypto/bcrypt"
)

func (s *UserService) Login(usernameOrEmail string, password string) (model.User, string, error) {
	if usernameOrEmail == "" {
		return model.User{}, "", stdErrors.New("user or email is required")
	}
	if password == "" {
		return model.User{}, "", stdErrors.New("password is required")
	}
	// TODO add validations step , length check basic format etc..

	user, err := s.userRepo.FindByUsernameOrEmail(usernameOrEmail)
	if err != nil {
		return model.User{}, "", stdErrors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return model.User{}, "", stdErrors.New("invalid credentials")
	}

	return user, "", nil
}