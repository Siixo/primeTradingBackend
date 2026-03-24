package application

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
)

func (s *UserService) ChangePassword(req ChangePasswordInput) error {
	if req.OldPassword == "" {
		return errors.New("current password is required")
	}
	if req.NewPassword == "" {
		return errors.New("new password is required")
	}

	user, err := s.userRepo.FindByID(req.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("incorrect current password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update in repo
	return s.userRepo.UpdatePassword(req.UserID, string(hashedPassword))
}
