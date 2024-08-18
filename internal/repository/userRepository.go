package repository

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver
	"news-feed/internal/entity"
)

// UserRepositoryInterface defines the methods for user data operations.
type UserRepositoryInterface interface {
	GetByUserName(userName string) (entity.User, error)
	CreateUser(user entity.User) error
	UpdateUser(user entity.User) error
}

// UserRepository is a concrete implementation of UserRepositoryInterface.
type UserRepository struct {
	db *sql.DB
}

// GetByUserName retrieves a user by their username.
func (r *UserRepository) GetByUserName(userName string) (entity.User, error) {
	var user entity.User
	query := "SELECT id, username, email, first_name, last_name, birthday, password FROM users WHERE username = ?"
	err := r.db.QueryRow(query, userName).Scan(
		&user.ID, &user.UserName, &user.Email, &user.FirstName, &user.LastName, &user.Birthday, &user.Password,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, nil // User not found
		}
		return user, err
	}
	return user, nil
}

// CreateUser inserts a new user into the database.
func (r *UserRepository) CreateUser(user entity.User) error {
	query := "INSERT INTO users (username, email, first_name, last_name, birthday, password) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := r.db.Exec(query, user.UserName, user.Email, user.FirstName, user.LastName, user.Birthday, user.Password)
	return err
}

// UpdateUser updates an existing user in the database.
func (r *UserRepository) UpdateUser(user entity.User) error {
	query := "UPDATE users SET first_name = ?, last_name = ?, birthday = ?, password = ? WHERE username = ?"
	_, err := r.db.Exec(query, user.FirstName, user.LastName, user.Birthday, user.Password, user.UserName)
	return err
}
