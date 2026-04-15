package application

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func (s *UserService) ChangePassword(ctx context.Context, req ChangePasswordInput) error {
	if req.OldPassword == "" {
		return errors.New("current password is required")
	}
	if req.NewPassword == "" {
		return errors.New("new password is required")
	}

	// Validate new password strength
	if err := V.Password(req.NewPassword); err != nil {
		return err
	}

	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("incorrect current password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(ctx, req.UserID, string(hashedPassword))
}
