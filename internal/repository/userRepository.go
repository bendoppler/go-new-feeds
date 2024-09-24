package repository

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver
	"news-feed/internal/entity"
	"news-feed/pkg/logger"
	"strings"
)

// UserRepositoryInterface defines the methods for user data operations.
type UserRepositoryInterface interface {
	GetByUserName(userName string) (entity.User, error)
	CreateUser(user entity.User) error
	UpdateUser(user entity.User) error
	GetAllUserNames() ([]string, error)
	GetByUserID(userID int) (entity.User, error)
}

// UserRepository is a concrete implementation of UserRepositoryInterface.
type UserRepository struct {
	db *sql.DB
}

// GetByUserID retrieves a user by their user id.
func (r *UserRepository) GetByUserID(userID int) (entity.User, error) {
	query := `SELECT id, hashed_password, salt, first_name, last_name, email, user_name FROM user WHERE id = ?`
	row := r.db.QueryRow(query, userID)

	var user entity.User
	err := row.Scan(
		&user.ID, &user.HashedPassword, &user.Salt, &user.FirstName, &user.LastName, &user.Email,
		&user.Username,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.LogError(fmt.Sprintf("User not found"))
			return user, fmt.Errorf("user not found")
		}
		logger.LogError(fmt.Sprintf("error getting user: %v", err))
		return user, fmt.Errorf("error getting user: %v", err)
	}
	return user, nil
}

// GetByUserName retrieves a user by their username.
func (r *UserRepository) GetByUserName(userName string) (entity.User, error) {
	query := `SELECT id, hashed_password, salt, first_name, last_name, email, user_name FROM user WHERE user_name = ?`
	row := r.db.QueryRow(query, userName)

	var user entity.User
	err := row.Scan(
		&user.ID, &user.HashedPassword, &user.Salt, &user.FirstName, &user.LastName, &user.Email,
		&user.Username,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.LogError(fmt.Sprintf("User not found"))
			return user, fmt.Errorf("user not found")
		}
		logger.LogError(fmt.Sprintf("error getting user: %v", err))
		return user, fmt.Errorf("error getting user: %v", err)
	}
	return user, nil
}

func (r *UserRepository) CreateUser(user entity.User) error {
	query := `INSERT INTO user (hashed_password, salt, first_name, last_name, dob, email, user_name) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(
		query, user.HashedPassword, user.Salt, user.FirstName, user.LastName, user.Birthday, user.Email, user.Username,
	)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}
	return nil
}

// UpdateUser updates an existing user in the database.
func (r *UserRepository) UpdateUser(user entity.User) error {
	var updateFields []string
	var args []interface{}
	if user.FirstName != "" {
		updateFields = append(updateFields, "first_name = ?")
		args = append(args, user.FirstName)
	}
	if user.LastName != "" {
		updateFields = append(updateFields, "last_name = ?")
		args = append(args, user.LastName)
	}
	if !user.Birthday.IsZero() {
		updateFields = append(updateFields, "dob = ?")
		args = append(args, user.Birthday)
	}
	if user.HashedPassword != "" {
		updateFields = append(updateFields, "hashed_password = ?, salt = ?")
		args = append(args, user.HashedPassword, user.Salt)
	}

	if len(updateFields) == 0 {
		return nil // No fields to update
	}

	args = append(args, user.Username)
	updateQuery := "UPDATE user SET" + " " + strings.Join(updateFields, ", ") + " WHERE user_name = ?"
	_, err := r.db.Exec(updateQuery, args...)
	return err
}

func (r *UserRepository) GetAllUserNames() ([]string, error) {
	query := `SELECT user_name FROM user`

	// Execute the query
	rows, err := r.db.Query(query)
	if err != nil {
		logger.LogError(fmt.Sprintf("Error executing query to get all usernames: %v", err))
		return nil, fmt.Errorf("could not execute query: %v", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.LogError(fmt.Sprintf("Error closing rows: %v", err))
			return
		}
	}(rows)

	// Slice to hold the usernames
	var userNames []string

	// Iterate over the rows
	for rows.Next() {
		var username string
		// Scan the result into the username variable
		if err := rows.Scan(&username); err != nil {
			logger.LogError(fmt.Sprintf("Error scanning username: %v", err))
			return nil, fmt.Errorf("could not scan username: %v", err)
		}
		// Append the username to the slice
		userNames = append(userNames, username)
	}

	// Check if any errors occurred during the iteration
	if err := rows.Err(); err != nil {
		logger.LogError(fmt.Sprintf("Error during rows iteration: %v", err))
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	// Return the slice of usernames
	return userNames, nil
}
