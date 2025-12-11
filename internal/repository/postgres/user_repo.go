package postgres

import (
	"backend/internal/domain/model"
	"backend/internal/repository"
	"database/sql"
	"errors"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &UserRepository{db}
} 

func (p *UserRepository) Save(user model.User) error {
	query := `INSERT INTO users(id, username, email, password) VALUES ($1, $2, $3, $4)`
	_, err := p.db.Exec(query, user.ID, user.Username, user.Email, user.Password)
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