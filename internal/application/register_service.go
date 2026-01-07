package application

import (
	"backend/internal/domain/model"
	"backend/internal/errors"
	"backend/internal/handler/dto"
	"backend/utils"
	stdErrors "errors"

	"golang.org/x/crypto/bcrypt"
)

var V = utils.Validator{}

func (s *UserService) Register(req dto.RegisterRequest) error {
	// Let's check if the data is valid
	var errs errors.ValidationErrors

	if err := V.Username(req.Username); err != nil {
		errs = append(errs, err.(errors.ValidationError))
	}
	if err := V.Email(req.Email); err != nil {
		errs = append(errs, err.(errors.ValidationError))
	}
	if err := V.Password(req.Password); err != nil {
		errs = append(errs, err.(errors.ValidationError))
	}
	if len(errs) > 0 {
		return errs
	}

	// Let's check if user exists, checking by username and by email
	existing, err := s.userRepo.FindByUsernameOrEmail(req.Username)
	if err == nil && existing.ID != 0 {
		return stdErrors.New("username already taken")
	}
	existing, err = s.userRepo.FindByUsernameOrEmail(req.Email)
	if err == nil && existing.ID != 0 {
		return stdErrors.New("an account with this email already exists")
	}

	// If user is not found we hash the password before registering
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashed),
	}

	if err := s.userRepo.Save(user); err != nil {
		return err
	}
	return nil
}