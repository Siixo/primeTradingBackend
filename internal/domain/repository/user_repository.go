package repository

import (
	"backend/internal/domain/model"
	"context"
)

type UserRepository interface {
	Migrate() error
	Save(ctx context.Context, user model.User) error
	FindByUsernameOrEmail(ctx context.Context, identifier string) (model.User, error)
	FindByID(ctx context.Context, id uint) (model.User, error)
	UpdatePassword(ctx context.Context, id uint, hashedPassword string) error
}
