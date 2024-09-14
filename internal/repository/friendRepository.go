package repository

import (
	"database/sql"
	"fmt"
	"news-feed/internal/entity"
)

type FriendsRepositoryInterface interface {
	GetFriends(userID int) ([]entity.User, error)
	FollowUser(currentUserID int, followedUserID int) error
	UnfollowUser(currentUserID int, unfollowedUserID int) error
}

type FriendsRepository struct {
	db *sql.DB
}

// GetFriends retrieves the list of friends for a user.
func (r *FriendsRepository) GetFriends(userID int) ([]entity.User, error) {
	rows, err := r.db.Query(
		"SELECT id, first_name, last_name, email, user_name FROM user WHERE id IN (SELECT fk_follower_id FROM user_user WHERE fk_user_id = ?)",
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
func (r *FriendsRepository) FollowUser(currentUserID int, followedUserID int) error {
	// This example assumes you want to follow user ID 2, modify accordingly
	_, err := r.db.Exec(
		"INSERT INTO user_user (fk_user_id, fk_follower_id) VALUES (?, ?)", currentUserID, followedUserID,
	)
	return err
}

// UnfollowUser unfollows a user.
func (r *FriendsRepository) UnfollowUser(currentUserID int, unfollowedUserID int) error {
	_, err := r.db.Exec(
		"DELETE FROM user_user WHERE fk_user_id = ? AND fk_follower_id = ?", currentUserID, unfollowedUserID,
	)
	return err
}
