package application

import (
	"backend/internal/domain/model"
	"backend/internal/domain/repository"
)

type UserService struct {
	userRepo repository.UserRepository
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

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo}
}

func (s *UserService) FindByID(id uint) (model.User, error) {
	return s.userRepo.FindByID(id)
}
