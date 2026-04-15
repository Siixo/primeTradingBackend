package application

import (
	"backend/internal/domain/model"
	"backend/internal/domain/repository"
	"context"
)

type UserService struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
}

type LoginInput struct {
	Identifier string
	Password   string
}

type RegisterInput struct {
	Username  string
	Email     string
	Password  string
	Password2 string
}

type ChangePasswordInput struct {
	UserID      uint
	OldPassword string
	NewPassword string
}

func NewUserService(userRepo repository.UserRepository, jwtSecret string) *UserService {
	return &UserService{userRepo: userRepo, jwtSecret: []byte(jwtSecret)}
}

func (s *UserService) FindByID(ctx context.Context, id uint) (model.User, error) {
	return s.userRepo.FindByID(ctx, id)
}
