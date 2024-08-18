package repository

import (
	"database/sql"
	"fmt"
	"news-feed/internal/entity"
)

type FriendsRepositoryInterface interface {
	GetFriends(userID int) ([]entity.User, error)
	FollowUser(userID int) error
	UnfollowUser(userID int) error
}

type FriendsRepository struct {
	db *sql.DB
}

// GetFriends retrieves the list of friends for a user.
func (r *FriendsRepository) GetFriends(userID int) ([]entity.User, error) {
	rows, err := r.db.Query(
		"SELECT id, first_name, last_name, email, user_name FROM users WHERE id IN (SELECT friend_id FROM friends WHERE user_id = ?)",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("Failed to close rows: %v\n", err)
			return
		}
	}(rows)

	var users []entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Username); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// FollowUser follows a user.
func (r *FriendsRepository) FollowUser(userID int) error {
	// This example assumes you want to follow user ID 2, modify accordingly
	_, err := r.db.Exec("INSERT INTO friends (user_id, friend_id) VALUES (?, ?)", userID, 2)
	return err
}

// UnfollowUser unfollows a user.
func (r *FriendsRepository) UnfollowUser(userID int) error {
	_, err := r.db.Exec("DELETE FROM friends WHERE user_id = ? AND friend_id = ?", userID, 2)
	return err
}
