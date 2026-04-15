package application

import (
	"backend/internal/domain/model"
	"backend/internal/errors"
	"backend/utils"
	"context"
	stdErrors "errors"

	"golang.org/x/crypto/bcrypt"
)

var V = utils.Validator{}

func (s *UserService) Register(ctx context.Context, req RegisterInput) error {
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
	if req.Password != req.Password2 {
		errs = append(errs, errors.ValidationError{Field: "confirmPassword", Code: "passwords do not match"})
	}
	if len(errs) > 0 {
		return errs
	}

	existing, err := s.userRepo.FindByUsernameOrEmail(ctx, req.Username)
	if err == nil && existing.ID != 0 {
		return stdErrors.New("username already taken")
	}
	existing, err = s.userRepo.FindByUsernameOrEmail(ctx, req.Email)
	if err == nil && existing.ID != 0 {
		return stdErrors.New("an account with this email already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashed),
		Role:     "user",
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return err
	}
	return nil
}
