package application

import (
	"backend/internal/domain/model"
	"backend/internal/errors"
	"backend/utils"
	stdErrors "errors"

	"golang.org/x/crypto/bcrypt"
)

var V = utils.Validator{}

func (s *UserService) Register(user model.User) error {
	// Let's check if the data is valid
	var errs errors.ValidationErrors

	if err := V.Username(user.Username); err != nil {
		errs = append(errs, err.(errors.ValidationError))
	}
	if err := V.Email(user.Email); err != nil {
		errs = append(errs, err.(errors.ValidationError))
	}
	if err := V.Password(user.Password); err != nil {
		errs = append(errs, err.(errors.ValidationError))
	}
	if len(errs) > 0 {
		return errs
	}

	// Let's check if user exists, checking by username and by email
	existing, err := s.userRepo.FindByUsernameOrEmail(user.Username)
	if err == nil && existing.ID != 0 {
		return stdErrors.New("username already taken")
	}
	existing, err = s.userRepo.FindByUsernameOrEmail(user.Email)
	if err == nil && existing.ID != 0 {
		return stdErrors.New("an account with this email already exists")
	}

	// If user is not found we hash the password before registering
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashed)

	if err := s.userRepo.Save(user); err != nil {
		return err
	}
	return nil
}