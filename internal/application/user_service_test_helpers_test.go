package application

import (
	"backend/internal/domain/model"
	"context"
)

type fakeUserRepository struct {
	migrateFn               func() error
	saveFn                  func(user model.User) error
	findByUsernameOrEmailFn func(identifier string) (model.User, error)
	findByIDFn              func(id uint) (model.User, error)

	savedUsers []model.User
}

func (f *fakeUserRepository) Migrate() error {
	if f.migrateFn != nil {
		return f.migrateFn()
	}
	return nil
}

func (f *fakeUserRepository) Save(ctx context.Context, user model.User) error {
	f.savedUsers = append(f.savedUsers, user)
	if f.saveFn != nil {
		return f.saveFn(user)
	}
	return nil
}

func (f *fakeUserRepository) FindByUsernameOrEmail(ctx context.Context, identifier string) (model.User, error) {
	if f.findByUsernameOrEmailFn != nil {
		return f.findByUsernameOrEmailFn(identifier)
	}
	return model.User{}, nil
}

func (f *fakeUserRepository) FindByID(ctx context.Context, id uint) (model.User, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(id)
	}
	return model.User{}, nil
}

func (f *fakeUserRepository) UpdatePassword(ctx context.Context, id uint, hashedPassword string) error {
	return nil
}
