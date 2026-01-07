package postgres

import (
	"backend/internal/domain/model"
	"backend/internal/domain/repository"
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
	query := `CREATE TABLE users (
		id          SERIAL PRIMARY KEY,
		username    VARCHAR(50) NOT NULL UNIQUE,
		email       VARCHAR(100) NOT NULL UNIQUE,
		password    VARCHAR(255) NOT NULL,
		created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
		last_login  TIMESTAMP
	);`

	_, err := p.db.Exec(query)
	return err
}

func (p *UserRepository) Save(user model.User) error {
	query := `INSERT INTO users(username, email, password) VALUES ($1, $2, $3)`
	_, err := p.db.Exec(query, user.Username, user.Email, user.Password)
	return err
}

func (p *UserRepository) FindByUsernameOrEmail(identifier string) (model.User, error) {
	var user model.User
	query := `SELECT id, username, email, password FROM users WHERE username=$1 OR email=$2 LIMIT 1`
	row := p.db.QueryRow(query, identifier, identifier)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, errors.New("user not found")
		}
		return model.User{}, err
	}
	return user, nil
}

func (p *UserRepository) FindByID(id uint) (model.User, error) {
	var user model.User
	query := `SELECT id, username, email, password FROM users WHERE id=$1 LIMIT 1`
	row := p.db.QueryRow(query, id)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, errors.New("user not found")
		}
		return model.User{}, err
	}
	return user, nil
}
