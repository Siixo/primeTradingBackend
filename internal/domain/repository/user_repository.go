package repository

import "backend/internal/domain/model"

type UserRepository interface {
	Save(user model.User) error
	FindByUsernameOrEmail(identifier string) (model.User, error)
	FindByID(id uint) (model.User, error)
}