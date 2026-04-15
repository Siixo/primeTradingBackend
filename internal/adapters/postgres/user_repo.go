package postgres

import (
	"backend/internal/domain/model"
	"backend/internal/domain/repository"
	"context"
	"database/sql"
	"errors"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &UserRepository{db}
}

func (p *UserRepository) Migrate() error {
	query := `CREATE TABLE IF NOT EXISTS users (
		id          SERIAL PRIMARY KEY,
		username    VARCHAR(50) NOT NULL UNIQUE,
		email       VARCHAR(100) NOT NULL UNIQUE,
		password    VARCHAR(255) NOT NULL,
		role        VARCHAR(50) NOT NULL DEFAULT 'user',
		created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
		last_login  TIMESTAMP
	);`

	if _, err := p.db.Exec(query); err != nil {
		return err
	}

	// Add role column for existing tables
	_, _ = p.db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(50) NOT NULL DEFAULT 'user'`)
	return nil
}

func (p *UserRepository) Save(ctx context.Context, user model.User) error {
	query := `INSERT INTO users(username, email, password, role) VALUES ($1, $2, $3, $4)`
	_, err := p.db.ExecContext(ctx, query, user.Username, user.Email, user.Password, user.Role)
	return err
}

func (p *UserRepository) FindByUsernameOrEmail(ctx context.Context, identifier string) (model.User, error) {
	var user model.User
	query := `SELECT id, username, email, password, role FROM users WHERE username=$1 OR email=$2 LIMIT 1`
	row := p.db.QueryRowContext(ctx, query, identifier, identifier)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, errors.New("user not found")
		}
		return model.User{}, err
	}
	return user, nil
}

func (p *UserRepository) FindByID(ctx context.Context, id uint) (model.User, error) {
	var user model.User
	query := `SELECT id, username, email, password, role FROM users WHERE id=$1 LIMIT 1`
	row := p.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, errors.New("user not found")
		}
		return model.User{}, err
	}
	return user, nil
}

func (p *UserRepository) UpdatePassword(ctx context.Context, id uint, hashedPassword string) error {
	query := `UPDATE users SET password=$1 WHERE id=$2`
	_, err := p.db.ExecContext(ctx, query, hashedPassword, id)
	return err
}
